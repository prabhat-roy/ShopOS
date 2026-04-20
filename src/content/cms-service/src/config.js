'use strict';

require('dotenv').config();

const config = {
  port: parseInt(process.env.PORT || '8603', 10),
  mongoUri: process.env.MONGO_URI || 'mongodb://localhost:27017/cmsdb',
  nodeEnv: process.env.NODE_ENV || 'development',
  defaultLocale: process.env.DEFAULT_LOCALE || 'en',
};

module.exports = config;
