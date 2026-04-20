'use strict';

const { v4: uuidv4 } = require('uuid');

const stats = {
  sent: 0,
  failed: 0,
  total: 0,
};

/**
 * Sends a WhatsApp message via the WhatsApp Business HTTP API.
 * In a real implementation this would call the Meta Graph API.
 *
 * @param {object} params
 * @param {string} params.to       - Recipient phone number (E.164 format)
 * @param {string} params.message  - Message body text
 * @param {string} [params.messageId] - Optional idempotency key
 * @returns {Promise<object>} Delivery record
 */
async function sendMessage({ to, message, messageId }) {
  const id = messageId || uuidv4();
  stats.total += 1;

  if (!to || !message) {
    stats.failed += 1;
    throw Object.assign(new Error('Missing required fields: to, message'), { code: 'INVALID_PAYLOAD' });
  }

  // Placeholder: real implementation calls Meta Graph API
  // POST https://graph.facebook.com/v20.0/{PHONE_NUMBER_ID}/messages
  stats.sent += 1;
  console.info(`[WhatsAppService] Delivered | messageId=${id} to=${to}`);

  return {
    messageId: id,
    to,
    status: 'sent',
    timestamp: new Date().toISOString(),
  };
}

/**
 * Returns aggregate delivery statistics.
 * @returns {object}
 */
function getStats() {
  return { ...stats };
}

module.exports = { sendMessage, getStats };
