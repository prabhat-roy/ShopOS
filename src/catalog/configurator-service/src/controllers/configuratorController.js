'use strict';

const Joi = require('joi');
const service = require('../services/configuratorService');

// ─── Validation schemas ───────────────────────────────────────────────────────

const choiceSchema = Joi.object({
  value: Joi.string().required(),
  label: Joi.string().required(),
  priceAdj: Joi.number().default(0),
  available: Joi.boolean().default(true),
});

const optionSchema = Joi.object({
  name: Joi.string().required(),
  type: Joi.string().valid('select', 'radio', 'checkbox').required(),
  required: Joi.boolean().default(false),
  choices: Joi.array().items(choiceSchema).default([]),
});

const ruleSchema = Joi.object({
  condition: Joi.object({
    optionName: Joi.string().required(),
    value: Joi.string().required(),
  }).required(),
  effect: Joi.object({
    optionName: Joi.string().required(),
    allowedValues: Joi.array().items(Joi.string()).required(),
  }).required(),
});

const createUpdateBodySchema = Joi.object({
  options: Joi.array().items(optionSchema).default([]),
  rules: Joi.array().items(ruleSchema).default([]),
});

const validateBodySchema = Joi.object({
  selections: Joi.object().pattern(
    Joi.string(),
    Joi.alternatives().try(Joi.string(), Joi.array().items(Joi.string()))
  ).required(),
});

// ─── Helpers ──────────────────────────────────────────────────────────────────

function handleError(res, err) {
  const status = err.status || 500;
  res.status(status).json({ error: err.message });
}

// ─── Controllers ─────────────────────────────────────────────────────────────

async function getConfigurator(req, res) {
  try {
    const doc = await service.getConfigurator(req.params.productId);
    res.json(doc);
  } catch (err) {
    handleError(res, err);
  }
}

async function createConfigurator(req, res) {
  const { error, value } = createUpdateBodySchema.validate(req.body, { abortEarly: false });
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const doc = await service.createConfigurator(req.params.productId, value);
    res.status(201).json(doc);
  } catch (err) {
    handleError(res, err);
  }
}

async function updateConfigurator(req, res) {
  const { error, value } = createUpdateBodySchema.validate(req.body, { abortEarly: false });
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const doc = await service.updateConfigurator(req.params.productId, value);
    res.json(doc);
  } catch (err) {
    handleError(res, err);
  }
}

async function deleteConfigurator(req, res) {
  try {
    await service.deleteConfigurator(req.params.productId);
    res.status(204).send();
  } catch (err) {
    handleError(res, err);
  }
}

async function validateSelection(req, res) {
  const { error, value } = validateBodySchema.validate(req.body, { abortEarly: false });
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const result = await service.validateSelection(req.params.productId, value.selections);
    const statusCode = result.valid ? 200 : 422;
    res.status(statusCode).json(result);
  } catch (err) {
    handleError(res, err);
  }
}

module.exports = {
  getConfigurator,
  createConfigurator,
  updateConfigurator,
  deleteConfigurator,
  validateSelection,
};
