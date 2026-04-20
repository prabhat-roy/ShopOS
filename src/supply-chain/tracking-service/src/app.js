'use strict';

const express = require('express');
const createTrackingRouter = require('./routes/trackingRoutes');

/**
 * Factory that creates and configures the Express application.
 * Accepts an optional tracking controller (injected during tests).
 * @param {Object} [deps]
 * @param {import('./controllers/TrackingController')} [deps.controller]
 * @returns {express.Application}
 */
function createApp({ controller } = {}) {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: false }));

  // Health check
  app.get('/healthz', (_req, res) => {
    res.status(200).json({ status: 'ok', service: 'tracking-service' });
  });

  // Shipment routes
  app.use('/shipments', createTrackingRouter(controller));

  // 404 catch-all
  app.use((_req, res) => {
    res.status(404).json({ error: 'Not found' });
  });

  // Global error handler
  app.use((err, _req, res, _next) => {
    console.error('Unhandled error:', err);
    res.status(500).json({ error: 'Internal server error' });
  });

  return app;
}

module.exports = createApp;
