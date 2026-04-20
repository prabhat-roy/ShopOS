'use strict';

const { Router } = require('express');
const ctrl = require('../controllers/currencyController');

const router = Router();

// List all supported currencies with their rates
router.get('/', ctrl.listCurrencies);

// All rates (base USD)
router.get('/rates', ctrl.getAllRates);

// Single rate  — must come before /:code to avoid shadowing
router.get('/rate', ctrl.getRate);

// Convert to many — registered BEFORE /convert so Express matches the longer path first
router.post('/convert/many', ctrl.convertToMany);

// Convert single
router.post('/convert', ctrl.convertCurrency);

module.exports = router;
