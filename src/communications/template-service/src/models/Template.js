'use strict';

const mongoose = require('mongoose');

const variableSchema = new mongoose.Schema(
  {
    name: {
      type: String,
      required: true,
      trim: true,
    },
    required: {
      type: Boolean,
      default: false,
    },
    defaultValue: {
      type: String,
      default: null,
    },
  },
  { _id: false },
);

const templateSchema = new mongoose.Schema(
  {
    name: {
      type: String,
      required: true,
      unique: true,
      trim: true,
      index: true,
    },
    channel: {
      type: String,
      required: true,
      enum: ['email', 'sms', 'push', 'in_app'],
      index: true,
    },
    subject: {
      type: String,
      trim: true,
      default: null,
    },
    body: {
      type: String,
      required: true,
    },
    variables: {
      type: [variableSchema],
      default: [],
    },
    active: {
      type: Boolean,
      default: true,
      index: true,
    },
    version: {
      type: Number,
      default: 1,
      min: 1,
    },
    tags: {
      type: [String],
      default: [],
      index: true,
    },
  },
  {
    timestamps: true,
    versionKey: false,
  },
);

// Compound index for common list queries
templateSchema.index({ channel: 1, active: 1 });
templateSchema.index({ tags: 1, active: 1 });

const Template = mongoose.model('Template', templateSchema);

module.exports = Template;
