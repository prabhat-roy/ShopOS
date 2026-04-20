'use strict';

const { Router } = require('express');
const QAController = require('../controllers/QAController');
const QAService = require('../services/QAService');
const QuestionRepository = require('../repositories/QuestionRepository');

function createQARouter(qaService) {
  const router = Router();
  const service = qaService || new QAService(new QuestionRepository());
  const controller = new QAController(service);

  // Bind methods
  const ask = controller.askQuestion.bind(controller);
  const getOne = controller.getQuestion.bind(controller);
  const listAll = controller.listQuestions.bind(controller);
  const listByCustomer = controller.listCustomerQuestions.bind(controller);
  const answer = controller.answerQuestion.bind(controller);
  const markAnswered = controller.markAnswered.bind(controller);
  const close = controller.closeQuestion.bind(controller);
  const answerHelpful = controller.markAnswerHelpful.bind(controller);

  // POST /questions — ask a new question
  router.post('/', ask);

  // GET /questions?productId=&status= — list questions for a product
  router.get('/', listAll);

  // GET /questions/customer/:customerId — list questions by a customer
  // Must come before /:id to avoid misrouting
  router.get('/customer/:customerId', listByCustomer);

  // GET /questions/:id — get a single question (increments view)
  router.get('/:id', getOne);

  // POST /questions/:id/answers — add an answer to a question
  router.post('/:id/answers', answer);

  // PATCH /questions/:id/answered — mark question as answered
  router.patch('/:id/answered', markAnswered);

  // PATCH /questions/:id/close — close a question
  router.patch('/:id/close', close);

  // POST /questions/:questionId/answers/:answerId/helpful — mark answer as helpful
  router.post('/:questionId/answers/:answerId/helpful', answerHelpful);

  return router;
}

module.exports = { createQARouter };
