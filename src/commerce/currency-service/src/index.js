'use strict';

/**
 * Re-export the Express app for testing purposes.
 * The actual server is started in the root index.js.
 */
const express = require('express');
const currencyRoutes = require('./routes/currencyRoutes');

function createApp() {
  const app = express();

  app.use(express.json());

  // Health check
  app.get('/healthz', (_req, res) => {
    res.json({ status: 'ok' });
  });

  // Currency routes
  app.use('/currencies', currencyRoutes);

  // 404 handler
  app.use((_req, res) => {
    res.status(404).json({ error: 'Not found' });
  });

  // Global error handler
  // eslint-disable-next-line no-unused-vars
  app.use((err, _req, res, _next) => {
    console.error('Unhandled error:', err);
    res.status(500).json({ error: 'Internal server error' });
  });

  return app;
}

module.exports = { createApp };
