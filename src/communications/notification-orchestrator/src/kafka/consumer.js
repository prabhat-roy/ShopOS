'use strict';

const { Kafka } = require('kafkajs');
const config = require('../config');
const orchestrationService = require('../services/OrchestrationService');

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
 * Starts the Kafka consumer, subscribing to all configured notification topics.
 * Each message is parsed and routed through OrchestrationService.route().
 */
async function start() {
  if (running) {
    return;
  }

  const kafka = createKafkaInstance();
  consumer = kafka.consumer({ groupId: config.kafka.groupId });

  await consumer.connect();

  for (const topic of config.kafka.inputTopics) {
    await consumer.subscribe({ topic, fromBeginning: false });
  }

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const raw = message.value ? message.value.toString() : null;

      if (!raw) {
        console.warn(`[consumer] Received empty message on topic=${topic} partition=${partition}`);
        return;
      }

      let event;
      try {
        event = JSON.parse(raw);
      } catch (parseErr) {
        console.error(`[consumer] Failed to parse message on topic=${topic}: ${parseErr.message}`);
        return;
      }

      try {
        const result = await orchestrationService.route(topic, event);
        console.info(
          `[consumer] Routed messageId=${result.messageId} channel=${result.channel} topic=${topic}`,
        );
      } catch (routeErr) {
        console.error(
          `[consumer] Routing failed for topic=${topic} code=${routeErr.code || 'UNKNOWN'}: ${routeErr.message}`,
        );
      }
    },
  });

  running = true;
  console.info(`[consumer] Listening on topics: ${config.kafka.inputTopics.join(', ')}`);
}

/**
 * Gracefully stops the consumer.
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
