'use strict';

const ReviewRepository = require('../repositories/ReviewRepository');

class ReviewService {
  constructor(reviewRepository) {
    this.reviewRepository = reviewRepository || new ReviewRepository();
  }

  async submitReview({ productId, customerId, orderId, rating, title, body, verified, images }) {
    const existing = await this.reviewRepository.findByProductIdAndCustomerId(productId, customerId);
    if (existing) {
      const err = new Error('A review for this product by this customer already exists');
      err.status = 409;
      err.code = 'DUPLICATE_REVIEW';
      throw err;
    }

    const review = await this.reviewRepository.create({
      productId,
      customerId,
      orderId: orderId || null,
      rating,
      title: title || '',
      body: body || '',
      verified: verified || false,
      images: images || [],
      status: 'pending',
    });

    return review;
  }

  async getReview(id) {
    const review = await this.reviewRepository.findById(id);
    if (!review) {
      const err = new Error(`Review with id ${id} not found`);
      err.status = 404;
      err.code = 'REVIEW_NOT_FOUND';
      throw err;
    }
    return review;
  }

  async listProductReviews(productId, { status, limit, offset, sortBy } = {}) {
    if (!productId) {
      const err = new Error('productId is required');
      err.status = 400;
      err.code = 'MISSING_PRODUCT_ID';
      throw err;
    }
    return this.reviewRepository.findByProductId({ productId, status, limit, offset, sortBy });
  }

  async listCustomerReviews(customerId, { limit, offset } = {}) {
    if (!customerId) {
      const err = new Error('customerId is required');
      err.status = 400;
      err.code = 'MISSING_CUSTOMER_ID';
      throw err;
    }
    return this.reviewRepository.findByCustomerId(customerId, { limit, offset });
  }

  async approveReview(id) {
    const review = await this.reviewRepository.findById(id);
    if (!review) {
      const err = new Error(`Review with id ${id} not found`);
      err.status = 404;
      err.code = 'REVIEW_NOT_FOUND';
      throw err;
    }
    return this.reviewRepository.updateStatus(id, 'approved');
  }

  async rejectReview(id) {
    const review = await this.reviewRepository.findById(id);
    if (!review) {
      const err = new Error(`Review with id ${id} not found`);
      err.status = 404;
      err.code = 'REVIEW_NOT_FOUND';
      throw err;
    }
    return this.reviewRepository.updateStatus(id, 'rejected');
  }

  async markHelpful(id) {
    const review = await this.reviewRepository.findById(id);
    if (!review) {
      const err = new Error(`Review with id ${id} not found`);
      err.status = 404;
      err.code = 'REVIEW_NOT_FOUND';
      throw err;
    }
    return this.reviewRepository.incrementHelpful(id);
  }

  async getProductRatings(productId) {
    if (!productId) {
      const err = new Error('productId is required');
      err.status = 400;
      err.code = 'MISSING_PRODUCT_ID';
      throw err;
    }
    return this.reviewRepository.getProductStats(productId);
  }
}

module.exports = ReviewService;
