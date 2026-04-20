'use strict';

// dotenv is loaded by the entry point (index.js). When running in test mode
// the test runner may not have a .env file, so we do a best-effort load here.
try {
  require('dotenv').config({ path: require('path').join(__dirname, '..', '.env') });
} catch (_) {
  // dotenv not available or no .env file — environment variables are already set
}

module.exports = {
  httpPort: parseInt(process.env.HTTP_PORT || '8139', 10),
  grpcPort: parseInt(process.env.GRPC_PORT || '50085', 10),
  baseCurrency: (process.env.BASE_CURRENCY || 'USD').toUpperCase(),
  // How often to simulate a rate refresh (minutes)
  rateRefreshCronExpr: process.env.RATE_REFRESH_CRON || '0 * * * *',
};
