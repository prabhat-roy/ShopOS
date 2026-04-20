'use strict';

require('dotenv').config();

const config = {
  http: {
    port: parseInt(process.env.HTTP_PORT || '8501', 10),
  },
  mongodb: {
    uri: process.env.MONGODB_URI || 'mongodb://localhost:27017/template-service',
    options: {
      serverSelectionTimeoutMS: 5000,
      connectTimeoutMS: 10000,
    },
  },
  service: {
    name: 'template-service',
  },
};

module.exports = config;
