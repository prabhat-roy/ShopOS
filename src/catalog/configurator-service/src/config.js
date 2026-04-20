'use strict';

require('dotenv').config();

const config = {
  httpPort: parseInt(process.env.HTTP_PORT, 10) || 8117,
  mongoUri: process.env.MONGODB_URI || 'mongodb://localhost:27017',
  dbName: process.env.DB_NAME || 'configurator',
  nodeEnv: process.env.NODE_ENV || 'development',
};

module.exports = config;
