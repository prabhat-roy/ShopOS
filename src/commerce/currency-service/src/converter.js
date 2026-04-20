'use strict';

/**
 * Convert `amount` from one currency to another using the provided RateStore.
 *
 * @param {number} amount
 * @param {string} from
 * @param {string} to
 * @param {import('./rates').RateStore} rateStore
 * @returns {{ amount: number, from: string, to: string, converted: number, rate: number, timestamp: string }}
 */
function convert(amount, from, to, rateStore) {
  if (typeof amount !== 'number' || isNaN(amount)) {
    throw new Error('amount must be a valid number');
  }
  if (amount < 0) {
    throw new Error('amount must be non-negative');
  }

  const fromUpper = from.toUpperCase();
  const toUpper = to.toUpperCase();
  const rate = rateStore.getRate(fromUpper, toUpper);
  const converted = parseFloat((amount * rate).toFixed(6));

  return {
    amount,
    from: fromUpper,
    to: toUpper,
    converted,
    rate,
    timestamp: new Date().toISOString(),
  };
}

/**
 * Convert `amount` from one currency to many target currencies.
 *
 * @param {number} amount
 * @param {string} from
 * @param {string[]} targets
 * @param {import('./rates').RateStore} rateStore
 * @returns {Array<ReturnType<convert>>}
 */
function convertMany(amount, from, targets, rateStore) {
  if (!Array.isArray(targets) || targets.length === 0) {
    throw new Error('targets must be a non-empty array of currency codes');
  }

  return targets.map((to) => convert(amount, from, to, rateStore));
}

module.exports = { convert, convertMany };
