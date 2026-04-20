'use strict';

const Joi = require('joi');
const service = require('../services/seoService');

// ─── Validation schemas ───────────────────────────────────────────────────────

const productBodySchema = Joi.object({
  id: Joi.string().required(),
  name: Joi.string().required(),
  description: Joi.string().allow('', null).optional(),
  price: Joi.number().min(0).optional(),
  currency: Joi.string().length(3).uppercase().optional(),
  images: Joi.array().items(Joi.string().uri()).optional(),
  brand: Joi.string().allow('', null).optional(),
});

const categoryBodySchema = Joi.object({
  name: Joi.string().required(),
  slug: Joi.string().required(),
  description: Joi.string().allow('', null).optional(),
});

const brandBodySchema = Joi.object({
  name: Joi.string().required(),
  slug: Joi.string().required(),
  description: Joi.string().allow('', null).optional(),
});

const sitemapQuerySchema = Joi.object({
  urls: Joi.alternatives()
    .try(
      Joi.array().items(Joi.string().uri()),
      Joi.string().uri() // single URL passed as a string query param
    )
    .required(),
  priority: Joi.number().min(0).max(1).optional(),
  changefreq: Joi.string()
    .valid('always', 'hourly', 'daily', 'weekly', 'monthly', 'yearly', 'never')
    .optional(),
});

// ─── Helpers ──────────────────────────────────────────────────────────────────

function validate(schema, data, options = {}) {
  return schema.validate(data, { abortEarly: false, ...options });
}

function handleError(res, err) {
  const status = err.status || 500;
  res.status(status).json({ error: err.message });
}

// ─── Controllers ─────────────────────────────────────────────────────────────

function productMeta(req, res) {
  const { error, value } = validate(productBodySchema, req.body);
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const meta = service.generateProductMeta(value);
    res.json(meta);
  } catch (err) {
    handleError(res, err);
  }
}

function categoryMeta(req, res) {
  const { error, value } = validate(categoryBodySchema, req.body);
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const meta = service.generateCategoryMeta(value);
    res.json(meta);
  } catch (err) {
    handleError(res, err);
  }
}

function brandMeta(req, res) {
  const { error, value } = validate(brandBodySchema, req.body);
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }
  try {
    const meta = service.generateBrandMeta(value);
    res.json(meta);
  } catch (err) {
    handleError(res, err);
  }
}

function sitemap(req, res) {
  // urls can be passed as repeated query params (?urls=...&urls=...) or a single string
  const rawQuery = { ...req.query };

  const { error, value } = validate(sitemapQuerySchema, rawQuery);
  if (error) {
    return res.status(400).json({ error: error.details.map((d) => d.message).join('; ') });
  }

  try {
    const urls = Array.isArray(value.urls) ? value.urls : [value.urls];
    const priority = value.priority !== undefined ? value.priority : 0.5;
    const changefreq = value.changefreq || 'weekly';

    const entries = urls.map((url) => service.generateSitemapEntry(url, priority, changefreq));

    const xml =
      `<?xml version="1.0" encoding="UTF-8"?>\n` +
      `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n` +
      entries.join('\n') +
      `\n</urlset>`;

    res.set('Content-Type', 'application/xml');
    res.send(xml);
  } catch (err) {
    handleError(res, err);
  }
}

module.exports = { productMeta, categoryMeta, brandMeta, sitemap };
