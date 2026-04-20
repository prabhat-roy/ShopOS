'use strict';

const { Kafka } = require('kafkajs');
const config = require('../config');

let producer = null;
let isConnected = false;

function createKafkaInstance() {
  return new Kafka({
    clientId: `${config.kafka.clientId}-producer`,
    brokers: config.kafka.brokers,
    retry: {
      attempts: config.kafka.retryAttempts,
      initialRetryTime: config.kafka.retryInitialRetryTime,
    },
  });
}

async function getProducer() {
  if (!producer) {
    const kafka = createKafkaInstance();
    producer = kafka.producer({
      allowAutoTopicCreation: true,
      transactionTimeout: 30000,
    });
  }

  if (!isConnected) {
    await producer.connect();
    isConnected = true;
  }

  return producer;
}

/**
 * Publishes a message to the specified Kafka topic.
 * @param {string} topic - Target Kafka topic
 * @param {object} message - Message payload object (will be JSON-serialised)
 * @param {string} [key] - Optional partition key
 */
async function publish(topic, message, key = null) {
  const p = await getProducer();

  const record = {
    topic,
    messages: [
      {
        key: key ? String(key) : null,
        value: JSON.stringify(message),
        headers: {
          'content-type': 'application/json',
          'produced-by': config.service.name,
          timestamp: String(Date.now()),
        },
      },
    ],
  };

  await p.send(record);
}

async function disconnect() {
  if (producer && isConnected) {
    await producer.disconnect();
    isConnected = false;
    producer = null;
  }
}

module.exports = { publish, disconnect, getProducer };
