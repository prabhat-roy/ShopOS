'use strict';

const mongoose = require('mongoose');

const trackingEventSchema = new mongoose.Schema(
  {
    timestamp: {
      type: Date,
      default: Date.now,
    },
    location: {
      type: String,
      default: '',
    },
    description: {
      type: String,
      required: true,
    },
    status: {
      type: String,
      default: '',
    },
  },
  { _id: false }
);

const shipmentSchema = new mongoose.Schema(
  {
    trackingNumber: {
      type: String,
      required: true,
      unique: true,
      index: true,
      trim: true,
    },
    carrier: {
      type: String,
      required: true,
      trim: true,
    },
    status: {
      type: String,
      enum: ['created', 'in_transit', 'out_for_delivery', 'delivered', 'failed', 'returned'],
      default: 'created',
    },
    originAddress: {
      type: String,
      default: '',
    },
    destinationAddress: {
      type: String,
      default: '',
    },
    estimatedDelivery: {
      type: Date,
      default: null,
    },
    actualDelivery: {
      type: Date,
      default: null,
    },
    events: {
      type: [trackingEventSchema],
      default: [],
    },
    metadata: {
      type: mongoose.Schema.Types.Mixed,
      default: {},
    },
  },
  {
    timestamps: true,
    versionKey: false,
  }
);

shipmentSchema.index({ carrier: 1, status: 1 });
shipmentSchema.index({ createdAt: -1 });

const Shipment = mongoose.model('Shipment', shipmentSchema);

module.exports = Shipment;
