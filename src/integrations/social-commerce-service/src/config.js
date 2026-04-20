'use strict';

require('dotenv').config();

/**
 * Centralised configuration loaded from environment variables.
 * All values have safe defaults so the service can start in a local/test
 * environment without a fully populated .env file.
 */
const config = {
  /** HTTP server port */
  HTTP_PORT: parseInt(process.env.HTTP_PORT || '8902', 10),

  /** gRPC server port (reserved for Phase 1 gRPC server implementation) */
  GRPC_PORT: parseInt(process.env.GRPC_PORT || '50171', 10),

  /** Node environment */
  NODE_ENV: process.env.NODE_ENV || 'development',

  /**
   * Social platform API credentials.
   * All are optional — if absent, the service operates in "simulation" mode
   * and returns synthetic responses.
   */
  INSTAGRAM_API_KEY: process.env.INSTAGRAM_API_KEY || '',
  INSTAGRAM_CATALOG_ID: process.env.INSTAGRAM_CATALOG_ID || '',
  INSTAGRAM_BUSINESS_ACCOUNT_ID: process.env.INSTAGRAM_BUSINESS_ACCOUNT_ID || '',

  TIKTOK_API_KEY: process.env.TIKTOK_API_KEY || '',
  TIKTOK_SHOP_ID: process.env.TIKTOK_SHOP_ID || '',
  TIKTOK_APP_SECRET: process.env.TIKTOK_APP_SECRET || '',

  PINTEREST_API_KEY: process.env.PINTEREST_API_KEY || '',
  PINTEREST_AD_ACCOUNT_ID: process.env.PINTEREST_AD_ACCOUNT_ID || '',

  FACEBOOK_ACCESS_TOKEN: process.env.FACEBOOK_ACCESS_TOKEN || '',
  FACEBOOK_CATALOG_ID: process.env.FACEBOOK_CATALOG_ID || '',

  /** Maximum number of recent sync records to keep per platform (in-memory). */
  MAX_SYNC_HISTORY: parseInt(process.env.MAX_SYNC_HISTORY || '500', 10),
};

module.exports = config;
