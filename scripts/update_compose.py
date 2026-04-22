#!/usr/bin/env python3
"""Append new service entries to docker-compose.yml before the volumes: section."""
import re

BASE = r"c:/Users/prabh/Desktop/Project/ShopOS"
COMPOSE = f"{BASE}/docker-compose.yml"

# (name, domain, port, db_deps, extra_env)
# db_deps: list of docker-compose service names this depends on
SERVICES = [
    # ── Group 1 already done by agent — 8 platform/identity services ────────
    ("notification-preferences-service", "platform",            50210, ["postgres"],         "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("circuit-breaker-service",          "platform",            50211, ["redis"],             "REDIS_URL=redis://redis:6379"),
    ("idempotency-service",              "platform",            50212, ["redis"],             "REDIS_URL=redis://redis:6379"),
    ("correlation-id-service",           "platform",            50213, [],                   ""),
    ("data-masking-service",             "platform",            50214, [],                   ""),
    ("sso-service",                      "identity",            50215, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("password-policy-service",          "identity",            50216, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("bot-detection-service",            "identity",            50217, ["redis"],             "REDIS_URL=redis://redis:6379"),
    # ── Group 2: Catalog ────────────────────────────────────────────────────
    ("product-label-service",            "catalog",             50218, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("variant-service",                  "catalog",             50219, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("stock-reservation-service",        "catalog",             50220, ["redis"],             "REDIS_URL=redis://redis:6379"),
    # ── Group 2: Commerce ───────────────────────────────────────────────────
    ("split-payment-service",            "commerce",            50221, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("installment-service",              "commerce",            50222, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("dynamic-pricing-service",          "commerce",            50223, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("coupon-service",                   "commerce",            50224, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("order-amendment-service",          "commerce",            50225, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 3: Supply Chain ───────────────────────────────────────────────
    ("route-optimization-service",       "supply-chain",        50226, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("packaging-service",                "supply-chain",        50227, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("cross-dock-service",               "supply-chain",        50228, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("duty-drawback-service",            "supply-chain",        50229, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 3: Financial ──────────────────────────────────────────────────
    ("escrow-service",                   "financial",           50230, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("forex-service",                    "financial",           50231, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("audit-trail-service",              "financial",           50232, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("dunning-service",                  "financial",           50233, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 4: Customer Experience ────────────────────────────────────────
    ("loyalty-tier-service",             "customer-experience", 50234, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("accessibility-service",            "customer-experience", 50235, [],                   ""),
    ("return-portal-service",            "customer-experience", 50236, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 4: Communications ─────────────────────────────────────────────
    ("telegram-service",                 "communications",      8197,  ["kafka"],             "KAFKA_BROKERS=kafka:9092"),
    ("voice-service",                    "communications",      8198,  [],                   ""),
    ("webhook-delivery-service",         "communications",      8199,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 4: Content ────────────────────────────────────────────────────
    ("ab-content-service",               "content",             50240, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 5: B2B ────────────────────────────────────────────────────────
    ("rfp-service",                      "b2b",                 50241, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("vendor-onboarding-service",        "b2b",                 50242, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("purchase-requisition-service",     "b2b",                 50243, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── Group 5: Integrations ───────────────────────────────────────────────
    ("webhook-ingestion-service",        "integrations",        8200,  ["postgres", "kafka"], "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable\n      - KAFKA_BROKERS=kafka:9092"),
    ("etl-service",                      "integrations",        8201,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("data-sync-service",                "integrations",        50244, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("ipaas-connector-service",          "integrations",        8202,  [],                   ""),
    # ── Group 5: Affiliate ──────────────────────────────────────────────────
    ("click-tracking-service",           "affiliate",           8203,  ["redis"],             "REDIS_URL=redis://redis:6379"),
    ("fraud-prevention-affiliate-service","affiliate",          50248, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── NEW DOMAIN: marketplace ─────────────────────────────────────────────
    ("seller-registration-service",      "marketplace",         50250, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("listing-approval-service",         "marketplace",         50251, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("marketplace-commission-service",   "marketplace",         50252, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("dispute-resolution-service",       "marketplace",         50253, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("seller-analytics-service",         "marketplace",         8204,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("product-syndication-service",      "marketplace",         50254, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("storefront-service",               "marketplace",         8205,  [],                   ""),
    ("seller-payout-service",            "marketplace",         50255, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── NEW DOMAIN: gamification ────────────────────────────────────────────
    ("points-service",                   "gamification",        50260, ["redis"],             "REDIS_URL=redis://redis:6379"),
    ("badge-service",                    "gamification",        50261, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("leaderboard-service",              "gamification",        50262, ["redis"],             "REDIS_URL=redis://redis:6379"),
    ("challenge-service",                "gamification",        50263, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("reward-redemption-service",        "gamification",        50264, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("streak-service",                   "gamification",        50265, ["redis"],             "REDIS_URL=redis://redis:6379"),
    # ── NEW DOMAIN: developer-platform ──────────────────────────────────────
    ("api-management-service",           "developer-platform",  8206,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("sandbox-service",                  "developer-platform",  8207,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("developer-portal-backend",         "developer-platform",  8208,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("oauth-client-service",             "developer-platform",  50270, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("api-analytics-service",            "developer-platform",  8209,  ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("webhook-management-service",       "developer-platform",  50271, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── NEW DOMAIN: compliance ───────────────────────────────────────────────
    ("data-retention-service",           "compliance",          50280, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("consent-audit-service",            "compliance",          50281, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("privacy-request-service",          "compliance",          50282, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("compliance-reporting-service",     "compliance",          50283, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("data-lineage-service",             "compliance",          50284, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    # ── NEW DOMAIN: sustainability ───────────────────────────────────────────
    ("carbon-tracker-service",           "sustainability",      50290, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("eco-score-service",                "sustainability",      50291, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("green-shipping-service",           "sustainability",      50292, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("sustainability-reporting-service", "sustainability",      50293, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
    ("offset-service",                   "sustainability",      50294, ["postgres"],          "DATABASE_URL=postgres://postgres:postgres@postgres:5432/shopos?sslmode=disable"),
]

def make_entry(name, domain, port, db_deps, extra_env):
    lines = [
        f"  {name}:",
        f"    image: shopos/{name}:latest",
        f"    build:",
        f"      context: ./src/{domain}/{name}",
        f"      dockerfile: Dockerfile",
        f"    environment:",
        f"      - PORT={port}",
        f"      - LOG_LEVEL=info",
    ]
    if extra_env:
        for part in extra_env.split("\n"):
            lines.append(f"      - {part.strip()}")
    lines.append(f"    ports:")
    lines.append(f"      - \"{port}:{port}\"")
    if db_deps:
        lines.append(f"    depends_on:")
        for dep in db_deps:
            lines.append(f"      - {dep}")
    lines.append(f"    restart: unless-stopped")
    lines.append(f"    networks:")
    lines.append(f"      - shopos")
    lines.append("")
    return "\n".join(lines)

# Check which services already exist in compose
with open(COMPOSE, "r") as f:
    content = f.read()

# Build entries for services not already present
new_entries = []
for (name, domain, port, db_deps, extra_env) in SERVICES:
    key = f"  {name}:"
    if key not in content:
        new_entries.append(make_entry(name, domain, port, db_deps, extra_env))
    else:
        print(f"  SKIP (exists): {name}")

if not new_entries:
    print("Nothing to add.")
else:
    insertion = "\n".join(new_entries)
    # Insert before the volumes: section
    new_content = content.replace("\n# VOLUMES\n", f"\n{insertion}\n# VOLUMES\n", 1)
    with open(COMPOSE, "w", newline="\n") as f:
        f.write(new_content)
    print(f"Added {len(new_entries)} services to docker-compose.yml")
