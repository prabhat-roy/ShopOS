'use strict';

const { Pool } = require('pg');
const config = require('./config');

let pool;

function getPool() {
  if (!pool) {
    pool = new Pool({
      connectionString: config.DATABASE_URL,
      max: config.DB_POOL_MAX,
      min: config.DB_POOL_MIN,
      idleTimeoutMillis: config.DB_IDLE_TIMEOUT_MS,
      connectionTimeoutMillis: config.DB_CONNECTION_TIMEOUT_MS,
    });

    pool.on('error', (err) => {
      console.error('[db] Unexpected error on idle client', err);
    });

    pool.on('connect', () => {
      if (config.NODE_ENV !== 'test') {
        console.log('[db] New client connected to PostgreSQL');
      }
    });
  }
  return pool;
}

async function query(text, params) {
  const start = Date.now();
  const client = getPool();
  try {
    const result = await client.query(text, params);
    const duration = Date.now() - start;
    if (config.NODE_ENV === 'development') {
      console.log('[db] executed query', { text, duration, rows: result.rowCount });
    }
    return result;
  } catch (err) {
    console.error('[db] query error', { text, error: err.message });
    throw err;
  }
}

async function getClient() {
  return getPool().connect();
}

async function closePool() {
  if (pool) {
    await pool.end();
    pool = null;
  }
}

module.exports = { query, getClient, closePool, getPool };
