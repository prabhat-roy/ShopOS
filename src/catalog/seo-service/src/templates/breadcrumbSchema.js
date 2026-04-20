'use strict';

/**
 * Generates a schema.org/BreadcrumbList JSON-LD object.
 *
 * @param {Array<{name: string, url: string}>} items – ordered list of breadcrumb steps
 * @returns {object} JSON-LD object (not stringified)
 */
function buildBreadcrumbJsonLd(items) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: item.url,
    })),
  };
}

module.exports = { buildBreadcrumbJsonLd };
