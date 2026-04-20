'use strict';

const mongoose = require('mongoose');

const answerSchema = new mongoose.Schema(
  {
    customerId: {
      type: String,
      required: true,
    },
    body: {
      type: String,
      required: true,
      trim: true,
      minlength: 1,
      maxlength: 5000,
    },
    isStaff: {
      type: Boolean,
      default: false,
    },
    helpful: {
      type: Number,
      default: 0,
      min: 0,
    },
    createdAt: {
      type: Date,
      default: Date.now,
    },
  },
  { _id: true, versionKey: false }
);

const questionSchema = new mongoose.Schema(
  {
    productId: {
      type: String,
      required: true,
      index: true,
    },
    customerId: {
      type: String,
      required: true,
    },
    question: {
      type: String,
      required: true,
      trim: true,
      minlength: 10,
      maxlength: 1000,
    },
    status: {
      type: String,
      enum: ['open', 'answered', 'closed'],
      default: 'open',
    },
    answers: {
      type: [answerSchema],
      default: [],
    },
    tags: {
      type: [String],
      default: [],
    },
    viewCount: {
      type: Number,
      default: 0,
      min: 0,
    },
  },
  {
    timestamps: true,
    versionKey: false,
  }
);

// Compound index for product+status queries
questionSchema.index({ productId: 1, status: 1, createdAt: -1 });
questionSchema.index({ customerId: 1, createdAt: -1 });

const Question = mongoose.model('Question', questionSchema);

module.exports = Question;
