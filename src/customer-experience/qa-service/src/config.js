'use strict';

require('dotenv').config();

const config = {
  NODE_ENV: process.env.NODE_ENV || 'development',
  HTTP_PORT: parseInt(process.env.HTTP_PORT, 10) || 8401,
  GRPC_PORT: parseInt(process.env.GRPC_PORT, 10) || 50121,
  MONGODB_URI: process.env.MONGODB_URI || 'mongodb://localhost:27017/qa-service',
  LOG_LEVEL: process.env.LOG_LEVEL || 'info',
};

module.exports = config;
