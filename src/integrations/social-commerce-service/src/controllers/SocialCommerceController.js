'use strict';

const CatalogSyncService = require('../services/CatalogSyncService');

const catalogSyncService = new CatalogSyncService();

/**
 * POST /social/catalog/sync
 *
 * Synchronises a batch of products to a social commerce platform.
 *
 * Request body:
 * {
 *   "platform": "INSTAGRAM" | "TIKTOK" | "PINTEREST" | "FACEBOOK",
 *   "products": [ { productId, sku, title, description, price, currency, imageUrl, link, availability } ]
 * }
 *
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
function syncCatalog(req, res) {
  const { platform, products } = req.body;

  if (!platform) {
    return res.status(400).json({
      error: 'BAD_REQUEST',
      message: 'platform is required',
    });
  }

  if (!Array.isArray(products)) {
    return res.status(400).json({
      error: 'BAD_REQUEST',
      message: 'products must be an array',
    });
  }

  try {
    const result = catalogSyncService.syncCatalog(platform.toUpperCase(), products);
    return res.status(201).json(result);
  } catch (err) {
    if (err.message.startsWith('Unsupported platform')) {
      return res.status(400).json({ error: 'BAD_REQUEST', message: err.message });
    }
    return res.status(500).json({ error: 'INTERNAL_ERROR', message: err.message });
  }
}

/**
 * POST /social/orders/parse
 *
 * Parses an incoming social platform order webhook payload into a canonical SocialOrder.
 *
 * Request body:
 * {
 *   "platform": "INSTAGRAM" | "TIKTOK" | "PINTEREST" | "FACEBOOK",
 *   "rawOrder": { ... platform-specific order payload ... }
 * }
 *
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
function parseOrder(req, res) {
  const { platform, rawOrder } = req.body;

  if (!platform) {
    return res.status(400).json({ error: 'BAD_REQUEST', message: 'platform is required' });
  }
  if (!rawOrder || typeof rawOrder !== 'object') {
    return res.status(400).json({ error: 'BAD_REQUEST', message: 'rawOrder object is required' });
  }

  try {
    const socialOrder = catalogSyncService.parsePlatformOrder(platform.toUpperCase(), rawOrder);
    return res.status(200).json(socialOrder);
  } catch (err) {
    if (err.message.startsWith('Unsupported platform')) {
      return res.status(400).json({ error: 'BAD_REQUEST', message: err.message });
    }
    return res.status(500).json({ error: 'INTERNAL_ERROR', message: err.message });
  }
}

/**
 * GET /social/syncs
 *
 * Returns recent catalog sync records, optionally filtered by platform.
 *
 * Query params: platform (required), limit (optional, default 20)
 *
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
function getRecentSyncs(req, res) {
  const { platform, limit } = req.query;

  if (!platform) {
    return res.status(400).json({ error: 'BAD_REQUEST', message: 'platform query param is required' });
  }

  try {
    const syncs = catalogSyncService.getRecentSyncs(
      platform.toUpperCase(),
      parseInt(limit, 10) || 20
    );
    return res.status(200).json(syncs);
  } catch (err) {
    if (err.message.startsWith('Unsupported platform')) {
      return res.status(400).json({ error: 'BAD_REQUEST', message: err.message });
    }
    return res.status(500).json({ error: 'INTERNAL_ERROR', message: err.message });
  }
}

/**
 * GET /social/platforms
 *
 * Lists all social platforms supported by this service.
 *
 * @param {import('express').Request} _req
 * @param {import('express').Response} res
 */
function getSupportedPlatforms(_req, res) {
  const platforms = catalogSyncService.getSupportedPlatforms();
  return res.status(200).json({ platforms });
}

/**
 * GET /social/field-mappings/:platform
 *
 * Returns the field-name mapping between ShopOS canonical names and platform-native names.
 *
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
function getFieldMappings(req, res) {
  const { platform } = req.params;

  try {
    const mapping = catalogSyncService.getFieldMapping(platform.toUpperCase());
    return res.status(200).json({ platform: platform.toUpperCase(), fieldMapping: mapping });
  } catch (err) {
    if (err.message.startsWith('Unsupported platform')) {
      return res.status(400).json({ error: 'BAD_REQUEST', message: err.message });
    }
    return res.status(500).json({ error: 'INTERNAL_ERROR', message: err.message });
  }
}

/**
 * GET /healthz
 *
 * Kubernetes / load-balancer liveness and readiness check.
 *
 * @param {import('express').Request} _req
 * @param {import('express').Response} res
 */
function healthz(_req, res) {
  return res.status(200).json({ status: 'ok' });
}

module.exports = {
  syncCatalog,
  parseOrder,
  getRecentSyncs,
  getSupportedPlatforms,
  getFieldMappings,
  healthz,
};
