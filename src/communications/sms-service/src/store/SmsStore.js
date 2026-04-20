'use strict';

const config = require('../config');

/**
 * In-memory store for SMS message records.
 * Bounded to config.sms.maxLogSize entries — oldest entries are evicted when full.
 */
class SmsStore {
  constructor(maxSize) {
    this._maxSize = maxSize || config.sms.maxLogSize;
    // Map preserves insertion order, so the first key is always the oldest
    this._store = new Map();
  }

  /**
   * Persists an SMS record.
   * @param {string} messageId
   * @param {object} record
   */
  set(messageId, record) {
    // Evict oldest entry if we've hit capacity
    if (this._store.size >= this._maxSize && !this._store.has(messageId)) {
      const oldestKey = this._store.keys().next().value;
      this._store.delete(oldestKey);
    }

    this._store.set(messageId, record);
  }

  /**
   * Retrieves an SMS record by messageId.
   * @param {string} messageId
   * @returns {object|undefined}
   */
  get(messageId) {
    return this._store.get(messageId);
  }

  /**
   * Returns the current number of stored records.
   * @returns {number}
   */
  size() {
    return this._store.size;
  }

  /**
   * Clears all records (useful for testing).
   */
  clear() {
    this._store.clear();
  }
}

// Export a singleton for the application; tests can create their own instance
const defaultStore = new SmsStore();

module.exports = { SmsStore, defaultStore };
