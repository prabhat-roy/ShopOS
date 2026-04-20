'use strict';

require('dotenv').config();

const mongoose = require('mongoose');
const { createApp } = require('./app');
const config = require('./config');

async function start() {
  try {
    await mongoose.connect(config.MONGODB_URI);
    console.log(`[qa-service] Connected to MongoDB at ${config.MONGODB_URI}`);
  } catch (err) {
    console.warn('[qa-service] MongoDB unavailable — running without persistence:', err.message);
  }

  const app = createApp();

  const server = app.listen(config.HTTP_PORT, () => {
    console.log(`[qa-service] HTTP server listening on port ${config.HTTP_PORT}`);
  });

  const shutdown = async (signal) => {
    console.log(`[qa-service] Received ${signal}, shutting down gracefully...`);
    server.close(async () => {
      try { await mongoose.connection.close(); } catch (_) {}
      process.exit(0);
    });
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

start();