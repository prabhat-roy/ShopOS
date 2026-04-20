'use strict';

const { Router } = require('express');
const { isRunning } = require('../kafka/consumer');
const smsService = require('../services/SmsService');
const config = require('../config');

const router = Router();

// ─── Health ───────────────────────────────────────────────────────────────────

router.get('/healthz', (_req, res) => {
  res.status(200).json({ status: 'ok', service: config.service.name });
});

// ─── SMS Stats ────────────────────────────────────────────────────────────────

/**
 * GET /sms/stats
 * Returns aggregate delivery counters and consumer state.
 */
router.get('/sms/stats', (_req, res) => {
  const stats = smsService.getStats();
  const consumerRunning = isRunning();

  res.status(200).json({
    service: config.service.name,
    consumer: {
      running: consumerRunning,
      topic: config.kafka.topic,
      groupId: config.kafka.groupId,
    },
    stats,
    timestamp: new Date().toISOString(),
  });
});

// ─── Get SMS Record ───────────────────────────────────────────────────────────

/**
 * GET /sms/:messageId
 * Returns the stored record for a specific messageId.
 */
router.get('/sms/:messageId', (req, res) => {
  const record = smsService.getSmsLog(req.params.messageId);

  if (!record) {
    return res.status(404).json({
      error: `SMS record with messageId "${req.params.messageId}" not found`,
    });
  }

  res.status(200).json(record);
});

module.exports = router;
