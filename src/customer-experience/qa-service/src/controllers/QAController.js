'use strict';

const QAService = require('../services/QAService');

class QAController {
  constructor(qaService) {
    this.qaService = qaService || new QAService();
  }

  async askQuestion(req, res) {
    try {
      const { productId, customerId, question, tags } = req.body;

      if (!productId || !customerId || !question) {
        return res.status(400).json({
          error: 'productId, customerId, and question are required',
          code: 'VALIDATION_ERROR',
        });
      }

      const result = await this.qaService.askQuestion({ productId, customerId, question, tags });
      return res.status(201).json(result);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.askQuestion]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async getQuestion(req, res) {
    try {
      const question = await this.qaService.getQuestion(req.params.id);
      return res.status(200).json(question);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.getQuestion]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async listQuestions(req, res) {
    try {
      const { productId, status, limit, offset } = req.query;

      if (!productId) {
        return res.status(400).json({
          error: 'productId query parameter is required',
          code: 'VALIDATION_ERROR',
        });
      }

      const questions = await this.qaService.listProductQuestions(productId, {
        status,
        limit: limit ? parseInt(limit, 10) : undefined,
        offset: offset ? parseInt(offset, 10) : undefined,
      });

      return res.status(200).json({ data: questions, total: questions.length });
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.listQuestions]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async listCustomerQuestions(req, res) {
    try {
      const { customerId } = req.params;
      const { limit, offset } = req.query;

      const questions = await this.qaService.listCustomerQuestions(customerId, {
        limit: limit ? parseInt(limit, 10) : undefined,
        offset: offset ? parseInt(offset, 10) : undefined,
      });

      return res.status(200).json({ data: questions, total: questions.length });
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.listCustomerQuestions]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async answerQuestion(req, res) {
    try {
      const { id } = req.params;
      const { customerId, body, isStaff } = req.body;

      if (!customerId || !body) {
        return res.status(400).json({
          error: 'customerId and body are required',
          code: 'VALIDATION_ERROR',
        });
      }

      const result = await this.qaService.answerQuestion(id, { customerId, body, isStaff });
      return res.status(201).json(result);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.answerQuestion]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async markAnswered(req, res) {
    try {
      await this.qaService.markAnswered(req.params.id);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.markAnswered]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async closeQuestion(req, res) {
    try {
      await this.qaService.closeQuestion(req.params.id);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.closeQuestion]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async markAnswerHelpful(req, res) {
    try {
      const { questionId, answerId } = req.params;
      await this.qaService.markAnswerHelpful(questionId, answerId);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[QAController.markAnswerHelpful]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }
}

module.exports = QAController;
