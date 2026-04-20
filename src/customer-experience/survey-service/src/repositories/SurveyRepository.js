'use strict';

const { v4: uuidv4 } = require('uuid');

class SurveyRepository {
  constructor(db) {
    this.db = db;
  }

  async createSurvey({ title, description, questions }) {
    const id = uuidv4();
    const now = new Date();
    const result = await this.db.query(
      `INSERT INTO surveys (id, title, description, questions, status, created_at, updated_at)
       VALUES ($1, $2, $3, $4, 'DRAFT', $5, $6)
       RETURNING *`,
      [id, title, description || null, JSON.stringify(questions), now, now]
    );
    return result.rows[0];
  }

  async getSurvey(id) {
    const result = await this.db.query(
      `SELECT * FROM surveys WHERE id = $1`,
      [id]
    );
    return result.rows[0] || null;
  }

  async listSurveys({ status, limit = 20, offset = 0 } = {}) {
    const params = [];
    let whereClause = '';

    if (status) {
      params.push(status);
      whereClause = `WHERE status = $${params.length}`;
    }

    params.push(limit, offset);
    const result = await this.db.query(
      `SELECT * FROM surveys
       ${whereClause}
       ORDER BY created_at DESC
       LIMIT $${params.length - 1} OFFSET $${params.length}`,
      params
    );

    const countResult = await this.db.query(
      `SELECT COUNT(*) FROM surveys ${whereClause}`,
      status ? [status] : []
    );

    return {
      surveys: result.rows,
      total: parseInt(countResult.rows[0].count, 10),
    };
  }

  async updateSurveyStatus(id, status) {
    const result = await this.db.query(
      `UPDATE surveys SET status = $2, updated_at = NOW()
       WHERE id = $1
       RETURNING *`,
      [id, status]
    );
    return result.rows[0] || null;
  }

  async deleteSurvey(id) {
    const result = await this.db.query(
      `DELETE FROM surveys WHERE id = $1 RETURNING id`,
      [id]
    );
    return result.rows[0] || null;
  }

  async saveResponse({ surveyId, customerId, answers }) {
    const id = uuidv4();
    const now = new Date();
    const result = await this.db.query(
      `INSERT INTO survey_responses (id, survey_id, customer_id, answers, created_at)
       VALUES ($1, $2, $3, $4, $5)
       RETURNING *`,
      [id, surveyId, customerId || null, JSON.stringify(answers), now]
    );
    return result.rows[0];
  }

  async getResponses(surveyId, limit = 50) {
    const result = await this.db.query(
      `SELECT * FROM survey_responses
       WHERE survey_id = $1
       ORDER BY created_at DESC
       LIMIT $2`,
      [surveyId, limit]
    );
    return result.rows;
  }

  async getStats(surveyId) {
    const survey = await this.getSurvey(surveyId);
    if (!survey) return null;

    const countResult = await this.db.query(
      `SELECT COUNT(*) FROM survey_responses WHERE survey_id = $1`,
      [surveyId]
    );
    const totalResponses = parseInt(countResult.rows[0].count, 10);

    const questions = survey.questions || [];
    const ratingQuestions = questions.filter((q) => q.type === 'RATING');
    const npsQuestions = questions.filter((q) => q.type === 'NPS');

    const stats = { totalResponses };

    if (ratingQuestions.length > 0 && totalResponses > 0) {
      const responses = await this.db.query(
        `SELECT answers FROM survey_responses WHERE survey_id = $1`,
        [surveyId]
      );

      let ratingSum = 0;
      let ratingCount = 0;
      let npsScores = [];

      for (const row of responses.rows) {
        const answers = row.answers || {};
        for (const q of ratingQuestions) {
          const val = parseFloat(answers[q.id]);
          if (!isNaN(val)) {
            ratingSum += val;
            ratingCount++;
          }
        }
        for (const q of npsQuestions) {
          const val = parseInt(answers[q.id], 10);
          if (!isNaN(val)) {
            npsScores.push(val);
          }
        }
      }

      if (ratingCount > 0) {
        stats.avgRating = parseFloat((ratingSum / ratingCount).toFixed(2));
      }

      if (npsScores.length > 0) {
        const promoters = npsScores.filter((s) => s >= 9).length;
        const detractors = npsScores.filter((s) => s <= 6).length;
        stats.npsScore = parseFloat(
          (((promoters - detractors) / npsScores.length) * 100).toFixed(2)
        );
      }
    }

    return stats;
  }
}

module.exports = SurveyRepository;
