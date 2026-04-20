'use strict';

const request = require('supertest');
const { createApp } = require('../src/app');
const CatalogSyncService = require('../src/services/CatalogSyncService');
const SocialPlatformAdapter = require('../src/services/SocialPlatformAdapter');
const { Platform } = require('../src/models/social');

const app = createApp();

// ---- Shared fixtures ----

const sampleProduct = {
  productId: 'prod-001',
  sku: 'SKU-001',
  title: 'Wireless Headphones',
  description: 'Premium noise-cancelling headphones.',
  price: 149.99,
  currency: 'USD',
  imageUrl: 'https://cdn.example.com/images/prod-001.jpg',
  link: 'https://shop.example.com/products/prod-001',
  availability: 'in stock',
};

const sampleInstagramOrder = {
  order_id: 'ig-order-001',
  buyer_id: 'ig-user-789',
  products: [
    { retailer_id: 'SKU-001', quantity: 2, price: '149.99' },
  ],
  total_price: '299.98',
  status: 'CREATED',
};

const sampleTikTokOrder = {
  order_id: 'tt-order-001',
  buyer_uid: 'tt-user-123',
  item_list: [
    { product_id: 'prod-001', seller_sku: 'SKU-001', quantity: 1, sale_price: '149.99' },
  ],
  total_amount: '149.99',
  order_status: 'AWAITING_SHIPMENT',
};

// ============================================================================
// 1. syncCatalog — Instagram
// ============================================================================
describe('CatalogSyncService.syncCatalog', () => {
  let service;

  beforeEach(() => {
    service = new CatalogSyncService();
  });

  test('syncCatalog Instagram — returns SUCCESS with correct itemsSynced count', () => {
    const result = service.syncCatalog(Platform.INSTAGRAM, [sampleProduct]);

    expect(result.syncId).toBeDefined();
    expect(result.platform).toBe(Platform.INSTAGRAM);
    expect(result.itemsSynced).toBe(1);
    expect(result.errors).toHaveLength(0);
    expect(result.completedAt).toBeDefined();
  });

  // --------------------------------------------------------------------------
  // 2. syncCatalog — TikTok
  // --------------------------------------------------------------------------
  test('syncCatalog TikTok — handles multiple products', () => {
    const products = [
      sampleProduct,
      { ...sampleProduct, productId: 'prod-002', sku: 'SKU-002', title: 'Earbuds' },
    ];
    const result = service.syncCatalog(Platform.TIKTOK, products);

    expect(result.platform).toBe(Platform.TIKTOK);
    expect(result.itemsSynced).toBe(2);
    expect(result.errors).toHaveLength(0);
  });
});

// ============================================================================
// 3. formatForInstagram — correct field names
// ============================================================================
describe('SocialPlatformAdapter.formatForInstagram', () => {
  test('produces all required Instagram fields', () => {
    const adapter = new SocialPlatformAdapter();
    const { createProductCatalogItem } = require('../src/models/social');
    const item = createProductCatalogItem(sampleProduct);
    const formatted = adapter.formatForInstagram(item);

    expect(formatted).toHaveProperty('id', 'prod-001');
    expect(formatted).toHaveProperty('title', 'Wireless Headphones');
    expect(formatted).toHaveProperty('image_url', sampleProduct.imageUrl);
    expect(formatted).toHaveProperty('link', sampleProduct.link);
    expect(formatted).toHaveProperty('availability', 'in stock');
    expect(formatted).toHaveProperty('retailer_id', 'SKU-001');
    // price should be a formatted string
    expect(formatted.price).toMatch(/^\d+\.\d{2}\s[A-Z]{3}$/);
  });
});

// ============================================================================
// 4. formatForTikTok — correct field names
// ============================================================================
describe('SocialPlatformAdapter.formatForTikTok', () => {
  test('produces all required TikTok fields', () => {
    const adapter = new SocialPlatformAdapter();
    const { createProductCatalogItem } = require('../src/models/social');
    const item = createProductCatalogItem(sampleProduct);
    const formatted = adapter.formatForTikTok(item);

    expect(formatted).toHaveProperty('product_id', 'prod-001');
    expect(formatted).toHaveProperty('name', 'Wireless Headphones');
    expect(formatted).toHaveProperty('main_image', sampleProduct.imageUrl);
    expect(formatted).toHaveProperty('url', sampleProduct.link);
    expect(formatted).toHaveProperty('status', 'AVAILABLE');
    expect(formatted).toHaveProperty('seller_sku', 'SKU-001');
  });
});

// ============================================================================
// 5. parsePlatformOrder — Instagram
// ============================================================================
describe('CatalogSyncService.parsePlatformOrder', () => {
  let service;

  beforeEach(() => {
    service = new CatalogSyncService();
  });

  test('parsePlatformOrder Instagram — maps to canonical SocialOrder', () => {
    const order = service.parsePlatformOrder(Platform.INSTAGRAM, sampleInstagramOrder);

    expect(order.platformOrderId).toBe('ig-order-001');
    expect(order.platform).toBe(Platform.INSTAGRAM);
    expect(order.customerId).toBe('ig-user-789');
    expect(order.items).toHaveLength(1);
    expect(order.totalAmount).toBe(299.98);
    expect(order.shopOsOrderId).toBeNull();
  });

  // --------------------------------------------------------------------------
  // 6. parsePlatformOrder — TikTok
  // --------------------------------------------------------------------------
  test('parsePlatformOrder TikTok — maps item_list correctly', () => {
    const order = service.parsePlatformOrder(Platform.TIKTOK, sampleTikTokOrder);

    expect(order.platformOrderId).toBe('tt-order-001');
    expect(order.platform).toBe(Platform.TIKTOK);
    expect(order.customerId).toBe('tt-user-123');
    expect(order.items).toHaveLength(1);
    expect(order.items[0].sku).toBe('SKU-001');
    expect(order.status).toBe('AWAITING_SHIPMENT');
  });
});

// ============================================================================
// 7. getSupportedPlatforms
// ============================================================================
describe('CatalogSyncService.getSupportedPlatforms', () => {
  test('returns all four platforms', () => {
    const service = new CatalogSyncService();
    const platforms = service.getSupportedPlatforms();

    expect(platforms).toContain(Platform.INSTAGRAM);
    expect(platforms).toContain(Platform.TIKTOK);
    expect(platforms).toContain(Platform.PINTEREST);
    expect(platforms).toContain(Platform.FACEBOOK);
    expect(platforms).toHaveLength(4);
  });
});

// ============================================================================
// 8. getRecentSyncs — limit is respected
// ============================================================================
describe('CatalogSyncService.getRecentSyncs', () => {
  test('limit caps the number of results returned', () => {
    const service = new CatalogSyncService();
    for (let i = 0; i < 6; i++) {
      service.syncCatalog(Platform.INSTAGRAM, [sampleProduct]);
    }

    const results = service.getRecentSyncs(Platform.INSTAGRAM, 4);
    expect(results).toHaveLength(4);
  });
});

// ============================================================================
// 9. GET /social/field-mappings/:platform — HTTP endpoint
// ============================================================================
describe('GET /social/field-mappings/:platform', () => {
  test('returns Instagram field mapping with correct key names', async () => {
    const res = await request(app).get('/social/field-mappings/INSTAGRAM');

    expect(res.status).toBe(200);
    expect(res.body.platform).toBe('INSTAGRAM');
    expect(res.body.fieldMapping).toHaveProperty('imageUrl', 'image_url');
    expect(res.body.fieldMapping).toHaveProperty('productId', 'id');
  });
});

// ============================================================================
// 10. GET /healthz
// ============================================================================
describe('GET /healthz', () => {
  test('returns 200 with status ok', async () => {
    const res = await request(app).get('/healthz');

    expect(res.status).toBe(200);
    expect(res.body).toEqual({ status: 'ok' });
  });
});

// ============================================================================
// 11. POST /social/catalog/sync — invalid platform returns 400
// ============================================================================
describe('POST /social/catalog/sync — invalid platform', () => {
  test('returns 400 for unknown platform value', async () => {
    const res = await request(app)
      .post('/social/catalog/sync')
      .send({ platform: 'SNAPCHAT', products: [sampleProduct] });

    expect(res.status).toBe(400);
    expect(res.body.error).toBe('BAD_REQUEST');
    expect(res.body.message).toMatch(/Unsupported platform/i);
  });
});

// ============================================================================
// 12. GET /social/platforms — lists all platforms
// ============================================================================
describe('GET /social/platforms', () => {
  test('returns the four supported platform names', async () => {
    const res = await request(app).get('/social/platforms');

    expect(res.status).toBe(200);
    expect(res.body.platforms).toContain('INSTAGRAM');
    expect(res.body.platforms).toContain('TIKTOK');
    expect(res.body.platforms).toContain('PINTEREST');
    expect(res.body.platforms).toContain('FACEBOOK');
  });
});
