'use strict';

require('dotenv').config();

const config = require('./src/config');
const db = require('./src/db');
const createApp = require('./src/app');
const ConsentRepository = require('./src/repositories/ConsentRepository');
const ConsentService = require('./src/services/ConsentService');
const ConsentController = require('./src/controllers/ConsentController');

async function main() {
  // Wire up dependencies
  const consentRepository = new ConsentRepository(db);
  const consentService = new ConsentService(consentRepository);
  const consentController = new ConsentController(consentService);

  // Verify database connectivity
  try {
    await db.query('SELECT 1');
    console.log('[consent-management-service] Database connection established');
  } catch (err) {
    console.warn('[consent-management-service] Database unavailable — running without persistence:', err.message);
  }

  const app = createApp({ consentController });

  const server = app.listen(config.HTTP_PORT, () => {
    console.log(`[consent-management-service] HTTP server listening on port ${config.HTTP_PORT}`);
    console.log(`[consent-management-service] Environment: ${config.NODE_ENV}`);
  });

  // Graceful shutdown
  const shutdown = async (signal) => {
    console.log(`[consent-management-service] Received ${signal}, shutting down gracefully...`);
    server.close(async () => {
      console.log('[consent-management-service] HTTP server closed');
      try {
        await db.closePool();
        console.log('[consent-management-service] Database pool closed');
      } catch (err) {
        console.error('[consent-management-service] Error closing database pool:', err.message);
      }
      process.exit(0);
    });

    // Force exit after 10s
    setTimeout(() => {
      console.error('[consent-management-service] Forced shutdown after timeout');
      process.exit(1);
    }, 10000);
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));

  process.on('unhandledRejection', (reason) => {
    console.error('[consent-management-service] Unhandled rejection:', reason);
  });

  process.on('uncaughtException', (err) => {
    console.error('[consent-management-service] Uncaught exception:', err);
    process.exit(1);
  });
}

main();
