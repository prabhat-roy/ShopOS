'use strict';

const { Router } = require('express');

function createFeedbackRouter(feedbackController) {
  const router = Router();

  // Submit new feedback
  // POST /feedback
  router.post('/', feedbackController.submitFeedback);

  // Get NPS statistics (must come before /:id to avoid route shadowing)
  // GET /feedback/stats/nps
  router.get('/stats/nps', feedbackController.getNPSScore);

  // Get general statistics
  // GET /feedback/stats
  router.get('/stats', feedbackController.getStats);

  // Get all feedback (with optional ?type=NPS&status=NEW filters)
  // GET /feedback
  router.get('/', feedbackController.listFeedback);

  // Get a single feedback item
  // GET /feedback/:id
  router.get('/:id', feedbackController.getFeedback);

  // Mark as reviewed (NEW → REVIEWED)
  // PATCH /feedback/:id/review
  router.patch('/:id/review', feedbackController.reviewFeedback);

  // Mark as resolved
  // PATCH /feedback/:id/resolve  body: { note? }
  router.patch('/:id/resolve', feedbackController.resolveFeedback);

  // Mark as closed
  // PATCH /feedback/:id/close
  router.patch('/:id/close', feedbackController.closeFeedback);

  return router;
}

module.exports = createFeedbackRouter;
