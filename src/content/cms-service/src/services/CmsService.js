'use strict';

const { marked } = require('marked');
const { v4: uuidv4 } = require('uuid');
const ContentRepository = require('../repositories/ContentRepository');

class NotFoundError extends Error {
  constructor(message) {
    super(message);
    this.name = 'NotFoundError';
    this.statusCode = 404;
  }
}

class ValidationError extends Error {
  constructor(message) {
    super(message);
    this.name = 'ValidationError';
    this.statusCode = 400;
  }
}

function generateSlug(title) {
  return title
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

function renderMarkdown(body) {
  if (!body) return '';
  return marked.parse(body);
}

class CmsService {
  async createContent(data) {
    const { title, body } = data;
    if (!title) throw new ValidationError('title is required');
    if (!data.type) throw new ValidationError('type is required');

    const slug = data.slug ? data.slug : `${generateSlug(title)}-${uuidv4().slice(0, 8)}`;
    const htmlContent = renderMarkdown(body);

    const payload = {
      ...data,
      slug,
      htmlContent,
    };

    return ContentRepository.create(payload);
  }

  async getBySlug(slug, locale) {
    const item = await ContentRepository.findBySlug(slug, locale);
    if (!item) throw new NotFoundError(`Content not found for slug: ${slug}`);
    return item;
  }

  async getById(id) {
    const item = await ContentRepository.findById(id);
    if (!item) throw new NotFoundError(`Content not found: ${id}`);
    return item;
  }

  async listContent(filters) {
    return ContentRepository.list(filters);
  }

  async updateContent(id, data) {
    const existing = await ContentRepository.findById(id);
    if (!existing) throw new NotFoundError(`Content not found: ${id}`);

    const updatePayload = { ...data };

    if (data.body !== undefined) {
      updatePayload.htmlContent = renderMarkdown(data.body);
    }

    const updated = await ContentRepository.update(id, updatePayload);
    if (!updated) throw new NotFoundError(`Content not found: ${id}`);
    return updated;
  }

  async publishContent(id) {
    const existing = await ContentRepository.findById(id);
    if (!existing) throw new NotFoundError(`Content not found: ${id}`);

    const published = await ContentRepository.publish(id);
    return published;
  }

  async archiveContent(id) {
    const existing = await ContentRepository.findById(id);
    if (!existing) throw new NotFoundError(`Content not found: ${id}`);

    const archived = await ContentRepository.archive(id);
    return archived;
  }

  async deleteContent(id) {
    const existing = await ContentRepository.findById(id);
    if (!existing) throw new NotFoundError(`Content not found: ${id}`);
    if (existing.status !== 'draft') {
      throw new ValidationError('Only draft content can be deleted');
    }

    await ContentRepository.delete(id);
  }

  async searchContent(queryText, locale) {
    if (!queryText || queryText.trim().length === 0) {
      throw new ValidationError('Search query must not be empty');
    }
    return ContentRepository.search(queryText, locale);
  }
}

CmsService.NotFoundError = NotFoundError;
CmsService.ValidationError = ValidationError;

module.exports = new CmsService();
module.exports.CmsService = CmsService;
module.exports.NotFoundError = NotFoundError;
module.exports.ValidationError = ValidationError;
