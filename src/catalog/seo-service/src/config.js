'use strict';

require('dotenv').config();

const config = {
  httpPort: parseInt(process.env.HTTP_PORT, 10) || 8118,
  grpcPort: parseInt(process.env.GRPC_PORT, 10) || 50079,
  baseUrl: process.env.BASE_URL || 'https://shopos.com',
  nodeEnv: process.env.NODE_ENV || 'development',
};

module.exports = config;
