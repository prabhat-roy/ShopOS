'use strict';

const ReviewService = require('../src/services/ReviewService');

// Mock the repository
jest.mock('../src/repositories/ReviewRepository');
const ReviewRepository = require('../src/repositories/ReviewRepository');

const mockRepo = {
  create: jest.fn(),
  findById: jest.fn(),
  findByProductId: jest.fn(),
  findByCustomerId: jest.fn(),
  findByProductIdAndCustomerId: jest.fn(),
  updateStatus: jest.fn(),
  incrementHelpful: jest.fn(),
  getProductStats: jest.fn(),
};

ReviewRepository.mockImplementation(() => mockRepo);

const makeReview = (overrides = {}) => ({
  _id: 'review-id-1',
  productId: 'prod-123',
  customerId: 'cust-456',
  orderId: 'order-789',
  rating: 5,
  title: 'Excellent product',
  body: 'I really enjoyed using this product.',
  verified: true,
  status: 'pending',
  helpfulVotes: 0,
  images: [],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
});

describe('ReviewService', () => {
  let service;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new ReviewService(mockRepo);
  });

  // ── submitReview ──────────────────────────────────────────────────────────

  describe('submitReview', () => {
    it('should create and return a new review when no duplicate exists', async () => {
      const reviewData = {
        productId: 'prod-123',
        customerId: 'cust-456',
        orderId: 'order-789',
        rating: 5,
        title: 'Great',
        body: 'Loved it',
      };
      const expected = makeReview();

      mockRepo.findByProductIdAndCustomerId.mockResolvedValue(null);
      mockRepo.create.mockResolvedValue(expected);

      const result = await service.submitReview(reviewData);

      expect(mockRepo.findByProductIdAndCustomerId).toHaveBeenCalledWith('prod-123', 'cust-456');
      expect(mockRepo.create).toHaveBeenCalled();
      expect(result).toEqual(expected);
    });

    it('should throw 409 DUPLICATE_REVIEW when a review already exists for product+customer', async () => {
      mockRepo.findByProductIdAndCustomerId.mockResolvedValue(makeReview());

      await expect(
        service.submitReview({ productId: 'prod-123', customerId: 'cust-456', rating: 4 })
      ).rejects.toMatchObject({ status: 409, code: 'DUPLICATE_REVIEW' });

      expect(mockRepo.create).not.toHaveBeenCalled();
    });
  });

  // ── getReview ─────────────────────────────────────────────────────────────

  describe('getReview', () => {
    it('should return a review when it exists', async () => {
      const expected = makeReview();
      mockRepo.findById.mockResolvedValue(expected);

      const result = await service.getReview('review-id-1');

      expect(mockRepo.findById).toHaveBeenCalledWith('review-id-1');
      expect(result).toEqual(expected);
    });

    it('should throw 404 REVIEW_NOT_FOUND when review does not exist', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.getReview('nonexistent-id')).rejects.toMatchObject({
        status: 404,
        code: 'REVIEW_NOT_FOUND',
      });
    });
  });

  // ── listProductReviews ────────────────────────────────────────────────────

  describe('listProductReviews', () => {
    it('should return list of reviews for a product', async () => {
      const reviews = [makeReview(), makeReview({ _id: 'review-id-2', rating: 4 })];
      mockRepo.findByProductId.mockResolvedValue(reviews);

      const result = await service.listProductReviews('prod-123', { status: 'approved' });

      expect(mockRepo.findByProductId).toHaveBeenCalledWith({
        productId: 'prod-123',
        status: 'approved',
        limit: undefined,
        offset: undefined,
        sortBy: undefined,
      });
      expect(result).toHaveLength(2);
    });

    it('should throw 400 MISSING_PRODUCT_ID when productId is not provided', async () => {
      await expect(service.listProductReviews(null)).rejects.toMatchObject({
        status: 400,
        code: 'MISSING_PRODUCT_ID',
      });
    });
  });

  // ── listCustomerReviews ───────────────────────────────────────────────────

  describe('listCustomerReviews', () => {
    it('should return reviews for a specific customer', async () => {
      const reviews = [makeReview(), makeReview({ _id: 'review-id-3', productId: 'prod-999' })];
      mockRepo.findByCustomerId.mockResolvedValue(reviews);

      const result = await service.listCustomerReviews('cust-456');

      expect(mockRepo.findByCustomerId).toHaveBeenCalledWith('cust-456', {
        limit: undefined,
        offset: undefined,
      });
      expect(result).toHaveLength(2);
    });
  });

  // ── approveReview ─────────────────────────────────────────────────────────

  describe('approveReview', () => {
    it('should update status to approved for an existing review', async () => {
      const updated = makeReview({ status: 'approved' });
      mockRepo.findById.mockResolvedValue(makeReview());
      mockRepo.updateStatus.mockResolvedValue(updated);

      const result = await service.approveReview('review-id-1');

      expect(mockRepo.updateStatus).toHaveBeenCalledWith('review-id-1', 'approved');
      expect(result.status).toBe('approved');
    });

    it('should throw 404 when approving a non-existent review', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.approveReview('bad-id')).rejects.toMatchObject({
        status: 404,
        code: 'REVIEW_NOT_FOUND',
      });
    });
  });

  // ── rejectReview ──────────────────────────────────────────────────────────

  describe('rejectReview', () => {
    it('should update status to rejected for an existing review', async () => {
      const updated = makeReview({ status: 'rejected' });
      mockRepo.findById.mockResolvedValue(makeReview());
      mockRepo.updateStatus.mockResolvedValue(updated);

      const result = await service.rejectReview('review-id-1');

      expect(mockRepo.updateStatus).toHaveBeenCalledWith('review-id-1', 'rejected');
      expect(result.status).toBe('rejected');
    });

    it('should throw 404 when rejecting a non-existent review', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.rejectReview('bad-id')).rejects.toMatchObject({
        status: 404,
        code: 'REVIEW_NOT_FOUND',
      });
    });
  });

  // ── markHelpful ───────────────────────────────────────────────────────────

  describe('markHelpful', () => {
    it('should increment helpfulVotes for an existing review', async () => {
      const updated = makeReview({ helpfulVotes: 1 });
      mockRepo.findById.mockResolvedValue(makeReview());
      mockRepo.incrementHelpful.mockResolvedValue(updated);

      const result = await service.markHelpful('review-id-1');

      expect(mockRepo.incrementHelpful).toHaveBeenCalledWith('review-id-1');
      expect(result.helpfulVotes).toBe(1);
    });

    it('should throw 404 when marking helpful on a non-existent review', async () => {
      mockRepo.findById.mockResolvedValue(null);

      await expect(service.markHelpful('bad-id')).rejects.toMatchObject({
        status: 404,
        code: 'REVIEW_NOT_FOUND',
      });
    });
  });

  // ── getProductRatings ─────────────────────────────────────────────────────

  describe('getProductRatings', () => {
    it('should return aggregated rating stats for a product', async () => {
      const stats = {
        avgRating: 4.5,
        count: 10,
        distribution: { 1: 0, 2: 1, 3: 1, 4: 3, 5: 5 },
      };
      mockRepo.getProductStats.mockResolvedValue(stats);

      const result = await service.getProductRatings('prod-123');

      expect(mockRepo.getProductStats).toHaveBeenCalledWith('prod-123');
      expect(result.avgRating).toBe(4.5);
      expect(result.count).toBe(10);
      expect(result.distribution[5]).toBe(5);
    });

    it('should return zero stats when product has no approved reviews', async () => {
      const emptyStats = {
        avgRating: 0,
        count: 0,
        distribution: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 },
      };
      mockRepo.getProductStats.mockResolvedValue(emptyStats);

      const result = await service.getProductRatings('prod-no-reviews');

      expect(result.avgRating).toBe(0);
      expect(result.count).toBe(0);
    });

    it('should throw 400 MISSING_PRODUCT_ID when productId is not provided', async () => {
      await expect(service.getProductRatings(null)).rejects.toMatchObject({
        status: 400,
        code: 'MISSING_PRODUCT_ID',
      });
    });
  });
});
