'use strict';

require('dotenv').config();

const config = require('./src/config');
const db = require('./src/db');
const createApp = require('./src/app');
const SurveyRepository = require('./src/repositories/SurveyRepository');
const SurveyService = require('./src/services/SurveyService');
const SurveyController = require('./src/controllers/SurveyController');

async function main() {
  const surveyRepository = new SurveyRepository(db);
  const surveyService = new SurveyService(surveyRepository);
  const surveyController = new SurveyController(surveyService);

  try {
    await db.query('SELECT 1');
    console.log('[survey-service] Database connection established');
  } catch (err) {
    console.warn('[survey-service] Database unavailable — running without persistence:', err.message);
  }

  const app = createApp({ surveyController });

  const server = app.listen(config.HTTP_PORT, () => {
    console.log(`[survey-service] HTTP server listening on port ${config.HTTP_PORT}`);
    console.log(`[survey-service] Environment: ${config.NODE_ENV}`);
  });

  const shutdown = async (signal) => {
    console.log(`[survey-service] Received ${signal}, shutting down gracefully...`);
    server.close(async () => {
      console.log('[survey-service] HTTP server closed');
      try {
        await db.closePool();
        console.log('[survey-service] Database pool closed');
      } catch (err) {
        console.error('[survey-service] Error closing database pool:', err.message);
      }
      process.exit(0);
    });

    setTimeout(() => {
      console.error('[survey-service] Forced shutdown after timeout');
      process.exit(1);
    }, 10000);
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));

  process.on('unhandledRejection', (reason) => {
    console.error('[survey-service] Unhandled rejection:', reason);
  });

  process.on('uncaughtException', (err) => {
    console.error('[survey-service] Uncaught exception:', err);
    process.exit(1);
  });
}

main();
