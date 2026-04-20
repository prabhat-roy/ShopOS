'use strict';

const mongoose = require('mongoose');

const contentItemSchema = new mongoose.Schema(
  {
    slug: {
      type: String,
      required: true,
      unique: true,
      index: true,
      trim: true,
      lowercase: true,
    },
    type: {
      type: String,
      required: true,
      enum: ['page', 'blog_post', 'landing_page', 'content_block'],
    },
    title: {
      type: String,
      required: true,
      trim: true,
    },
    body: {
      type: String,
      default: '',
    },
    htmlContent: {
      type: String,
      default: '',
    },
    status: {
      type: String,
      enum: ['draft', 'published', 'archived'],
      default: 'draft',
    },
    locale: {
      type: String,
      default: 'en',
      index: true,
    },
    tags: {
      type: [String],
      default: [],
      index: true,
    },
    author: {
      type: String,
      default: '',
    },
    metadata: {
      type: mongoose.Schema.Types.Mixed,
      default: {},
    },
    publishedAt: {
      type: Date,
      default: null,
    },
  },
  {
    timestamps: true,
    versionKey: false,
  }
);

contentItemSchema.index({ status: 1, locale: 1 });
contentItemSchema.index({ type: 1, status: 1 });
contentItemSchema.index({ title: 'text', body: 'text' });

const ContentItem = mongoose.model('ContentItem', contentItemSchema);

module.exports = ContentItem;
