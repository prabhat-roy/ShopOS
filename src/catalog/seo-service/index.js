'use strict';

const express = require('express');
const config = require('./src/config');
const seoRoutes = require('./src/routes/seoRoutes');

const app = express();

app.use(express.json());

// Routes
app.use('/', seoRoutes);

// 404 handler
app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

// Global error handler
app.use((err, req, res, next) => {
  console.error('[seo-service] error:', err);
  const status = err.status || 500;
  res.status(status).json({ error: err.message || 'Internal server error' });
});

async function start() {
  try {
    const server = app.listen(config.httpPort, () => {
      console.log(`[seo-service] HTTP server listening on port ${config.httpPort}`);
      console.log(`[seo-service] BASE_URL: ${config.baseUrl}`);
    });

    const shutdown = (signal) => {
      console.log(`[seo-service] ${signal} received — shutting down`);
      server.close(() => {
        console.log('[seo-service] Bye.');
        process.exit(0);
      });
    };

    process.on('SIGTERM', () => shutdown('SIGTERM'));
    process.on('SIGINT', () => shutdown('SIGINT'));
  } catch (err) {
    console.error('[seo-service] startup failed:', err);
    process.exit(1);
  }
}

// Only start the server when this file is the entry point, not when required by tests
if (require.main === module) {
  start();
}

module.exports = app; // exported for testing
