'use strict';

const express = require('express');
const createFeedbackRouter = require('./routes/feedbackRoutes');

function createApp(deps = {}) {
  const { feedbackController } = deps;

  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: false }));

  app.use((req, res, next) => {
    if (process.env.NODE_ENV !== 'test') {
      console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
    }
    next();
  });

  // Health check
  app.get('/healthz', (req, res) => {
    res.status(200).json({ status: 'ok', service: 'feedback-service' });
  });

  // Readiness check
  app.get('/readyz', (req, res) => {
    res.status(200).json({ status: 'ready', service: 'feedback-service' });
  });

  // Metrics placeholder (Phase 4)
  app.get('/metrics', (req, res) => {
    res.set('Content-Type', 'text/plain');
    res.status(200).send('# feedback-service metrics\n');
  });

  // Feedback routes
  if (feedbackController) {
    app.use('/feedback', createFeedbackRouter(feedbackController));
  }

  // 404 handler
  app.use((req, res) => {
    res.status(404).json({ error: 'Not Found', message: `Route ${req.method} ${req.path} not found` });
  });

  // Global error handler
  app.use((err, req, res, next) => {
    console.error('[app] Unhandled error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: process.env.NODE_ENV === 'production' ? 'An unexpected error occurred' : err.message,
    });
  });

  return app;
}

module.exports = createApp;
