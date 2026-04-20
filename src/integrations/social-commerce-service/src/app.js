'use strict';

const express = require('express');
const socialRoutes = require('./routes/socialRoutes');
const { healthz } = require('./controllers/SocialCommerceController');

/**
 * Factory function that creates and configures the Express application.
 *
 * Separating app creation from server startup allows the test suite to
 * import the app without binding to a port.
 *
 * @returns {import('express').Application}
 */
function createApp() {
  const app = express();

  // ---- Middleware ----
  app.use(express.json({ limit: '5mb' }));
  app.use(express.urlencoded({ extended: true }));

  // Request logging in development
  if (process.env.NODE_ENV !== 'test') {
    app.use((req, _res, next) => {
      console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
      next();
    });
  }

  // ---- Routes ----
  app.get('/healthz', healthz);
  app.use('/social', socialRoutes);

  // ---- 404 handler ----
  app.use((_req, res) => {
    res.status(404).json({
      error: 'NOT_FOUND',
      message: 'The requested endpoint does not exist.',
    });
  });

  // ---- Global error handler ----
  // eslint-disable-next-line no-unused-vars
  app.use((err, _req, res, _next) => {
    console.error('[social-commerce-service] Unhandled error:', err);
    res.status(500).json({
      error: 'INTERNAL_ERROR',
      message: 'An unexpected error occurred.',
    });
  });

  return app;
}

module.exports = { createApp };
