'use strict';

const { Router } = require('express');
const templateService = require('../services/TemplateService');
const config = require('../config');

const router = Router();

// ─── Health ───────────────────────────────────────────────────────────────────

router.get('/healthz', (_req, res) => {
  res.status(200).json({ status: 'ok', service: config.service.name });
});

// ─── Create template ──────────────────────────────────────────────────────────

router.post('/templates', async (req, res) => {
  try {
    const template = await templateService.createTemplate(req.body);
    res.status(201).json(template);
  } catch (err) {
    if (err.code === 11000 || (err.message && err.message.includes('duplicate key'))) {
      return res.status(409).json({ error: 'Template name already exists' });
    }
    if (err.name === 'ValidationError') {
      return res.status(400).json({ error: 'Validation failed', details: err.message });
    }
    console.error('[TemplateController] createTemplate error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── Get template by id ───────────────────────────────────────────────────────

router.get('/templates/:id', async (req, res) => {
  try {
    const template = await templateService.getTemplate(req.params.id);
    res.status(200).json(template);
  } catch (err) {
    if (err.code === 'NOT_FOUND' || err.name === 'CastError') {
      return res.status(404).json({ error: err.message || 'Template not found' });
    }
    console.error('[TemplateController] getTemplate error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── Get template by name ─────────────────────────────────────────────────────

router.get('/templates/name/:name', async (req, res) => {
  try {
    const template = await templateService.getByName(req.params.name);
    res.status(200).json(template);
  } catch (err) {
    if (err.code === 'NOT_FOUND') {
      return res.status(404).json({ error: err.message });
    }
    console.error('[TemplateController] getByName error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── List templates ───────────────────────────────────────────────────────────

router.get('/templates', async (req, res) => {
  try {
    const filters = {
      channel: req.query.channel,
      active: req.query.active,
      tags: req.query.tags,
    };

    // Strip undefined keys
    Object.keys(filters).forEach((k) => filters[k] === undefined && delete filters[k]);

    const templates = await templateService.listTemplates(filters);
    res.status(200).json({ data: templates, count: templates.length });
  } catch (err) {
    console.error('[TemplateController] listTemplates error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── Update template ──────────────────────────────────────────────────────────

router.patch('/templates/:id', async (req, res) => {
  try {
    const updated = await templateService.updateTemplate(req.params.id, req.body);
    res.status(200).json(updated);
  } catch (err) {
    if (err.code === 'NOT_FOUND' || err.name === 'CastError') {
      return res.status(404).json({ error: err.message || 'Template not found' });
    }
    if (err.name === 'ValidationError') {
      return res.status(400).json({ error: 'Validation failed', details: err.message });
    }
    console.error('[TemplateController] updateTemplate error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── Archive (soft-delete) template ──────────────────────────────────────────

router.delete('/templates/:id', async (req, res) => {
  try {
    await templateService.archiveTemplate(req.params.id);
    res.status(204).send();
  } catch (err) {
    if (err.code === 'NOT_FOUND' || err.name === 'CastError') {
      return res.status(404).json({ error: err.message || 'Template not found' });
    }
    console.error('[TemplateController] archiveTemplate error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ─── Render template ──────────────────────────────────────────────────────────

router.post('/templates/render', async (req, res) => {
  const { name, variables } = req.body;

  if (!name) {
    return res.status(400).json({ error: 'Field "name" is required' });
  }

  try {
    const result = await templateService.renderTemplate(name, variables || {});
    res.status(200).json(result);
  } catch (err) {
    if (err.code === 'NOT_FOUND') {
      return res.status(404).json({ error: err.message });
    }
    if (err.code === 'MISSING_REQUIRED_VARIABLE') {
      return res.status(422).json({
        error: 'Missing required template variable',
        variable: err.variable,
        message: err.message,
      });
    }
    console.error('[TemplateController] renderTemplate error:', err.message);
    res.status(500).json({ error: 'Internal server error' });
  }
});

module.exports = router;
