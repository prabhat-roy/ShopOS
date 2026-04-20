'use strict';

/**
 * RateStore — in-memory exchange-rate store.
 *
 * All rates are relative to USD (base).
 * getRate(from, to) performs a two-step conversion via USD when needed.
 */

const SEED_RATES = {
  USD: 1.0,
  EUR: 0.92,
  GBP: 0.79,
  JPY: 149.5,
  CAD: 1.36,
  AUD: 1.53,
  INR: 83.2,
  CNY: 7.24,
  BRL: 4.97,
  MXN: 17.2,
  SGD: 1.34,
  AED: 3.67,
  CHF: 0.89,
  KRW: 1325.0,
  SEK: 10.42,
  NOK: 10.55,
};

class RateStore {
  constructor() {
    // Deep copy so tests can mutate without polluting the module cache
    this._rates = { ...SEED_RATES };
    this._updatedAt = new Date();
  }

  /**
   * Returns all rates keyed by currency code (base USD).
   * @returns {{ [currency: string]: number }}
   */
  getRates() {
    return { ...this._rates };
  }

  /**
   * Returns the direct conversion rate from `from` → `to`.
   * Uses USD as pivot when neither currency is USD.
   *
   * @param {string} from  - Source currency code (e.g. 'EUR')
   * @param {string} to    - Target currency code (e.g. 'JPY')
   * @returns {number}     - Conversion rate
   * @throws {Error}       - If either currency is unsupported
   */
  getRate(from, to) {
    const fromUpper = from.toUpperCase();
    const toUpper = to.toUpperCase();

    if (!(fromUpper in this._rates)) {
      throw new Error(`Unsupported currency: ${from}`);
    }
    if (!(toUpper in this._rates)) {
      throw new Error(`Unsupported currency: ${to}`);
    }

    if (fromUpper === toUpper) {
      return 1;
    }

    // Convert from → USD → to
    const fromToUsd = 1 / this._rates[fromUpper];
    const usdToTarget = this._rates[toUpper];
    return fromToUsd * usdToTarget;
  }

  /**
   * Update or add a rate for a given currency (relative to USD base).
   * @param {string} currency
   * @param {number} rate
   */
  updateRate(currency, rate) {
    if (typeof rate !== 'number' || rate <= 0) {
      throw new Error('Rate must be a positive number');
    }
    this._rates[currency.toUpperCase()] = rate;
    this._updatedAt = new Date();
  }

  /**
   * Returns an array of all supported currency codes.
   * @returns {string[]}
   */
  getSupportedCurrencies() {
    return Object.keys(this._rates);
  }

  /**
   * Timestamp of the last update.
   * @returns {Date}
   */
  getUpdatedAt() {
    return this._updatedAt;
  }
}

// Export a singleton instance
module.exports = new RateStore();
module.exports.RateStore = RateStore;
