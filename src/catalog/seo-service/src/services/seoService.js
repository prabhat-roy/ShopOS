'use strict';

const config = require('../config');
const { buildProductJsonLd } = require('../templates/productSchema');
const { buildBreadcrumbJsonLd } = require('../templates/breadcrumbSchema');

const BASE_URL = config.baseUrl.replace(/\/$/, ''); // strip trailing slash

/**
 * Truncate a string to maxLen characters, appending '…' when cut.
 */
function truncate(str, maxLen) {
  if (!str) return '';
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen - 1) + '\u2026';
}

/**
 * Strip HTML tags from a string (for use in meta descriptions).
 */
function stripHtml(str) {
  if (!str) return '';
  return str.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim();
}

// ─── Product ─────────────────────────────────────────────────────────────────

/**
 * Generate SEO metadata for a product page.
 *
 * @param {object} product
 * @param {string} product.id
 * @param {string} product.name
 * @param {string} [product.description]
 * @param {number} [product.price]
 * @param {string} [product.currency]
 * @param {string[]} [product.images]
 * @param {string} [product.brand]
 * @returns {{ title, description, keywords, canonical, jsonLd }}
 */
function generateProductMeta(product) {
  const brandPart = product.brand ? ` | ${product.brand}` : '';
  const title = truncate(`${product.name}${brandPart} | ShopOS`, 70);

  const rawDesc = stripHtml(product.description || '');
  const pricePart =
    product.price !== undefined
      ? ` Starting at ${product.currency || 'USD'} ${product.price}.`
      : '';
  const description = truncate(
    rawDesc ? `${rawDesc}${pricePart}` : `Buy ${product.name} online at ShopOS.${pricePart}`,
    160
  );

  const keywords = [product.name, product.brand, 'buy online', 'ShopOS']
    .filter(Boolean)
    .join(', ');

  const canonical = `${BASE_URL}/products/${product.id}`;

  const jsonLd = buildProductJsonLd({ ...product, url: canonical });

  return { title, description, keywords, canonical, jsonLd };
}

// ─── Category ────────────────────────────────────────────────────────────────

/**
 * Generate SEO metadata for a category page.
 *
 * @param {object} category
 * @param {string} category.name
 * @param {string} category.slug
 * @param {string} [category.description]
 * @returns {{ title, description, canonical, jsonLd }}
 */
function generateCategoryMeta(category) {
  const title = truncate(`${category.name} | ShopOS`, 70);

  const rawDesc = stripHtml(category.description || '');
  const description = truncate(
    rawDesc || `Shop the best ${category.name} products at ShopOS.`,
    160
  );

  const canonical = `${BASE_URL}/categories/${category.slug}`;

  const jsonLd = buildBreadcrumbJsonLd([
    { name: 'Home', url: BASE_URL },
    { name: category.name, url: canonical },
  ]);

  return { title, description, canonical, jsonLd };
}

// ─── Brand ───────────────────────────────────────────────────────────────────

/**
 * Generate SEO metadata for a brand page.
 *
 * @param {object} brand
 * @param {string} brand.name
 * @param {string} brand.slug
 * @param {string} [brand.description]
 * @returns {{ title, description, canonical }}
 */
function generateBrandMeta(brand) {
  const title = truncate(`${brand.name} Products | ShopOS`, 70);

  const rawDesc = stripHtml(brand.description || '');
  const description = truncate(
    rawDesc || `Explore all ${brand.name} products available at ShopOS.`,
    160
  );

  const canonical = `${BASE_URL}/brands/${brand.slug}`;

  return { title, description, canonical };
}

// ─── Sitemap ─────────────────────────────────────────────────────────────────

/**
 * Generates a single <url> XML entry for a sitemap.
 *
 * @param {string} url       – absolute URL of the page
 * @param {number} priority  – sitemap priority (0.0 – 1.0)
 * @param {string} changefreq – sitemap changefreq value
 * @returns {string} XML string (not including the surrounding <urlset> tags)
 */
function generateSitemapEntry(url, priority = 0.5, changefreq = 'weekly') {
  const today = new Date().toISOString().split('T')[0];
  return (
    `  <url>\n` +
    `    <loc>${escapeXml(url)}</loc>\n` +
    `    <lastmod>${today}</lastmod>\n` +
    `    <changefreq>${escapeXml(changefreq)}</changefreq>\n` +
    `    <priority>${priority.toFixed(1)}</priority>\n` +
    `  </url>`
  );
}

/**
 * Escape special XML characters in a string.
 */
function escapeXml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

module.exports = {
  generateProductMeta,
  generateCategoryMeta,
  generateBrandMeta,
  generateSitemapEntry,
};
