// Neo4j graph schema for ShopOS product recommendations and relationships

// ── Constraints ───────────────────────────────────────────────────────────────
CREATE CONSTRAINT product_id_unique IF NOT EXISTS
  FOR (p:Product) REQUIRE p.productId IS UNIQUE;

CREATE CONSTRAINT user_id_unique IF NOT EXISTS
  FOR (u:User) REQUIRE u.userId IS UNIQUE;

CREATE CONSTRAINT category_id_unique IF NOT EXISTS
  FOR (c:Category) REQUIRE c.categoryId IS UNIQUE;

CREATE CONSTRAINT brand_id_unique IF NOT EXISTS
  FOR (b:Brand) REQUIRE b.brandId IS UNIQUE;

CREATE CONSTRAINT order_id_unique IF NOT EXISTS
  FOR (o:Order) REQUIRE o.orderId IS UNIQUE;

// ── Indexes ───────────────────────────────────────────────────────────────────
CREATE INDEX product_tenant_idx IF NOT EXISTS FOR (p:Product) ON (p.tenantId);
CREATE INDEX user_tenant_idx    IF NOT EXISTS FOR (u:User)    ON (u.tenantId);
CREATE INDEX product_name_idx   IF NOT EXISTS FOR (p:Product) ON (p.name);
CREATE FULLTEXT INDEX product_search IF NOT EXISTS
  FOR (p:Product) ON EACH [p.name, p.description, p.tags];

// ── Node property definitions ─────────────────────────────────────────────────
// Product node
// (:Product { productId, name, description, price, inStock, tenantId, createdAt })

// User node
// (:User { userId, tenantId, segment, createdAt })

// Category node
// (:Category { categoryId, name, path, tenantId })

// Brand node
// (:Brand { brandId, name, tenantId })

// Order node
// (:Order { orderId, userId, tenantId, totalAmount, createdAt })

// ── Relationships ─────────────────────────────────────────────────────────────
// Product taxonomy
// (:Product)-[:IN_CATEGORY]->(:Category)
// (:Product)-[:MADE_BY]->(:Brand)
// (:Category)-[:PARENT_OF]->(:Category)

// User behaviour
// (:User)-[:VIEWED { viewedAt, durationSec }]->(:Product)
// (:User)-[:PURCHASED { purchasedAt, quantity, price }]->(:Product)
// (:User)-[:ADDED_TO_CART { addedAt }]->(:Product)
// (:User)-[:WISHLISTED { addedAt }]->(:Product)
// (:User)-[:REVIEWED { rating, reviewedAt }]->(:Product)
// (:User)-[:PLACED]->(:Order)

// Product co-occurrence (derived, updated by Flink)
// (:Product)-[:FREQUENTLY_BOUGHT_WITH { weight, support, confidence }]->(:Product)
// (:Product)-[:SIMILAR_TO { cosineSimilarity }]->(:Product)
// (:Product)-[:ALSO_VIEWED { weight }]->(:Product)

// Order contents
// (:Order)-[:CONTAINS { quantity, unitPrice }]->(:Product)

// ── Sample recommendation queries ─────────────────────────────────────────────
// Collaborative filtering — users who bought X also bought:
//   MATCH (p:Product {productId: $pid})<-[:PURCHASED]-(u:User)-[:PURCHASED]->(rec:Product)
//   WHERE rec.productId <> $pid AND rec.tenantId = $tid
//   RETURN rec, count(u) AS score ORDER BY score DESC LIMIT 10

// Personalised — based on user purchase history:
//   MATCH (u:User {userId: $uid})-[:PURCHASED]->(bought:Product)-[:FREQUENTLY_BOUGHT_WITH]->(rec:Product)
//   WHERE NOT (u)-[:PURCHASED]->(rec) AND rec.tenantId = $tid
//   RETURN rec, sum(r.weight) AS score ORDER BY score DESC LIMIT 20

// Category affinities:
//   MATCH (u:User {userId: $uid})-[:PURCHASED]->(p:Product)-[:IN_CATEGORY]->(c:Category)<-[:IN_CATEGORY]-(rec:Product)
//   WHERE NOT (u)-[:PURCHASED]->(rec)
//   RETURN rec, count(*) AS catScore ORDER BY catScore DESC LIMIT 20
