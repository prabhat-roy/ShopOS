'use strict';

const express = require('express');
const mongoose = require('mongoose');
const config = require('./src/config');
const configuratorRoutes = require('./src/routes/configuratorRoutes');

const app = express();

app.use(express.json());

// Routes
app.use('/', configuratorRoutes);

// 404 handler
app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

// Global error handler
app.use((err, req, res, next) => {
  console.error('[configurator-service] error:', err);
  const status = err.status || 500;
  res.status(status).json({ error: err.message || 'Internal server error' });
});

async function start() {
  try {
    await mongoose.connect(config.mongoUri, { dbName: config.dbName });
    console.log(`[configurator-service] MongoDB connected: ${config.mongoUri}/${config.dbName}`);
  } catch (err) {
    console.warn('[configurator-service] MongoDB unavailable — running without persistence:', err.message);
  }

  const server = app.listen(config.httpPort, () => {
    console.log(`[configurator-service] HTTP server listening on port ${config.httpPort}`);
  });

  const shutdown = async (signal) => {
    console.log(`[configurator-service] ${signal} received — shutting down`);
    server.close(async () => {
      try { await mongoose.disconnect(); } catch (_) {}
      console.log('[configurator-service] Bye.');
      process.exit(0);
    });
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

// Only start the server when this file is the entry point, not when required by tests
if (require.main === module) {
  start();
}

module.exports = app; // exported for testing
