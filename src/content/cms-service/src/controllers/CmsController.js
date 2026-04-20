'use strict';

const CmsService = require('../services/CmsService');
const { NotFoundError, ValidationError } = require('../services/CmsService');

function handleError(res, err) {
  if (err instanceof NotFoundError) {
    return res.status(404).json({ error: err.message });
  }
  if (err instanceof ValidationError || err.name === 'ValidationError') {
    return res.status(400).json({ error: err.message });
  }
  if (err.code === 11000) {
    return res.status(409).json({ error: 'Slug already exists' });
  }
  console.error('[CmsController] Unhandled error:', err);
  return res.status(500).json({ error: 'Internal server error' });
}

async function createContent(req, res) {
  try {
    const item = await CmsService.createContent(req.body);
    res.status(201).json(item);
  } catch (err) {
    handleError(res, err);
  }
}

async function getById(req, res) {
  try {
    const item = await CmsService.getById(req.params.id);
    res.json(item);
  } catch (err) {
    handleError(res, err);
  }
}

async function getBySlug(req, res) {
  try {
    const locale = req.query.locale;
    const item = await CmsService.getBySlug(req.params.slug, locale);
    res.json(item);
  } catch (err) {
    handleError(res, err);
  }
}

async function listContent(req, res) {
  try {
    const { type, status, locale, tags, limit, offset } = req.query;
    const result = await CmsService.listContent({
      type,
      status,
      locale,
      tags: tags ? tags.split(',') : undefined,
      limit: limit ? parseInt(limit, 10) : 20,
      offset: offset ? parseInt(offset, 10) : 0,
    });
    res.json(result);
  } catch (err) {
    handleError(res, err);
  }
}

async function updateContent(req, res) {
  try {
    const item = await CmsService.updateContent(req.params.id, req.body);
    res.json(item);
  } catch (err) {
    handleError(res, err);
  }
}

async function publishContent(req, res) {
  try {
    await CmsService.publishContent(req.params.id);
    res.status(204).send();
  } catch (err) {
    handleError(res, err);
  }
}

async function archiveContent(req, res) {
  try {
    await CmsService.archiveContent(req.params.id);
    res.status(204).send();
  } catch (err) {
    handleError(res, err);
  }
}

async function deleteContent(req, res) {
  try {
    await CmsService.deleteContent(req.params.id);
    res.status(204).send();
  } catch (err) {
    handleError(res, err);
  }
}

async function searchContent(req, res) {
  try {
    const { q, locale } = req.query;
    const results = await CmsService.searchContent(q, locale);
    res.json(results);
  } catch (err) {
    handleError(res, err);
  }
}

module.exports = {
  createContent,
  getById,
  getBySlug,
  listContent,
  updateContent,
  publishContent,
  archiveContent,
  deleteContent,
  searchContent,
};
