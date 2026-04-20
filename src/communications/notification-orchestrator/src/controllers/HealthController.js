'use strict';

const { Router } = require('express');
const { isRunning } = require('../kafka/consumer');
const { getStats } = require('../services/OrchestrationService');
const config = require('../config');

const router = Router();

/**
 * GET /healthz
 * Liveness probe — always returns 200 when the process is alive.
 */
router.get('/healthz', (_req, res) => {
  res.status(200).json({ status: 'ok', service: config.service.name });
});

/**
 * GET /status
 * Returns consumer running state and orchestration stats.
 */
router.get('/status', (_req, res) => {
  const consumerRunning = isRunning();
  const stats = getStats();

  const statusCode = consumerRunning ? 200 : 503;

  res.status(statusCode).json({
    service: config.service.name,
    consumer: {
      running: consumerRunning,
      topics: config.kafka.inputTopics,
      groupId: config.kafka.groupId,
    },
    stats,
    timestamp: new Date().toISOString(),
  });
});

module.exports = router;
