'use strict';

const Configurator = require('../models/configuratorModel');

/**
 * Retrieve a configurator by productId.
 * Throws a 404 error if not found.
 */
async function getConfigurator(productId) {
  const doc = await Configurator.findOne({ productId }).lean();
  if (!doc) {
    const err = new Error(`Configurator not found for productId: ${productId}`);
    err.status = 404;
    throw err;
  }
  return doc;
}

/**
 * Create or upsert a configurator for the given productId.
 * Returns the saved document.
 */
async function createConfigurator(productId, data) {
  const doc = await Configurator.findOneAndUpdate(
    { productId },
    { productId, options: data.options || [], rules: data.rules || [] },
    { upsert: true, new: true, runValidators: true, setDefaultsOnInsert: true }
  ).lean();
  return doc;
}

/**
 * Validate a set of selections against a configurator's options and rules.
 *
 * selections: { [optionName]: value }
 *
 * Returns { valid: boolean, errors: string[], totalPriceAdj: number }
 */
async function validateSelection(productId, selections) {
  const configurator = await getConfigurator(productId);

  const errors = [];
  let totalPriceAdj = 0;

  // Build a fast lookup map: optionName -> option definition
  const optionMap = {};
  for (const opt of configurator.options) {
    optionMap[opt.name] = opt;
  }

  // 1. Check required options are present
  for (const opt of configurator.options) {
    if (opt.required && !(opt.name in selections)) {
      errors.push(`Required option "${opt.name}" is missing.`);
    }
  }

  // 2. Validate each provided selection
  for (const [optionName, selectedValue] of Object.entries(selections)) {
    const opt = optionMap[optionName];

    if (!opt) {
      errors.push(`Unknown option "${optionName}".`);
      continue;
    }

    // Normalise to array for uniform processing (checkbox can be multi-value)
    const selectedValues = Array.isArray(selectedValue) ? selectedValue : [selectedValue];

    for (const val of selectedValues) {
      const choice = opt.choices.find((c) => c.value === val);
      if (!choice) {
        errors.push(`Invalid value "${val}" for option "${optionName}".`);
        continue;
      }
      if (!choice.available) {
        errors.push(`Value "${val}" for option "${optionName}" is currently unavailable.`);
        continue;
      }
      totalPriceAdj += choice.priceAdj || 0;
    }
  }

  // 3. Apply rules — if a condition matches, check that the affected option's
  //    selected value is in the allowed set.
  for (const rule of configurator.rules) {
    const { condition, effect } = rule;
    const condValue = selections[condition.optionName];

    // Normalise condition value too
    const condValues = Array.isArray(condValue) ? condValue : [condValue];

    if (condValues.includes(condition.value)) {
      // Rule fires — validate the effect option
      const affectedValues = selections[effect.optionName];
      if (affectedValues !== undefined) {
        const affectedArr = Array.isArray(affectedValues) ? affectedValues : [affectedValues];
        for (const v of affectedArr) {
          if (!effect.allowedValues.includes(v)) {
            errors.push(
              `When "${condition.optionName}" is "${condition.value}", ` +
                `"${effect.optionName}" must be one of [${effect.allowedValues.join(', ')}]. ` +
                `Got "${v}".`
            );
          }
        }
      }
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    totalPriceAdj,
  };
}

/**
 * Update an existing configurator's options and/or rules.
 * Throws 404 if the configurator does not exist.
 */
async function updateConfigurator(productId, data) {
  const update = {};
  if (data.options !== undefined) update.options = data.options;
  if (data.rules !== undefined) update.rules = data.rules;

  const doc = await Configurator.findOneAndUpdate(
    { productId },
    { $set: update },
    { new: true, runValidators: true }
  ).lean();

  if (!doc) {
    const err = new Error(`Configurator not found for productId: ${productId}`);
    err.status = 404;
    throw err;
  }
  return doc;
}

/**
 * Delete a configurator by productId.
 * Throws 404 if not found.
 */
async function deleteConfigurator(productId) {
  const result = await Configurator.findOneAndDelete({ productId });
  if (!result) {
    const err = new Error(`Configurator not found for productId: ${productId}`);
    err.status = 404;
    throw err;
  }
}

module.exports = {
  getConfigurator,
  createConfigurator,
  validateSelection,
  updateConfigurator,
  deleteConfigurator,
};
