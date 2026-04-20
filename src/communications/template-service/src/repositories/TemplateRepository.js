'use strict';

const Template = require('../models/Template');

/**
 * Creates a new template document.
 * @param {object} data
 * @returns {Promise<object>}
 */
async function create(data) {
  const template = new Template(data);
  return template.save();
}

/**
 * Finds a template by its MongoDB ObjectId.
 * @param {string} id
 * @returns {Promise<object|null>}
 */
async function findById(id) {
  return Template.findById(id).lean();
}

/**
 * Finds an active template by its unique name.
 * @param {string} name
 * @returns {Promise<object|null>}
 */
async function findByName(name) {
  return Template.findOne({ name, active: true }).lean();
}

/**
 * Lists templates with optional filters.
 * @param {object} [filters]
 * @param {string} [filters.channel]
 * @param {boolean} [filters.active]
 * @param {string[]} [filters.tags]
 * @returns {Promise<object[]>}
 */
async function list(filters = {}) {
  const query = {};

  if (filters.channel !== undefined) {
    query.channel = filters.channel;
  }

  if (filters.active !== undefined) {
    query.active = filters.active;
  } else {
    // Default to showing active templates only
    query.active = true;
  }

  if (filters.tags && filters.tags.length > 0) {
    query.tags = { $in: filters.tags };
  }

  return Template.find(query).sort({ createdAt: -1 }).lean();
}

/**
 * Updates a template by id, merging the provided data.
 * @param {string} id
 * @param {object} data
 * @returns {Promise<object|null>}
 */
async function update(id, data) {
  return Template.findByIdAndUpdate(id, { $set: data }, { new: true, runValidators: true }).lean();
}

/**
 * Archives (soft-deletes) a template by setting active = false.
 * @param {string} id
 * @returns {Promise<object|null>}
 */
async function archive(id) {
  return Template.findByIdAndUpdate(id, { $set: { active: false } }, { new: true }).lean();
}

/**
 * Renders a template by replacing {{variable}} placeholders with provided values.
 * Uses defaultValue for missing optional variables.
 * Throws if a required variable is missing and has no default.
 *
 * @param {object} template - Template document (must have .body and optionally .subject)
 * @param {object} data - Key-value pairs of variable substitutions
 * @returns {{ subject: string|null, body: string }}
 */
function render(template, data) {
  const values = { ...data };

  // Resolve defaults and check required variables
  for (const variable of template.variables || []) {
    const hasValue = values[variable.name] !== undefined && values[variable.name] !== null;

    if (!hasValue) {
      if (variable.required) {
        const err = new Error(
          `Required template variable "{{${variable.name}}}" is missing and has no default`,
        );
        err.code = 'MISSING_REQUIRED_VARIABLE';
        err.variable = variable.name;
        throw err;
      }

      if (variable.defaultValue !== null && variable.defaultValue !== undefined) {
        values[variable.name] = variable.defaultValue;
      } else {
        // Leave unresolved placeholder as empty string
        values[variable.name] = '';
      }
    }
  }

  function substitute(text) {
    if (!text) return text;
    return text.replace(/\{\{(\w+)\}\}/g, (_match, key) => {
      return values[key] !== undefined ? String(values[key]) : '';
    });
  }

  return {
    subject: substitute(template.subject),
    body: substitute(template.body),
  };
}

module.exports = { create, findById, findByName, list, update, archive, render };
