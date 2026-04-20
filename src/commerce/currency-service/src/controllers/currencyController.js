'use strict';

const rateStore = require('../rates');
const { convert, convertMany } = require('../converter');

/**
 * GET /currencies
 * Returns list of supported currency codes with their USD-base rates.
 */
function listCurrencies(req, res) {
  const currencies = rateStore.getSupportedCurrencies().map((code) => ({
    code,
    rate_vs_usd: rateStore.getRates()[code],
  }));

  res.json({
    base: 'USD',
    currencies,
    updated_at: rateStore.getUpdatedAt().toISOString(),
  });
}

/**
 * GET /currencies/rates
 * Returns all exchange rates (base USD).
 */
function getAllRates(req, res) {
  res.json({
    base: 'USD',
    rates: rateStore.getRates(),
    updated_at: rateStore.getUpdatedAt().toISOString(),
  });
}

/**
 * GET /currencies/rate?from=USD&to=EUR
 * Returns a single conversion rate.
 */
function getRate(req, res) {
  const { from, to } = req.query;

  if (!from || !to) {
    return res.status(400).json({ error: 'Query params "from" and "to" are required' });
  }

  try {
    const rate = rateStore.getRate(from, to);
    return res.json({
      from: from.toUpperCase(),
      to: to.toUpperCase(),
      rate,
      updated_at: rateStore.getUpdatedAt().toISOString(),
    });
  } catch (err) {
    return res.status(400).json({ error: err.message });
  }
}

/**
 * POST /currencies/convert
 * Body: { amount, from, to }
 * Returns conversion result.
 */
function convertCurrency(req, res) {
  const { amount, from, to } = req.body;

  if (amount === undefined || amount === null) {
    return res.status(400).json({ error: '"amount" is required' });
  }
  if (!from) {
    return res.status(400).json({ error: '"from" currency is required' });
  }
  if (!to) {
    return res.status(400).json({ error: '"to" currency is required' });
  }

  const numAmount = Number(amount);
  if (isNaN(numAmount)) {
    return res.status(400).json({ error: '"amount" must be a valid number' });
  }

  try {
    const result = convert(numAmount, from, to, rateStore);
    return res.json(result);
  } catch (err) {
    return res.status(400).json({ error: err.message });
  }
}

/**
 * POST /currencies/convert/many
 * Body: { amount, from, targets: string[] }
 * Returns array of conversion results.
 */
function convertToMany(req, res) {
  const { amount, from, targets } = req.body;

  if (amount === undefined || amount === null) {
    return res.status(400).json({ error: '"amount" is required' });
  }
  if (!from) {
    return res.status(400).json({ error: '"from" currency is required' });
  }
  if (!Array.isArray(targets) || targets.length === 0) {
    return res.status(400).json({ error: '"targets" must be a non-empty array' });
  }

  const numAmount = Number(amount);
  if (isNaN(numAmount)) {
    return res.status(400).json({ error: '"amount" must be a valid number' });
  }

  try {
    const results = convertMany(numAmount, from, targets, rateStore);
    return res.json({ results });
  } catch (err) {
    return res.status(400).json({ error: err.message });
  }
}

module.exports = {
  listCurrencies,
  getAllRates,
  getRate,
  convertCurrency,
  convertToMany,
};
