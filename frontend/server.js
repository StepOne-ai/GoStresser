// frontend/server.js
const express = require('express');
const axios = require('axios');
const path = require('path');
const expressLayouts = require('express-ejs-layouts');

const app = express();
const PORT = 3000;

const USE_MOCK_API = true;
const API_BASE_URL = 'http://localhost:8080/api';

let mockElapsedTime = 0;

app.use(expressLayouts);
app.set('layout', 'layout'); // default layout file (views/layout.ejs)
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));
app.use(express.static(path.join(__dirname, 'public')));
app.use(express.urlencoded({ extended: true }));

// ========= API CLIENT =========
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

  getMetricsUrl(runId) {
    return `/api-proxy/metrics/${runId}`;
  },
};

// Then inside the route:
app.get('/api-proxy/metrics/:runId', async (req, res) => {
  try {
    if (USE_MOCK_API) {
      // Increment time every call
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

// ========= ROUTES =========
app.get('/', async (req, res) => {
  try {
    const scenarios = await apiClient.getScenarios();
    res.render('dashboard', { 
      scenarios,
      path: '/' // ← for active nav
    });
  } catch (err) {
    res.status(500).send('Failed to load scenarios');
  }
});

app.get('/scenario/new', (req, res) => {
  res.render('create_scenario', { 
    path: '/scenario/new' // ← for active nav
  });
});

app.get('/run/:id', async (req, res) => {
  try {
    const scenarios = await apiClient.getScenarios();
    const scenario = scenarios.find(s => s.id === req.params.id);
    if (!scenario) return res.status(404).send('Scenario not found');

    const { runId } = await apiClient.startTest(scenario.id);

    res.render('live_test', { 
      scenario, 
      runId,
      metricsUrl: apiClient.getMetricsUrl(runId),
      path: '/run/' + req.params.id // ← optional, or just leave as '/'
    });
  } catch (err) {
    res.status(500).send('Failed to start test');
  }
});

app.get('/report/:runId', (req, res) => {
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
    path: '/' // or '/report'
  });
});

app.listen(PORT, () => {
  console.log(`✅ GoStresser Frontend running at http://localhost:${PORT}`);
  console.log(`🔌 Backend API: ${USE_MOCK_API ? 'MOCK MODE' : API_BASE_URL}`);
});