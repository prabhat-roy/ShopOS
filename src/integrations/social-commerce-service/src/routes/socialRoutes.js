'use strict';

const { Router } = require('express');
const controller = require('../controllers/SocialCommerceController');

const router = Router();

/**
 * POST /social/catalog/sync
 * Sync a batch of ShopOS products to a social platform catalog.
 * Body: { platform: string, products: ProductCatalogItem[] }
 * Response: 201 SyncResult
 */
router.post('/catalog/sync', controller.syncCatalog);

/**
 * POST /social/orders/parse
 * Parse a raw social platform order webhook into a canonical SocialOrder.
 * Body: { platform: string, rawOrder: object }
 * Response: 200 SocialOrder
 */
router.post('/orders/parse', controller.parseOrder);

/**
 * GET /social/syncs?platform=INSTAGRAM&limit=20
 * List recent catalog sync records for a given platform.
 * Response: 200 SyncRecord[]
 */
router.get('/syncs', controller.getRecentSyncs);

/**
 * GET /social/platforms
 * List all supported social commerce platforms.
 * Response: 200 { platforms: string[] }
 */
router.get('/platforms', controller.getSupportedPlatforms);

/**
 * GET /social/field-mappings/:platform
 * Return the field-name mapping for the given platform.
 * Response: 200 { platform: string, fieldMapping: object }
 */
router.get('/field-mappings/:platform', controller.getFieldMappings);

module.exports = router;
