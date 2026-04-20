'use strict';

require('dotenv').config();

const { createApp } = require('./src/app');
const config = require('./src/config');
const consumer = require('./src/kafka/consumer');
const { disconnect: disconnectProducer } = require('./src/kafka/producer');

const app = createApp();

async function main() {
  // Start Kafka consumer
  try {
    await consumer.start();
    console.info('[main] Kafka consumer started');
  } catch (err) {
    console.error('[main] Failed to start Kafka consumer:', err.message);
    // Continue — HTTP server still useful for health checks even if Kafka is down
  }

  // Start HTTP server
  const server = app.listen(config.http.port, () => {
    console.info(`[main] ${config.service.name} HTTP server listening on port ${config.http.port}`);
  });

  // Graceful shutdown
  async function shutdown(signal) {
    console.info(`[main] Received ${signal}, shutting down gracefully...`);

    server.close(async () => {
      try {
        await consumer.stop();
        await disconnectProducer();
        console.info('[main] Clean shutdown complete');
        process.exit(0);
      } catch (err) {
        console.error('[main] Error during shutdown:', err.message);
        process.exit(1);
      }
    });

    // Force exit after 15 seconds
    setTimeout(() => {
      console.error('[main] Forced shutdown after timeout');
      process.exit(1);
    }, 15000);
  }

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

main().catch((err) => {
  console.error('[main] Fatal startup error:', err.message);
  process.exit(1);
});
