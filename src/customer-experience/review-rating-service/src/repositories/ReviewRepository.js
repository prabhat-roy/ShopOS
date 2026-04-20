'use strict';

const Review = require('../models/Review');

class ReviewRepository {
  async create(data) {
    const review = new Review(data);
    return review.save();
  }

  async findById(id) {
    return Review.findById(id).lean();
  }

  async findByProductId({ productId, status, limit = 20, offset = 0, sortBy = 'createdAt' }) {
    const query = { productId };
    if (status) {
      query.status = status;
    }

    const allowedSortFields = {
      createdAt: { createdAt: -1 },
      rating_asc: { rating: 1 },
      rating_desc: { rating: -1 },
      helpful: { helpfulVotes: -1 },
    };

    const sort = allowedSortFields[sortBy] || { createdAt: -1 };

    return Review.find(query).sort(sort).skip(offset).limit(limit).lean();
  }

  async findByCustomerId(customerId, { limit = 20, offset = 0 } = {}) {
    return Review.find({ customerId }).sort({ createdAt: -1 }).skip(offset).limit(limit).lean();
  }

  async findByProductIdAndCustomerId(productId, customerId) {
    return Review.findOne({ productId, customerId }).lean();
  }

  async updateStatus(id, status) {
    return Review.findByIdAndUpdate(id, { status }, { new: true }).lean();
  }

  async incrementHelpful(id) {
    return Review.findByIdAndUpdate(id, { $inc: { helpfulVotes: 1 } }, { new: true }).lean();
  }

  async getProductStats(productId) {
    const pipeline = [
      { $match: { productId, status: 'approved' } },
      {
        $group: {
          _id: null,
          avgRating: { $avg: '$rating' },
          count: { $sum: 1 },
          count1: { $sum: { $cond: [{ $eq: ['$rating', 1] }, 1, 0] } },
          count2: { $sum: { $cond: [{ $eq: ['$rating', 2] }, 1, 0] } },
          count3: { $sum: { $cond: [{ $eq: ['$rating', 3] }, 1, 0] } },
          count4: { $sum: { $cond: [{ $eq: ['$rating', 4] }, 1, 0] } },
          count5: { $sum: { $cond: [{ $eq: ['$rating', 5] }, 1, 0] } },
        },
      },
      {
        $project: {
          _id: 0,
          avgRating: { $round: ['$avgRating', 2] },
          count: 1,
          distribution: {
            1: '$count1',
            2: '$count2',
            3: '$count3',
            4: '$count4',
            5: '$count5',
          },
        },
      },
    ];

    const result = await Review.aggregate(pipeline);

    if (!result.length) {
      return {
        avgRating: 0,
        count: 0,
        distribution: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 },
      };
    }

    return result[0];
  }
}

module.exports = ReviewRepository;
