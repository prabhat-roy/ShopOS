'use strict';

require('dotenv').config();

const config = require('./src/config');
const { app, connectDb } = require('./src/app');

async function main() {
  try {
    await connectDb();
  } catch (err) {
    console.warn('[cms-service] MongoDB unavailable — running without persistence:', err.message);
  }
  app.listen(config.port, () => {
    console.log(`[cms-service] Listening on port ${config.port}`);
  });
}

main();
