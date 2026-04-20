'use strict';

const express = require('express');
const { Kafka } = require('kafkajs');

const HEALTH_PORT   = process.env.PORT || 8001;
const KAFKA_BROKERS = (process.env.KAFKA_BROKERS || 'kafka:9092').split(',');
const SERVICE_NAME  = 'sms-service';

// --- Express health endpoint ---
const app = express();
app.use(express.json());

app.get('/healthz', (_req, res) => {
  res.json({ status: 'ok' });
});

const server = app.listen(HEALTH_PORT, () => {
  console.log(`${SERVICE_NAME} healthz listening on port ${HEALTH_PORT}`);
});

// --- Kafka consumer ---
const kafka    = new Kafka({ clientId: SERVICE_NAME, brokers: KAFKA_BROKERS });
const consumer = kafka.consumer({ groupId: `${SERVICE_NAME}-group` });

async function startConsumer() {
  await consumer.connect();
  console.log(`${SERVICE_NAME} connected to Kafka brokers: ${KAFKA_BROKERS.join(', ')}`);

  await consumer.subscribe({ topic: 'notification.sms.requested', fromBeginning: false });
  console.log('Subscribed to topic: notification.sms.requested');

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const value = message.value ? message.value.toString() : null;
      console.log(`[${topic}] partition=${partition} offset=${message.offset} value=${value}`);
      // TODO: Integrate SMS provider (e.g., Twilio, Vonage, AWS SNS)
      //       Parse message payload, build SMS body, call provider SDK
    },
  });
}

startConsumer().catch((err) => {
  console.error('Failed to start Kafka consumer:', err);
  process.exit(1);
});

// Graceful shutdown
async function shutdown() {
  console.log('SIGTERM received — shutting down gracefully');
  await consumer.disconnect();
  server.close(() => {
    console.log('HTTP server closed');
    process.exit(0);
  });
}

process.on('SIGTERM', shutdown);
process.on('SIGINT',  shutdown);
