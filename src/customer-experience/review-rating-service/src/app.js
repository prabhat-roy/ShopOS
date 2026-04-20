'use strict';

const express = require('express');
const { createReviewRouter, createProductRatingRouter } = require('./routes/reviewRoutes');

function createApp(reviewService) {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: false }));

  // Health check
  app.get('/healthz', (req, res) => {
    res.status(200).json({ status: 'ok', service: 'review-rating-service' });
  });

  // API routes
  app.use('/reviews', createReviewRouter(reviewService));
  app.use('/products/:productId/ratings', createProductRatingRouter(reviewService));

  // 404 handler
  app.use((req, res) => {
    res.status(404).json({ error: 'Route not found' });
  });

  // Global error handler
  app.use((err, req, res, next) => { // eslint-disable-line no-unused-vars
    console.error('[review-rating-service] Unhandled error:', err);
    const status = err.status || 500;
    res.status(status).json({
      error: err.message || 'Internal server error',
      code: err.code || 'INTERNAL_ERROR',
    });
  });

  return app;
}

module.exports = { createApp };
