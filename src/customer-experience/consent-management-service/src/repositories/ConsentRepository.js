'use strict';

const { v4: uuidv4 } = require('uuid');

class ConsentRepository {
  constructor(db) {
    this.db = db;
  }

  async upsertConsent(customerId, type, granted, source, ipAddress) {
    const now = new Date();
    const result = await this.db.query(
      `INSERT INTO consents (customer_id, type, granted, source, ip_address, created_at, updated_at)
       VALUES ($1, $2, $3, $4, $5, $6, $7)
       ON CONFLICT (customer_id, type)
       DO UPDATE SET
         granted = EXCLUDED.granted,
         source = EXCLUDED.source,
         ip_address = EXCLUDED.ip_address,
         updated_at = EXCLUDED.updated_at
       RETURNING *`,
      [customerId, type, granted, source, ipAddress, now, now]
    );

    const action = granted ? 'grant' : 'revoke';
    await this.db.query(
      `INSERT INTO consent_history (id, customer_id, type, action, source, ip_address, created_at)
       VALUES ($1, $2, $3, $4, $5, $6, $7)`,
      [uuidv4(), customerId, type, action, source, ipAddress, now]
    );

    return result.rows[0];
  }

  async getConsent(customerId, type) {
    const result = await this.db.query(
      `SELECT * FROM consents WHERE customer_id = $1 AND type = $2`,
      [customerId, type]
    );
    return result.rows[0] || null;
  }

  async getAllConsents(customerId) {
    const result = await this.db.query(
      `SELECT * FROM consents WHERE customer_id = $1 ORDER BY type ASC`,
      [customerId]
    );
    return result.rows;
  }

  async revokeAll(customerId, source, ipAddress) {
    const now = new Date();
    const result = await this.db.query(
      `UPDATE consents
       SET granted = false, source = $2, ip_address = $3, updated_at = $4
       WHERE customer_id = $1 AND granted = true
       RETURNING *`,
      [customerId, source, ipAddress, now]
    );

    if (result.rows.length > 0) {
      // Each row uses 7 params: id, customer_id, type, action, source, ip_address, created_at
      const PARAMS_PER_ROW = 7;
      const historyValues = result.rows
        .map((_, idx) => {
          const base = idx * PARAMS_PER_ROW;
          return `($${base + 1}, $${base + 2}, $${base + 3}, $${base + 4}, $${base + 5}, $${base + 6}, $${base + 7})`;
        })
        .join(', ');

      const historyParams = result.rows.flatMap((row) => [
        uuidv4(),
        customerId,
        row.type,
        'revoke',
        source,
        ipAddress,
        now,
      ]);

      await this.db.query(
        `INSERT INTO consent_history (id, customer_id, type, action, source, ip_address, created_at)
         VALUES ${historyValues}`,
        historyParams
      );
    }

    return result.rows;
  }

  async getHistory(customerId, type, limit = 50) {
    const params = [customerId];
    let whereClause = 'WHERE customer_id = $1';

    if (type) {
      params.push(type);
      whereClause += ` AND type = $${params.length}`;
    }

    params.push(limit);
    const result = await this.db.query(
      `SELECT * FROM consent_history
       ${whereClause}
       ORDER BY created_at DESC
       LIMIT $${params.length}`,
      params
    );
    return result.rows;
  }

  async hasConsent(customerId, type) {
    const row = await this.getConsent(customerId, type);
    return row ? row.granted === true : false;
  }
}

module.exports = ConsentRepository;
