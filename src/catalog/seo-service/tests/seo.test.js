'use strict';

/**
 * Tests for seo-service.
 *
 * No database — all logic is pure computation, so no mocking required.
 */

const request = require('supertest');
const app = require('../index');
const seoService = require('../src/services/seoService');

// Suppress console output during tests
beforeAll(() => {
  jest.spyOn(console, 'log').mockImplementation(() => {});
  jest.spyOn(console, 'error').mockImplementation(() => {});
});

afterAll(() => {
  jest.restoreAllMocks();
});

// ─── /healthz ─────────────────────────────────────────────────────────────────

describe('GET /healthz', () => {
  it('returns 200 with status ok', async () => {
    const res = await request(app).get('/healthz');
    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: 'ok' });
  });
});

// ─── seoService.generateProductMeta ──────────────────────────────────────────

describe('seoService.generateProductMeta', () => {
  const product = {
    id: 'laptop-pro-x1',
    name: 'Pro X1 Laptop',
    description: 'High performance laptop for professionals.',
    price: 1299.99,
    currency: 'USD',
    images: ['https://cdn.shopos.com/products/laptop-pro-x1.jpg'],
    brand: 'TechBrand',
  };

  it('returns a title containing the product name', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.title).toContain(product.name);
  });

  it('returns a title containing the brand name', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.title).toContain(product.brand);
  });

  it('title does not exceed 70 characters', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.title.length).toBeLessThanOrEqual(70);
  });

  it('description does not exceed 160 characters', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.description.length).toBeLessThanOrEqual(160);
  });

  it('canonical URL contains the product id', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.canonical).toContain(product.id);
  });

  it('returns a JSON-LD object with @type Product', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.jsonLd['@type']).toBe('Product');
    expect(meta.jsonLd.name).toBe(product.name);
  });

  it('JSON-LD includes offers with the correct price', () => {
    const meta = seoService.generateProductMeta(product);
    expect(meta.jsonLd.offers).toBeDefined();
    expect(meta.jsonLd.offers.price).toBe(product.price);
    expect(meta.jsonLd.offers.priceCurrency).toBe('USD');
  });

  it('works without optional fields (no brand, no description, no price)', () => {
    const minimal = { id: 'prod-min', name: 'Minimal Product' };
    const meta = seoService.generateProductMeta(minimal);
    expect(meta.title).toContain('Minimal Product');
    expect(meta.canonical).toContain('prod-min');
    expect(meta.jsonLd.offers).toBeUndefined();
    expect(meta.jsonLd.brand).toBeUndefined();
  });
});

// ─── POST /seo/product ────────────────────────────────────────────────────────

describe('POST /seo/product', () => {
  it('returns 200 with SEO metadata for a valid product', async () => {
    const res = await request(app).post('/seo/product').send({
      id: 'laptop-001',
      name: 'SuperBook Pro',
      description: 'The best laptop money can buy.',
      price: 1999,
      currency: 'USD',
      brand: 'SuperTech',
    });

    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('title');
    expect(res.body.title).toContain('SuperBook Pro');
    expect(res.body).toHaveProperty('description');
    expect(res.body).toHaveProperty('canonical');
    expect(res.body).toHaveProperty('jsonLd');
    expect(res.body.jsonLd['@type']).toBe('Product');
  });

  it('returns 400 when required fields are missing', async () => {
    const res = await request(app).post('/seo/product').send({ description: 'No name or id' });
    expect(res.status).toBe(400);
    expect(res.body).toHaveProperty('error');
  });
});

// ─── seoService.generateCategoryMeta ─────────────────────────────────────────

describe('seoService.generateCategoryMeta', () => {
  const category = {
    name: 'Laptops',
    slug: 'laptops',
    description: 'Browse our wide range of laptops.',
  };

  it('returns title containing category name', () => {
    const meta = seoService.generateCategoryMeta(category);
    expect(meta.title).toContain(category.name);
  });

  it('returns the correct canonical URL', () => {
    const meta = seoService.generateCategoryMeta(category);
    expect(meta.canonical).toContain('/categories/laptops');
  });

  it('returns a BreadcrumbList JSON-LD object', () => {
    const meta = seoService.generateCategoryMeta(category);
    expect(meta.jsonLd['@type']).toBe('BreadcrumbList');
    expect(meta.jsonLd.itemListElement).toHaveLength(2);
    expect(meta.jsonLd.itemListElement[1].name).toBe('Laptops');
  });
});

// ─── POST /seo/category ───────────────────────────────────────────────────────

describe('POST /seo/category', () => {
  it('returns 200 with correct canonical URL', async () => {
    const res = await request(app).post('/seo/category').send({
      name: 'Smartphones',
      slug: 'smartphones',
      description: 'Latest smartphones.',
    });

    expect(res.status).toBe(200);
    expect(res.body.canonical).toContain('/categories/smartphones');
  });

  it('returns 400 when slug is missing', async () => {
    const res = await request(app).post('/seo/category').send({ name: 'No Slug' });
    expect(res.status).toBe(400);
  });
});

// ─── seoService.generateBrandMeta ────────────────────────────────────────────

describe('seoService.generateBrandMeta', () => {
  it('returns title containing brand name', () => {
    const meta = seoService.generateBrandMeta({ name: 'TechBrand', slug: 'techbrand' });
    expect(meta.title).toContain('TechBrand');
  });

  it('returns the correct canonical URL', () => {
    const meta = seoService.generateBrandMeta({ name: 'TechBrand', slug: 'techbrand' });
    expect(meta.canonical).toContain('/brands/techbrand');
  });
});

// ─── POST /seo/brand ──────────────────────────────────────────────────────────

describe('POST /seo/brand', () => {
  it('returns 200 with brand metadata', async () => {
    const res = await request(app)
      .post('/seo/brand')
      .send({ name: 'Acme Corp', slug: 'acme-corp' });

    expect(res.status).toBe(200);
    expect(res.body.title).toContain('Acme Corp');
    expect(res.body.canonical).toContain('/brands/acme-corp');
  });

  it('returns 400 when name is missing', async () => {
    const res = await request(app).post('/seo/brand').send({ slug: 'no-name' });
    expect(res.status).toBe(400);
  });
});

// ─── seoService.generateSitemapEntry ─────────────────────────────────────────

describe('seoService.generateSitemapEntry', () => {
  it('returns a string containing the URL in a <loc> tag', () => {
    const entry = seoService.generateSitemapEntry('https://shopos.com/products/p1', 0.8, 'daily');
    expect(entry).toContain('<loc>https://shopos.com/products/p1</loc>');
  });

  it('returns the specified priority', () => {
    const entry = seoService.generateSitemapEntry('https://shopos.com/', 1.0, 'always');
    expect(entry).toContain('<priority>1.0</priority>');
  });

  it('returns the specified changefreq', () => {
    const entry = seoService.generateSitemapEntry('https://shopos.com/cat', 0.5, 'monthly');
    expect(entry).toContain('<changefreq>monthly</changefreq>');
  });
});

// ─── GET /seo/sitemap ────────────────────────────────────────────────────────

describe('GET /seo/sitemap', () => {
  it('returns XML with correct content-type and <urlset> wrapper', async () => {
    const res = await request(app)
      .get('/seo/sitemap')
      .query({ urls: 'https://shopos.com/products/p1', priority: 0.8, changefreq: 'daily' });

    expect(res.status).toBe(200);
    expect(res.headers['content-type']).toMatch(/application\/xml/);
    expect(res.text).toContain('<urlset');
    expect(res.text).toContain('<loc>https://shopos.com/products/p1</loc>');
  });

  it('returns 400 when urls query param is missing', async () => {
    const res = await request(app).get('/seo/sitemap');
    expect(res.status).toBe(400);
  });
});
