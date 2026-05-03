# ShopOS AI/ML & Autonomous Agent Operations — Complete Architecture Document

Scope: Architecture, implementation process, and technology decisions only.

---

## 1. Executive Summary

ShopOS is a 154-microservice, 13-domain, 8-language enterprise commerce platform with a mature DevSecOps stack (50+ security tools, 15 CI/CD platforms, full GitOps). This document extends it with three layers:

Layer 1 — AI/ML integration across all business domains. 65 use cases spanning every domain from semantic search to fraud detection to autonomous customer support. Each use case identifies the ML discipline, model type, data sources, affected services, and implementation approach.

Layer 2 — Self-hosted model infrastructure. All LLM and ML models run on Kubernetes using open-source serving stacks. Zero paid subscriptions. Full data sovereignty. A tiered model architecture serves different workload types from flagship reasoning to fast classification to embeddings.

Layer 3 — Autonomous agent operations. 238 specialized AI agents organized into four divisions (Developer, DevOps, DevSecOps, AI) that manage the entire ShopOS lifecycle: writing code, deploying infrastructure, monitoring systems, responding to incidents, managing ML models, and documenting everything. Each agent is a deep specialist in exactly one domain, tool, language, or service.

### Key numbers

| Dimension | Count |
|---|---|
| AI/ML use cases | 65 across 13 domains |
| Services directly enhanced by ML | 45 of 154 (29%) |
| Self-hosted LLM/ML models | 11 models across 5 tiers |
| GPU nodes | 7 (15 GPUs total) |
| Autonomous agents | 238 (stabilizes to 120–150 after tuning) |
| Agent divisions | 4 (Developer, DevOps, DevSecOps, AI) |
| Paid subscriptions required | 0 |
| Monthly infrastructure cost | $3,100–5,500 |
| Implementation timeline | 8 months |

---

## 2. AI/ML Integration — All 13 Domains

### 2.1 Platform Domain (22 services, 5 AI/ML use cases)

Use case 2.1.1 — Intelligent adaptive rate limiting. The existing `rate-limiter-service` uses static Redis counters. ML replaces this with an XGBoost classifier that scores each incoming request as legitimate, bot, or abuser based on features including request rate in a sliding window, endpoint access entropy (how randomly the caller hits different paths), device fingerprint trust score from `device-fingerprint-service`, geographic anomaly (is this IP in a country the user has never logged in from), session age, and TLS JA3/JA4 fingerprint. The classifier outputs a risk score from 0 to 1 which maps to three rate limit tiers: generous (score below 0.3), standard (0.3 to 0.7), and aggressive throttle (above 0.7). The model trains on historical access logs labeled as bot, legitimate, or abuser. It runs as a lightweight Python sidecar to `rate-limiter-service`, loading a pre-trained ONNX model from MLflow for sub-10ms scoring.

Use case 2.1.2 — Predictive autoscaling. KEDA currently scales on Kafka consumer lag and CPU utilization. An ML model predicts traffic spikes 15 to 30 minutes before they arrive. The model is a time-series forecaster (Prophet or NeuralProphet) trained on historical Prometheus `istio_requests_total` metrics combined with event data from `scheduler-service` (knows about scheduled events), `flash-sale-service` (knows about upcoming flash sales), and an external marketing campaign calendar. The forecaster outputs predicted requests-per-second per service for the next 30 minutes. This prediction feeds into a KEDA external scaler that pre-provisions pods before the traffic arrives rather than reacting after latency has already spiked. Implementation is a CronJob running every 5 minutes that pushes predicted metrics to a Prometheus pushgateway.

Use case 2.1.3 — Smart dead letter queue triage. The `dead-letter-service` currently collects failed messages for manual replay. A text classifier (TF-IDF plus logistic regression, or a small transformer) auto-classifies DLQ messages by failure type: transient infrastructure errors (retry immediately), permanent data errors (route to data quality team), schema mismatches (route to Proto/Avro Agent), and unknown errors (escalate to human). The classifier trains on historical DLQ messages labeled by their eventual resolution type.

Use case 2.1.4 — Anomaly detection for service health. The `health-check-service` currently performs binary healthy/unhealthy checks. An Isolation Forest or LSTM autoencoder trained on normal service behavior metrics (response time distribution, error rate, memory growth pattern, garbage collection frequency) detects "degrading" patterns where a service is technically healthy but trending toward failure. This allows pre-emptive action (restart, scale, alert) before the service actually crashes.

Use case 2.1.5 — GraphQL query cost prediction. The `graphql-gateway` aggregates queries across domains. A gradient-boosted regression model predicts query latency from query structural features (depth, field count, number of involved downstream services, presence of nested lists) and pre-warms caches via `cache-warming-service` for queries predicted to be expensive. This reduces p99 latency for complex queries.

### 2.2 Identity Domain (8 services, 4 AI/ML use cases)

Use case 2.2.1 — Login anomaly detection with adaptive MFA step-up. This is the most security-critical ML application. Every login attempt is scored by a risk model before authentication completes. The model evaluates: is this a known device (from `device-fingerprint-service`)? Is the geolocation consistent with the user's history? Is it a normal login time for this user? How many login attempts have occurred recently from this IP? Is the browser/OS fingerprint consistent? Is the IP residential, datacenter, or a known Tor exit node? The model is a dual approach: an XGBoost binary classifier trained on successful logins as the positive class with engineered features, plus an Isolation Forest for unsupervised anomaly detection to catch novel attack patterns. The risk score determines the authentication flow: low risk (score 0 to 0.3) issues a token directly, medium risk (0.3 to 0.7) triggers MFA via `mfa-service`, high risk (0.7 to 0.9) requires MFA plus email confirmation, and very high risk (above 0.9) blocks the attempt and notifies the user. This runs as a Python sidecar on `auth-service` or as a dedicated `login-risk-service` using a pre-loaded ONNX model for sub-10ms scoring.

Use case 2.2.2 — Bot detection on registration. A gradient-boosted classifier detects bot registrations using features including registration velocity per IP, form fill timing (bots fill forms instantly while humans take 5 to 30 seconds), mouse movement entropy when available from the frontend, email domain reputation, and username pattern analysis (random strings versus human-like names).

Use case 2.2.3 — Behavioral biometrics for continuous authentication. After login, a Siamese neural network continuously compares the current session's behavioral patterns (typing cadence, mouse movement, scrolling behavior) to the enrolled user's behavioral profile. If behavior shifts mid-session (indicating possible session hijacking), the system triggers a re-authentication challenge. This model is trained per-user, with the enrollment profile built over the first few sessions.

Use case 2.2.4 — GDPR request auto-classification. The `gdpr-service` receives data subject requests via free-text submissions. A text classifier (fine-tuned BERT or TF-IDF plus SVM) auto-classifies request type (access, erasure, portability, rectification, objection) and routes to the correct automated workflow, eliminating manual triage.

### 2.3 Catalog Domain (12 services, 8 AI/ML use cases)

Use case 2.3.1 — Semantic product search (hybrid BM25 plus vector). This is the highest-impact catalog use case. The existing `search-service` uses Elasticsearch BM25 keyword matching only. The enhancement adds semantic vector search so a query like "something to keep my coffee hot" returns thermoses and insulated mugs even though those words never appear in the query. The architecture is hybrid: the user's query goes to both Elasticsearch (BM25 keyword match producing results ranked by keyword overlap) and to a sentence-transformer embedding model that converts the query to a 768-dimensional vector and searches Weaviate (nearVector search producing results ranked by semantic similarity). The two result sets are merged using Reciprocal Rank Fusion or a cross-encoder reranker into a single ranked list. All product titles and descriptions are pre-embedded into Weaviate via a batch indexing pipeline running as a Flink job or Argo Workflow. The embedding model is BGE-Large-en-v1.5 (self-hosted via TEI on a T4 GPU). The cross-encoder reranker is FlashRank or a fine-tuned cross-encoder model.

Use case 2.3.2 — Visual product search (search by image). Users upload a photo and the system finds visually similar products. The uploaded image goes through `image-processing-service` for resizing and normalization, then through a CLIP model (self-hosted) that encodes it into a 512-dimensional vector. This vector is searched against product image embeddings pre-indexed in Weaviate. All catalog product images are pre-encoded through the same CLIP model during a batch indexing pipeline. This enables "I saw this chair at a friend's house, find me something similar" experiences.

Use case 2.3.3 — Auto-categorization of new products. When a merchant uploads a new product with a title like "Nike Air Max 90 Men's Running Shoe Size 10" and a description, an NLP classifier extracts category (Shoes, Running Shoes, Men's), brand (Nike), attributes (color, size 10, type running), and tags (lightweight, cushioned, running). The classifier is a fine-tuned BERT for multi-label text classification plus named entity recognition for attribute extraction. An image classifier (ResNet or EfficientNet) serves as a fallback when text descriptions are sparse, classifying product category from the product images alone.

Use case 2.3.4 — Dynamic pricing intelligence. The `pricing-service` currently applies manual pricing rules. ML adds a price elasticity model that estimates the demand curve for each product based on historical sales at different price points (from ClickHouse analytics), current inventory levels (from `inventory-service`), competitor pricing (from `marketplace-connector-service` data), and time features (day of week, seasonality, holiday proximity). An optimization layer then finds the price that maximizes revenue (or margin, or units — configurable business objective) subject to constraints including minimum margin percentage, maximum price change per day, and competitor price ceiling. The model is XGBoost regression for demand estimation plus scipy constrained optimization for price selection. Suggested prices flow to `pricing-service` where they can be applied automatically or surfaced for human override.

Use case 2.3.5 — Smart inventory replenishment. The `inventory-service` tracks stock levels but relies on manual reorder points. A Prophet time-series forecast per SKU predicts when each product will run out factoring in seasonality, active promotions, and trend. The system recommends reorder quantities and timing, feeding directly into `demand-forecast-service` and `purchase-order-service`.

Use case 2.3.6 — Product similarity and recommendations. A hybrid approach combining Neo4j graph traversal (co-purchase edges: "customers who bought X also bought Y") with Weaviate vector similarity (products with similar embeddings) powers "similar products," "customers also viewed," and "frequently bought together" experiences. The recommendation engine in `recommendation-service` merges both signals, weighted by the user's browsing context.

Use case 2.3.7 — Auto-generated product descriptions. An LLM generates or improves product descriptions from structured attributes. This is especially valuable for marketplace sellers with sparse or low-quality listings. The LLM (Llama 70B, self-hosted) takes product title, specifications, brand, and category as input and produces SEO-optimized descriptions. This runs as a batch job or on-demand when a merchant requests description generation.

Use case 2.3.8 — Size and fit recommendation. ML learns from return reasons and review text to recommend sizes, reducing returns. A collaborative filtering model on user-size-item triples combined with NLP extraction of size mentions from reviews ("this shirt runs small," "order one size up") produces size recommendations displayed on product pages. The model is trained on historical order and return data from `return-refund-service` and review text from `review-rating-service`.

### 2.4 Commerce Domain (23 services, 9 AI/ML use cases)

Use case 2.4.1 — Transaction-level fraud detection. Every order is scored for fraud risk before payment capture. The `fraud-detection-service` evaluates features including order amount versus the user's historical average, shipping address not matching billing address, new account combined with high-value order, device fingerprint trust score, payment method risk (prepaid cards score higher), velocity (orders per hour from this IP or device), whether the shipping address is a known freight forwarder, product mix (high-resale-value items like electronics score higher), and session behavior (time on site before purchase — bots often have very short sessions). The model is an XGBoost ensemble combined with a rule engine for known fraud patterns. Training data comes from historical chargebacks (chargeback equals fraud) with SMOTE or class weighting to handle the severe class imbalance (fraud is typically less than 1% of transactions). The model outputs a risk score: low risk triggers auto-approve, medium risk holds for a human review queue, and high risk triggers auto-decline plus account flagging.

Use case 2.4.2 — Cart abandonment prediction and recovery. A Flink streaming model scores active carts in real-time every 60 seconds. Features include cart age (minutes since first item added), items in cart and total cart value, page views since last cart interaction, user's historical conversion rate, price sensitivity signals (has the user visited comparison sites), and stock pressure (are limited-stock items in the cart). When the model predicts abandonment probability above 0.7 and the user hasn't started checkout in 10 minutes, it triggers the `notification-orchestrator` to send a personalized recovery message: an email with "Your cart is waiting" plus a targeted 10% discount from `promotions-service`, a push notification with "Items in your cart are selling fast," or an on-site banner offering free shipping. The model is a logistic regression or LightGBM on streaming features computed in Flink.

Use case 2.4.3 — Delivery time estimation. Accurate "arrives by" dates on product pages and checkout improve trust and conversion. A gradient-boosted regression model trained on historical shipment data predicts actual delivery days from features including origin warehouse, destination zip code, carrier, package weight, day of week, and season. The model replaces the current static carrier-provided estimates which are often inaccurate.

Use case 2.4.4 — Buy-now-pay-later credit scoring. The `bnpl-service` needs instant credit decisions. An ML model scores creditworthiness from in-platform behavioral signals without requiring a credit bureau pull. Features include account age, order history, return rate, payment history (late payments on previous BNPL), CLV score from `clv-service`, and current cart composition. The model is logistic regression (required for interpretability in lending decisions) with SHAP explanations attached to each decision for regulatory compliance.

Use case 2.4.5 — Smart coupon and promotion targeting. Instead of blanket discounts that waste margin on customers who would have purchased anyway, a causal uplift model identifies which users need a discount to convert (persuadable) versus which would buy regardless (sure things). For each user the model predicts probability of purchase without a discount and probability of purchase with a 10% discount. If the uplift (the difference) is significant, the user sees the discount. If the user would purchase anyway, the margin is preserved. Implementation uses the CausalML library or a custom two-model approach integrated with `promotions-service` and `ab-testing-service`.

Use case 2.4.6 — Cross-sell and upsell at checkout. "Customers who bought X also bought Y" but personalized to the specific user. A graph neural network on the Neo4j co-purchase graph combined with user embedding from Weaviate predicts which add-on product the specific user is most likely to want given their current cart and history. A simpler alternative is association rules (Apriori algorithm) reranked by a user preference model.

Use case 2.4.7 — Returns prediction. At checkout, a LightGBM classifier predicts whether an order is likely to be returned. Features include product category return rate, user's personal return history, size confidence (did the user use size recommendations), price point, whether it's a multi-item order, and whether it's marked as a gift. High-return-risk orders can be flagged for operational planning (for example, ship from the nearest warehouse to reduce return shipping cost rather than from a cheaper but distant fulfillment center).

Use case 2.4.8 — Dynamic shipping option ranking. Shipping options at checkout are ordered by predicted user preference rather than just price. A learning-to-rank model (XGBoost ranking) trained on the user's historical shipping choices learns that some users always pick the fastest option while others always pick the cheapest. The user's likely preference becomes the default selection.

Use case 2.4.9 — Subscription churn prediction. For `subscription-billing-service`, a survival analysis model (Lifetimes library) or gradient-boosted classifier predicts which subscribers will cancel. Features include usage frequency trend, support ticket count, payment failure count, engagement decline over time, and time since last non-subscription purchase. High-churn-risk subscribers trigger retention offers via `notification-orchestrator` before they cancel.

### 2.5 Supply Chain Domain (13 services, 6 AI/ML use cases)

Use case 2.5.1 — Demand forecasting. This is the most impactful supply chain use case. The `demand-forecast-service` predicts demand per SKU per warehouse per week, driving purchasing decisions, warehouse staffing, and inventory placement. The model is a hierarchical time-series forecaster: aggregate forecasts at the category level (more data, more stable) are reconciled down to SKU level. Inputs include historical sales data from ClickHouse, the promotional calendar, seasonal patterns, external signals (weather, holidays), and web traffic trends from the analytics domain. The recommended approach is LightGBM or NeuralProphet. For cold-start SKUs (new products with no history), the model uses category-level analogues — predicting demand based on similar products in the same category.

Use case 2.5.2 — Warehouse slotting optimization. Products frequently ordered together should be stored near each other in the warehouse for faster picking. K-means clustering on the order co-occurrence matrix identifies product groupings, and a linear programming optimizer assigns products to warehouse bins to minimize average pick path length.

Use case 2.5.3 — Carrier selection optimization. For each shipment, a multi-armed bandit (Vowpal Wabbit) selects the optimal carrier by exploring options and exploiting the best-performing carrier per route. The optimization target is not just cost but "promised delivery date met" — reliability weighted alongside price and speed.

Use case 2.5.4 — Supplier risk scoring. The `supplier-rating-service` currently uses manual scoring. An XGBoost regression model predicts future on-time delivery rate from features including delivery history, quality defect rates, order volume trend, payment terms compliance, and optionally NLP sentiment analysis on news and web mentions of the supplier.

Use case 2.5.5 — Returns route optimization. When a customer initiates a return, the system decides the optimal return path: send back to the origin warehouse, redirect to the nearest warehouse, or route directly to a liquidation partner. A decision tree enhanced with ML prediction of "will this item be resellable" (based on return reason, product category, and product age) determines the routing.

Use case 2.5.6 — Cold chain anomaly detection. The `cold-chain-service` monitors temperature for perishable goods. An LSTM autoencoder trained on normal temperature time-series detects anomalous patterns (unusual temperature trends, cooling equipment degradation) before they breach thresholds, enabling preventive maintenance.

### 2.6 Financial Domain (11 services, 5 AI/ML use cases)

Use case 2.6.1 — Automated reconciliation matching. The `reconciliation-service` matches payments to orders. ML handles fuzzy matches that rule-based systems miss: partial payments, split payments, and bank reference mismatches. A Siamese network trained on payment-order feature pairs learns similarity. Alternatively, a gradient-boosted classifier on handcrafted match features (amount similarity, date proximity, reference string similarity) scores potential matches.

Use case 2.6.2 — Cash flow forecasting. A time-series regression model (Prophet or LightGBM) predicts daily and weekly cash flow from the order pipeline, payment terms, payout schedules from `payout-service`, subscription renewal patterns from `subscription-billing-service`, and refund patterns from `return-refund-service`.

Use case 2.6.3 — Chargeback prediction. An XGBoost classifier predicts which orders will result in chargebacks before the chargeback is filed. Pre-emptive refund (proactively refunding a flagged order) is cheaper than losing a chargeback dispute and paying the dispute fee plus penalty.

Use case 2.6.4 — KYC document verification. The `kyc-aml-service` reviews identity documents using a multi-model pipeline: PaddleOCR or Tesseract for text extraction from documents, a CNN classifier (ResNet) for document type classification (passport versus driver's license versus utility bill), anomaly detection for document authenticity (detecting manipulated documents), and ArcFace embeddings for face matching between document photo and selfie.

Use case 2.6.5 — Anomaly detection in financial transactions. An Isolation Forest on transaction features flags unusual patterns in internal financial flows without needing labeled data: unexpected journal entries in `accounting-service`, unusual payout amounts in `payout-service`, vendor payment spikes in `expense-management-service`. This is unsupervised and requires no labeled training data.

### 2.7 Customer Experience Domain (14 services, 7 AI/ML use cases)

Use case 2.7.1 — Review sentiment analysis and auto-moderation. A multi-model pipeline processes every submitted review through four stages: a toxicity classifier (fine-tuned BERT on toxic comment datasets) detects abusive content, a spam detector (XGBoost on features like review length, reviewer history, review submission velocity) detects fake or incentivized reviews, a sentiment scorer (fine-tuned RoBERTa or distilBERT on product review corpus) provides fine-grained sentiment beyond simple positive/negative, and a helpfulness predictor (regression on features like review specificity, length, and image count) estimates how useful other shoppers will find the review. Clean reviews are auto-approved. Borderline reviews queue for human moderation. Spam and toxic content is auto-rejected.

Use case 2.7.2 — Intelligent support ticket routing. When a customer writes "I ordered 3 items but only received 2, the blue shirt is missing from my package," an NLP pipeline auto-classifies the category (Fulfillment, Missing Item), assigns priority (Medium — partial shipment, not total loss), detects sentiment (frustrated), extracts entities (blue shirt, order details from customer context), and routes to the correct team queue with pre-populated agent screen and suggested response template.

Use case 2.7.3 — AI customer support agent (chatbot). The `chatbot-service` uses an LLM (Llama 3.1 70B self-hosted via vLLM, or external API as fallback) combined with RAG (LlamaIndex retrieving product catalog and policy documents from Weaviate) and tool use (structured API calls to ShopOS services like `order-service`, `tracking-service`, `return-refund-service`). The chatbot handles order tracking, returns processing, address changes, product questions, and account support. When confidence is low, it seamlessly escalates to `live-chat-service` (human agent) with full conversation context. Guardrails (LLM Guard sidecar plus NeMo Guardrails application-layer flows) protect against prompt injection, PII leakage, and unauthorized actions. For tool use, OpenFGA authorization ensures the chatbot can only access the authenticated user's data and can only perform actions within dollar-amount thresholds (refunds above a certain amount require human approval).

Use case 2.7.4 — Smart FAQ generation. BERTopic clustering on support ticket text identifies recurring question patterns. An LLM auto-generates FAQ entries for the most common patterns, which are published to `cms-service` after human review. This proactively deflects future tickets.

Use case 2.7.5 — Personalized product Q&A. RAG (LlamaIndex over product data in Weaviate) generates answers to product questions from product specifications, reviews, and existing Q&A pairs rather than showing only pre-written answers.

Use case 2.7.6 — Price drop alert optimization. The `price-alert-service` sends notifications when wishlist items drop in price. A propensity model (logistic regression on user engagement features plus price sensitivity score) predicts which users will actually buy at the new price, targeting notifications only at high-intent users instead of spamming everyone.

Use case 2.7.7 — NPS prediction. An ordinal regression model predicts NPS score from order experience signals (delivery speed relative to promise, support interactions, product review sentiment) without waiting for the survey response, enabling proactive outreach to likely detractors before they submit negative scores.

### 2.8 Communications Domain (9 services, 5 AI/ML use cases)

Use case 2.8.1 — Send-time optimization. The `notification-orchestrator` currently sends notifications immediately or at a fixed time. A per-user time-of-day preference model (histogram of historical open times with Bayesian smoothing for users with little data) delays delivery to the optimal window. A user who opens emails at 7:30 AM local time on weekdays gets their notification scheduled for 7:30 AM, not 2:00 AM when it would be buried.

Use case 2.8.2 — Message content personalization. A contextual bandit (Vowpal Wabbit) selects which email template variant and subject line will perform best for each user segment. Each template variant is an arm, user features are the context. The system automatically learns which variant works for which segment without explicit A/B test design.

Use case 2.8.3 — Channel preference prediction. A multi-class classifier predicts which notification channel (email, SMS, push, WhatsApp, in-app) has the highest open probability for each user and notification type combination, routing each notification through the channel most likely to reach the user.

Use case 2.8.4 — Smart digest aggregation. A reinforcement learning agent (multi-armed bandit) determines the optimal digest frequency per user (real-time versus daily versus weekly) and content selection (which notifications to include versus suppress), optimizing for engagement without causing notification fatigue.

Use case 2.8.5 — Chatbot escalation intelligence. When the chatbot's confidence score drops below 0.6, it seamlessly hands off to a human agent via `live-chat-service` with full conversation context, detected intent, extracted entities, and suggested resolution — all transmitted via NATS real-time message.

### 2.9 Content Domain (8 services, 6 AI/ML use cases)

Use case 2.9.1 — Automatic image tagging and alt text. When a product image is uploaded to `media-asset-service`, a CLIP model generates image embeddings and tags ("red," "dress," "floral," "summer," "women's"), an LLM generates accessibility alt text ("A red floral summer dress with a V-neck and short sleeves on a white background"), and k-means clustering on pixel values extracts dominant colors. Tags, alt text, and colors are stored alongside the image in MinIO metadata and auto-populate product attributes.

Use case 2.9.2 — Image quality gating. A CNN classifier trained on good and bad product image pairs auto-rejects low-quality uploads: blurry images, too-dark photos, images with watermarks, wrong aspect ratios, and cluttered backgrounds. Quality score thresholds are configurable per marketplace.

Use case 2.9.3 — NSFW and prohibited content moderation. YOLO v8 or EfficientNet classifiers detect NSFW or policy-violating images before they appear on the platform. Open-source models like NudeNet provide pre-trained detection capabilities.

Use case 2.9.4 — SEO content generation. An LLM with structured output (using Instructor or Outlines for guaranteed JSON output) auto-generates meta titles, meta descriptions, and JSON-LD structured data from product attributes. This runs as a batch job across all products, improving search engine visibility.

Use case 2.9.5 — Blog and content recommendation. Content-based filtering using TF-IDF or embedding similarity between blog posts and the product or user context recommends relevant articles on product pages and in the CMS.

Use case 2.9.6 — Translation quality scoring. A quality estimation model (COMET or similar) scores machine translations in `i18n-l10n-service` and flags poor translations for human review. An LLM generates initial draft translations that the quality model evaluates before publishing.

### 2.10 Analytics and AI Domain (13 services)

All 13 services in this domain are ML-native by design. The key architectural principle is that all models read features from `ml-feature-store` (backed by Feast) rather than directly from databases. This ensures feature consistency between training time and inference time, which is the number one cause of ML production bugs. The data flow is: raw events from Kafka flow through `data-pipeline-service` (ETL to ClickHouse and MinIO), then to `ml-feature-store` (Feast: offline store in PostgreSQL for batch training, online store in Redis for real-time inference), then to individual model services (recommendation, sentiment, pricing, fraud, demand, personalization, CLV, attribution, search analytics) which consume features and produce predictions.

### 2.11 B2B Domain (7 services, 4 AI/ML use cases)

Use case 2.11.1 — Contract clause extraction and risk scoring. LayoutLM or Donut (document layout analysis models) extract key clauses from uploaded contract PDFs in `contract-service`. A text classifier scores each clause for risk: unfavorable terms, unusual penalties, missing SLAs, non-standard payment terms.

Use case 2.11.2 — B2B demand prediction. Survival analysis on `purchase-order-service` data predicts reorder timing: "Company X typically reorders 500 units every 6 weeks — they are due in 5 days." This enables proactive outreach from sales teams.

Use case 2.11.3 — Dynamic credit limit adjustment. The `b2b-credit-limit-service` currently sets limits manually. A regression model predicts probability of default from payment behavior, order volume trend, and external business signals. Credit limits are set as a function of risk score multiplied by revenue potential.

Use case 2.11.4 — RFQ response generation. An LLM with structured output auto-generates initial quote responses in `quote-rfq-service` based on historical pricing, customer tier, product margins, and volume discounts.

### 2.12 Integrations Domain (10 services, 3 AI/ML use cases)

Use case 2.12.1 — Intelligent data field mapping. When integrating with a new ERP or CRM system, embedding similarity (encoding field names plus sample values, then matching by cosine similarity) auto-suggests field mappings, reducing integration setup time from days to hours.

Use case 2.12.2 — Entity resolution across systems. When syncing data between ShopOS and external systems, a deduplication model (Python Dedupe library or custom Siamese network on name/address/email features) matches entities even when names and addresses differ slightly between systems.

Use case 2.12.3 — Marketplace listing optimization. An LLM generates platform-specific listing titles and descriptions optimized for each marketplace's search algorithm (Amazon A9, eBay Cassini), combined with regression-based pricing optimization per marketplace.

### 2.13 Affiliate Domain (4 services, 3 AI/ML use cases)

Use case 2.13.1 — Affiliate fraud detection. An Isolation Forest plus rule engine detects click fraud, cookie stuffing, and self-referral. Features include click-to-conversion ratio, conversion time distribution, IP address clustering, referral chain analysis, and coupon code distribution patterns.

Use case 2.13.2 — Influencer performance prediction. A regression model predicts which influencer will drive the highest ROI for a given product category based on audience size, engagement rate, niche match, and historical conversion rate.

Use case 2.13.3 — Dynamic commission optimization. A multi-armed bandit explores different commission rates per affiliate per product category to find the rate that maximizes profit: higher commissions on hard-to-sell products, lower commissions on products that sell themselves.

---

## 3. Self-Hosted Model Infrastructure

### 3.1 Architecture decision: zero subscriptions

All LLM and ML models run on the ShopOS Kubernetes cluster using open-source serving software. No paid subscriptions to Anthropic, OpenAI, GitHub Copilot, Cursor, or any other proprietary AI tool. The rationale is threefold: cost efficiency at scale (238 agents would exhaust any subscription plan), data sovereignty (source code and operational data never leave the cluster), and elimination of rate limits (agents can fire requests simultaneously without hitting per-minute caps).

### 3.2 Five-tier model architecture

The models are organized into five tiers by capability and resource requirements.

Tier 1 — Flagship reasoning. DeepSeek-R1 (671 billion parameter mixture-of-experts, approximately 50 billion active parameters per inference). This is the most capable open-source reasoning model, matching or exceeding GPT-4o on many benchmarks. It serves squad leads and handles complex architecture decisions, cross-domain reasoning, and incident synthesis. It requires 4 A100 80GB GPUs with tensor parallelism and produces approximately 50 to 80 tokens per second. The fallback is Qwen 2.5 72B which requires only one A100.

Tier 2 — Coding specialist. DeepSeek-Coder-V3 (236 billion parameter mixture-of-experts) is the primary coding model serving all 58 coding agents across 8 programming languages. It is quantized to AWQ 4-bit to fit on 2 A100 80GB GPUs. The fallback is Qwen 2.5 Coder 32B (GPTQ 4-bit, fits on a single A100 or even an L4 24GB). Additional coding models available as fallbacks include CodeLlama 70B and StarCoder2 33B.

Tier 3 — Fast inference. Llama 3.1 8B serves high-volume, low-complexity tasks: agent task routing, classification, simple text generation, and the Agent Health Watchdog (which needs to check 238 agents every 15 minutes and must be fast). It requires only a single T4 16GB GPU and produces over 200 tokens per second. Alternatives include Qwen 2.5 7B and Phi-3.5 Mini (3.8B) for even faster classification.

Tier 4 — Embedding models. BGE-Large-en-v1.5 (335 million parameters, 1024-dimensional output) is the primary text embedding model for RAG retrieval, semantic search, and the agent knowledge base search in Weaviate. Nomic-Embed-Text v1.5 handles long-context embeddings (up to 8192 tokens) for embedding entire product descriptions and documentation. CLIP ViT-L/14 provides cross-modal (image plus text) embeddings for visual product search. All three embedding models share a single T4 GPU, served by HuggingFace Text Embeddings Inference (TEI).

Tier 5 — Vision models. LLaVA-1.6 7B handles vision-language tasks (image analysis, product photo description). YOLO v8 handles real-time object detection and NSFW content moderation. LayoutLM v3 and PaddleOCR handle document layout analysis and OCR for KYC and contract processing. These share a single L4 24GB GPU.

### 3.3 GPU cluster layout

The cluster has 7 nodes with 15 GPUs total. Five inference nodes are always-on: Node 1 with 4 A100 80GB GPUs running DeepSeek-R1, Node 2 with 2 A100 80GB running DeepSeek-Coder-V3 and Qwen 2.5 Coder 32B, Node 3 with 2 A100 80GB running Llama 3.1 70B and Qwen 2.5 72B, Node 4 with 2 T4 16GB running embeddings and Llama 3.1 8B, and Node 5 with 1 L4 24GB running vision models. Two training nodes (each with 2 A100 80GB) provision on-demand as spot instances via Karpenter, auto-terminate after training job completion, and run approximately 20 hours per week for model training and fine-tuning.

### 3.4 Model serving stack

The serving architecture has three layers. At the top is LiteLLM Proxy, which provides a single OpenAI-compatible API endpoint for all agents. Every agent and every ShopOS ML service calls the same URL. LiteLLM handles model routing (directing requests to the correct vLLM instance based on model name), load balancing across fallback instances, automatic failover (if the primary coding model is down, requests route to the backup), token usage tracking per API key (each agent has its own key), cost estimation, and logging to Langfuse for observability.

In the middle layer is vLLM, the high-performance LLM serving engine. Each model tier has its own vLLM instance (or TEI instance for embeddings). vLLM provides continuous batching, PagedAttention for efficient GPU memory usage, tensor parallelism across multiple GPUs, and an OpenAI-compatible API. Each vLLM instance is deployed as a Kubernetes Deployment with a PersistentVolumeClaim for model weight caching (avoiding re-downloading 200+ GB models on pod restarts), HorizontalPodAutoscaler scaling on request queue depth, PodDisruptionBudget ensuring at least one replica is always available, and ServiceMonitor for Prometheus metrics collection.

At the bottom layer is the NVIDIA GPU Operator, which manages GPU drivers, device plugins, and runtime components on all GPU nodes.

### 3.5 Open-source tool replacements

Claude Code (proprietary, $200/month) is replaced by OpenHands (MIT license, 77.6% SWE-bench score) for complex multi-file coding tasks and by Aider (Apache 2.0, git-native) for smaller changes. Both connect to local models via the LiteLLM endpoint. Codex CLI (proprietary, $200/month) is replaced by OpenCode (MIT, 100,000+ GitHub stars) for parallel bulk operations. Cursor (proprietary, $20/month) is replaced by Cline (Apache 2.0, VS Code extension) and RooCode (MIT) for IDE-based coding. GitHub Copilot (proprietary, $20/month) is replaced by Continue (Apache 2.0) for autocomplete. ChatGPT and Claude chat subscriptions are replaced by Open WebUI (MIT) which provides a web interface to all local models. All tools connect via the same LiteLLM endpoint with zero code changes — they all speak the OpenAI-compatible API.

### 3.6 Model versioning and update strategy

Every model is pinned by HuggingFace commit SHA in a `model-registry.yaml` file stored in git. The format is exclusively SafeTensors (never pickle — this eliminates the entire pickle remote code execution attack vector). Model updates follow a five-step pipeline: the Research Agent flags a new model version, the Benchmark Agent evaluates it on ShopOS-specific tasks (a custom coding eval across Go, Java, Kotlin, Python tasks plus operations tasks for Terraform, Helm, kubectl), if the new model scores better a shadow deployment serves 10% of traffic for 48 hours while quality metrics are compared, if shadow passes a canary deployment serves 50% for 24 hours, and if canary passes the new model is promoted to primary and the old model becomes the fallback.

Future enhancement: fine-tuning a coding model on ShopOS-specific data. After 6 months of agent operation, 10,000+ approved PRs provide high-quality training pairs (prompt plus code change). QLoRA fine-tuning of Qwen 2.5 Coder 32B on this data is expected to produce 5 to 15% quality improvement on ShopOS-specific tasks. Further ahead, distilling DeepSeek-Coder-V3 (236B) into a 7B ShopOS-specialized model would provide 10x faster inference at 90% of the quality, servable on a single T4 GPU.

---

## 4. Unified Agent Army — Four Divisions

### 4.1 Design principles

One agent equals one specialty. Every agent is an expert in exactly one thing. The Go Agent knows Go. The Terraform Agent knows Terraform. The order-service Agent knows order-service. No agent is a generalist.

Self-documenting. Every agent maintains its own knowledge base: SOUL.md (identity, expertise, constraints), MEMORY.md (current context and session state), ISSUES.md (every issue encountered with root cause and resolution), RESEARCH.md (findings and best practices), RUNBOOK.md (step-by-step procedures for common tasks), and METRICS.md (self-tracking of tasks completed, token spend, escalation rate). Documentation is a byproduct of work, not a separate task.

Minimal prompting. Agents are pre-loaded with deep context via their SOUL.md file (which contains coding standards, escalation rules, and worked examples), their SKILL.md files (which contain step-by-step procedures for every task type), and RAG over relevant documentation (architecture docs, ADRs, past issues from the Weaviate knowledge index). When a human says "fix the failing test in order-service," the agent already knows the language (Kotlin), the framework (Spring Boot), the database (PostgreSQL), the test framework (JUnit), and the CI pipeline (Jenkins).

Hierarchical delegation. The chain of command is Human, then Division Lead, then Squad Lead, then Specialist Agent. Each level adds context and routes to the right specialist. No specialist agent talks to the human directly unless escalating.

Everything logged. Every action, decision, reasoning chain, tool call, and outcome is logged to Langfuse (LLM trace observability) and Paperclip (task management audit trail). This becomes the organization's institutional memory.

### 4.2 Developer Division (58 agents)

This division owns all application source code across 154 services in 8 languages. The division lead is an OpenHands instance using the DeepSeek-R1 flagship model.

Language squads. Each programming language has a squad lead agent that is a deep expert in that language's idioms, toolchain, test framework, dependency management, and common pitfalls. The Go Squad Lead manages 80+ Go services and has sub-agents organized by domain group: Go-Platform Agent (18 platform Go services), Go-Commerce Agent (15 commerce Go services), Go-Catalog Agent (7 catalog Go services), and so on across all domains. Similarly, the Java Squad Lead manages approximately 20 Java services with sub-agents per domain, the Kotlin Squad Lead manages 8 services, the Python Squad Lead manages 15 services (particularly important for the Analytics and AI domain), the Node.js Squad Lead manages 12 services, and individual squad leads cover Rust (2 services), C# (2 services), and Scala (1 service). All coding agents use OpenHands or Aider with the DeepSeek-Coder-V3 model via LiteLLM.

Protocol and schema agents. A dedicated Proto/gRPC Agent manages all 58 proto files using Buf CLI for linting, breaking-change detection, and code generation. An Avro/Events Agent manages all 20 Avro event schemas and validates schema evolution rules through Schema Registry.

Service specialist agents. The top 20 most critical services each get a dedicated agent that knows everything about that one service: its language, database schema, Kafka events, proto definition, Helm chart, test suite, and every bug ever fixed (recorded in its ISSUES.md). These agents include order-service, payment-service, fraud-detection-service, chatbot-service, search-service, api-gateway, checkout-service, cart-service, recommendation-service, auth-service, inventory-service, pricing-service, fulfillment-service, notification-orchestrator, product-catalog-service, warehouse-service, review-rating-service, subscription-billing-service, invoice-service, and graphql-gateway.

Quality agents. A Code Review Agent (OpenHands in read-only mode) reviews every PR from every coding agent, checking coding standards, architecture compliance, and test coverage. A Test Generation Agent (OpenCode running parallel tasks) generates unit, integration, and end-to-end tests across all services. Both use the coding model tier.

### 4.3 DevOps Division (88 agents)

This division owns infrastructure, CI/CD, GitOps, monitoring, databases, messaging, networking, and cloud resources. The division lead is an OpenClaw instance using the Llama 3.1 70B general model.

CI/CD squad (17 agents). One agent per CI platform: Jenkins Agent, GitHub Actions Agent, GitLab CI Agent, Tekton Agent, ArgoCD/Flux Agent (GitOps), Drone CI Agent, Woodpecker CI Agent, Dagger Agent, Concourse Agent, CircleCI Agent, GoCD Agent, Travis CI Agent, Harness Agent, Azure DevOps Agent, AWS CodePipeline Agent, and GCP Cloud Build Agent. Each agent knows its platform's syntax, credential management, and debugging patterns. The ArgoCD/Flux Agent also handles GitOps: App-of-Apps management, ApplicationSets, sync operations, rollbacks, and Argo Rollouts canary deployments.

Infrastructure squad (5 agents). Terraform Agent manages all Terraform modules for EKS, GKE, and AKS provisioning and can plan, apply (with human approval for destructive changes), import, and handle state management. OpenTofu Agent covers the same targets using the open-source alternative. Crossplane Agent manages Kubernetes-native IaC compositions and claims. Ansible Agent generates and manages playbooks for node bootstrapping using Ansible Lightspeed. Helm Agent manages all 184 Helm charts (154 services plus 30 tools) including linting, templating, upgrading, and rollback.

Kubernetes squad (5 agents). The Kubernetes Core Agent manages namespaces, RBAC, resource quotas, and pod disruption budgets. The KEDA Agent manages all ScaledObjects for Kafka and Redis trigger-based autoscaling. The Velero Agent manages backup schedules and disaster recovery procedures. The Network Policy Agent manages default-deny patterns and Cilium L7 policies. The GPU Operator Agent manages NVIDIA GPU drivers, device plugins, and MIG partitioning.

Database squad (13 agents). One agent per database technology: PostgreSQL (managing 100+ service schemas), MongoDB (catalog, CMS, reviews), Redis (cache, sessions, pub/sub), Cassandra (time-series analytics), ScyllaDB (high-throughput analytics), Elasticsearch (full-text search), OpenSearch (log analytics and security events), ClickHouse (OLAP aggregation), Weaviate (vector search and embeddings), Neo4j (graph recommendations), MinIO (object storage), etcd (distributed configuration), and Memcached (high-throughput caching). Each agent knows its database's administration, performance tuning, backup and recovery, schema management, and monitoring.

Messaging squad (5 agents). Kafka Agent manages topics, partitions, consumer groups, and integrates with Confluent, Strimzi, ksqlDB, and Schema Registry. RabbitMQ Agent manages queues, exchanges, bindings, and dead-letter handling. NATS Agent manages JetStream streams, consumers, and clustering. Debezium Agent manages CDC connectors and offset management. Flink Agent manages FlinkDeployments, job lifecycle, and savepoints.

Networking squad (5 agents). Istio Agent manages the service mesh, mTLS, and traffic management rules. Cilium Agent manages eBPF-based network policies and Hubble observability. Traefik Agent manages ingress routing, TLS termination, and routing rules. Consul Agent manages service discovery, health checks, and KV configuration. Kong Agent manages API gateway plugins and rate limiting.

Monitoring squad (7 agents). Prometheus Agent manages recording rules, alert rules, and Thanos long-term storage. Grafana Agent manages dashboards, provisioning, and data sources. Loki Agent manages log queries, alert rules, and retention policies. Tempo/Jaeger Agent handles distributed trace analysis and service dependency mapping. Sentry Agent manages error tracking and release management. Uptime Kuma Agent manages status pages and health probes. OpenTelemetry Agent manages collector configurations and instrumentation.

Cloud squads (31 agents total). Each cloud provider gets a dedicated squad with one agent per cloud service. The AWS Cloud Squad has 11 agents covering EKS, S3, RDS, MSK, ECR, IAM, VPC, Bedrock, SageMaker, Cost Explorer, and GuardDuty/SecurityHub. The GCP Cloud Squad has 10 agents covering GKE, GCS, CloudSQL, Vertex AI, Artifact Registry, IAM, VPC, DLP, Billing, and Security Command Center. The Azure Cloud Squad has 10 agents covering AKS, Blob Storage, Azure DB, Azure OpenAI, Azure ML, ACR, Identity/Entra, VNet, Cost Management, and Defender for Cloud/AI.

### 4.4 DevSecOps Division (37 agents)

This division secures everything. All agents are predominantly read-only, sandboxed via NemoClaw, and escalate to humans for remediation actions.

SAST squad (5 agents). Semgrep Agent runs custom security rules and auto-triages results. SonarQube Agent manages quality gates, code smells, and technical debt tracking. Checkov Agent scans Terraform, Helm, and Kubernetes configurations. KICS Agent provides an extended IaC scanning rule set. CodeQL Agent performs deep semantic analysis on GitHub-hosted repositories.

DAST squad (2 agents). OWASP ZAP Agent performs dynamic application scanning and API fuzzing. Nuclei Agent runs CVE template-based scanning against deployed services.

SCA squad (8 agents). Trivy Agent scans container images and IaC for vulnerabilities. Grype Agent provides CVE scanning for container images. Snyk Agent (open-source edition) handles dependency vulnerability analysis and license compliance. Cosign Agent manages image and model artifact signing via Sigstore. Syft Agent generates CycloneDX SBOMs. modelscan Agent scans ML model artifacts for malicious code. picklescan Agent specifically targets pickle deserialization risks. pip-audit Agent audits Python dependency security.

Runtime security squad (4 agents). Falco Agent manages runtime threat detection rules and alert triage. Tetragon Agent manages eBPF enforcement policies and process monitoring. Tracee Agent handles eBPF event collection and analysis. Coraza WAF Agent manages ModSecurity rules and LLM-specific attack pattern detection.

Policy squad (4 agents). OPA/Gatekeeper Agent manages Rego policies, testing, and admission control. Kyverno Agent manages Kubernetes admission controller policies. Kubewarden Agent manages WebAssembly-based policy engine rules. OpenFGA Agent manages relationship-based authorization (Google Zanzibar model).

Secrets and identity squad (5 agents). Vault Agent manages secrets, PKI, and dynamic credentials. SPIFFE/SPIRE Agent manages workload identity and X.509 SVIDs for mTLS. Keycloak Agent manages IAM, SSO, and realm configuration. cert-manager Agent manages TLS certificate lifecycle and automatic renewal. Dex Agent manages OIDC federation.

Penetration testing squad (4 agents). kube-bench Agent runs CIS Kubernetes benchmark assessments. kube-hunter Agent performs Kubernetes penetration testing. Garak Agent probes LLM services for prompt injection, jailbreak, and data leakage vulnerabilities. PyRIT Agent orchestrates red-team scenarios against GenAI applications.

AI Security squad (5 agents). LLM Guard Agent monitors guardrail block rates and PII redaction events. Egress Proxy Agent monitors external LLM API traffic patterns. Drift Monitor Agent watches for model drift across all production models. ART Agent (Adversarial Robustness Toolbox) detects adversarial inputs to fraud and recommendation models. Presidio Agent monitors PII detection in data pipelines.

### 4.5 AI Division (30 agents)

This division manages all ML models, LLM services, data pipelines, and AI infrastructure. The division lead is an OpenHands instance using the DeepSeek-R1 flagship model.

ML Platform squad (6 agents). MLflow Agent manages experiment tracking, model registry, and model serving configuration. KServe Agent manages InferenceService resources and autoscaling. Feast Agent manages the feature store (offline in PostgreSQL, online in Redis). DVC Agent manages dataset versioning and data pipeline tracking. Optuna Agent orchestrates hyperparameter tuning. Argo Training Agent manages training DAG workflows on the burst GPU pool.

Model squad (9 agents, one per production model). Each agent owns its model's complete lifecycle: monitoring drift (via Evidently sidecar metrics), triggering retraining when drift exceeds thresholds, evaluating retrained models against baselines, running the OPA promotion gate (checking signature, eval pass, fairness pass, dataset consent), promoting to production in MLflow, and updating Helm values to deploy the new model version. The nine model agents cover fraud detection, recommendations, sentiment analysis, price optimization, demand forecasting, personalization, CLV scoring, attribution modeling, and search ranking.

LLM squad (6 agents). vLLM Agent manages model serving instances, GPU allocation, and batching configuration. Embedding Server Agent manages TEI instances for text and image embeddings. RAG Pipeline Agent manages the LlamaIndex retrieval pipeline, chunking strategies, and retrieval quality. Prompt Engineering Agent manages prompt versions, A/B testing via Promptfoo, and prompt regression testing. Guardrails Agent manages LLM Guard and NeMo Guardrails configurations. Fine-tuning Agent manages LoRA/QLoRA training, dataset preparation, and evaluation.

Data Pipeline squad (4 agents). Flink Agent manages streaming jobs for feature computation and PII scrubbing. Data Quality Agent runs Great Expectations validations on datasets before they enter the feature store. Data Lineage Agent manages OpenLineage and Marquez for end-to-end data traceability. ETL Agent manages the data-pipeline-service for batch processing from Kafka to ClickHouse and MinIO.

Evaluation squad (5 agents). Promptfoo Agent runs prompt regression tests on every chatbot prompt change. Ragas Agent evaluates RAG retrieval quality (faithfulness, relevance, context recall). DeepEval Agent evaluates LLM output quality for hallucination and toxicity. Fairlearn Agent runs bias detection on all production models. Benchmark Agent runs SWE-bench-style evaluations on coding agents to measure their quality.

### 4.6 Cross-Cutting Agents (25 agents)

Documentation squad (6 agents). Architecture Doc Agent monitors every merged PR and updates relevant architecture documents. ADR Agent detects significant design decisions in PRs and drafts Architecture Decision Records. API Doc Agent regenerates API documentation when proto files or REST endpoints change. Runbook Agent creates or updates runbook procedures after incidents are resolved. Changelog Agent updates CHANGELOG.md with human-readable entries for every merged PR. README Agent updates relevant README files when new services, tools, or configurations are added.

Research squad (4 agents). Technology Research Agent scans tech news, Hacker News, and GitHub trending daily for tools and developments relevant to ShopOS. Vulnerability Research Agent monitors CVE feeds and GitHub Security Advisories every 6 hours, checking every ShopOS dependency against new CVEs. Best Practices Agent scans CNCF blogs, Kubernetes release notes, and cloud provider updates weekly for applicable recommendations. Competitive Intelligence Agent (optional) scans open-source commerce platforms for feature comparisons.

Issue Management squad (3 agents). Issue Triage Agent auto-classifies new issues (bug, feature, question), assigns priority, and routes to the correct division, squad, and agent. Issue Resolution Tracker monitors all open issues, pings stale issues, auto-closes resolved ones, and produces weekly health reports. Post-Incident Review Agent generates a post-incident review document after every incident resolution, covering timeline, root cause, resolution steps, and prevention measures.

Cost Management squad (4 agents). Cloud Cost Agent (one per cloud provider, already counted in DevOps cloud squads) tracks compute spend and identifies optimization opportunities. Agent Cost Agent tracks LLM token spend per agent via LiteLLM and Langfuse, identifies wasteful loops and inefficient agents, and produces daily agent cost reports. License Cost Agent tracks all SaaS subscriptions and open-source license compliance. Resource Optimization Agent uses OpenCost and Goldilocks (both already deployed in ShopOS) to generate right-sizing recommendations for all pods.

Quality Assurance squad (4 agents). PR Review Agent (cross-division) reviews every PR from every agent for code quality, test presence, security, and documentation updates. Integration Test Agent runs cross-service integration tests when multiple services change in the same time window. Load Test Agent manages k6, Locust, and Gatling test suites and triggers performance tests before major deployments. Chaos Test Agent manages Chaos Mesh and LitmusChaos experiments and runs scheduled weekly chaos game days.

System agents (4 agents). Agent Health Watchdog checks all agents every 15 minutes for heartbeat, productivity, error rate, and token spend, auto-restarting unresponsive agents and alerting on anomalies. Retrospective Agent runs weekly analysis of rejected PRs, escalations, and recurring issues, then generates SOUL.md amendments (which require human approval) to improve agent behavior over time. Feedback Agent collects human thumbs-up/thumbs-down reactions on agent output and proposes preference rules after detecting patterns. Daily Digest Agent aggregates all agent activity into a daily Slack report covering tasks completed, issues resolved, escalations, cost, and agent health.

---

## 5. Agent System — 12 Critical Improvements

Improvement 1 — Inter-agent event mesh. Agents publish events (code.changed, test.failed, model.retrained, alert.fired, deploy.started, schema.changed, config.changed, issue.resolved) to NATS JetStream. Other agents subscribe to events relevant to their domain. This eliminates the isolation problem where one agent's change breaks another agent's service without notification. Events are durable (JetStream) so offline agents catch up when they wake.

Improvement 2 — Distributed locks for conflict prevention. Before modifying any file or Kubernetes resource, an agent acquires an advisory lock via Redis (SETNX with 600-second TTL). If the lock is held by another agent, it waits or notifies the other agent via the event mesh. For semantic conflicts (no file overlap but logical conflict), agents run a change impact analysis by querying the Neo4j service dependency graph before every PR to identify downstream services affected by the change.

Improvement 3 — Agent Health Watchdog. A dedicated agent checks all 238 agents every 15 minutes. It verifies each agent's heartbeat (is it alive), productivity (is it completing tasks), error rate (is it failing), token spend (is it looping), and memory freshness (is it persisting state). Agents that are unresponsive are auto-restarted. Agents with error rates above 30% are disabled and escalated. Agents with token spend exceeding 3x their daily average are throttled and investigated. A Grafana heatmap dashboard shows all agents color-coded by health status.

Improvement 4 — Self-improvement via Retrospective Agent. A weekly analysis of each agent's rejected PRs, escalations, recurring issues, and token-expensive sessions identifies behavioral patterns. The Retrospective Agent generates SOUL.md amendments (for example: "ALWAYS wrap Go errors with fmt.Errorf" after the Go Agent was rejected 4 times for bare error returns). Amendments require human approval before they take effect. This creates a flywheel: mistakes lead to patterns, patterns lead to rules, rules lead to fewer mistakes.

Improvement 5 — Agent staging pipeline. No agent configuration change goes live untested. The pipeline has four stages: EVAL (run agent against synthetic test scenarios), SHADOW (deploy alongside production agent, both receive same inputs, shadow outputs are logged but not executed for 48 hours), CANARY (10% of real tasks for 24 hours with auto-rollback if error rate exceeds baseline plus 5%), and PROMOTE (full rollout with old config retained for 7-day rollback).

Improvement 6 — Priority tiers with SLAs. P0 (production down, data loss, security breach): 5-minute response SLA, multiple agents swarm, immediate human notification, unlimited token budget. P1 (degraded service, SLO breach): 30-minute response, escalation after 1 hour. P2 (bug fix, feature request): 4-hour pickup, escalation after 24 hours. P3 (documentation, research): 48-hour SLA. P4 (nice-to-have, exploration): best effort. P0 tasks bypass normal queues via dedicated high-priority NATS stream.

Improvement 7 — Multi-model fallback. Every agent defines a model fallback chain in its SOUL.md. Since all models are self-hosted, failover is handled by LiteLLM routing and is instantaneous (under 1 millisecond switch time). If the primary coding model is down, requests automatically route to the backup coding model, then to the general model, then to the fast model (degraded capability), then to pause-and-escalate. This is a significant advantage over API-based models where switching from Anthropic to OpenAI requires different authentication, different API formats, and different pricing.

Improvement 8 — Human preference learning. A Feedback Agent collects thumbs-up/thumbs-down reactions on agent output via Slack emoji reactions and GitHub PR review comments tagged with a feedback marker. When 3 or more similar feedbacks form a pattern (for example: "use table-driven tests in Go" appearing in 5 reviews), the agent generates a SOUL.md amendment that, after human approval, permanently adjusts the agent's behavior. Effectiveness is tracked: if the amendment doesn't reduce corrections within 2 weeks, it's revised or removed.

Improvement 9 — Four collaboration patterns. CHAIN pattern for sequential handoffs across domains (Proto Agent adds field, then Go Agent updates service, then Java Agent updates client, then Helm Agent updates config, then QA Agent reviews all). SWARM pattern for P0 incidents where multiple agents work simultaneously on different hypotheses (Go Agent checks recent code changes while PostgreSQL Agent checks query performance while Kafka Agent checks consumer lag while Kubernetes Agent checks pod health, all reporting to SRE Agent who synthesizes). REVIEW pattern requiring minimum 2 agent approvals for any PR (language expert plus QA, with human approval required for security and infrastructure changes). DELEGATE pattern for when an agent encounters something outside its expertise (Go Agent hits a database performance issue, creates a task that routes to PostgreSQL Agent, then resumes its own work after receiving the resolution).

Improvement 10 — Agent utilization review and consolidation. Monthly review flags agents with zero tasks in 30 days for retirement and agents with fewer than 3 tasks for consolidation with their squad lead. Expected consolidation: ScyllaDB plus Cassandra merge into a single wide-column DB agent, Drone plus Woodpecker plus Travis merge into a single secondary CI agent, Tempo plus Jaeger plus Zipkin merge into a single tracing agent. Target: start at 238, stabilize at 120 to 150 after 3 months.

Improvement 11 — Automated reporting. Daily digest delivered to Slack at 06:00 UTC covering tasks completed (by division), issues resolved (with one-line summaries), escalations, daily cost, agent health status, and new research findings. Weekly report in markdown covering top issues, knowledge base growth, agent performance rankings, consolidation recommendations, and cost trends. Live Grafana dashboard with agent heatmap, task throughput by division, token cost trends, and escalation rate.

Improvement 12 — Disaster recovery. Git is the source of truth for all agent knowledge. All SOUL.md, MEMORY.md, ISSUES.md, RESEARCH.md, RUNBOOK.md, and METRICS.md files are stored in the git repository. As long as git survives, the entire agent system rebuilds from scratch. Paperclip state is backed up hourly to S3/GCS/Blob via PostgreSQL backup. The Weaviate knowledge index is snapshot every 6 hours to MinIO. Model weights are cached on PersistentVolumeClaims and can be re-downloaded from HuggingFace if lost. Single agent crash recovery takes less than 15 minutes (Health Watchdog detects, restarts pod, agent reads MEMORY.md to recover context). Full cluster recovery takes 30 to 60 minutes after Kubernetes is back (all state reconstructed from git and backups).

---

## 6. Knowledge Management System

Every agent maintains six files that persist across sessions. SOUL.md defines the agent's identity, expertise, constraints, coding standards, and escalation rules — it is only modified by the Retrospective Agent with human approval. MEMORY.md is the agent's current working context: priorities, ongoing investigations, recent changes, and notes from the last session — it is auto-updated after every heartbeat. ISSUES.md records every issue the agent encounters with symptoms, root cause analysis, resolution steps, prevention measures, related issues, and PR references — this is the primary learning artifact and grows to hundreds of entries over time. RESEARCH.md captures findings, best practices, tool evaluations, and recommendations — it grows steadily as the agent researches topics relevant to its domain. RUNBOOK.md contains step-by-step procedures for common tasks — it grows then stabilizes at 30 to 50 procedures. METRICS.md tracks the agent's own performance: tasks completed, PRs created, escalations, and token spend — it is a rolling weekly summary.

All knowledge files are indexed in Weaviate for cross-agent search. Any agent can query "has any agent seen this error before?" and get results from any other agent's ISSUES.md. This means a fix by the Redis Agent 3 months ago automatically helps the PostgreSQL Agent when it encounters a similar connection pool issue today. Knowledge files of retired agents are archived (not deleted) and remain searchable. Replacing agents inherit their predecessor's ISSUES.md and RESEARCH.md via RAG.

---

## 7. AI/ML Security Architecture

The security layer covers five surfaces. Supply chain security uses modelscan and picklescan in CI to block model artifacts containing malicious code, Cosign to sign model artifacts after evaluation, CycloneDX ML-BOM generation, HuggingFace model revision pinning by commit SHA, and an ML dependency allowlist with Renovate rules for safe auto-bumping. Admission control uses Kyverno policies requiring model signatures, evaluation pass tags, and fairness pass tags before pods can load models, plus a validating webhook that queries MLflow to verify model metadata. LLM runtime protection uses LLM Guard as a sidecar (for self-hosted models) or egress proxy (for external API fallback) providing prompt injection detection, PII egress filtering, jailbreak detection, and output validation. An Envoy token-budget filter rate-limits by token cost, and Coraza WAF rules detect LLM-specific attack patterns. Training data protection uses a Flink PII scrubbing job between raw Kafka event topics and feature store topics, a poisoning anomaly detector on review ingestion, OpenLineage for dataset lineage tracking, and a GDPR erasure fanout workflow that propagates right-to-be-forgotten requests to MLflow (retrain trigger), Weaviate (vector delete), and Neo4j (node delete). Inference protection uses Evidently drift monitoring sidecars on all production models, IBM ART adversarial input detection on fraud-detection-service, and anti-extraction rate limiting with query-entropy anomaly detection.

---

## 8. Cloud Infrastructure (AWS, GCP, Azure)

Each cloud is designed independently. All three run the same Kubernetes-native ML stack (vLLM, LiteLLM, KServe, Feast, MLflow) on top.

AWS uses EKS with Karpenter for GPU node provisioning (automatic spot instance selection from a priority list of g5, p4d, inf2, and trn1 instances), S3 for model artifacts and datasets, ECR for container and OCI model images, RDS for MLflow and Feast backend, ElastiCache for Feast online store, MSK for Kafka, and IRSA for pod-level IAM. AWS-specific AI security includes Amazon Macie scanning S3 dataset buckets for PII, GuardDuty for infrastructure threat detection, and SecurityHub for findings aggregation. Amazon Bedrock is available as a fallback LLM API path if self-hosted models are insufficient.

GCP uses GKE (Autopilot mode for zero-management GPU scheduling where pods with GPU requests are scheduled on-demand without persistent node pools), GCS for storage, Artifact Registry for images, CloudSQL for databases, Memorystore for Redis, and Workload Identity for pod-level IAM. GCP has the cheapest GPU spot pricing of the three clouds. GCP-specific AI security includes the DLP API (the most comprehensive PII detection available) and Security Command Center. Vertex AI is available as a fallback for training and serving.

Azure uses AKS with spot GPU node pools, Blob Storage for artifacts, ACR for container registry, Azure Database for PostgreSQL, Azure Cache for Redis, Event Hubs (Kafka mode) for messaging, and Managed Identity for pod-level IAM. Azure has a unique advantage: Microsoft Defender for AI, the only cloud with a dedicated AI security product that detects prompt injection, jailbreaks, sensitive data leakage, and token cost attacks inline on Azure OpenAI traffic. Microsoft Purview provides data catalog, lineage, and PII discovery across all data stores. Azure OpenAI is available as a fallback LLM API path.

---

## 9. Future Enhancements

Near-term (3 to 6 months). Agent-to-agent teaching where solving a novel problem publishes a lesson event that related agents immediately incorporate. Custom model fine-tuning on ShopOS codebase using QLoRA. Model distillation to create a 7B ShopOS-specialized coding model. Voice interface for P0 incident response via OpenClaw's ElevenLabs integration. Mobile companion app for approvals on the go. Publishing ShopOS-specific skills to OpenClaw ClawHub and Paperclip Clipmart marketplaces.

Medium-term (6 to 12 months). Multi-cluster agent federation across production, staging, and DR. Autonomous chaos engineering where agents design, execute, and evaluate resilience experiments. Self-healing infrastructure with Terraform drift auto-correction. AI-driven capacity planning predicting infrastructure needs 30 days ahead. Autonomous A/B testing with automatic experiment design and statistical analysis. Local model RLHF using agent interaction data as reward signal for continuous model improvement.

Long-term (12+ months). Fully autonomous feature development from high-level business requirements. Agent-designed architecture with automated ADR creation and implementation. Digital twin simulation of the full ShopOS platform for risk-free experimentation. Multi-agent debate for complex decisions. On-device ML for privacy-first customer personalization. Custom commerce foundation model pre-trained on all ShopOS data. Self-replicating agents that spawn clones during load spikes and auto-terminate when load subsides.

---

## 10. Implementation Roadmap

Month 1 — GPU infrastructure and model serving. Week 1: GPU node pools via Terraform with Karpenter provisioner, NVIDIA GPU Operator. Week 2: vLLM deployment for all 5 model tiers, model weights downloaded to PVCs. Week 3: LiteLLM proxy with routing rules, TEI embeddings, Ollama for dev. Week 4: Open WebUI for human interaction, Prometheus monitoring for all model servers, Grafana dashboards, alert rules.

Month 2 — Agent foundation. Week 5: Paperclip orchestrator, NemoClaw sandbox with tiered policies, NATS JetStream event mesh, Redis advisory locks. Week 6: First 5 read-only agents (Monitoring, Security, 3 cloud cost agents). Week 7: Langfuse and AgentOps for agent observability, Grafana agent heatmap. Week 8: Agent Health Watchdog, Daily Digest Agent.

Month 3 — Developer Division. Week 9: Go Squad Lead plus 3 domain sub-agents. Week 10: Java and Python Squad Leads plus first service specialists (order-service, payment-service). Week 11: Code Review Agent and Test Generation Agent. Week 12: More service specialists (fraud, chatbot, search) plus remaining language squads.

Month 4 — DevOps Division. Week 13: Jenkins, GitHub Actions, and ArgoCD agents. Week 14: Terraform, Helm, and Kubernetes Core agents. Week 15: PostgreSQL, Redis, and Kafka agents. Week 16: Primary cloud squad plus remaining database and messaging agents.

Month 5 — DevSecOps and AI Division. Week 17: Trivy, Semgrep, Falco, and Vault security agents. Week 18: AI Security squad (LLM Guard, modelscan, Garak). Week 19: MLflow, Feast, and KServe platform agents. Week 20: Model lifecycle agents (fraud, recommendation, sentiment).

Month 6 — Cross-cutting agents and improvements. Week 21: Documentation and Research squads. Week 22: Issue Management, Retrospective Agent, and Feedback Agent. Week 23: Agent staging pipeline (eval, shadow, canary, promote). Week 24: Priority/SLA system, first agent consolidation review.

Month 7 — Scale and tune. Remaining cloud squads, remaining language squads, remaining CI/CD agents, networking and monitoring agents. Tune heartbeat frequencies and budgets based on actual usage data. Second consolidation review.

Month 8 — Full operations. All agents deployed and tuned. Third consolidation review (target: 238 down to 120 to 150 active agents). Begin fine-tuning local models on ShopOS agent interaction data. Comprehensive ROI measurement: MTTR target under 1 hour (from 4 to 6 hours baseline), deployment frequency target 5 to 10 per day (from 2 to 3 per week), change failure rate target under 5% (from 15 to 20%), PR review turnaround target under 30 minutes (from 4 to 8 hours), security patch time target under 4 hours (from 2 to 5 days).

Cost at full operations: $3,100 to $5,500 per month for all GPU compute, agent infrastructure, and storage. Zero paid subscriptions. Equivalent human team cost for the same coverage: approximately $32,000 per month (2 DevOps engineers, 1 security engineer, 1 ML engineer). Agent team cost is 10 to 17% of the human equivalent with 24/7 coverage and broader scope.

---

## 11. Sibling Platform Adoption

The Paperclip orchestrator, OpenClaw agent platform, and NemoClaw sandbox defined here are shared across all 15 sibling projects in this monorepo group. Each sibling has its own `AI_PLAN.md` describing the project-specific adoption — domain use cases, hierarchical agent architecture (Tier 0 architect → Tier 1 division leads → Tier 2 per-language/tool/service → Tier 3 ephemeral workers), separate AI infrastructure namespace and GPU pool, and project-specific compliance constraints (HIPAA, PCI-DSS, NIST 800-53, EU AI Act high-risk, ISA/IEC 62443, NERC CIP, DRM/content licensing, etc.).

| Project | AI plane | Headline AI surfaces | Primary regulatory constraint |
|---|---|---|---|
| [CivicLink](../CivicLink/AI_PLAN.md) | `civic-ai-*` | Citizen voice assistant, doc-AI, fraud detection | NIST 800-53, India IT Act, GDPR |
| [EstateIQ](../EstateIQ/AI_PLAN.md) | `estate-ai-*` | AVM, demand heatmap, KYC doc-AI | AML/KYC, GDPR/CCPA, RICS |
| [FactoryMind](../FactoryMind/AI_PLAN.md) | `factory-ai-*` (+ edge) | Predictive maintenance, defect CV, scheduling | ISA/IEC 62443, ISO 9001 |
| [FarmPulse](../FarmPulse/AI_PLAN.md) | `farm-ai-*` (+ mobile/edge) | Crop disease CV, yield, voice advisor | DGCA/FAA, GDPR/DPDP |
| [FreightForce](../FreightForce/AI_PLAN.md) | `freight-ai-*` (+ edge) | Routing, ETA, customs doc-AI, yard CV | C-TPAT/AEO, FMCSA HoS, GDPR |
| [GridForge](../GridForge/AI_PLAN.md) | `grid-ai-*` (+ edge) | Load forecast, SCADA anomaly, theft detection | NERC CIP, IEC 61850/62443 |
| [HelixCare](../HelixCare/AI_PLAN.md) | `helix-ai-*` (HIPAA-region) | CDS, imaging triage, ambient scribe | HIPAA, MDR/IVDR, FDA SaMD |
| [LedgerX](../LedgerX/AI_PLAN.md) | `ledger-ai-*` (PCI-segmented) | Fraud, AML graph, credit, robo-advisor | PCI-DSS, SOX, SR 11-7, MAR |
| [MatchDay](../MatchDay/AI_PLAN.md) | `match-ai-*` (+ edge) | Win-prob, tracking CV, highlights | Sports licensing, COPPA |
| [RiskShield](../RiskShield/AI_PLAN.md) | `risk-ai-*` (+ air-gap option) | SIEM triage, malware CV, underwriting | ISO 27001, MITRE D3FEND |
| [ScholarPath](../ScholarPath/AI_PLAN.md) | `scholar-ai-*` | AI tutor, essay scoring, plagiarism | COPPA, FERPA, GDPR-K |
| [SignalGrid](../SignalGrid/AI_PLAN.md) | `signal-ai-*` (+ MEC) | Network anomaly, RAN energy, fraud | GSMA NESAS, ETSI ZSM |
| [StayNest](../StayNest/AI_PLAN.md) | `stay-ai-*` | Demand forecast, dynamic pricing, concierge | PCI-DSS, GDPR, HTNG |
| [StreamVault](../StreamVault/AI_PLAN.md) | `stream-ai-*` (+ encode pool) | Recs, encode, captions, moderation | DRM, geo-licensing, CVAA/EAA |
| [TalentBridge](../TalentBridge/AI_PLAN.md) | `talent-ai-*` (region-sharded) | CV parsing, attrition, pay-equity | EU AI Act high-risk, GDPR, EEOC |

The agent-platform code (Paperclip, OpenClaw, NemoClaw) is built once in ShopOS and consumed as a versioned Helm chart by each sibling. Project-specific guardrail policies (clinical-safe, financial-safe, youth-safe, etc.) are implemented as NemoClaw policy bundles per sibling and never cross project boundaries.
