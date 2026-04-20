'use strict';

const emailChannel = require('../channels/emailChannel');
const smsChannel = require('../channels/smsChannel');
const pushChannel = require('../channels/pushChannel');
const config = require('../config');

// Mapping of Kafka input topic → channel handler
const TOPIC_CHANNEL_MAP = {
  'notification.email.requested': emailChannel,
  'notification.sms.requested': smsChannel,
  'notification.push.requested': pushChannel,
};

// Internal stats counters
let stats = {
  processed: 0,
  succeeded: 0,
  failed: 0,
  byChannel: { email: 0, sms: 0, push: 0 },
};

/**
 * Routes an incoming Kafka event to the appropriate channel handler.
 *
 * @param {string} topic - The Kafka topic the event arrived on
 * @param {object} event - Parsed JSON payload from the Kafka message
 * @returns {Promise<object>} - The routed envelope returned by the channel handler
 * @throws {Error} - If the topic is unknown or validation/publishing fails
 */
async function route(topic, event) {
  stats.processed++;

  const channel = TOPIC_CHANNEL_MAP[topic];

  if (!channel) {
    const err = new Error(`Unknown notification topic: "${topic}". Supported topics: ${Object.keys(TOPIC_CHANNEL_MAP).join(', ')}`);
    err.code = 'UNKNOWN_TOPIC';
    stats.failed++;
    throw err;
  }

  try {
    const result = await channel.handle(event);

    stats.succeeded++;

    const channelName = result.channel;
    if (channelName && stats.byChannel[channelName] !== undefined) {
      stats.byChannel[channelName]++;
    }

    return result;
  } catch (err) {
    stats.failed++;
    throw err;
  }
}

/**
 * Returns current orchestration statistics.
 * @returns {object}
 */
function getStats() {
  return { ...stats, byChannel: { ...stats.byChannel } };
}

/**
 * Resets statistics counters (useful for testing).
 */
function resetStats() {
  stats = {
    processed: 0,
    succeeded: 0,
    failed: 0,
    byChannel: { email: 0, sms: 0, push: 0 },
  };
}

/**
 * Returns the list of supported input topics.
 * @returns {string[]}
 */
function getSupportedTopics() {
  return Object.keys(TOPIC_CHANNEL_MAP);
}

module.exports = { route, getStats, resetStats, getSupportedTopics };
