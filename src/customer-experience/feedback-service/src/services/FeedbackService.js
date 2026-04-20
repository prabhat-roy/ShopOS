'use strict';

const {
  FeedbackType,
  FeedbackStatus,
  isValidFeedbackType,
  isValidNpsScore,
  ALL_FEEDBACK_TYPES,
} = require('../models/feedback');

class FeedbackService {
  constructor(feedbackRepository) {
    this.feedbackRepository = feedbackRepository;
  }

  async submitFeedback({ customerId, type, score, title, body, contactEmail, metadata }) {
    if (!type || !isValidFeedbackType(type)) {
      throw Object.assign(
        new Error(`Invalid feedback type: ${type}. Valid types: ${ALL_FEEDBACK_TYPES.join(', ')}`),
        { statusCode: 400 }
      );
    }

    if (type === FeedbackType.NPS) {
      if (score === undefined || score === null) {
        throw Object.assign(
          new Error('score is required for NPS feedback (0-10)'),
          { statusCode: 400 }
        );
      }
      if (!isValidNpsScore(score)) {
        throw Object.assign(
          new Error(`NPS score must be an integer between 0 and 10, got: ${score}`),
          { statusCode: 400 }
        );
      }
    }

    if (!body && !title && type !== FeedbackType.NPS) {
      throw Object.assign(
        new Error('At least one of title or body is required'),
        { statusCode: 400 }
      );
    }

    return this.feedbackRepository.create({
      customerId,
      type,
      score: score !== undefined ? parseInt(score, 10) : null,
      title,
      body,
      contactEmail,
      metadata,
    });
  }

  async getFeedback(id) {
    if (!id) {
      throw Object.assign(new Error('id is required'), { statusCode: 400 });
    }
    const record = await this.feedbackRepository.findById(id);
    if (!record) {
      throw Object.assign(new Error(`Feedback ${id} not found`), { statusCode: 404 });
    }
    return record;
  }

  async listFeedback({ type, status, limit, offset } = {}) {
    const parsedLimit = Math.min(parseInt(limit, 10) || 20, 100);
    const parsedOffset = parseInt(offset, 10) || 0;

    if (type && !isValidFeedbackType(type)) {
      throw Object.assign(
        new Error(`Invalid feedback type: ${type}`),
        { statusCode: 400 }
      );
    }

    return this.feedbackRepository.list({ type, status, limit: parsedLimit, offset: parsedOffset });
  }

  async reviewFeedback(id) {
    const record = await this.getFeedback(id);
    if (record.status !== FeedbackStatus.NEW) {
      throw Object.assign(
        new Error(`Cannot review feedback in status ${record.status}. Feedback must be NEW`),
        { statusCode: 422 }
      );
    }
    return this.feedbackRepository.updateStatus(id, FeedbackStatus.REVIEWED);
  }

  async resolveFeedback(id, note) {
    const record = await this.getFeedback(id);
    const allowedStatuses = [FeedbackStatus.NEW, FeedbackStatus.REVIEWED, FeedbackStatus.IN_PROGRESS];
    if (!allowedStatuses.includes(record.status)) {
      throw Object.assign(
        new Error(`Cannot resolve feedback in status ${record.status}`),
        { statusCode: 422 }
      );
    }
    return this.feedbackRepository.updateStatus(id, FeedbackStatus.RESOLVED, note || null);
  }

  async closeFeedback(id) {
    const record = await this.getFeedback(id);
    const closableStatuses = [
      FeedbackStatus.NEW,
      FeedbackStatus.REVIEWED,
      FeedbackStatus.IN_PROGRESS,
      FeedbackStatus.RESOLVED,
    ];
    if (!closableStatuses.includes(record.status)) {
      throw Object.assign(
        new Error(`Cannot close feedback already in status ${record.status}`),
        { statusCode: 422 }
      );
    }
    return this.feedbackRepository.updateStatus(id, FeedbackStatus.CLOSED);
  }

  async getNPSScore() {
    return this.feedbackRepository.getNpsScore();
  }

  async getStats() {
    return this.feedbackRepository.getStats();
  }
}

module.exports = FeedbackService;
