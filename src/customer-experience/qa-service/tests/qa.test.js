'use strict';

const QAService = require('../src/services/QAService');

// Mock the repository
jest.mock('../src/repositories/QuestionRepository');
const QuestionRepository = require('../src/repositories/QuestionRepository');

const mockRepo = {
  create: jest.fn(),
  findById: jest.fn(),
  findByProductId: jest.fn(),
  findByCustomerId: jest.fn(),
  addAnswer: jest.fn(),
  incrementView: jest.fn(),
  updateStatus: jest.fn(),
  markAnswerHelpful: jest.fn(),
};

QuestionRepository.mockImplementation(() => mockRepo);

const makeAnswer = (overrides = {}) => ({
  _id: 'answer-id-1',
  customerId: 'staff-001',
  body: 'This product is compatible with all versions.',
  isStaff: true,
  helpful: 0,
  createdAt: new Date(),
  ...overrides,
});

const makeQuestion = (overrides = {}) => ({
  _id: 'question-id-1',
  productId: 'prod-123',
  customerId: 'cust-456',
  question: 'Is this product compatible with older versions?',
  status: 'open',
  answers: [],
  tags: ['compatibility'],
  viewCount: 0,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
});

describe('QAService', () => {
  let service;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new QAService(mockRepo);
  });

  // ── askQuestion ───────────────────────────────────────────────────────────

  describe('askQuestion', () => {
    it('should create and return a new question', async () => {
      const input = {
        productId: 'prod-123',
        customerId: 'cust-456',
        question: 'Is this product compatible with older versions?',
        tags: ['compatibility'],
      };
      const expected = makeQuestion();
      mockRepo.create.mockResolvedValue(expected);

      const result = await service.askQuestion(input);

      expect(mockRepo.create).toHaveBeenCalledWith({
        productId: 'prod-123',
        customerId: 'cust-456',
        question: 'Is this product compatible with older versions?',
        tags: ['compatibility'],
        status: 'open',
      });
      expect(result).toEqual(expected);
    });

    it('should throw 400 VALIDATION_ERROR when question is too short', async () => {
      await expect(
        service.askQuestion({ productId: 'p1', customerId: 'c1', question: 'Short?' })
      ).rejects.toMatchObject({ status: 400, code: 'VALIDATION_ERROR' });

      expect(mockRepo.create).not.toHaveBeenCalled();
    });

    it('should throw 400 VALIDATION_ERROR when required fields are missing', async () => {
      await expect(
        service.askQuestion({ productId: 'p1', question: 'Is this product good enough to buy?' })
      ).rejects.toMatchObject({ status: 400, code: 'VALIDATION_ERROR' });
    });
  });

  // ── getQuestion ───────────────────────────────────────────────────────────

  describe('getQuestion', () => {
    it('should return a question and trigger view increment', async () => {
      const expected = makeQuestion({ viewCount: 5 });
      mockRepo.findById.mockResolvedValue(expected);
      mockRepo.incrementView.mockResolvedValue({ ...expected, viewCount: 6 });

      const result = await service.getQuestion('question-id-1');

      expect(mockRepo.findById).toHaveBeenCalledWith('question-id-1');
      expect(result).toEqual(expected);
      // Give the fire-and-forget a tick to run
      await new Promise(setImmediate);
      expect(mockRepo.incrementView).toHaveBeenCalledWith('question-id-1');
    });

    it('should throw 404 QUESTION_NOT_FOUND when question does not exist', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.getQuestion('nonexistent-id')).rejects.toMatchObject({
        status: 404,
        code: 'QUESTION_NOT_FOUND',
      });
    });
  });

  // ── listProductQuestions ──────────────────────────────────────────────────

  describe('listProductQuestions', () => {
    it('should return questions for a product', async () => {
      const questions = [makeQuestion(), makeQuestion({ _id: 'q2', status: 'answered' })];
      mockRepo.findByProductId.mockResolvedValue(questions);

      const result = await service.listProductQuestions('prod-123', { status: 'open' });

      expect(mockRepo.findByProductId).toHaveBeenCalledWith({
        productId: 'prod-123',
        status: 'open',
        limit: undefined,
        offset: undefined,
      });
      expect(result).toHaveLength(2);
    });

    it('should apply status filter when provided', async () => {
      const answered = [makeQuestion({ status: 'answered' })];
      mockRepo.findByProductId.mockResolvedValue(answered);

      const result = await service.listProductQuestions('prod-123', { status: 'answered' });

      expect(mockRepo.findByProductId).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'answered' })
      );
      expect(result).toHaveLength(1);
    });
  });

  // ── listCustomerQuestions ─────────────────────────────────────────────────

  describe('listCustomerQuestions', () => {
    it('should return questions asked by a customer', async () => {
      const questions = [makeQuestion()];
      mockRepo.findByCustomerId.mockResolvedValue(questions);

      const result = await service.listCustomerQuestions('cust-456');

      expect(mockRepo.findByCustomerId).toHaveBeenCalledWith('cust-456', {
        limit: undefined,
        offset: undefined,
      });
      expect(result).toHaveLength(1);
    });
  });

  // ── answerQuestion ────────────────────────────────────────────────────────

  describe('answerQuestion', () => {
    it('should add an answer to an open question', async () => {
      const answerPayload = { customerId: 'staff-001', body: 'Yes, it is compatible.', isStaff: true };
      const updated = makeQuestion({ answers: [makeAnswer()] });
      mockRepo.findById.mockResolvedValue(makeQuestion());
      mockRepo.addAnswer.mockResolvedValue(updated);

      const result = await service.answerQuestion('question-id-1', answerPayload);

      expect(mockRepo.addAnswer).toHaveBeenCalledWith(
        'question-id-1',
        expect.objectContaining({ customerId: 'staff-001', isStaff: true })
      );
      expect(result.answers).toHaveLength(1);
    });

    it('should throw 422 QUESTION_CLOSED when answering a closed question', async () => {
      mockRepo.findById.mockResolvedValue(makeQuestion({ status: 'closed' }));

      await expect(
        service.answerQuestion('question-id-1', { customerId: 'c1', body: 'Some answer text here.' })
      ).rejects.toMatchObject({ status: 422, code: 'QUESTION_CLOSED' });
    });

    it('should throw 404 when question does not exist', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(
        service.answerQuestion('bad-id', { customerId: 'c1', body: 'Some answer text.' })
      ).rejects.toMatchObject({ status: 404, code: 'QUESTION_NOT_FOUND' });
    });
  });

  // ── markAnswered ──────────────────────────────────────────────────────────

  describe('markAnswered', () => {
    it('should update question status to answered', async () => {
      const updated = makeQuestion({ status: 'answered' });
      mockRepo.findById.mockResolvedValue(makeQuestion());
      mockRepo.updateStatus.mockResolvedValue(updated);

      const result = await service.markAnswered('question-id-1');

      expect(mockRepo.updateStatus).toHaveBeenCalledWith('question-id-1', 'answered');
      expect(result.status).toBe('answered');
    });

    it('should throw 404 when question does not exist', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.markAnswered('bad-id')).rejects.toMatchObject({
        status: 404,
        code: 'QUESTION_NOT_FOUND',
      });
    });
  });

  // ── closeQuestion ─────────────────────────────────────────────────────────

  describe('closeQuestion', () => {
    it('should update question status to closed', async () => {
      const updated = makeQuestion({ status: 'closed' });
      mockRepo.findById.mockResolvedValue(makeQuestion());
      mockRepo.updateStatus.mockResolvedValue(updated);

      const result = await service.closeQuestion('question-id-1');

      expect(mockRepo.updateStatus).toHaveBeenCalledWith('question-id-1', 'closed');
      expect(result.status).toBe('closed');
    });
  });

  // ── markAnswerHelpful ─────────────────────────────────────────────────────

  describe('markAnswerHelpful', () => {
    it('should increment helpful count for an answer', async () => {
      const answer = makeAnswer({ _id: { toString: () => 'answer-id-1' } });
      const question = makeQuestion({ answers: [answer] });
      const updated = makeQuestion({ answers: [{ ...answer, helpful: 1 }] });

      mockRepo.findById.mockResolvedValue(question);
      mockRepo.markAnswerHelpful.mockResolvedValue(updated);

      const result = await service.markAnswerHelpful('question-id-1', 'answer-id-1');

      expect(mockRepo.markAnswerHelpful).toHaveBeenCalledWith('question-id-1', 'answer-id-1');
      expect(result.answers[0].helpful).toBe(1);
    });

    it('should throw 404 ANSWER_NOT_FOUND when answer does not exist', async () => {
      const question = makeQuestion({ answers: [] });
      mockRepo.findById.mockResolvedValue(question);

      await expect(
        service.markAnswerHelpful('question-id-1', 'nonexistent-answer')
      ).rejects.toMatchObject({ status: 404, code: 'ANSWER_NOT_FOUND' });
    });
  });
});
