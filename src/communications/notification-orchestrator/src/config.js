'use strict';

require('dotenv').config();

const config = {
  http: {
    port: parseInt(process.env.HTTP_PORT || '8500', 10),
  },
  kafka: {
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(',').map((b) => b.trim()),
    groupId: process.env.KAFKA_GROUP_ID || 'notification-orchestrator',
    clientId: process.env.KAFKA_CLIENT_ID || 'notification-orchestrator',
    inputTopics: (
      process.env.INPUT_TOPICS ||
      'notification.email.requested,notification.sms.requested,notification.push.requested'
    )
      .split(',')
      .map((t) => t.trim()),
    outputTopics: {
      email: process.env.EMAIL_TOPIC || 'email.send',
      sms: process.env.SMS_TOPIC || 'sms.send',
      push: process.env.PUSH_TOPIC || 'push.send',
    },
    retryAttempts: parseInt(process.env.KAFKA_RETRY_ATTEMPTS || '8', 10),
    retryInitialRetryTime: parseInt(process.env.KAFKA_RETRY_INITIAL_TIME || '300', 10),
  },
  service: {
    name: 'notification-orchestrator',
  },
};

module.exports = config;
