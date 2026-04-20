'use strict';

require('dotenv').config();

const config = require('./src/config');
const { createApp } = require('./src/app');

const app = createApp();

const server = app.listen(config.HTTP_PORT, () => {
  console.log(`[social-commerce-service] HTTP server listening on port ${config.HTTP_PORT}`);
  console.log(`[social-commerce-service] gRPC port reserved at ${config.GRPC_PORT}`);
  console.log(`[social-commerce-service] Environment: ${config.NODE_ENV}`);
});

/**
 * Graceful shutdown handler.
 * Stops accepting new connections, waits for in-flight requests to drain,
 * then exits.  Kubernetes sends SIGTERM before SIGKILL, giving the service
 * time to finish outstanding work.
 */
function shutdown(signal) {
  console.log(`[social-commerce-service] Received ${signal}, shutting down gracefully...`);
  server.close((err) => {
    if (err) {
      console.error('[social-commerce-service] Error during shutdown:', err);
      process.exit(1);
    }
    console.log('[social-commerce-service] Server closed. Exiting.');
    process.exit(0);
  });

  // Force exit after 10 seconds if graceful shutdown stalls
  setTimeout(() => {
    console.error('[social-commerce-service] Forced exit after 10s shutdown timeout.');
    process.exit(1);
  }, 10_000).unref();
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT',  () => shutdown('SIGINT'));

module.exports = { app, server };
