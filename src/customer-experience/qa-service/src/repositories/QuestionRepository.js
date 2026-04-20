'use strict';

const Question = require('../models/Question');

class QuestionRepository {
  async create(data) {
    const question = new Question(data);
    return question.save();
  }

  async findById(id) {
    return Question.findById(id).lean();
  }

  async findByProductId({ productId, status, limit = 20, offset = 0 }) {
    const query = { productId };
    if (status) {
      query.status = status;
    }
    return Question.find(query)
      .sort({ createdAt: -1 })
      .skip(offset)
      .limit(limit)
      .lean();
  }

  async findByCustomerId(customerId, { limit = 20, offset = 0 } = {}) {
    return Question.find({ customerId })
      .sort({ createdAt: -1 })
      .skip(offset)
      .limit(limit)
      .lean();
  }

  async addAnswer(questionId, answer) {
    return Question.findByIdAndUpdate(
      questionId,
      { $push: { answers: answer } },
      { new: true }
    ).lean();
  }

  async incrementView(questionId) {
    return Question.findByIdAndUpdate(
      questionId,
      { $inc: { viewCount: 1 } },
      { new: true }
    ).lean();
  }

  async updateStatus(questionId, status) {
    return Question.findByIdAndUpdate(
      questionId,
      { status },
      { new: true }
    ).lean();
  }

  async markAnswerHelpful(questionId, answerId) {
    return Question.findOneAndUpdate(
      { _id: questionId, 'answers._id': answerId },
      { $inc: { 'answers.$.helpful': 1 } },
      { new: true }
    ).lean();
  }
}

module.exports = QuestionRepository;
