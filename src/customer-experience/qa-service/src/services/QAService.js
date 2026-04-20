'use strict';

const QuestionRepository = require('../repositories/QuestionRepository');

class QAService {
  constructor(questionRepository) {
    this.questionRepository = questionRepository || new QuestionRepository();
  }

  async askQuestion({ productId, customerId, question, tags }) {
    if (!productId || !customerId || !question) {
      const err = new Error('productId, customerId, and question are required');
      err.status = 400;
      err.code = 'VALIDATION_ERROR';
      throw err;
    }

    if (question.length < 10) {
      const err = new Error('question must be at least 10 characters long');
      err.status = 400;
      err.code = 'VALIDATION_ERROR';
      throw err;
    }

    return this.questionRepository.create({
      productId,
      customerId,
      question,
      tags: tags || [],
      status: 'open',
    });
  }

  async getQuestion(id) {
    const question = await this.questionRepository.findById(id);
    if (!question) {
      const err = new Error(`Question with id ${id} not found`);
      err.status = 404;
      err.code = 'QUESTION_NOT_FOUND';
      throw err;
    }

    // Increment view count (fire-and-forget pattern; result not awaited for response)
    this.questionRepository.incrementView(id).catch(() => {});

    return question;
  }

  async listProductQuestions(productId, { status, limit, offset } = {}) {
    if (!productId) {
      const err = new Error('productId is required');
      err.status = 400;
      err.code = 'MISSING_PRODUCT_ID';
      throw err;
    }
    return this.questionRepository.findByProductId({ productId, status, limit, offset });
  }

  async listCustomerQuestions(customerId, { limit, offset } = {}) {
    if (!customerId) {
      const err = new Error('customerId is required');
      err.status = 400;
      err.code = 'MISSING_CUSTOMER_ID';
      throw err;
    }
    return this.questionRepository.findByCustomerId(customerId, { limit, offset });
  }

  async answerQuestion(questionId, { customerId, body, isStaff }) {
    if (!customerId || !body) {
      const err = new Error('customerId and body are required');
      err.status = 400;
      err.code = 'VALIDATION_ERROR';
      throw err;
    }

    const question = await this.questionRepository.findById(questionId);
    if (!question) {
      const err = new Error(`Question with id ${questionId} not found`);
      err.status = 404;
      err.code = 'QUESTION_NOT_FOUND';
      throw err;
    }

    if (question.status === 'closed') {
      const err = new Error('Cannot answer a closed question');
      err.status = 422;
      err.code = 'QUESTION_CLOSED';
      throw err;
    }

    const answer = {
      customerId,
      body,
      isStaff: isStaff || false,
      helpful: 0,
      createdAt: new Date(),
    };

    return this.questionRepository.addAnswer(questionId, answer);
  }

  async markAnswered(questionId) {
    const question = await this.questionRepository.findById(questionId);
    if (!question) {
      const err = new Error(`Question with id ${questionId} not found`);
      err.status = 404;
      err.code = 'QUESTION_NOT_FOUND';
      throw err;
    }
    return this.questionRepository.updateStatus(questionId, 'answered');
  }

  async closeQuestion(questionId) {
    const question = await this.questionRepository.findById(questionId);
    if (!question) {
      const err = new Error(`Question with id ${questionId} not found`);
      err.status = 404;
      err.code = 'QUESTION_NOT_FOUND';
      throw err;
    }
    return this.questionRepository.updateStatus(questionId, 'closed');
  }

  async markAnswerHelpful(questionId, answerId) {
    const question = await this.questionRepository.findById(questionId);
    if (!question) {
      const err = new Error(`Question with id ${questionId} not found`);
      err.status = 404;
      err.code = 'QUESTION_NOT_FOUND';
      throw err;
    }

    const answerExists = question.answers && question.answers.some(
      (a) => a._id && a._id.toString() === answerId
    );
    if (!answerExists) {
      const err = new Error(`Answer with id ${answerId} not found`);
      err.status = 404;
      err.code = 'ANSWER_NOT_FOUND';
      throw err;
    }

    return this.questionRepository.markAnswerHelpful(questionId, answerId);
  }
}

module.exports = QAService;
