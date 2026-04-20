'use strict';

const request = require('supertest');
const createApp = require('../src/app');
const SurveyService = require('../src/services/SurveyService');
const { SurveyStatus, QuestionType } = require('../src/models/survey');

jest.mock('../src/services/SurveyService');

const SAMPLE_QUESTIONS = [
  { id: 'q1', type: QuestionType.RATING, text: 'How satisfied are you?' },
  { id: 'q2', type: QuestionType.TEXT, text: 'Any comments?' },
];

const SAMPLE_SURVEY = {
  id: 'survey-001',
  title: 'Customer Satisfaction',
  description: 'Q4 2026 satisfaction survey',
  questions: SAMPLE_QUESTIONS,
  status: SurveyStatus.DRAFT,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('Survey Service', () => {
  let app;
  let mockSurveyService;

  beforeEach(() => {
    jest.clearAllMocks();
    mockSurveyService = new SurveyService();

    const SurveyController = require('../src/controllers/SurveyController');
    const controller = new SurveyController(mockSurveyService);

    app = createApp({ surveyController: controller });
  });

  // Test 1: Health check
  test('GET /healthz returns status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.service).toBe('survey-service');
  });

  // Test 2: Create survey
  test('POST /surveys creates a new survey and returns 201', async () => {
    mockSurveyService.createSurvey = jest.fn().mockResolvedValue(SAMPLE_SURVEY);

    const res = await request(app)
      .post('/surveys')
      .send({
        title: 'Customer Satisfaction',
        description: 'Q4 2026 satisfaction survey',
        questions: SAMPLE_QUESTIONS,
      });

    expect(res.status).toBe(201);
    expect(res.body.id).toBe('survey-001');
    expect(res.body.status).toBe(SurveyStatus.DRAFT);
    expect(mockSurveyService.createSurvey).toHaveBeenCalledTimes(1);
  });

  // Test 3: Create survey - missing title
  test('POST /surveys returns 400 when title is missing', async () => {
    const res = await request(app)
      .post('/surveys')
      .send({ questions: SAMPLE_QUESTIONS });

    expect(res.status).toBe(400);
    expect(res.body.error).toBe('Bad Request');
  });

  // Test 4: Get survey by ID
  test('GET /surveys/:id returns the survey', async () => {
    mockSurveyService.getSurvey = jest.fn().mockResolvedValue(SAMPLE_SURVEY);

    const res = await request(app).get('/surveys/survey-001');

    expect(res.status).toBe(200);
    expect(res.body.id).toBe('survey-001');
    expect(mockSurveyService.getSurvey).toHaveBeenCalledWith('survey-001');
  });

  // Test 5: Get survey - not found
  test('GET /surveys/:id returns 404 when survey not found', async () => {
    mockSurveyService.getSurvey = jest.fn().mockRejectedValue(
      Object.assign(new Error('Survey not-exist not found'), { statusCode: 404 })
    );

    const res = await request(app).get('/surveys/not-exist');
    expect(res.status).toBe(404);
    expect(res.body.error).toBe('Not Found');
  });

  // Test 6: List surveys (active only)
  test('GET /surveys?status=ACTIVE returns active surveys', async () => {
    const activeSurvey = { ...SAMPLE_SURVEY, status: SurveyStatus.ACTIVE };
    mockSurveyService.listSurveys = jest.fn().mockResolvedValue({
      surveys: [activeSurvey],
      total: 1,
    });

    const res = await request(app).get('/surveys?status=ACTIVE');

    expect(res.status).toBe(200);
    expect(res.body.surveys).toHaveLength(1);
    expect(res.body.surveys[0].status).toBe(SurveyStatus.ACTIVE);
    expect(mockSurveyService.listSurveys).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'ACTIVE' })
    );
  });

  // Test 7: Activate a survey
  test('PATCH /surveys/:id/activate activates a DRAFT survey', async () => {
    mockSurveyService.activateSurvey = jest.fn().mockResolvedValue({
      ...SAMPLE_SURVEY,
      status: SurveyStatus.ACTIVE,
    });

    const res = await request(app).patch('/surveys/survey-001/activate');
    expect(res.status).toBe(204);
    expect(mockSurveyService.activateSurvey).toHaveBeenCalledWith('survey-001');
  });

  // Test 8: Cannot activate non-DRAFT survey
  test('PATCH /surveys/:id/activate returns 422 for non-DRAFT survey', async () => {
    mockSurveyService.activateSurvey = jest.fn().mockRejectedValue(
      Object.assign(
        new Error('Cannot activate survey in status CLOSED'),
        { statusCode: 422 }
      )
    );

    const res = await request(app).patch('/surveys/survey-001/activate');
    expect(res.status).toBe(422);
  });

  // Test 9: Close an active survey
  test('PATCH /surveys/:id/close closes an ACTIVE survey', async () => {
    mockSurveyService.closeSurvey = jest.fn().mockResolvedValue({
      ...SAMPLE_SURVEY,
      status: SurveyStatus.CLOSED,
    });

    const res = await request(app).patch('/surveys/survey-001/close');
    expect(res.status).toBe(204);
    expect(mockSurveyService.closeSurvey).toHaveBeenCalledWith('survey-001');
  });

  // Test 10: Delete a DRAFT survey
  test('DELETE /surveys/:id deletes a DRAFT survey', async () => {
    mockSurveyService.deleteSurvey = jest.fn().mockResolvedValue({ id: 'survey-001' });

    const res = await request(app).delete('/surveys/survey-001');
    expect(res.status).toBe(204);
    expect(mockSurveyService.deleteSurvey).toHaveBeenCalledWith('survey-001');
  });

  // Test 11: Cannot delete non-DRAFT survey
  test('DELETE /surveys/:id returns 422 for non-DRAFT survey', async () => {
    mockSurveyService.deleteSurvey = jest.fn().mockRejectedValue(
      Object.assign(
        new Error('Cannot delete survey in status ACTIVE'),
        { statusCode: 422 }
      )
    );

    const res = await request(app).delete('/surveys/survey-001');
    expect(res.status).toBe(422);
  });

  // Test 12: Submit a response
  test('POST /surveys/:id/responses submits a response and returns 201', async () => {
    const mockResponse = {
      id: 'resp-001',
      survey_id: 'survey-001',
      customer_id: 'cust-001',
      answers: { q1: 4, q2: 'Great service!' },
      created_at: new Date().toISOString(),
    };
    mockSurveyService.submitResponse = jest.fn().mockResolvedValue(mockResponse);

    const res = await request(app)
      .post('/surveys/survey-001/responses')
      .send({
        customerId: 'cust-001',
        answers: { q1: 4, q2: 'Great service!' },
      });

    expect(res.status).toBe(201);
    expect(res.body.id).toBe('resp-001');
    expect(mockSurveyService.submitResponse).toHaveBeenCalledWith(
      'survey-001',
      'cust-001',
      { q1: 4, q2: 'Great service!' }
    );
  });

  // Test 13: Get survey responses
  test('GET /surveys/:id/responses returns response list', async () => {
    mockSurveyService.getSurvey = jest.fn().mockResolvedValue(SAMPLE_SURVEY);
    mockSurveyService.getResponses = jest.fn().mockResolvedValue([
      { id: 'r1', survey_id: 'survey-001', answers: { q1: 5 } },
      { id: 'r2', survey_id: 'survey-001', answers: { q1: 3 } },
    ]);

    const res = await request(app).get('/surveys/survey-001/responses');

    expect(res.status).toBe(200);
    expect(res.body.responses).toHaveLength(2);
    expect(res.body.surveyId).toBe('survey-001');
  });

  // Test 14: Get survey stats
  test('GET /surveys/:id/stats returns aggregate statistics', async () => {
    const stats = {
      totalResponses: 100,
      avgRating: 4.2,
      npsScore: 45.0,
    };
    mockSurveyService.getSurveyStats = jest.fn().mockResolvedValue(stats);

    const res = await request(app).get('/surveys/survey-001/stats');

    expect(res.status).toBe(200);
    expect(res.body.totalResponses).toBe(100);
    expect(res.body.avgRating).toBe(4.2);
    expect(res.body.npsScore).toBe(45.0);
    expect(mockSurveyService.getSurveyStats).toHaveBeenCalledWith('survey-001');
  });
});
