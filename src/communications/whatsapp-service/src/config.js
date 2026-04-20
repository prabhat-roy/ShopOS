'use strict';

require('dotenv').config();

const config = {
  http: {
    port: parseInt(process.env.HTTP_PORT || '3020', 10),
  },
  kafka: {
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(',').map((b) => b.trim()),
    groupId: process.env.KAFKA_GROUP_ID || 'whatsapp-service',
    clientId: process.env.KAFKA_CLIENT_ID || 'whatsapp-service',
    topic: process.env.KAFKA_TOPIC || 'notification.whatsapp.requested',
    retryAttempts: parseInt(process.env.KAFKA_RETRY_ATTEMPTS || '8', 10),
    retryInitialRetryTime: parseInt(process.env.KAFKA_RETRY_INITIAL_TIME || '300', 10),
  },
  whatsapp: {
    apiUrl: process.env.WHATSAPP_API_URL || 'https://graph.facebook.com/v20.0',
    phoneNumberId: process.env.WHATSAPP_PHONE_NUMBER_ID || '',
    accessToken: process.env.WHATSAPP_ACCESS_TOKEN || '',
  },
  service: {
    name: 'whatsapp-service',
  },
};

module.exports = config;
