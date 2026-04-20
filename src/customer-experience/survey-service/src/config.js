'use strict';

require('dotenv').config();

module.exports = {
  HTTP_PORT: parseInt(process.env.HTTP_PORT || '8409', 10),
  GRPC_PORT: parseInt(process.env.GRPC_PORT || '50129', 10),
  DATABASE_URL: process.env.DATABASE_URL || 'postgresql://postgres:postgres@localhost:5432/survey_db',
  NODE_ENV: process.env.NODE_ENV || 'development',
  LOG_LEVEL: process.env.LOG_LEVEL || 'info',
  DB_POOL_MAX: parseInt(process.env.DB_POOL_MAX || '10', 10),
  DB_POOL_MIN: parseInt(process.env.DB_POOL_MIN || '2', 10),
  DB_IDLE_TIMEOUT_MS: parseInt(process.env.DB_IDLE_TIMEOUT_MS || '30000', 10),
  DB_CONNECTION_TIMEOUT_MS: parseInt(process.env.DB_CONNECTION_TIMEOUT_MS || '2000', 10),
};
