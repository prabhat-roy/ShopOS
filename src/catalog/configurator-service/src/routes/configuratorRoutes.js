'use strict';

const express = require('express');
const ctrl = require('../controllers/configuratorController');

const router = express.Router();

// Health check
router.get('/healthz', (req, res) => {
  res.json({ status: 'ok' });
});

// Get configurator for a product
router.get('/configurators/:productId', ctrl.getConfigurator);

// Create / upsert configurator for a product
router.post('/configurators/:productId', ctrl.createConfigurator);

// Update configurator for a product
router.put('/configurators/:productId', ctrl.updateConfigurator);

// Delete configurator for a product
router.delete('/configurators/:productId', ctrl.deleteConfigurator);

// Validate a set of selections against a product's configurator rules
router.post('/configurators/:productId/validate', ctrl.validateSelection);

module.exports = router;
