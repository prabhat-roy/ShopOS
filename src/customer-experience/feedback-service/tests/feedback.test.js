'use strict';

const request = require('supertest');
const createApp = require('../src/app');
const FeedbackService = require('../src/services/FeedbackService');
const { FeedbackType, FeedbackStatus } = require('../src/models/feedback');

jest.mock('../src/services/FeedbackService');

const SAMPLE_FEEDBACK = {
  id: 'fb-001',
  customer_id: 'cust-001',
  type: FeedbackType.NPS,
  status: FeedbackStatus.NEW,
  score: 9,
  title: null,
  body: null,
  contact_email: null,
  metadata: null,
  note: null,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('Feedback Service', () => {
  let app;
  let mockFeedbackService;

  beforeEach(() => {
    jest.clearAllMocks();
    mockFeedbackService = new FeedbackService();

    const FeedbackController = require('../src/controllers/FeedbackController');
    const controller = new FeedbackController(mockFeedbackService);

    app = createApp({ feedbackController: controller });
  });

  // Test 1: Health check
  test('GET /healthz returns status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.service).toBe('feedback-service');
  });

  // Test 2: Submit NPS feedback
  test('POST /feedback submits NPS feedback and returns 201', async () => {
    mockFeedbackService.submitFeedback = jest.fn().mockResolvedValue(SAMPLE_FEEDBACK);

    const res = await request(app)
      .post('/feedback')
      .send({ customerId: 'cust-001', type: FeedbackType.NPS, score: 9 });

    expect(res.status).toBe(201);
    expect(res.body.id).toBe('fb-001');
    expect(res.body.type).toBe(FeedbackType.NPS);
    expect(mockFeedbackService.submitFeedback).toHaveBeenCalledTimes(1);
  });

  // Test 3: Submit feedback - missing type
  test('POST /feedback returns 400 when type is missing', async () => {
    const res = await request(app)
      .post('/feedback')
      .send({ customerId: 'cust-001', title: 'Feedback without type' });

    expect(res.status).toBe(400);
    expect(res.body.error).toBe('Bad Request');
  });

  // Test 4: Submit feature request
  test('POST /feedback submits FEATURE_REQUEST feedback', async () => {
    const featureRecord = {
      ...SAMPLE_FEEDBACK,
      id: 'fb-002',
      type: FeedbackType.FEATURE_REQUEST,
      score: null,
      title: 'Add dark mode',
      body: 'Please add dark mode support',
    };
    mockFeedbackService.submitFeedback = jest.fn().mockResolvedValue(featureRecord);

    const res = await request(app)
      .post('/feedback')
      .send({ type: FeedbackType.FEATURE_REQUEST, title: 'Add dark mode', body: 'Please add dark mode support' });

    expect(res.status).toBe(201);
    expect(res.body.type).toBe(FeedbackType.FEATURE_REQUEST);
    expect(res.body.title).toBe('Add dark mode');
  });

  // Test 5: Get feedback by ID
  test('GET /feedback/:id returns the feedback item', async () => {
    mockFeedbackService.getFeedback = jest.fn().mockResolvedValue(SAMPLE_FEEDBACK);

    const res = await request(app).get('/feedback/fb-001');

    expect(res.status).toBe(200);
    expect(res.body.id).toBe('fb-001');
    expect(mockFeedbackService.getFeedback).toHaveBeenCalledWith('fb-001');
  });

  // Test 6: Get feedback - not found
  test('GET /feedback/:id returns 404 when not found', async () => {
    mockFeedbackService.getFeedback = jest.fn().mockRejectedValue(
      Object.assign(new Error('Feedback fb-999 not found'), { statusCode: 404 })
    );

    const res = await request(app).get('/feedback/fb-999');
    expect(res.status).toBe(404);
    expect(res.body.error).toBe('Not Found');
  });

  // Test 7: List feedback with filters
  test('GET /feedback lists feedback with type and status filters', async () => {
    const feedbackList = {
      feedback: [SAMPLE_FEEDBACK],
      total: 1,
    };
    mockFeedbackService.listFeedback = jest.fn().mockResolvedValue(feedbackList);

    const res = await request(app).get('/feedback?type=NPS&status=NEW');

    expect(res.status).toBe(200);
    expect(res.body.feedback).toHaveLength(1);
    expect(res.body.total).toBe(1);
    expect(mockFeedbackService.listFeedback).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'NPS', status: 'NEW' })
    );
  });

  // Test 8: Review feedback
  test('PATCH /feedback/:id/review marks feedback as reviewed', async () => {
    mockFeedbackService.reviewFeedback = jest.fn().mockResolvedValue({
      ...SAMPLE_FEEDBACK,
      status: FeedbackStatus.REVIEWED,
    });

    const res = await request(app).patch('/feedback/fb-001/review');
    expect(res.status).toBe(204);
    expect(mockFeedbackService.reviewFeedback).toHaveBeenCalledWith('fb-001');
  });

  // Test 9: Resolve feedback with note
  test('PATCH /feedback/:id/resolve resolves feedback with note', async () => {
    mockFeedbackService.resolveFeedback = jest.fn().mockResolvedValue({
      ...SAMPLE_FEEDBACK,
      status: FeedbackStatus.RESOLVED,
      note: 'Issue fixed in v2.1',
    });

    const res = await request(app)
      .patch('/feedback/fb-001/resolve')
      .send({ note: 'Issue fixed in v2.1' });

    expect(res.status).toBe(204);
    expect(mockFeedbackService.resolveFeedback).toHaveBeenCalledWith('fb-001', 'Issue fixed in v2.1');
  });

  // Test 10: Close feedback
  test('PATCH /feedback/:id/close closes the feedback item', async () => {
    mockFeedbackService.closeFeedback = jest.fn().mockResolvedValue({
      ...SAMPLE_FEEDBACK,
      status: FeedbackStatus.CLOSED,
    });

    const res = await request(app).patch('/feedback/fb-001/close');
    expect(res.status).toBe(204);
    expect(mockFeedbackService.closeFeedback).toHaveBeenCalledWith('fb-001');
  });

  // Test 11: Get NPS score
  test('GET /feedback/stats/nps returns calculated NPS score', async () => {
    const npsData = {
      score: 42.5,
      totalResponses: 80,
      promoters: 50,
      passives: 14,
      detractors: 16,
    };
    mockFeedbackService.getNPSScore = jest.fn().mockResolvedValue(npsData);

    const res = await request(app).get('/feedback/stats/nps');

    expect(res.status).toBe(200);
    expect(res.body.score).toBe(42.5);
    expect(res.body.promoters).toBe(50);
    expect(res.body.detractors).toBe(16);
    expect(res.body.totalResponses).toBe(80);
    expect(mockFeedbackService.getNPSScore).toHaveBeenCalledTimes(1);
  });

  // Test 12: Get general stats
  test('GET /feedback/stats returns aggregate stats by type and status', async () => {
    const statsData = {
      total: 250,
      byType: {
        NPS: 100,
        FEATURE_REQUEST: 60,
        BUG_REPORT: 50,
        GENERAL: 30,
        COMPLAINT: 10,
      },
      byStatus: {
        NEW: 80,
        REVIEWED: 50,
        IN_PROGRESS: 40,
        RESOLVED: 60,
        CLOSED: 20,
      },
      nps: { score: 38.0, totalResponses: 100, promoters: 55, passives: 28, detractors: 17 },
    };
    mockFeedbackService.getStats = jest.fn().mockResolvedValue(statsData);

    const res = await request(app).get('/feedback/stats');

    expect(res.status).toBe(200);
    expect(res.body.total).toBe(250);
    expect(res.body.byType.NPS).toBe(100);
    expect(res.body.byStatus.NEW).toBe(80);
    expect(res.body.nps.score).toBe(38.0);
    expect(mockFeedbackService.getStats).toHaveBeenCalledTimes(1);
  });
});
