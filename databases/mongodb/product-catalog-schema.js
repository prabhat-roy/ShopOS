/**
 * MongoDB schema documentation for product-catalog-service
 * Run this to set up validators and indexes on the shopos_catalog database.
 *
 * Usage: mongosh shopos_catalog product-catalog-schema.js
 */

// Products collection
db.createCollection("products", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["name", "sku", "price", "currency", "status"],
      properties: {
        _id: { bsonType: "objectId" },
        name: { bsonType: "string", minLength: 1, maxLength: 500 },
        slug: { bsonType: "string" },
        description: { bsonType: "string" },
        short_description: { bsonType: "string", maxLength: 500 },
        sku: { bsonType: "string" },
        price: { bsonType: "decimal", minimum: 0 },
        compare_at_price: { bsonType: "decimal", minimum: 0 },
        currency: { bsonType: "string", minLength: 3, maxLength: 3 },
        cost_price: { bsonType: "decimal", minimum: 0 },
        category_id: { bsonType: "string" },
        category_path: {
          bsonType: "array",
          items: { bsonType: "string" },
        },
        brand_id: { bsonType: "string" },
        brand_name: { bsonType: "string" },
        status: {
          bsonType: "string",
          enum: ["draft", "active", "archived", "discontinued"],
        },
        stock_status: {
          bsonType: "string",
          enum: ["in_stock", "out_of_stock", "backorder", "preorder"],
        },
        stock_quantity: { bsonType: "int", minimum: 0 },
        weight_grams: { bsonType: "int", minimum: 0 },
        dimensions: {
          bsonType: "object",
          properties: {
            length_cm: { bsonType: "decimal" },
            width_cm: { bsonType: "decimal" },
            height_cm: { bsonType: "decimal" },
          },
        },
        images: {
          bsonType: "array",
          items: {
            bsonType: "object",
            required: ["url"],
            properties: {
              url: { bsonType: "string" },
              alt: { bsonType: "string" },
              position: { bsonType: "int" },
              is_primary: { bsonType: "bool" },
            },
          },
        },
        variants: {
          bsonType: "array",
          items: {
            bsonType: "object",
            required: ["sku"],
            properties: {
              sku: { bsonType: "string" },
              attributes: { bsonType: "object" },
              price: { bsonType: "decimal" },
              stock_quantity: { bsonType: "int" },
            },
          },
        },
        attributes: { bsonType: "object" },
        tags: { bsonType: "array", items: { bsonType: "string" } },
        seo: {
          bsonType: "object",
          properties: {
            title: { bsonType: "string" },
            description: { bsonType: "string" },
            keywords: { bsonType: "array", items: { bsonType: "string" } },
          },
        },
        ratings: {
          bsonType: "object",
          properties: {
            average: { bsonType: "decimal", minimum: 0, maximum: 5 },
            count: { bsonType: "int", minimum: 0 },
          },
        },
        is_digital: { bsonType: "bool" },
        requires_shipping: { bsonType: "bool" },
        seller_id: { bsonType: "string" },
        created_at: { bsonType: "date" },
        updated_at: { bsonType: "date" },
      },
    },
  },
});

// Indexes
db.products.createIndex({ sku: 1 }, { unique: true });
db.products.createIndex({ slug: 1 }, { unique: true, sparse: true });
db.products.createIndex({ category_id: 1, status: 1 });
db.products.createIndex({ brand_id: 1, status: 1 });
db.products.createIndex({ status: 1, created_at: -1 });
db.products.createIndex({ seller_id: 1, status: 1 });
db.products.createIndex({ "ratings.average": -1, status: 1 });
db.products.createIndex({ tags: 1 });
db.products.createIndex(
  { name: "text", description: "text", brand_name: "text", tags: "text" },
  { weights: { name: 10, brand_name: 5, tags: 3, description: 1 } }
);

print("Products collection created with validators and indexes");

// Reviews collection (belongs to catalog, not customer-experience DB)
db.createCollection("reviews", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["product_id", "user_id", "rating"],
      properties: {
        product_id: { bsonType: "string" },
        user_id: { bsonType: "string" },
        order_id: { bsonType: "string" },
        rating: { bsonType: "int", minimum: 1, maximum: 5 },
        title: { bsonType: "string", maxLength: 200 },
        body: { bsonType: "string", maxLength: 5000 },
        verified_purchase: { bsonType: "bool" },
        status: {
          bsonType: "string",
          enum: ["pending", "approved", "rejected", "flagged"],
        },
        helpful_votes: { bsonType: "int", minimum: 0 },
        images: { bsonType: "array", items: { bsonType: "string" } },
        created_at: { bsonType: "date" },
        updated_at: { bsonType: "date" },
      },
    },
  },
});

db.reviews.createIndex({ product_id: 1, status: 1, created_at: -1 });
db.reviews.createIndex({ user_id: 1, created_at: -1 });
db.reviews.createIndex({ product_id: 1, user_id: 1 }, { unique: true });
db.reviews.createIndex(
  { title: "text", body: "text" },
  { weights: { title: 5, body: 1 } }
);

print("Reviews collection created");
print("Catalog schema setup complete.");
