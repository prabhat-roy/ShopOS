'use strict';

/**
 * Platform enum — the social commerce platforms supported by this adapter.
 * @readonly
 * @enum {string}
 */
const Platform = Object.freeze({
  INSTAGRAM: 'INSTAGRAM',
  TIKTOK: 'TIKTOK',
  PINTEREST: 'PINTEREST',
  FACEBOOK: 'FACEBOOK',
});

/**
 * SyncStatus enum — lifecycle states of a catalog synchronisation operation.
 * @readonly
 * @enum {string}
 */
const SyncStatus = Object.freeze({
  PENDING: 'PENDING',
  SUCCESS: 'SUCCESS',
  FAILED: 'FAILED',
});

/**
 * Factory function for a ProductCatalogItem.
 * Represents a single product entry to be pushed to a social platform's catalog.
 *
 * @param {object} params
 * @param {string} params.productId      ShopOS product identifier.
 * @param {string} params.sku            Stock-keeping unit code.
 * @param {string} params.title          Product title / display name.
 * @param {string} params.description    Product description.
 * @param {number} params.price          Unit price as a number.
 * @param {string} params.currency       ISO 4217 currency code (e.g. "USD").
 * @param {string} params.imageUrl       Publicly accessible product image URL.
 * @param {string} params.link           Deep link to the product page.
 * @param {string} params.availability   Stock status: "in stock" | "out of stock" | "preorder".
 * @returns {ProductCatalogItem}
 */
function createProductCatalogItem({
  productId,
  sku,
  title,
  description,
  price,
  currency,
  imageUrl,
  link,
  availability,
}) {
  if (!productId) throw new Error('productId is required');
  if (!sku) throw new Error('sku is required');
  if (!title) throw new Error('title is required');

  return Object.freeze({
    productId,
    sku,
    title,
    description: description || '',
    price: typeof price === 'number' ? price : parseFloat(price) || 0,
    currency: currency || 'USD',
    imageUrl: imageUrl || '',
    link: link || '',
    availability: availability || 'in stock',
  });
}

/**
 * Factory function for a SocialOrder.
 * Represents an order received from a social commerce platform.
 *
 * @param {object} params
 * @param {string}   params.platformOrderId  Order identifier assigned by the social platform.
 * @param {string}   [params.shopOsOrderId]  ShopOS order identifier after mapping (may be null initially).
 * @param {string}   params.platform         One of the Platform enum values.
 * @param {string}   params.customerId       Platform-side customer/buyer identifier.
 * @param {object[]} params.items            Array of line items { productId, sku, quantity, unitPrice }.
 * @param {number}   params.totalAmount      Order total.
 * @param {string}   params.status           Raw platform order status string.
 * @returns {SocialOrder}
 */
function createSocialOrder({
  platformOrderId,
  shopOsOrderId,
  platform,
  customerId,
  items,
  totalAmount,
  status,
}) {
  if (!platformOrderId) throw new Error('platformOrderId is required');
  if (!Platform[platform]) throw new Error(`Unknown platform: ${platform}`);

  return Object.freeze({
    platformOrderId,
    shopOsOrderId: shopOsOrderId || null,
    platform,
    customerId: customerId || 'UNKNOWN',
    items: Array.isArray(items) ? items : [],
    totalAmount: typeof totalAmount === 'number' ? totalAmount : parseFloat(totalAmount) || 0,
    status: status || 'CREATED',
  });
}

module.exports = {
  Platform,
  SyncStatus,
  createProductCatalogItem,
  createSocialOrder,
};
