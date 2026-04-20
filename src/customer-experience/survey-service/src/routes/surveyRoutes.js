'use strict';

const { Router } = require('express');

function createSurveyRouter(surveyController) {
  const router = Router();

  // Create a new survey (DRAFT status)
  // POST /surveys
  router.post('/', surveyController.createSurvey);

  // List all surveys (filterable by ?status=ACTIVE)
  // GET /surveys
  router.get('/', surveyController.listSurveys);

  // Get a survey by ID
  // GET /surveys/:id
  router.get('/:id', surveyController.getSurvey);

  // Activate a survey (DRAFT → ACTIVE)
  // PATCH /surveys/:id/activate
  router.patch('/:id/activate', surveyController.activateSurvey);

  // Close a survey (ACTIVE → CLOSED)
  // PATCH /surveys/:id/close
  router.patch('/:id/close', surveyController.closeSurvey);

  // Delete a survey (only DRAFT)
  // DELETE /surveys/:id
  router.delete('/:id', surveyController.deleteSurvey);

  // Submit a response to a survey
  // POST /surveys/:id/responses
  router.post('/:id/responses', surveyController.submitResponse);

  // Get responses for a survey
  // GET /surveys/:id/responses
  router.get('/:id/responses', surveyController.getResponses);

  // Get aggregate stats for a survey
  // GET /surveys/:id/stats
  router.get('/:id/stats', surveyController.getSurveyStats);

  return router;
}

module.exports = createSurveyRouter;
