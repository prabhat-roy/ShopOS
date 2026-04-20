'use strict';

const repo = require('../repositories/TemplateRepository');

/**
 * Creates a new template.
 * @param {object} data
 * @returns {Promise<object>}
 */
async function createTemplate(data) {
  return repo.create(data);
}

/**
 * Retrieves a template by its id.
 * Throws 404-style error if not found.
 * @param {string} id
 * @returns {Promise<object>}
 */
async function getTemplate(id) {
  const template = await repo.findById(id);

  if (!template) {
    const err = new Error(`Template with id "${id}" not found`);
    err.code = 'NOT_FOUND';
    err.statusCode = 404;
    throw err;
  }

  return template;
}

/**
 * Retrieves an active template by name.
 * Throws 404-style error if not found.
 * @param {string} name
 * @returns {Promise<object>}
 */
async function getByName(name) {
  const template = await repo.findByName(name);

  if (!template) {
    const err = new Error(`Template with name "${name}" not found or is not active`);
    err.code = 'NOT_FOUND';
    err.statusCode = 404;
    throw err;
  }

  return template;
}

/**
 * Lists templates, optionally filtered by channel, active flag, and tags.
 * @param {object} filters
 * @returns {Promise<object[]>}
 */
async function listTemplates(filters = {}) {
  // Normalise the `active` filter coming in from query params (strings)
  if (filters.active !== undefined) {
    if (typeof filters.active === 'string') {
      filters.active = filters.active === 'true';
    }
  }

  if (filters.tags && typeof filters.tags === 'string') {
    filters.tags = filters.tags.split(',').map((t) => t.trim()).filter(Boolean);
  }

  return repo.list(filters);
}

/**
 * Updates a template by id.
 * Bumps the version number on every update.
 * @param {string} id
 * @param {object} data
 * @returns {Promise<object>}
 */
async function updateTemplate(id, data) {
  const existing = await getTemplate(id);

  // Bump version
  const updatePayload = { ...data, version: existing.version + 1 };

  const updated = await repo.update(id, updatePayload);

  if (!updated) {
    const err = new Error(`Template with id "${id}" not found`);
    err.code = 'NOT_FOUND';
    err.statusCode = 404;
    throw err;
  }

  return updated;
}

/**
 * Archives (soft-deletes) a template.
 * @param {string} id
 * @returns {Promise<object>}
 */
async function archiveTemplate(id) {
  const archived = await repo.archive(id);

  if (!archived) {
    const err = new Error(`Template with id "${id}" not found`);
    err.code = 'NOT_FOUND';
    err.statusCode = 404;
    throw err;
  }

  return archived;
}

/**
 * Renders a template by name or id with the provided variable values.
 * @param {string} nameOrId - Template name or MongoDB id
 * @param {object} variables - Key-value substitution map
 * @returns {Promise<{ templateId: string, name: string, channel: string, subject: string|null, body: string }>}
 */
async function renderTemplate(nameOrId, variables = {}) {
  let template;

  // Try name first (more common usage), then fall back to id lookup
  try {
    template = await repo.findByName(nameOrId);
  } catch (_e) {
    template = null;
  }

  if (!template) {
    try {
      template = await repo.findById(nameOrId);
    } catch (_e) {
      template = null;
    }
  }

  if (!template) {
    const err = new Error(`Template "${nameOrId}" not found`);
    err.code = 'NOT_FOUND';
    err.statusCode = 404;
    throw err;
  }

  const rendered = repo.render(template, variables);

  return {
    templateId: String(template._id),
    name: template.name,
    channel: template.channel,
    subject: rendered.subject,
    body: rendered.body,
  };
}

module.exports = {
  createTemplate,
  getTemplate,
  getByName,
  listTemplates,
  updateTemplate,
  archiveTemplate,
  renderTemplate,
};
