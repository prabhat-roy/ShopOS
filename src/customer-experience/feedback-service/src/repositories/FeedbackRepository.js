'use strict';

const { v4: uuidv4 } = require('uuid');

class FeedbackRepository {
  constructor(db) {
    this.db = db;
  }

  async create({ customerId, type, score, title, body, contactEmail, metadata }) {
    const id = uuidv4();
    const now = new Date();
    const result = await this.db.query(
      `INSERT INTO feedback
         (id, customer_id, type, status, score, title, body, contact_email, metadata, created_at, updated_at)
       VALUES ($1, $2, $3, 'NEW', $4, $5, $6, $7, $8, $9, $10)
       RETURNING *`,
      [
        id,
        customerId || null,
        type,
        score !== undefined && score !== null ? score : null,
        title || null,
        body || null,
        contactEmail || null,
        metadata ? JSON.stringify(metadata) : null,
        now,
        now,
      ]
    );
    return result.rows[0];
  }

  async findById(id) {
    const result = await this.db.query(
      `SELECT * FROM feedback WHERE id = $1`,
      [id]
    );
    return result.rows[0] || null;
  }

  async list({ type, status, limit = 20, offset = 0 } = {}) {
    const conditions = [];
    const params = [];

    if (type) {
      params.push(type);
      conditions.push(`type = $${params.length}`);
    }
    if (status) {
      params.push(status);
      conditions.push(`status = $${params.length}`);
    }

    const whereClause = conditions.length > 0 ? `WHERE ${conditions.join(' AND ')}` : '';

    params.push(limit, offset);
    const result = await this.db.query(
      `SELECT * FROM feedback
       ${whereClause}
       ORDER BY created_at DESC
       LIMIT $${params.length - 1} OFFSET $${params.length}`,
      params
    );

    const countParams = conditions.length > 0 ? params.slice(0, conditions.length) : [];
    const countResult = await this.db.query(
      `SELECT COUNT(*) FROM feedback ${whereClause}`,
      countParams
    );

    return {
      feedback: result.rows,
      total: parseInt(countResult.rows[0].count, 10),
    };
  }

  async updateStatus(id, status, note = null) {
    const result = await this.db.query(
      `UPDATE feedback
       SET status = $2, note = COALESCE($3, note), updated_at = NOW()
       WHERE id = $1
       RETURNING *`,
      [id, status, note]
    );
    return result.rows[0] || null;
  }

  async getStats() {
    const typeCountResult = await this.db.query(
      `SELECT type, COUNT(*) as count FROM feedback GROUP BY type ORDER BY type`
    );

    const statusCountResult = await this.db.query(
      `SELECT status, COUNT(*) as count FROM feedback GROUP BY status ORDER BY status`
    );

    const npsResult = await this.db.query(
      `SELECT score FROM feedback WHERE type = 'NPS' AND score IS NOT NULL`
    );

    const byType = {};
    for (const row of typeCountResult.rows) {
      byType[row.type] = parseInt(row.count, 10);
    }

    const byStatus = {};
    for (const row of statusCountResult.rows) {
      byStatus[row.status] = parseInt(row.count, 10);
    }

    const total = Object.values(byType).reduce((sum, c) => sum + c, 0);

    const stats = { total, byType, byStatus };

    if (npsResult.rows.length > 0) {
      const scores = npsResult.rows.map((r) => parseInt(r.score, 10));
      const promoters = scores.filter((s) => s >= 9).length;
      const detractors = scores.filter((s) => s <= 6).length;
      const npsScore = parseFloat(
        (((promoters - detractors) / scores.length) * 100).toFixed(2)
      );
      stats.nps = {
        totalResponses: scores.length,
        promoters,
        passives: scores.filter((s) => s >= 7 && s <= 8).length,
        detractors,
        score: npsScore,
      };
    }

    return stats;
  }

  async getNpsScore() {
    const result = await this.db.query(
      `SELECT score FROM feedback WHERE type = 'NPS' AND score IS NOT NULL`
    );

    if (result.rows.length === 0) {
      return { score: null, totalResponses: 0, promoters: 0, passives: 0, detractors: 0 };
    }

    const scores = result.rows.map((r) => parseInt(r.score, 10));
    const promoters = scores.filter((s) => s >= 9).length;
    const passives = scores.filter((s) => s >= 7 && s <= 8).length;
    const detractors = scores.filter((s) => s <= 6).length;
    const npsScore = parseFloat(
      (((promoters - detractors) / scores.length) * 100).toFixed(2)
    );

    return {
      score: npsScore,
      totalResponses: scores.length,
      promoters,
      passives,
      detractors,
    };
  }
}

module.exports = FeedbackRepository;
