'use strict';

const { Platform, createSocialOrder } = require('../models/social');

/**
 * SocialPlatformAdapter
 *
 * Handles platform-specific field-name transformations between ShopOS's canonical
 * product model and each social platform's catalog API schema.
 *
 * Each social platform has a different JSON schema for product uploads and
 * incoming orders.  This adapter is the single place where those mappings live.
 */
class SocialPlatformAdapter {
  /**
   * Returns all platforms supported by this adapter.
   *
   * @returns {string[]} Array of Platform enum values.
   */
  getSupportedPlatforms() {
    return Object.values(Platform);
  }

  /**
   * Returns the field-name mapping from ShopOS canonical names to platform-native names.
   *
   * @param {string} platform - One of the Platform enum values.
   * @returns {Object} Map of { shopOsField: platformField }.
   * @throws {Error} If the platform is unsupported.
   */
  getFieldMapping(platform) {
    switch (platform) {
      case Platform.INSTAGRAM:
        return {
          productId: 'id',
          title: 'title',
          description: 'description',
          price: 'price',
          currency: 'currency',
          imageUrl: 'image_url',
          link: 'link',
          availability: 'availability',
          sku: 'retailer_id',
        };

      case Platform.TIKTOK:
        return {
          productId: 'product_id',
          title: 'name',
          description: 'description',
          price: 'price',
          currency: 'currency',
          imageUrl: 'main_image',
          link: 'url',
          availability: 'status',
          sku: 'seller_sku',
        };

      case Platform.PINTEREST:
        return {
          productId: 'id',
          title: 'title',
          description: 'description',
          price: 'price',
          currency: 'currency',
          imageUrl: 'media.images.original.url',
          link: 'link',
          availability: 'availability',
          sku: 'catalog_item_id',
        };

      case Platform.FACEBOOK:
        return {
          productId: 'id',
          title: 'name',
          description: 'description',
          price: 'price',
          currency: 'currency',
          imageUrl: 'image_url',
          link: 'url',
          availability: 'availability',
          sku: 'retailer_id',
        };

      default:
        throw new Error(`Unsupported platform: ${platform}`);
    }
  }

  /**
   * Formats a ShopOS product for Instagram Shopping catalog upload.
   *
   * Instagram Product Catalog API expects flat JSON with specific field names.
   * Price must be a string in "X.XX USD" format for the API.
   *
   * @param {import('../models/social').ProductCatalogItem} product
   * @returns {object} Instagram-formatted product object.
   */
  formatForInstagram(product) {
    return {
      id: product.productId,
      title: product.title,
      description: product.description,
      price: `${product.price.toFixed(2)} ${product.currency}`,
      currency: product.currency,
      image_url: product.imageUrl,
      link: product.link,
      availability: product.availability,
      retailer_id: product.sku,
      condition: 'new',
      brand: '',
    };
  }

  /**
   * Formats a ShopOS product for TikTok Shop product catalog.
   *
   * TikTok expects price as a string and uses `name` instead of `title`.
   *
   * @param {import('../models/social').ProductCatalogItem} product
   * @returns {object} TikTok-formatted product object.
   */
  formatForTikTok(product) {
    return {
      product_id: product.productId,
      name: product.title,
      description: product.description,
      price: product.price.toFixed(2),
      currency: product.currency,
      main_image: product.imageUrl,
      url: product.link,
      status: product.availability === 'in stock' ? 'AVAILABLE' : 'SOLDOUT',
      seller_sku: product.sku,
    };
  }

  /**
   * Formats a ShopOS product for Pinterest catalog.
   *
   * Pinterest uses a nested media object for images and `catalog_item_id` for SKU.
   *
   * @param {import('../models/social').ProductCatalogItem} product
   * @returns {object} Pinterest-formatted product object.
   */
  formatForPinterest(product) {
    return {
      id: product.productId,
      title: product.title,
      description: product.description,
      price: product.price.toFixed(2),
      currency: product.currency,
      media: {
        images: {
          original: {
            url: product.imageUrl,
          },
        },
      },
      link: product.link,
      availability: product.availability,
      catalog_item_id: product.sku,
    };
  }

  /**
   * Formats a ShopOS product for Facebook Commerce catalog.
   *
   * Facebook uses `name` for the title and `url` for the product link.
   *
   * @param {import('../models/social').ProductCatalogItem} product
   * @returns {object} Facebook-formatted product object.
   */
  formatForFacebook(product) {
    return {
      id: product.productId,
      name: product.title,
      description: product.description,
      price: `${product.price.toFixed(2)} ${product.currency}`,
      currency: product.currency,
      image_url: product.imageUrl,
      url: product.link,
      availability: product.availability,
      retailer_id: product.sku,
      condition: 'new',
    };
  }

  /**
   * Parses a raw order payload from a social platform into a canonical SocialOrder.
   *
   * Each platform structures their order webhook differently; this method
   * normalises them all into a single SocialOrder shape.
   *
   * @param {string} platform    - One of the Platform enum values.
   * @param {object} rawOrder    - The raw order object received from the platform webhook.
   * @returns {import('../models/social').SocialOrder}
   * @throws {Error} If the platform is unsupported or rawOrder is missing required fields.
   */
  parseIncomingOrder(platform, rawOrder) {
    switch (platform) {
      case Platform.INSTAGRAM:
        return createSocialOrder({
          platformOrderId: rawOrder.order_id || rawOrder.id,
          shopOsOrderId: null,
          platform,
          customerId: rawOrder.buyer_id || rawOrder.user_id,
          items: (rawOrder.products || rawOrder.items || []).map((item) => ({
            productId: item.retailer_id || item.product_id,
            sku: item.retailer_id || '',
            quantity: item.quantity || 1,
            unitPrice: parseFloat(item.price) || 0,
          })),
          totalAmount: parseFloat(rawOrder.total_price || rawOrder.total_amount) || 0,
          status: rawOrder.status || 'CREATED',
        });

      case Platform.TIKTOK:
        return createSocialOrder({
          platformOrderId: rawOrder.order_id || rawOrder.orderId,
          shopOsOrderId: null,
          platform,
          customerId: rawOrder.buyer_uid || rawOrder.buyer_id,
          items: (rawOrder.item_list || rawOrder.items || []).map((item) => ({
            productId: item.product_id,
            sku: item.seller_sku || item.sku || '',
            quantity: item.quantity || 1,
            unitPrice: parseFloat(item.sale_price || item.price) || 0,
          })),
          totalAmount: parseFloat(rawOrder.payment_info?.total_amount || rawOrder.total_amount) || 0,
          status: rawOrder.order_status || rawOrder.status || 'CREATED',
        });

      case Platform.PINTEREST:
        return createSocialOrder({
          platformOrderId: rawOrder.checkout_id || rawOrder.order_id,
          shopOsOrderId: null,
          platform,
          customerId: rawOrder.customer_info?.email || rawOrder.customer_id || 'UNKNOWN',
          items: (rawOrder.line_items || rawOrder.items || []).map((item) => ({
            productId: item.catalog_item_id || item.product_id,
            sku: item.catalog_item_id || '',
            quantity: item.quantity || 1,
            unitPrice: parseFloat(item.price) || 0,
          })),
          totalAmount: parseFloat(rawOrder.total_price || rawOrder.total_amount) || 0,
          status: rawOrder.status || 'CREATED',
        });

      case Platform.FACEBOOK:
        return createSocialOrder({
          platformOrderId: rawOrder.id || rawOrder.order_id,
          shopOsOrderId: null,
          platform,
          customerId: rawOrder.buyer_details?.id || rawOrder.buyer_id || 'UNKNOWN',
          items: (rawOrder.items?.data || rawOrder.items || []).map((item) => ({
            productId: item.retailer_id || item.id,
            sku: item.retailer_id || '',
            quantity: item.quantity || 1,
            unitPrice: parseFloat(item.price_per_unit?.amount || item.price) || 0,
          })),
          totalAmount: parseFloat(
            rawOrder.estimated_payment_details?.total_amount?.amount || rawOrder.total_amount
          ) || 0,
          status: rawOrder.state || rawOrder.status || 'CREATED',
        });

      default:
        throw new Error(`Unsupported platform for order parsing: ${platform}`);
    }
  }

  /**
   * Formats a ShopOS product for any supported platform using a dispatch table.
   *
   * @param {string} platform
   * @param {import('../models/social').ProductCatalogItem} product
   * @returns {object} Platform-formatted product.
   * @throws {Error} If the platform is unsupported.
   */
  formatProduct(platform, product) {
    switch (platform) {
      case Platform.INSTAGRAM: return this.formatForInstagram(product);
      case Platform.TIKTOK:    return this.formatForTikTok(product);
      case Platform.PINTEREST: return this.formatForPinterest(product);
      case Platform.FACEBOOK:  return this.formatForFacebook(product);
      default:
        throw new Error(`Unsupported platform: ${platform}`);
    }
  }
}

module.exports = SocialPlatformAdapter;
