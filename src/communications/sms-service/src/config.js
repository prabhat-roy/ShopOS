'use strict';

require('dotenv').config();

const config = {
  http: {
    port: parseInt(process.env.HTTP_PORT || '8502', 10),
  },
  kafka: {
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(',').map((b) => b.trim()),
    groupId: process.env.KAFKA_GROUP_ID || 'sms-service',
    clientId: process.env.KAFKA_CLIENT_ID || 'sms-service',
    topic: process.env.KAFKA_TOPIC || 'sms.send',
    retryAttempts: parseInt(process.env.KAFKA_RETRY_ATTEMPTS || '8', 10),
    retryInitialRetryTime: parseInt(process.env.KAFKA_RETRY_INITIAL_TIME || '300', 10),
  },
  sms: {
    // Simulated delivery success rate (0.0–1.0)
    successRate: parseFloat(process.env.SMS_SUCCESS_RATE || '0.9'),
    // Maximum messages to keep in the in-memory log
    maxLogSize: parseInt(process.env.SMS_MAX_LOG_SIZE || '10000', 10),
  },
  service: {
    name: 'sms-service',
  },
};

module.exports = config;
