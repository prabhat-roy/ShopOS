'use strict';

require('dotenv').config();

const config = {
  NODE_ENV: process.env.NODE_ENV || 'development',
  HTTP_PORT: parseInt(process.env.HTTP_PORT, 10) || 8400,
  GRPC_PORT: parseInt(process.env.GRPC_PORT, 10) || 50120,
  MONGODB_URI: process.env.MONGODB_URI || 'mongodb://localhost:27017/review-rating-service',
  LOG_LEVEL: process.env.LOG_LEVEL || 'info',
};

module.exports = config;
