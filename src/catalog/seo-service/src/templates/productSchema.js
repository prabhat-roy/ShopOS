'use strict';

/**
 * Generates a schema.org/Product JSON-LD object.
 *
 * @param {object} product
 * @param {string} product.id
 * @param {string} product.name
 * @param {string} [product.description]
 * @param {number} [product.price]
 * @param {string} [product.currency]
 * @param {string[]} [product.images]
 * @param {string} [product.brand]
 * @param {string} product.url  – canonical URL for this product page
 * @returns {object} JSON-LD object (not stringified)
 */
function buildProductJsonLd(product) {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: product.name,
    url: product.url,
  };

  if (product.description) {
    jsonLd.description = product.description;
  }

  if (product.images && product.images.length > 0) {
    jsonLd.image = product.images.length === 1 ? product.images[0] : product.images;
  }

  if (product.brand) {
    jsonLd.brand = {
      '@type': 'Brand',
      name: product.brand,
    };
  }

  if (product.price !== undefined && product.price !== null) {
    jsonLd.offers = {
      '@type': 'Offer',
      price: product.price,
      priceCurrency: product.currency || 'USD',
      availability: 'https://schema.org/InStock',
      url: product.url,
    };
  }

  return jsonLd;
}

module.exports = { buildProductJsonLd };
