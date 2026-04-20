'use strict';

const { Kafka } = require('kafkajs');
const config = require('../config');
const whatsappService = require('../services/WhatsAppService');

let consumer = null;
let running = false;

function createKafkaInstance() {
  return new Kafka({
    clientId: `${config.kafka.clientId}-consumer`,
    brokers: config.kafka.brokers,
    retry: {
      attempts: config.kafka.retryAttempts,
      initialRetryTime: config.kafka.retryInitialRetryTime,
    },
  });
}

/**
 * Starts the Kafka consumer, subscribing to notification.whatsapp.requested.
 * Each message is parsed and dispatched to WhatsAppService.sendMessage().
 */
async function start() {
  if (running) {
    return;
  }

  const kafka = createKafkaInstance();
  consumer = kafka.consumer({ groupId: config.kafka.groupId });

  await consumer.connect();
  await consumer.subscribe({ topic: config.kafka.topic, fromBeginning: false });

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const raw = message.value ? message.value.toString() : null;

      if (!raw) {
        console.warn(`[consumer] Empty message on topic=${topic} partition=${partition}`);
        return;
      }

      let event;
      try {
        event = JSON.parse(raw);
      } catch (parseErr) {
        console.error(`[consumer] Failed to parse message on ${topic}: ${parseErr.message}`);
        return;
      }

      try {
        const result = await whatsappService.sendMessage({
          to: event.to,
          message: event.message,
          messageId: event.messageId,
        });

        console.info(
          `[consumer] Processed WhatsApp message | messageId=${result.messageId} status=${result.status}`,
        );
      } catch (err) {
        console.error(
          `[consumer] WhatsApp delivery failed code=${err.code || 'UNKNOWN'}: ${err.message}`,
        );
      }
    },
  });

  running = true;
  console.info(`[consumer] Subscribed to topic: ${config.kafka.topic}`);
}

/**
 * Gracefully stops the Kafka consumer.
 */
async function stop() {
  if (consumer) {
    await consumer.disconnect();
    consumer = null;
    running = false;
    console.info('[consumer] Disconnected');
  }
}

function isRunning() {
  return running;
}

module.exports = { start, stop, isRunning };
