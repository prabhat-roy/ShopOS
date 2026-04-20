'use strict';

require('dotenv').config();

const mongoose = require('mongoose');
const createApp = require('./src/app');
const config = require('./src/config');

const app = createApp();

async function start() {
  try {
    await mongoose.connect(config.mongodbUri, {
      serverSelectionTimeoutMS: 5000,
    });
    console.log(`[tracking-service] MongoDB connected: ${config.mongodbUri}`);
  } catch (err) {
    console.warn('[tracking-service] MongoDB unavailable — running without persistence:', err.message);
  }

  const server = app.listen(config.httpPort, () => {
    console.log(`[tracking-service] HTTP server listening on port ${config.httpPort}`);
    console.log(`[tracking-service] gRPC port configured: ${config.grpcPort}`);
    console.log(`[tracking-service] Environment: ${config.nodeEnv}`);
  });

  // Graceful shutdown
  const shutdown = async (signal) => {
    console.log(`[tracking-service] Received ${signal}. Shutting down gracefully...`);
    server.close(async () => {
      try {
        await mongoose.connection.close();
        console.log('[tracking-service] MongoDB connection closed.');
        console.log('[tracking-service] HTTP server closed.');
        process.exit(0);
      } catch (err) {
        console.error('[tracking-service] Error during shutdown:', err.message);
        process.exit(1);
      }
    });

    // Force exit after 10 seconds if graceful shutdown stalls
    setTimeout(() => {
      console.error('[tracking-service] Forced shutdown after timeout.');
      process.exit(1);
    }, 10000);
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

start();
