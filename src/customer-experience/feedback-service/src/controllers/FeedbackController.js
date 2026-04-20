'use strict';

class FeedbackController {
  constructor(feedbackService) {
    this.feedbackService = feedbackService;
    this.submitFeedback = this.submitFeedback.bind(this);
    this.getFeedback = this.getFeedback.bind(this);
    this.listFeedback = this.listFeedback.bind(this);
    this.reviewFeedback = this.reviewFeedback.bind(this);
    this.resolveFeedback = this.resolveFeedback.bind(this);
    this.closeFeedback = this.closeFeedback.bind(this);
    this.getNPSScore = this.getNPSScore.bind(this);
    this.getStats = this.getStats.bind(this);
  }

  async submitFeedback(req, res) {
    const { customerId, type, score, title, body, contactEmail, metadata } = req.body;

    if (!type) {
      return res.status(400).json({ error: 'Bad Request', message: 'type is required' });
    }

    try {
      const record = await this.feedbackService.submitFeedback({
        customerId,
        type,
        score,
        title,
        body,
        contactEmail,
        metadata,
      });
      return res.status(201).json(record);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getFeedback(req, res) {
    const { id } = req.params;
    try {
      const record = await this.feedbackService.getFeedback(id);
      return res.status(200).json(record);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async listFeedback(req, res) {
    const { type, status, limit, offset } = req.query;
    try {
      const result = await this.feedbackService.listFeedback({ type, status, limit, offset });
      return res.status(200).json(result);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async reviewFeedback(req, res) {
    const { id } = req.params;
    try {
      await this.feedbackService.reviewFeedback(id);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async resolveFeedback(req, res) {
    const { id } = req.params;
    const { note } = req.body || {};
    try {
      await this.feedbackService.resolveFeedback(id, note);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async closeFeedback(req, res) {
    const { id } = req.params;
    try {
      await this.feedbackService.closeFeedback(id);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getNPSScore(req, res) {
    try {
      const nps = await this.feedbackService.getNPSScore();
      return res.status(200).json(nps);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getStats(req, res) {
    try {
      const stats = await this.feedbackService.getStats();
      return res.status(200).json(stats);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  _handleError(res, err) {
    const statusCode = err.statusCode || 500;
    const message = err.message || 'Internal server error';

    if (statusCode === 500) {
      console.error('[FeedbackController] Internal error:', err);
    }

    return res.status(statusCode).json({
      error: statusCode === 400 ? 'Bad Request'
        : statusCode === 404 ? 'Not Found'
        : statusCode === 422 ? 'Unprocessable Entity'
        : 'Internal Server Error',
      message,
    });
  }
}

module.exports = FeedbackController;
