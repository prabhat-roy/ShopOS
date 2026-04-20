'use strict';

const ContentItem = require('../models/ContentItem');

class ContentRepository {
  async create(data) {
    const item = new ContentItem(data);
    return item.save();
  }

  async findBySlug(slug, locale) {
    const query = { slug: slug.toLowerCase() };
    if (locale) query.locale = locale;
    return ContentItem.findOne(query).lean();
  }

  async findById(id) {
    return ContentItem.findById(id).lean();
  }

  async list({ type, status, locale, tags, limit = 20, offset = 0 } = {}) {
    const query = {};
    if (type) query.type = type;
    if (status) query.status = status;
    if (locale) query.locale = locale;
    if (tags && tags.length > 0) query.tags = { $in: Array.isArray(tags) ? tags : [tags] };

    const [items, total] = await Promise.all([
      ContentItem.find(query)
        .sort({ createdAt: -1 })
        .skip(Number(offset))
        .limit(Number(limit))
        .lean(),
      ContentItem.countDocuments(query),
    ]);

    return { items, total, limit: Number(limit), offset: Number(offset) };
  }

  async update(id, data) {
    return ContentItem.findByIdAndUpdate(id, { $set: data }, { new: true, runValidators: true }).lean();
  }

  async publish(id) {
    return ContentItem.findByIdAndUpdate(
      id,
      { $set: { status: 'published', publishedAt: new Date() } },
      { new: true }
    ).lean();
  }

  async archive(id) {
    return ContentItem.findByIdAndUpdate(
      id,
      { $set: { status: 'archived' } },
      { new: true }
    ).lean();
  }

  async delete(id) {
    return ContentItem.findByIdAndDelete(id).lean();
  }

  async search(queryText, locale) {
    const filter = { $text: { $search: queryText } };
    if (locale) filter.locale = locale;
    return ContentItem.find(filter, { score: { $meta: 'textScore' } })
      .sort({ score: { $meta: 'textScore' } })
      .limit(50)
      .lean();
  }
}

module.exports = new ContentRepository();
