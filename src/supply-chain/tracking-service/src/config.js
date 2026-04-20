'use strict';

require('dotenv').config();

const config = {
  httpPort: parseInt(process.env.HTTP_PORT, 10) || 8204,
  grpcPort: parseInt(process.env.GRPC_PORT, 10) || 50104,
  mongodbUri: process.env.MONGODB_URI || 'mongodb://localhost:27017/tracking-service',
  nodeEnv: process.env.NODE_ENV || 'development',
};

module.exports = config;
