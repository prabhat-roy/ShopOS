'use strict';

require('dotenv').config();

const config = require('./src/config');
const db = require('./src/db');
const createApp = require('./src/app');
const FeedbackRepository = require('./src/repositories/FeedbackRepository');
const FeedbackService = require('./src/services/FeedbackService');
const FeedbackController = require('./src/controllers/FeedbackController');

async function main() {
  const feedbackRepository = new FeedbackRepository(db);
  const feedbackService = new FeedbackService(feedbackRepository);
  const feedbackController = new FeedbackController(feedbackService);

  try {
    await db.query('SELECT 1');
    console.log('[feedback-service] Database connection established');
  } catch (err) {
    console.warn('[feedback-service] Database unavailable — running without persistence:', err.message);
  }

  const app = createApp({ feedbackController });

  const server = app.listen(config.HTTP_PORT, () => {
    console.log(`[feedback-service] HTTP server listening on port ${config.HTTP_PORT}`);
    console.log(`[feedback-service] Environment: ${config.NODE_ENV}`);
  });

  const shutdown = async (signal) => {
    console.log(`[feedback-service] Received ${signal}, shutting down gracefully...`);
    server.close(async () => {
      console.log('[feedback-service] HTTP server closed');
      try {
        await db.closePool();
        console.log('[feedback-service] Database pool closed');
      } catch (err) {
        console.error('[feedback-service] Error closing database pool:', err.message);
      }
      process.exit(0);
    });

    setTimeout(() => {
      console.error('[feedback-service] Forced shutdown after timeout');
      process.exit(1);
    }, 10000);
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));

  process.on('unhandledRejection', (reason) => {
    console.error('[feedback-service] Unhandled rejection:', reason);
  });

  process.on('uncaughtException', (err) => {
    console.error('[feedback-service] Uncaught exception:', err);
    process.exit(1);
  });
}

main();
