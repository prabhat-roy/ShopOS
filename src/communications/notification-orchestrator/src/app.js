'use strict';

const express = require('express');
const healthController = require('./controllers/HealthController');

/**
 * Creates and configures the Express application.
 * Separated from index.js so it can be imported in tests without starting the server.
 *
 * @returns {express.Application}
 */
function createApp() {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: false }));

  // Request logging middleware
  app.use((req, _res, next) => {
    console.info(`[http] ${req.method} ${req.path}`);
    next();
  });

  // Routes
  app.use('/', healthController);

  // 404 handler
  app.use((_req, res) => {
    res.status(404).json({ error: 'Not found' });
  });

  // Global error handler
  app.use((err, _req, res, _next) => {
    console.error('[http] Unhandled error:', err.message);
    res.status(500).json({ error: 'Internal server error', message: err.message });
  });

  return app;
}

module.exports = { createApp };
