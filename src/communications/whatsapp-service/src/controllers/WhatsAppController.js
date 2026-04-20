'use strict';

const { Router } = require('express');
const { isRunning } = require('../kafka/consumer');
const whatsappService = require('../services/WhatsAppService');
const config = require('../config');

const router = Router();

// ─── Health ───────────────────────────────────────────────────────────────────

router.get('/healthz', (_req, res) => {
  res.status(200).json({ status: 'ok', service: config.service.name });
});

// ─── Stats ────────────────────────────────────────────────────────────────────

/**
 * GET /whatsapp/stats
 * Returns aggregate delivery counters and consumer state.
 */
router.get('/whatsapp/stats', (_req, res) => {
  const stats = whatsappService.getStats();
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

module.exports = router;
