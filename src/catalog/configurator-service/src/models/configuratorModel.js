'use strict';

const mongoose = require('mongoose');

const choiceSchema = new mongoose.Schema(
  {
    value: { type: String, required: true },
    label: { type: String, required: true },
    priceAdj: { type: Number, default: 0 },
    available: { type: Boolean, default: true },
  },
  { _id: false }
);

const optionSchema = new mongoose.Schema(
  {
    name: { type: String, required: true },
    type: {
      type: String,
      enum: ['select', 'radio', 'checkbox'],
      required: true,
    },
    required: { type: Boolean, default: false },
    choices: { type: [choiceSchema], default: [] },
  },
  { _id: false }
);

const ruleConditionSchema = new mongoose.Schema(
  {
    optionName: { type: String, required: true },
    value: { type: String, required: true },
  },
  { _id: false }
);

const ruleEffectSchema = new mongoose.Schema(
  {
    optionName: { type: String, required: true },
    allowedValues: { type: [String], default: [] },
  },
  { _id: false }
);

const ruleSchema = new mongoose.Schema(
  {
    condition: { type: ruleConditionSchema, required: true },
    effect: { type: ruleEffectSchema, required: true },
  },
  { _id: false }
);

const configuratorSchema = new mongoose.Schema(
  {
    productId: { type: String, required: true, unique: true, index: true },
    options: { type: [optionSchema], default: [] },
    rules: { type: [ruleSchema], default: [] },
  },
  {
    timestamps: true,
    collection: 'configurators',
  }
);

const Configurator = mongoose.model('Configurator', configuratorSchema);

module.exports = Configurator;
