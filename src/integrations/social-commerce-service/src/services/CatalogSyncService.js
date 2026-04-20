'use strict';

const { v4: uuidv4 } = require('uuid');
const { Platform, SyncStatus, createProductCatalogItem } = require('../models/social');
const SocialPlatformAdapter = require('./SocialPlatformAdapter');
const config = require('../config');

/**
 * CatalogSyncService
 *
 * Orchestrates catalog synchronisation to social platforms.
 *
 * This service is stateless with respect to persistent storage.  In-memory
 * ring buffers (per-platform arrays) hold the most recent MAX_SYNC_HISTORY
 * sync records for observability during a single service-instance lifetime.
 * A production deployment would back these structures with Redis or a database.
 *
 * The actual social platform API call is simulated; in production, replace
 * the `_simulatePlatformCall` method with a real HTTP call via the `axios`
 * client (credentials loaded from config).
 */
class CatalogSyncService {
  constructor() {
    this.adapter = new SocialPlatformAdapter();

    /**
     * Per-platform sync history: Map<Platform, SyncRecord[]>
     * Arrays are bounded to config.MAX_SYNC_HISTORY items (newest at index 0).
     */
    this._history = new Map(
      Object.values(Platform).map((p) => [p, []])
    );
  }

  // -------------------------------------------------------------------------
  // Public API
  // -------------------------------------------------------------------------

  /**
   * Synchronises a batch of products to the specified social platform.
   *
   * @param {string}   platform  - One of the Platform enum values.
   * @param {object[]} products  - Array of raw product objects (ShopOS catalog shape).
   * @returns {{ syncId: string, platform: string, itemsSynced: number, errors: string[], completedAt: string }}
   * @throws {Error} If the platform is unsupported.
   */
  syncCatalog(platform, products) {
    if (!Platform[platform]) {
      throw new Error(`Unsupported platform: ${platform}`);
    }
    if (!Array.isArray(products)) {
      throw new Error('products must be an array');
    }

    const syncId = uuidv4();
    const errors = [];
    let itemsSynced = 0;

    for (const rawProduct of products) {
      try {
        const catalogItem = createProductCatalogItem(rawProduct);
        const formatted = this.adapter.formatProduct(platform, catalogItem);

        // Simulate outbound API call — replace with real HTTP call in production
        this._simulatePlatformCall(platform, 'catalog', formatted);
        itemsSynced++;
      } catch (err) {
        errors.push(`Product ${rawProduct.productId || 'unknown'}: ${err.message}`);
      }
    }

    const record = {
      syncId,
      platform,
      status: errors.length === 0
        ? SyncStatus.SUCCESS
        : itemsSynced === 0
          ? SyncStatus.FAILED
          : SyncStatus.SUCCESS, // partial — still SUCCESS at operation level for catalog
      itemsSynced,
      errors,
      completedAt: new Date().toISOString(),
    };

    this._recordSync(platform, record);
    return record;
  }

  /**
   * Parses a raw incoming order webhook payload from a social platform
   * into a canonical SocialOrder.
   *
   * @param {string} platform  - One of the Platform enum values.
   * @param {object} rawOrder  - Raw order object from the platform webhook.
   * @returns {import('../models/social').SocialOrder}
   * @throws {Error} If the platform is unsupported or the order is malformed.
   */
  parsePlatformOrder(platform, rawOrder) {
    if (!Platform[platform]) {
      throw new Error(`Unsupported platform: ${platform}`);
    }
    return this.adapter.parseIncomingOrder(platform, rawOrder);
  }

  /**
   * Returns recent catalog sync records for the specified platform.
   *
   * @param {string} platform - One of the Platform enum values.
   * @param {number} limit    - Maximum number of records to return.
   * @returns {object[]} Array of sync records, newest first.
   * @throws {Error} If the platform is unsupported.
   */
  getRecentSyncs(platform, limit) {
    if (!Platform[platform]) {
      throw new Error(`Unsupported platform: ${platform}`);
    }
    const effectiveLimit = Math.min(Math.max(1, limit || 20), config.MAX_SYNC_HISTORY);
    const history = this._history.get(platform) || [];
    return history.slice(0, effectiveLimit);
  }

  /**
   * Returns the list of platforms supported by this service.
   *
   * @returns {string[]}
   */
  getSupportedPlatforms() {
    return this.adapter.getSupportedPlatforms();
  }

  /**
   * Returns the field mapping for the given platform.
   *
   * @param {string} platform
   * @returns {object}
   */
  getFieldMapping(platform) {
    if (!Platform[platform]) {
      throw new Error(`Unsupported platform: ${platform}`);
    }
    return this.adapter.getFieldMapping(platform);
  }

  // -------------------------------------------------------------------------
  // Private helpers
  // -------------------------------------------------------------------------

  /**
   * Records a sync operation in the per-platform ring buffer.
   * Evicts the oldest record if the buffer exceeds MAX_SYNC_HISTORY.
   *
   * @param {string} platform
   * @param {object} record
   */
  _recordSync(platform, record) {
    const history = this._history.get(platform);
    history.unshift(record); // newest first
    if (history.length > config.MAX_SYNC_HISTORY) {
      history.pop();
    }
  }

  /**
   * Simulates an outbound platform API call.
   * Replace with a real axios.post() call in production.
   *
   * @param {string} platform
   * @param {string} entityType
   * @param {object} payload
   */
  _simulatePlatformCall(platform, entityType, payload) {
    // In production:
    // const baseUrl = config[`${platform}_BASE_URL`];
    // await axios.post(`${baseUrl}/${entityType}`, payload, { headers: { Authorization: `Bearer ${apiKey}` } });
    void platform;
    void entityType;
    void payload;
  }
}

module.exports = CatalogSyncService;
