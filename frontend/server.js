// frontend/server.js
const express = require('express');
const axios = require('axios');
const path = require('path');
const session = require('express-session');
const expressLayouts = require('express-ejs-layouts');

const app = express();
const PORT = 3000;

const USE_MOCK_API = false;
const API_BASE_URL = 'http://localhost:8080/api/v1';

// Mock data for development
let mockElapsedTime = 0;
const mockScenarios = [
  { id: "1", name: "Users API Load", target: "https://api.example.com/users", method: "GET", rps: 100, duration: 30 },
  { id: "2", name: "Auth Stress", target: "https://api.example.com/login", method: "POST", rps: 50, duration: 60 },
];

app.use(session({
  secret: 'your-secret-key-change-in-production',
  resave: false,
  saveUninitialized: false,
  cookie: { 
    secure: process.env.NODE_ENV === 'production', // Only use secure cookies in production
    httpOnly: true,
    maxAge: 24 * 60 * 60 * 1000 // 24 hours
  }
}));

app.use(expressLayouts);
app.set('layout', 'layout');
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));
app.use(express.static(path.join(__dirname, 'public')));
app.use(express.json());

// Helper function to make authenticated API calls to backend
async function callBackendApi(method, endpoint, data = null, session = null) {
  const headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  };

  // Add authorization header if we have a token from session
  if (session?.token) {
    headers['Authorization'] = `Bearer ${session.token}`;
  }

  try {
    const url = `${API_BASE_URL}${endpoint}`;
    let response;

    switch (method.toUpperCase()) {
      case 'GET':
        response = await axios.get(url, { headers });
        break;
      case 'POST':
        response = await axios.post(url, data, { headers });
        break;
      case 'PUT':
        response = await axios.put(url, data, { headers });
        break;
      case 'DELETE':
        response = await axios.delete(url, { headers });
        break;
      default:
        throw new Error(`Unsupported method: ${method}`);
    }

    return { success: true, data: response.data };
  } catch (error) {
    console.error(`Backend API error (${method} ${endpoint}):`, error.response?.data || error.message);
    
    const errorMessage = error.response?.data?.error || 
                        error.response?.data?.message || 
                        error.message || 
                        'API request failed';
    
    return { 
      success: false, 
      error: errorMessage,
      statusCode: error.response?.status
    };
  }
}

// Authentication middleware
function requireAuth(req, res, next) {
  if (!req.session.userID) {
    return res.redirect('/login');
  }
  next();
}

// Mock API functions (for development only)
function mockGetScenarios() {
  return Promise.resolve(mockScenarios);
}

function mockCreateScenario(data) {
  const newScenario = {
    id: Date.now().toString(),
    ...data
  };
  mockScenarios.push(newScenario);
  return Promise.resolve(newScenario);
}

function mockStartTest(scenarioId) {
  return Promise.resolve({ runId: `mock_run_${Date.now()}` });
}

// Routes
app.get('/', requireAuth, async (req, res) => {
  try {
    let scenarios;
    if (USE_MOCK_API) {
      scenarios = await mockGetScenarios();
    } else {
      const result = await callBackendApi('GET', '/scenarios', null, req.session);
      if (!result.success) throw new Error(result.error);
      scenarios = result.data;
    }

    res.render('dashboard', { 
      scenarios, 
      path: '/',
      userEmail: req.session.userEmail,
      name: req.session.userName || req.session.userEmail.split('@')[0]
    });
  } catch (err) {
    console.error('Dashboard error:', err);
    res.status(500).send('Failed to load dashboard');
  }
});

app.get('/login', (req, res) => {
  if (req.session.userID) {
    return res.redirect('/');
  }
  res.render('login', { path: '/login' });
});

app.get('/register', ((req, res) => {
  if (req.session.userID) {
    return res.redirect('/');
  }
  res.render('register', { path: '/register' });
}));

// Authentication endpoints - proxy to backend
app.post('/api/v1/auth/register', async (req, res) => {
  try {
    const result = USE_MOCK_API 
      ? { success: true, data: { user: { id: 'mock_user', email: req.body.email }, token: 'mock_token' } }
      : await callBackendApi('POST', '/auth/register', req.body);

    if (result.success) {
      // For mock mode, set session immediately
      if (USE_MOCK_API) {
        req.session.userID = 'mock_user';
        req.session.userEmail = req.body.email;
        req.session.userName = req.body.name || req.body.email.split('@')[0];
        req.session.token = 'mock_token';
      }
      res.json(result);
    } else {
      res.status(result.statusCode || 400).json(result);
    }
  } catch (error) {
    console.error('Registration error:', error);
    res.status(500).json({ success: false, error: 'Registration failed' });
  }
});

app.post('/api/v1/auth/login', async (req, res) => {
  try {
    const result = USE_MOCK_API 
      ? { 
          success: true, 
          data: { 
            user: { 
              id: 'mock_user', 
              email: req.body.email,
              name: req.body.email.split('@')[0]
            }, 
            token: 'mock_token'
          } 
        }
      : await callBackendApi('POST', '/auth/login', req.body);

    if (result.success) {
      // Set session data
      req.session.userID = result.data.user.id;
      req.session.userEmail = result.data.user.email;
      req.session.userName = result.data.user.name || result.data.user.email.split('@')[0];
      req.session.token = result.data.token;
      
      res.json({ 
        success: true, 
        redirect: '/',
        user: {
          email: req.session.userEmail,
          name: req.session.userName
        }
      });
    } else {
      res.status(result.statusCode || 401).json(result);
    }
  } catch (error) {
    console.error('Login error:', error);
    res.status(500).json({ success: false, error: 'Login failed' });
  }
});

app.post('/api/v1/auth/logout', async (req, res) => {
  try {
    if (!USE_MOCK_API && req.session.token) {
      await callBackendApi('POST', '/auth/logout', {}, req.session);
    }
    
    req.session.destroy(err => {
      if (err) {
        console.error('Session destroy error:', err);
        return res.status(500).json({ success: false, error: 'Logout failed' });
      }
      res.json({ success: true, redirect: '/login' });
    });
  } catch (error) {
    console.error('Logout error:', error);
    // Still destroy session even if backend logout fails
    req.session.destroy(() => {
      res.json({ success: true, redirect: '/login' });
    });
  }
});

// API Proxy endpoints
app.get('/api-proxy/metrics/:runId', async (req, res) => {
  try {
    if (USE_MOCK_API) {
      mockElapsedTime = (mockElapsedTime + 1) % 11;

      setTimeout(() => {
        res.json({
          rps: Math.floor(90 + Math.random() * 20),
          latency: Math.floor(100 + Math.random() * 50),
          cpu: Math.floor(50 + Math.random() * 30),
          ram: Math.floor(40 + Math.random() * 30),
          errors: Math.random() > 0.95 ? 1 : 0,
          timeElapsed: mockElapsedTime,
          duration: 10,
        });
      }, 100);
    } else {
      const result = await callBackendApi('GET', `/run/${req.params.runId}/metrics`, null, req.session);
      if (!result.success) throw new Error(result.error);
      res.json(result.data);
    }
  } catch (err) {
    console.error('Metrics error:', err);
    res.status(500).json({ error: 'Failed to fetch metrics' });
  }
});

app.post('/api/scenarios', requireAuth, async (req, res) => {
  try {
    const scenarioData = {
      user_id: req.session.userID,
      name: req.body.name,
      target_url: req.body.target.trim(),
      method: req.body.method,
      rps: parseInt(req.body.rps),
      duration: parseInt(req.body.duration),
    };

    const result = USE_MOCK_API
      ? await mockCreateScenario(scenarioData)
      : await callBackendApi('POST', '/scenarios', scenarioData, req.session);

    if (!USE_MOCK_API && !result.success) {
      return res.status(result.statusCode || 400).json({
        success: false,
        error: result.error || 'Failed to create scenario'
      });
    }

    const scenario = USE_MOCK_API ? result : result.data;

    // ✅ Success: send response and RETURN
    return res.json({ success: true, scenario });

  } catch (error) {
    console.error('Create scenario error:', error);

    if (!res.headersSent) {
      res.status(500).json({
        success: false,
        error: error.message || 'Internal server error'
      });
    }
  }
});

app.get('/scenario/new', requireAuth, (req, res) => {
  res.render('create_scenario', { 
    path: '/scenario/new',
    userEmail: req.session.userEmail,
    name: req.session.userName
  });
});

app.get('/run/:id', requireAuth, async (req, res) => {
  const scenarioId = req.params.id;

  try {
    // ✅ Fetch SINGLE scenario by ID
    let scenario;
    if (USE_MOCK_API) {
      const allScenarios = await mockGetScenarios();
      scenario = allScenarios.find(s => s.id === scenarioId);
    } else {
      // Real API: fetch single scenario directly
      const result = await callBackendApi('GET', `/scenarios/${scenarioId}`, null, req.session);
      if (!result.success) {
        throw new Error(result.error || 'не получилось получить сценарий');
      }
      scenario = result.data;
    }

    if (!scenario) {
      return res.status(404).send('сценарий не найден');
    }

    // ✅ Start the test WITH AUTHENTICATION
    // Your backend requires auth middleware, so we need to pass the token
    const startResult = USE_MOCK_API
      ? await mockStartTest(scenario.id)
      : await callBackendApi('POST', `/run/${scenarioId}`, null, req.session); // ✅ req.session includes token

      console.log(startResult);
      if (!USE_MOCK_API && !startResult.success) {
      if (startResult.statusCode === 404) {
        return res.status(404).send('Scenario not found');
      }
      if (startResult.statusCode === 401 || startResult.statusCode === 403) {
        return res.status(401).send('Authentication required');
      }
      throw new Error(startResult.error || 'Failed to start test');
    }

    // ✅ Get run_id EXACTLY as your backend returns it
    const runId = USE_MOCK_API 
      ? startResult.runId 
      : startResult.data.run_id; // ✅ snake_case matches your Go backend

    if (!runId) {
      throw new Error('Test started but no run_id returned');
    }

    // ✅ Render with DIRECT metrics URL
    res.render('live_test', { 
      scenario, 
      runId,
      metricsUrl: `/api-proxy/metrics/${runId}`,
      path: `/run/${scenarioId}`,
      userEmail: req.session.userEmail,
      name: req.session.userName
    });

  } catch (err) {
    console.error('❌ Run scenario error:', err);
    
    if (!res.headersSent) {
      res.status(500).send('Failed to start test: ' + (err.message || 'Unknown error'));
    }
  }
});

app.get('/delete/:id', requireAuth, async (req, res) => {
  try {
    const result = USE_MOCK_API
      ? await mockDeleteScenario(req.params.id)
      : await callBackendApi('DELETE', `/scenarios/${req.params.id}`, null, req.session);

    if (!USE_MOCK_API && !result.success) {
      throw new Error(result.error);
    }

    res.redirect('/');
  } catch (err) {
    console.error('Delete scenario error:', err);
    res.status(500).send('Failed to delete scenario');
  }
});

app.get('/report/:runId', requireAuth, (req, res) => {
  const mockReport = {
    runId: req.params.runId,
    scenarioName: "Users API Load",
    totalRequests: 3000,
    avgLatency: 142,
    errorRate: 0.2,
    peakCPU: 82,
    recommendations: [
      "Latency stable — no bottlenecks detected",
      "CPU near limit — consider scaling at >500 RPS",
    ],
  };
  res.render('report', { 
    report: mockReport,
    path: '/',
    userEmail: req.session.userEmail,
    name: req.session.userName
  });
});

app.put('/edit/:id', async (req, res) => {
  try {
    const scenarioId = req.params.id; // Get the ID from the URL parameter
    const updateData = req.body; // Get the new data from the request body

    // Example validation
    if (!scenarioId) {
      return res.status(400).json({ error: 'Scenario ID is required' });
    }

    const updatedScenario = await callBackendApi('PUT', `/scenarios/${scenarioId}`, updateData, req.session);
    if (!updatedScenario.success) {
      throw new Error(updatedScenario.error);
    }

    // Example success response
    res.status(200).json({ 
      message: 'Успешное обновление сценария', 
      scenario: updatedScenario.data
    });

  } catch (error) {
    console.error('Ошибка обновления сценария', error.message || error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

app.get('/edit/:id', async (req, res) => {
  try {
    const scenarioId = req.params.id;

    // Example: Fetch the current scenario data to pre-populate the update form
    // const scenario = await ScenarioModel.findById(scenarioId);
    result = await callBackendApi('GET', `/scenarios/${scenarioId}`, null, req.session);

    if (!result.success) {
      throw new Error(result.error);
    }

    const scenario = result.data;

    console.log(`Fetching scenario data for ID: ${scenarioId} to render update form`);
    res.render('update_scenario', { scenario: scenario });

  } catch (error) {
    console.error('Error fetching scenario for update:', error);
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    timestamp: new Date().toISOString(),
    mockMode: USE_MOCK_API,
    backendUrl: API_BASE_URL 
  });
});

app.listen(PORT, () => {
  console.log(`✅ GoStresser Frontend running at http://localhost:${PORT}`);
  console.log(`🔐 Session management: ${USE_MOCK_API ? 'MOCK MODE' : 'Production'}`);
  console.log(`🔌 Backend API: ${USE_MOCK_API ? 'MOCK MODE' : API_BASE_URL}`);
  console.log(`📊 Environment: ${process.env.NODE_ENV || 'development'}`);
});