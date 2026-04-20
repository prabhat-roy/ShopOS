'use strict';

const { Router } = require('express');
const ReviewController = require('../controllers/ReviewController');
const ReviewService = require('../services/ReviewService');
const ReviewRepository = require('../repositories/ReviewRepository');

function createReviewRouter(reviewService) {
  const router = Router();
  const service = reviewService || new ReviewService(new ReviewRepository());
  const controller = new ReviewController(service);

  // Bind controller methods so `this` is always the controller instance
  const submit = controller.submitReview.bind(controller);
  const getOne = controller.getReview.bind(controller);
  const listAll = controller.listReviews.bind(controller);
  const listByCustomer = controller.listCustomerReviews.bind(controller);
  const approve = controller.approveReview.bind(controller);
  const reject = controller.rejectReview.bind(controller);
  const helpful = controller.markHelpful.bind(controller);
  const ratings = controller.getProductRatings.bind(controller);

  // POST /reviews — submit a new review
  router.post('/', submit);

  // GET /reviews?productId=&status=&sortBy= — list product reviews
  // GET /reviews/:id — get single review
  // Order matters: specific routes before param routes
  router.get('/', listAll);
  router.get('/customer/:customerId', listByCustomer);
  router.get('/:id', getOne);

  // PATCH /reviews/:id/approve — moderate: approve
  router.patch('/:id/approve', approve);

  // PATCH /reviews/:id/reject — moderate: reject
  router.patch('/:id/reject', reject);

  // POST /reviews/:id/helpful — increment helpful votes
  router.post('/:id/helpful', helpful);

  return router;
}

function createProductRatingRouter(reviewService) {
  const router = Router({ mergeParams: true });
  const service = reviewService || new ReviewService(new ReviewRepository());
  const controller = new ReviewController(service);

  // GET /products/:productId/ratings
  router.get('/', controller.getProductRatings.bind(controller));

  return router;
}

module.exports = { createReviewRouter, createProductRatingRouter };
