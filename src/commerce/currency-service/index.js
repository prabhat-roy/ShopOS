'use strict';

require('dotenv').config();

const cron = require('node-cron');
const config = require('./src/config');
const { createApp } = require('./src/index');
const rateStore = require('./src/rates');

const app = createApp();

// Simulated periodic rate refresh.
// In production this would call an external FX rate provider.
try {
  cron.schedule(config.rateRefreshCronExpr, () => {
    console.log('[cron] Refreshing exchange rates...');

    // Apply a tiny random jitter (±0.5%) to simulate live rates
    const currentRates = rateStore.getRates();
    for (const [currency, rate] of Object.entries(currentRates)) {
      if (currency === config.baseCurrency) continue;
      const jitter = 1 + (Math.random() - 0.5) * 0.01;
      rateStore.updateRate(currency, parseFloat((rate * jitter).toFixed(6)));
    }

    console.log('[cron] Exchange rates updated at', new Date().toISOString());
  });
} catch (err) {
  console.warn('[cron] Invalid cron expression, rate refresh disabled:', err.message);
}

app.listen(config.httpPort, () => {
  console.log(`currency-service listening on port ${config.httpPort}`);
  console.log(`Base currency: ${config.baseCurrency}`);
  console.log(`Supported currencies: ${rateStore.getSupportedCurrencies().join(', ')}`);
});
