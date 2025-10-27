// frontend/server.js
const express = require('express');
const axios = require('axios');
const path = require('path');
const session = require('express-session');
const expressLayouts = require('express-ejs-layouts');

const app = express();
const PORT = 3000;

const USE_MOCK_API = true;
const API_BASE_URL = 'http://localhost:8080/api/v1';

let mockElapsedTime = 0;

app.use(session({
  secret: 'your-secret-key-change-in-production',
  resave: false,
  saveUninitialized: false,
  cookie: { secure: false }
}));

app.use(expressLayouts);
app.set('layout', 'layout');
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));
app.use(express.static(path.join(__dirname, 'public')));
app.use(express.json());

const apiClient = {
  async getScenarios() {
    if (USE_MOCK_API) {
      return Promise.resolve([
        { id: "1", name: "Users API Load", target: "https://api.example.com/users", method: "GET", rps: 100, duration: 30 },
        { id: "2", name: "Auth Stress", target: "https://api.example.com/login", method: "POST", rps: 50, duration: 60 },
      ]);
    }
    const res = await axios.get(`${API_BASE_URL}/scenarios`);
    return res.data;
  },

  async createScenario(data) {
    if (USE_MOCK_API) {
      return Promise.resolve({ ...data, id: Date.now().toString() });
    }
    const res = await axios.post(`${API_BASE_URL}/scenarios`, data);
    return res.data;
  },

  async startTest(scenarioId) {
    if (USE_MOCK_API) {
      return Promise.resolve({ runId: `mock_run_${Date.now()}` });
    }
    const res = await axios.post(`${API_BASE_URL}/run/${scenarioId}`);
    return res.data;
  },
  
  async register(credentials) {
    try {
      const res = await axios.post(`${API_BASE_URL}/auth/register`, credentials);
      return { success: true, data: res.data };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Registration failed'
      };
    }
  },

  async login(credentials) {
    try {
      const res = await axios.post(`${API_BASE_URL}/auth/login`, credentials);
      return { success: true, data: res.data };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Login failed'
      };
    }
  },

  async logout() {
    try {
      await axios.post(`${API_BASE_URL}/auth/logout`);
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: 'Logout failed'
      };
    }
  },


  getMetricsUrl(runId) {
    return `/api-proxy/metrics/${runId}`;
  },
};

function requireAuth(req, res, next) {
  if (!req.session || !req.session.userId) {
    return res.redirect('/login');
  }
  next();
}

app.get('/', requireAuth, async (req, res) => {
  try {
    const scenarios = await apiClient.getScenarios(req.session.token);
    res.render('dashboard', { 
      scenarios, 
      path: '/',
      userEmail: req.session.userEmail
    });
  } catch (err) {
    res.status(500).send('Failed to load scenarios');
  }
});

app.get('/login', (req, res) => {
  res.render('login', { path: '/login' });
});

app.get('/register', (req, res) => {
  res.render('register', { path: '/register' });
});

app.post('/api/v1/auth/register', async (req, res) => {
  console.log(req.body);
  const result = await apiClient.register(req.body);
  res.json(result);
});

app.post('/api/v1/auth/login', async (req, res) => {
  console.log(req);
  const result = await apiClient.login(req.body);
  if (result.success) {
    req.session.token = result.data.token;
    req.session.userId = result.data.user.id;
    req.session.userEmail = result.data.user.email; 
  }
  res.json(result);
});

app.post('/api/v1/auth/logout', async (req, res) => {
  console.log(req.body);
  const result = await apiClient.logout();
  if (result.success) {
    req.session.destroy();
  }
  res.json(result);
});

app.get('/api-proxy/metrics/:runId', async (req, res) => {
  try {
    if (USE_MOCK_API) {
      mockElapsedTime = (mockElapsedTime + 1) % 11; // 0 → 10 → 0...

      setTimeout(() => {
        res.json({
          rps: Math.floor(90 + Math.random() * 20),
          latency: Math.floor(100 + Math.random() * 50),
          cpu: Math.floor(50 + Math.random() * 30),
          ram: Math.floor(40 + Math.random() * 30),
          errors: Math.random() > 0.95 ? 1 : 0,
          timeElapsed: mockElapsedTime, // ✅ Now increments: 0,1,2,3...10
          duration: 10,
        });
      }, 100); // ↓ faster response for smoother UI
    } else {
      const apiRes = await axios.get(`${API_BASE_URL}/run/${req.params.runId}/metrics`);
      res.json(apiRes.data);
    }
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/scenario/new', requireAuth, (req, res) => {
  res.render('create_scenario', { 
    path: '/scenario/new',
    userEmail: req.session.userEmail
  });
});

app.get('/run/:id', requireAuth, async (req, res) => {
  try {
    const scenarios = await apiClient.getScenarios();
    const scenario = scenarios.find(s => s.id === req.params.id);
    if (!scenario) return res.status(404).send('Scenario not found');

    const { runId } = await apiClient.startTest(scenario.id);

    res.render('live_test', { 
      scenario, 
      runId,
      metricsUrl: apiClient.getMetricsUrl(runId),
      path: '/run/' + req.params.id,
      userEmail: req.session.userEmail
    });
  } catch (err) {
    res.status(500).send('Failed to start test');
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
    userEmail: req.session.userEmail
  });
});

app.listen(PORT, () => {
  console.log(`✅ GoStresser Frontend running at http://localhost:${PORT}`);
  console.log(`🔌 Backend API: ${USE_MOCK_API ? 'MOCK MODE' : API_BASE_URL}`);
});