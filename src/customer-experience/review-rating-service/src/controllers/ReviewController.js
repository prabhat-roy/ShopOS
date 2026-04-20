'use strict';

const ReviewService = require('../services/ReviewService');

class ReviewController {
  constructor(reviewService) {
    this.reviewService = reviewService || new ReviewService();
  }

  async submitReview(req, res) {
    try {
      const { productId, customerId, orderId, rating, title, body, verified, images } = req.body;

      if (!productId || !customerId || rating === undefined) {
        return res.status(400).json({
          error: 'productId, customerId, and rating are required',
          code: 'VALIDATION_ERROR',
        });
      }

      const review = await this.reviewService.submitReview({
        productId,
        customerId,
        orderId,
        rating,
        title,
        body,
        verified,
        images,
      });

      return res.status(201).json(review);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      // Mongoose duplicate key error
      if (err.code === 11000) {
        return res.status(409).json({
          error: 'A review for this product by this customer already exists',
          code: 'DUPLICATE_REVIEW',
        });
      }
      console.error('[ReviewController.submitReview]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async getReview(req, res) {
    try {
      const review = await this.reviewService.getReview(req.params.id);
      return res.status(200).json(review);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.getReview]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async listReviews(req, res) {
    try {
      const { productId, status, sortBy, limit, offset } = req.query;

      if (!productId) {
        return res.status(400).json({ error: 'productId query parameter is required', code: 'VALIDATION_ERROR' });
      }

      const reviews = await this.reviewService.listProductReviews(productId, {
        status,
        sortBy,
        limit: limit ? parseInt(limit, 10) : undefined,
        offset: offset ? parseInt(offset, 10) : undefined,
      });

      return res.status(200).json({ data: reviews, total: reviews.length });
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.listReviews]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async listCustomerReviews(req, res) {
    try {
      const { customerId } = req.params;
      const { limit, offset } = req.query;

      const reviews = await this.reviewService.listCustomerReviews(customerId, {
        limit: limit ? parseInt(limit, 10) : undefined,
        offset: offset ? parseInt(offset, 10) : undefined,
      });

      return res.status(200).json({ data: reviews, total: reviews.length });
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.listCustomerReviews]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async approveReview(req, res) {
    try {
      await this.reviewService.approveReview(req.params.id);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.approveReview]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async rejectReview(req, res) {
    try {
      await this.reviewService.rejectReview(req.params.id);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.rejectReview]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async markHelpful(req, res) {
    try {
      await this.reviewService.markHelpful(req.params.id);
      return res.status(204).send();
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.markHelpful]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }

  async getProductRatings(req, res) {
    try {
      const stats = await this.reviewService.getProductRatings(req.params.productId);
      return res.status(200).json(stats);
    } catch (err) {
      if (err.status) {
        return res.status(err.status).json({ error: err.message, code: err.code });
      }
      console.error('[ReviewController.getProductRatings]', err);
      return res.status(500).json({ error: 'Internal server error' });
    }
  }
}

module.exports = ReviewController;
