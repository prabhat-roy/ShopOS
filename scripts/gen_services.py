#!/usr/bin/env python3
"""Generate skeleton microservices + Helm charts for ShopOS."""
import os

BASE = r"c:/Users/prabh/Desktop/Project/ShopOS"

def w(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", newline="\n") as f:
        f.write(content)

# (name, domain, port, db, language, description, k8s_namespace, domain_label)
SERVICES = [
    # ── Group 2: Catalog ────────────────────────────────────────────────────
    ("product-label-service",          "catalog",             50218, "postgres", "go",     "Product eco/cert label management",               "shopos-catalog",           "catalog"),
    ("variant-service",                "catalog",             50219, "postgres", "go",     "Product variant (size/colour matrix) management", "shopos-catalog",           "catalog"),
    ("stock-reservation-service",      "catalog",             50220, "redis",    "go",     "Soft-reserve inventory during checkout TTL",       "shopos-catalog",           "catalog"),
    # ── Group 2: Commerce ───────────────────────────────────────────────────
    ("split-payment-service",          "commerce",            50221, "postgres", "go",     "Split bill across multiple payment methods",       "shopos-commerce",          "commerce"),
    ("installment-service",            "commerce",            50222, "postgres", "go",     "EMI calculation and scheduling",                   "shopos-commerce",          "commerce"),
    ("dynamic-pricing-service",        "commerce",            50223, "postgres", "go",     "Real-time demand-based price adjustment",          "shopos-commerce",          "commerce"),
    ("coupon-service",                 "commerce",            50224, "postgres", "go",     "Coupon code lifecycle management",                 "shopos-commerce",          "commerce"),
    ("order-amendment-service",        "commerce",            50225, "postgres", "go",     "Post-order edits before fulfillment",              "shopos-commerce",          "commerce"),
    # ── Group 3: Supply Chain ───────────────────────────────────────────────
    ("route-optimization-service",     "supply-chain",        50226, "postgres", "go",     "Last-mile delivery route optimization",            "shopos-supply-chain",      "supply-chain"),
    ("packaging-service",              "supply-chain",        50227, "postgres", "go",     "Packaging selection and eco/gift wrapping rules",  "shopos-supply-chain",      "supply-chain"),
    ("cross-dock-service",             "supply-chain",        50228, "postgres", "go",     "Cross-docking workflow management",                "shopos-supply-chain",      "supply-chain"),
    ("duty-drawback-service",          "supply-chain",        50229, "postgres", "go",     "Export duty refund claims management",             "shopos-supply-chain",      "supply-chain"),
    # ── Group 3: Financial ──────────────────────────────────────────────────
    ("escrow-service",                 "financial",           50230, "postgres", "go",     "Marketplace escrow — hold funds until delivery",   "shopos-financial",         "financial"),
    ("forex-service",                  "financial",           50231, "postgres", "go",     "Foreign exchange rate feeds and conversion",       "shopos-financial",         "financial"),
    ("audit-trail-service",            "financial",           50232, "postgres", "java",   "Financial audit trail with immutable ledger",      "shopos-financial",         "financial"),
    ("dunning-service",                "financial",           50233, "postgres", "go",     "Failed payment retry and escalation logic",        "shopos-financial",         "financial"),
    # ── Group 4: Customer Experience ────────────────────────────────────────
    ("loyalty-tier-service",           "customer-experience", 50234, "postgres", "go",     "Loyalty tier calculation and upgrade logic",       "shopos-cx",                "customer-experience"),
    ("accessibility-service",          "customer-experience", 50235, "none",     "nodejs", "WCAG compliance hints and alt-text generation",    "shopos-cx",                "customer-experience"),
    ("return-portal-service",          "customer-experience", 50236, "postgres", "go",     "Customer-facing self-serve returns portal",        "shopos-cx",                "customer-experience"),
    # ── Group 4: Communications ─────────────────────────────────────────────
    ("telegram-service",               "communications",      8197,  "none",     "nodejs", "Telegram Bot API notifications",                   "shopos-communications",    "communications"),
    ("voice-service",                  "communications",      8198,  "none",     "go",     "IVR and voice call notifications",                 "shopos-communications",    "communications"),
    ("webhook-delivery-service",       "communications",      8199,  "postgres", "go",     "Partner/merchant webhook delivery with retry",     "shopos-communications",    "communications"),
    # ── Group 4: Content ────────────────────────────────────────────────────
    ("ab-content-service",             "content",             50240, "postgres", "go",     "A/B test content variants",                        "shopos-content",           "content"),
    # ── Group 5: B2B ────────────────────────────────────────────────────────
    ("rfp-service",                    "b2b",                 50241, "postgres", "go",     "Request for Proposal lifecycle management",        "shopos-b2b",               "b2b"),
    ("vendor-onboarding-service",      "b2b",                 50242, "postgres", "java",   "Vendor KYC, document verification and setup",      "shopos-b2b",               "b2b"),
    ("purchase-requisition-service",   "b2b",                 50243, "postgres", "kotlin", "Internal purchase request approval workflow",       "shopos-b2b",               "b2b"),
    # ── Group 5: Integrations ───────────────────────────────────────────────
    ("webhook-ingestion-service",      "integrations",        8200,  "postgres", "go",     "Inbound webhooks from Stripe/Shopify to Kafka",    "shopos-integrations",      "integrations"),
    ("etl-service",                    "integrations",        8201,  "postgres", "python", "Batch ETL from external systems to internal DBs",  "shopos-integrations",      "integrations"),
    ("data-sync-service",              "integrations",        50244, "postgres", "go",     "Bidirectional sync between internal and partner",  "shopos-integrations",      "integrations"),
    ("ipaas-connector-service",        "integrations",        8202,  "none",     "go",     "Generic iPaaS adapter (Zapier/Make/n8n protocol)", "shopos-integrations",      "integrations"),
    # ── Group 5: Affiliate ──────────────────────────────────────────────────
    ("click-tracking-service",         "affiliate",           8203,  "redis",    "go",     "Affiliate link click capture with UTM tracking",   "shopos-affiliate",         "affiliate"),
    ("fraud-prevention-affiliate-service","affiliate",        50248, "postgres", "go",     "Affiliate fraud detection (click stuffing etc.)",  "shopos-affiliate",         "affiliate"),
    # ── NEW DOMAIN: marketplace ─────────────────────────────────────────────
    ("seller-registration-service",    "marketplace",         50250, "postgres", "go",     "Seller onboarding and registration",               "shopos-marketplace",       "marketplace"),
    ("listing-approval-service",       "marketplace",         50251, "postgres", "go",     "Product listing review and approval workflow",     "shopos-marketplace",       "marketplace"),
    ("marketplace-commission-service", "marketplace",         50252, "postgres", "go",     "Marketplace commission rules and calculation",     "shopos-marketplace",       "marketplace"),
    ("dispute-resolution-service",     "marketplace",         50253, "postgres", "java",   "Buyer-seller dispute resolution workflow",         "shopos-marketplace",       "marketplace"),
    ("seller-analytics-service",       "marketplace",         8204,  "postgres", "go",     "Seller-facing sales and performance analytics",    "shopos-marketplace",       "marketplace"),
    ("product-syndication-service",    "marketplace",         50254, "postgres", "go",     "Product feed syndication to external channels",    "shopos-marketplace",       "marketplace"),
    ("storefront-service",             "marketplace",         8205,  "none",     "nodejs", "Multi-seller storefront rendering and routing",    "shopos-marketplace",       "marketplace"),
    ("seller-payout-service",          "marketplace",         50255, "postgres", "go",     "Seller payout scheduling and disbursement",        "shopos-marketplace",       "marketplace"),
    # ── NEW DOMAIN: gamification ────────────────────────────────────────────
    ("points-service",                 "gamification",        50260, "redis",    "go",     "User points ledger and transaction log",           "shopos-gamification",      "gamification"),
    ("badge-service",                  "gamification",        50261, "postgres", "go",     "Badge definition, award and display logic",        "shopos-gamification",      "gamification"),
    ("leaderboard-service",            "gamification",        50262, "redis",    "go",     "Real-time global and segment leaderboards",        "shopos-gamification",      "gamification"),
    ("challenge-service",              "gamification",        50263, "postgres", "go",     "Gamified challenge definition and completion",     "shopos-gamification",      "gamification"),
    ("reward-redemption-service",      "gamification",        50264, "postgres", "go",     "Reward catalogue and points redemption",           "shopos-gamification",      "gamification"),
    ("streak-service",                 "gamification",        50265, "redis",    "go",     "Daily/weekly activity streak tracking",            "shopos-gamification",      "gamification"),
    # ── NEW DOMAIN: developer-platform ──────────────────────────────────────
    ("api-management-service",         "developer-platform",  8206,  "postgres", "go",     "External API lifecycle, versioning and keys",      "shopos-developer",         "developer-platform"),
    ("sandbox-service",                "developer-platform",  8207,  "postgres", "go",     "Isolated sandbox environment for partner testing", "shopos-developer",         "developer-platform"),
    ("developer-portal-backend",       "developer-platform",  8208,  "postgres", "nodejs", "Backend API for developer documentation portal",   "shopos-developer",         "developer-platform"),
    ("oauth-client-service",           "developer-platform",  50270, "postgres", "go",     "OAuth 2.0 client registration and token issuance", "shopos-developer",         "developer-platform"),
    ("api-analytics-service",          "developer-platform",  8209,  "postgres", "go",     "Per-key API usage metrics and quota enforcement",  "shopos-developer",         "developer-platform"),
    ("webhook-management-service",     "developer-platform",  50271, "postgres", "go",     "Developer webhook subscription and retry config",  "shopos-developer",         "developer-platform"),
    # ── NEW DOMAIN: compliance ───────────────────────────────────────────────
    ("data-retention-service",         "compliance",          50280, "postgres", "go",     "Automated data retention and purge scheduling",    "shopos-compliance",        "compliance"),
    ("consent-audit-service",          "compliance",          50281, "postgres", "go",     "Consent event audit log with immutable trail",     "shopos-compliance",        "compliance"),
    ("privacy-request-service",        "compliance",          50282, "postgres", "go",     "GDPR/CCPA subject access and erasure requests",    "shopos-compliance",        "compliance"),
    ("compliance-reporting-service",   "compliance",          50283, "postgres", "java",   "Regulatory compliance report generation",          "shopos-compliance",        "compliance"),
    ("data-lineage-service",           "compliance",          50284, "postgres", "go",     "Data flow lineage tracking across services",       "shopos-compliance",        "compliance"),
    # ── NEW DOMAIN: sustainability ───────────────────────────────────────────
    ("carbon-tracker-service",         "sustainability",      50290, "postgres", "go",     "Carbon footprint calculation per order/shipment",  "shopos-sustainability",    "sustainability"),
    ("eco-score-service",              "sustainability",      50291, "postgres", "go",     "Product eco-score computation and labelling",      "shopos-sustainability",    "sustainability"),
    ("green-shipping-service",         "sustainability",      50292, "postgres", "go",     "Green shipping option selection and optimisation", "shopos-sustainability",    "sustainability"),
    ("sustainability-reporting-service","sustainability",     50293, "postgres", "go",     "ESG and sustainability KPI reporting",             "shopos-sustainability",    "sustainability"),
    ("offset-service",                 "sustainability",      50294, "postgres", "go",     "Carbon offset purchase and certificate issuance",  "shopos-sustainability",    "sustainability"),
]

# ─── Go templates ────────────────────────────────────────────────────────────

def go_main(name, port):
    return f'''package main

import (
\t"context"
\t"encoding/json"
\t"log/slog"
\t"net/http"
\t"os"
\t"os/signal"
\t"syscall"
\t"time"
)

func main() {{
\tport := os.Getenv("PORT")
\tif port == "" {{
\t\tport = "{port}"
\t}}

\tmux := http.NewServeMux()
\tmux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {{
\t\tw.Header().Set("Content-Type", "application/json")
\t\t_ = json.NewEncoder(w).Encode(map[string]string{{"status": "ok"}})
\t}})
\tmux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {{
\t\tw.Header().Set("Content-Type", "text/plain")
\t\t_, _ = w.Write([]byte("# placeholder metrics\\n"))
\t}})

\tsrv := &http.Server{{
\t\tAddr:         ":" + port,
\t\tHandler:      mux,
\t\tReadTimeout:  15 * time.Second,
\t\tWriteTimeout: 15 * time.Second,
\t\tIdleTimeout:  60 * time.Second,
\t}}

\tctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
\tdefer stop()

\tgo func() {{
\t\t<-ctx.Done()
\t\tshutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
\t\tdefer cancel()
\t\t_ = srv.Shutdown(shutCtx)
\t}}()

\tslog.Info("{name} listening", "port", port)
\tif err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {{
\t\tslog.Error("server error", "err", err)
\t\tos.Exit(1)
\t}}
}}
'''

def go_mod(name):
    return f'''module github.com/shopos/{name}

go 1.24

require (
\tgithub.com/joho/godotenv v1.5.1
\tgo.uber.org/zap v1.27.0
)
'''

def go_dockerfile(name, port):
    return f'''FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \\
    rm -f go.sum && GONOSUMDB=* go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \\
    --mount=type=cache,target=/root/.cache/go-build \\
    rm -f go.sum && GONOSUMDB=* go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o {name} .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/{name} /{name}
USER nonroot:nonroot
EXPOSE {port}
ENTRYPOINT ["/{name}"]
'''

# ─── Java templates ──────────────────────────────────────────────────────────

def java_pkg(name):
    return name.replace("-", "").lower()

def java_main(name):
    pkg = java_pkg(name)
    cls = "".join(p.capitalize() for p in name.split("-"))
    return f'''package com.enterprise.{pkg};

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class {cls}Application {{
    public static void main(String[] args) {{
        SpringApplication.run({cls}Application.class, args);
    }}
}}
'''

def java_health(name):
    pkg = java_pkg(name)
    return f'''package com.enterprise.{pkg};

import org.springframework.web.bind.annotation.*;
import java.util.Map;

@RestController
public class HealthController {{
    @GetMapping("/healthz")
    public Map<String, String> health() {{
        return Map.of("status", "ok");
    }}

    @GetMapping("/metrics")
    public String metrics() {{
        return "# placeholder metrics\\n";
    }}
}}
'''

def java_pom(name):
    return f'''<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.3.4</version>
    </parent>
    <groupId>com.enterprise</groupId>
    <artifactId>{name}</artifactId>
    <version>1.0.0</version>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-actuator</artifactId>
        </dependency>
    </dependencies>
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
'''

def java_dockerfile(port):
    return f'''FROM maven:3.9-eclipse-temurin-21-alpine AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -q
COPY src ./src
RUN mvn package -DskipTests -q

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE {port}
ENTRYPOINT ["java", "-jar", "app.jar"]
'''

def java_makefile(name):
    return f'''SERVICE_NAME := {name}
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build test docker-build docker-push

build:
\tmvn package -DskipTests

test:
\tmvn test

docker-build:
\tdocker build -t $(IMAGE) .

docker-push:
\tdocker push $(IMAGE)
'''

# ─── Kotlin templates ────────────────────────────────────────────────────────

def kotlin_pkg(name):
    return name.replace("-", "").lower()

def kotlin_main(name):
    pkg = kotlin_pkg(name)
    cls = "".join(p.capitalize() for p in name.split("-"))
    return f'''package com.enterprise.{pkg}

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication
class {cls}Application

fun main(args: Array<String>) {{
    runApplication<{cls}Application>(*args)
}}
'''

def kotlin_health(name):
    pkg = kotlin_pkg(name)
    return f'''package com.enterprise.{pkg}

import org.springframework.web.bind.annotation.*

@RestController
class HealthController {{
    @GetMapping("/healthz")
    fun health() = mapOf("status" to "ok")

    @GetMapping("/metrics")
    fun metrics() = "# placeholder metrics\\n"
}}
'''

def kotlin_gradle(name):
    return f'''plugins {{
    id("org.springframework.boot") version "3.3.4"
    id("io.spring.dependency-management") version "1.1.6"
    kotlin("jvm") version "2.0.20"
    kotlin("plugin.spring") version "2.0.20"
}}

group = "com.enterprise"
version = "1.0.0"

repositories {{
    mavenCentral()
}}

dependencies {{
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.jetbrains.kotlin:kotlin-reflect")
}}
'''

def kotlin_dockerfile(port):
    return f'''FROM gradle:8.10-jdk21-alpine AS builder
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts* ./
RUN gradle dependencies --no-daemon -q || true
COPY src ./src
RUN gradle bootJar --no-daemon -q

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar app.jar
EXPOSE {port}
ENTRYPOINT ["java", "-jar", "app.jar"]
'''

def kotlin_makefile(name):
    return f'''SERVICE_NAME := {name}
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build test docker-build docker-push

build:
\tgradle bootJar --no-daemon

test:
\tgradle test --no-daemon

docker-build:
\tdocker build -t $(IMAGE) .

docker-push:
\tdocker push $(IMAGE)
'''

# ─── Node.js templates ───────────────────────────────────────────────────────

def nodejs_index(name, port):
    return f'''\'use strict\'

const express = require(\'express\')
const app = express()
const port = process.env.PORT || {port}

app.use(express.json())

app.get(\'/healthz\', (_req, res) => res.json({{ status: \'ok\' }}))
app.get(\'/metrics\', (_req, res) => {{
  res.set(\'Content-Type\', \'text/plain\')
  res.send(\'# placeholder metrics\\n\')
}})

app.listen(port, () => console.log(`{name} listening on port ${{port}}`))
'''

def nodejs_package(name):
    return f'''{{"name": "{name}", "version": "1.0.0", "main": "index.js", "scripts": {{"start": "node index.js", "dev": "nodemon index.js", "test": "jest"}}, "dependencies": {{"express": "^4.21.0"}}, "devDependencies": {{"nodemon": "^3.1.7", "jest": "^29.7.0"}}}}
'''

def nodejs_dockerfile(name, port):
    return f'''FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
USER node
EXPOSE {port}
CMD ["node", "index.js"]
'''

def nodejs_makefile(name):
    return f'''SERVICE_NAME := {name}
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: install start test docker-build docker-push

install:
\tnpm install

start:
\tnode index.js

test:
\tnpm test

docker-build:
\tdocker build -t $(IMAGE) .

docker-push:
\tdocker push $(IMAGE)
'''

# ─── Python templates ────────────────────────────────────────────────────────

def python_main(name, port):
    return f'''from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
import uvicorn
import os

app = FastAPI(title="{name}")


@app.get("/healthz")
def healthz():
    return {{"status": "ok"}}


@app.get("/metrics", response_class=PlainTextResponse)
def metrics():
    return "# placeholder metrics\\n"


if __name__ == "__main__":
    port = int(os.getenv("PORT", "{port}"))
    uvicorn.run(app, host="0.0.0.0", port=port)
'''

def python_requirements():
    return '''fastapi==0.115.0
uvicorn[standard]==0.30.6
python-dotenv==1.0.1
'''

def python_dockerfile(port):
    return f'''FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /root/.local /root/.local
COPY . .
ENV PATH=/root/.local/bin:$PATH
EXPOSE {port}
CMD ["python", "main.py"]
'''

def python_makefile(name):
    return f'''SERVICE_NAME := {name}
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: install run test docker-build docker-push

install:
\tpip install -r requirements.txt

run:
\tpython main.py

test:
\tpytest

docker-build:
\tdocker build -t $(IMAGE) .

docker-push:
\tdocker push $(IMAGE)
'''

# ─── Shared .env.example ────────────────────────────────────────────────────

def env_example(port, db):
    lines = [f"PORT={port}", "LOG_LEVEL=info", "LOG_FORMAT=json"]
    if db == "postgres":
        lines.append("DATABASE_URL=postgres://postgres:postgres@localhost:5432/shopos?sslmode=disable")
    if db == "redis":
        lines.append("REDIS_URL=redis://localhost:6379")
    lines += [
        "KAFKA_BROKERS=localhost:9092",
        "OTEL_ENABLED=false",
        "OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318",
    ]
    return "\n".join(lines) + "\n"

# ─── Makefile (Go) ───────────────────────────────────────────────────────────

def go_makefile(name):
    return f'''SERVICE_NAME := {name}
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build run test lint docker-build docker-push

build:
\tgo build -o $(SERVICE_NAME) .

run:
\tgo run .

test:
\tgo test ./...

lint:
\tgolangci-lint run ./...

docker-build:
\tdocker build -t $(IMAGE) .

docker-push:
\tdocker push $(IMAGE)
'''

# ─── Helm chart templates ────────────────────────────────────────────────────

def helm_chart_yaml(name, desc, domain_label, lang):
    return f'''apiVersion: v2
name: {name}
description: ShopOS {name} — {desc}
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - shopos
  - {name}
  - {domain_label}
home: https://github.com/your-org/shopos
sources:
  - https://github.com/your-org/shopos/tree/main/src/{domain_label}/{name}
maintainers:
  - name: ShopOS Platform Team
    email: platform@shopos.io
annotations:
  app.kubernetes.io/domain: {domain_label}
  app.kubernetes.io/language: {lang}
'''

def helm_values(name, port, ns, domain_label):
    return f'''replicaCount: 1

image:
  repository: shopos/{name}
  tag: latest
  pullPolicy: IfNotPresent

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  name: shopos-service-account
  create: false
  annotations: {{}}

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "{port}"
  prometheus.io/path: "/metrics"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

service:
  type: ClusterIP
  port: {port}
  targetPort: {port}
  protocol: TCP
  annotations: {{}}

ingress:
  enabled: false
  className: nginx
  annotations: {{}}
  hosts: []
  tls: []

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
    scaleUp:
      stabilizationWindowSeconds: 60

livenessProbe:
  httpGet:
    path: /healthz
    port: {port}
  initialDelaySeconds: 15
  periodSeconds: 20
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /healthz
    port: {port}
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3

strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0

terminationGracePeriodSeconds: 30

config:
  environment: production
  kafkaBrokers: "kafka-0.kafka-headless.shopos-infra.svc.cluster.local:9092"
  redisUrl: "redis://redis-master.shopos-infra.svc.cluster.local:6379"
  logLevel: info
  logFormat: json
  otelEnabled: "false"
  otelEndpoint: "http://otel-collector.shopos-infra.svc.cluster.local:4318"

namespaceOverride: "{ns}"

extraEnv: []
extraVolumes: []
extraVolumeMounts: []
nodeSelector: {{}}
tolerations: []
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          topologyKey: kubernetes.io/hostname
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: {name}
'''

def helm_helpers(name, domain_label):
    return f'''{{{{/*
ShopOS — {name} Helm helper templates
*/}}}}

{{{{- define "{name}.name" -}}}}
{{{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}}}
{{{{- end }}}}

{{{{- define "{name}.fullname" -}}}}
{{{{- if .Values.fullnameOverride }}}}
{{{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}}}
{{{{- else }}}}
{{{{- $name := default .Chart.Name .Values.nameOverride }}}}
{{{{- if contains $name .Release.Name }}}}
{{{{- .Release.Name | trunc 63 | trimSuffix "-" }}}}
{{{{- else }}}}
{{{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}}}
{{{{- end }}}}
{{{{- end }}}}
{{{{- end }}}}

{{{{- define "{name}.chart" -}}}}
{{{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}}}
{{{{- end }}}}

{{{{- define "{name}.labels" -}}}}
helm.sh/chart: {{{{- include "{name}.chart" . }}}}
{{{{- include "{name}.selectorLabels" . }}}}
{{{{- if .Chart.AppVersion }}}}
app.kubernetes.io/version: {{{{- .Chart.AppVersion | quote }}}}
{{{{- end }}}}
app.kubernetes.io/managed-by: {{{{- .Release.Service }}}}
app.kubernetes.io/domain: {domain_label}
{{{{- end }}}}

{{{{- define "{name}.selectorLabels" -}}}}
app.kubernetes.io/name: {{{{- include "{name}.name" . }}}}
app.kubernetes.io/instance: {{{{- .Release.Name }}}}
{{{{- end }}}}
'''

def helm_deployment(name):
    return f'''apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{{{- include "{name}.fullname" . }}}}
  namespace: {{{{- .Values.namespaceOverride | default .Release.Namespace }}}}
  labels:
    {{{{- include "{name}.labels" . | nindent 4 }}}}
spec:
  {{{{- if not .Values.autoscaling.enabled }}}}
  replicas: {{{{- .Values.replicaCount }}}}
  {{{{- end }}}}
  selector:
    matchLabels:
      {{{{- include "{name}.selectorLabels" . | nindent 6 }}}}
  strategy:
    type: {{{{- .Values.strategy.type }}}}
    {{{{- if eq .Values.strategy.type "RollingUpdate" }}}}
    rollingUpdate:
      maxSurge: {{{{- .Values.strategy.rollingUpdate.maxSurge }}}}
      maxUnavailable: {{{{- .Values.strategy.rollingUpdate.maxUnavailable }}}}
    {{{{- end }}}}
  template:
    metadata:
      annotations:
        {{{{- with .Values.podAnnotations }}}}
        {{{{- toYaml . | nindent 8 }}}}
        {{{{- end }}}}
      labels:
        {{{{- include "{name}.selectorLabels" . | nindent 8 }}}}
    spec:
      serviceAccountName: {{{{- .Values.serviceAccount.name }}}}
      securityContext:
        {{{{- toYaml .Values.podSecurityContext | nindent 8 }}}}
      terminationGracePeriodSeconds: {{{{- .Values.terminationGracePeriodSeconds }}}}
      containers:
        - name: {{{{- .Chart.Name }}}}
          image: "{{{{- .Values.image.repository }}}}:{{{{- .Values.image.tag | default .Chart.AppVersion }}}}"
          imagePullPolicy: {{{{- .Values.image.pullPolicy }}}}
          ports:
            - name: http
              containerPort: {{{{- .Values.service.targetPort }}}}
              protocol: {{{{- .Values.service.protocol }}}}
          env:
            - name: PORT
              value: {{{{- .Values.service.targetPort | quote }}}}
            - name: ENVIRONMENT
              value: {{{{- .Values.config.environment | quote }}}}
            - name: LOG_LEVEL
              value: {{{{- .Values.config.logLevel | quote }}}}
            - name: OTEL_ENABLED
              value: {{{{- .Values.config.otelEnabled | quote }}}}
            - name: OTEL_EXPORTER_OTLP_ENDPOINT
              value: {{{{- .Values.config.otelEndpoint | quote }}}}
            {{{{- with .Values.extraEnv }}}}
            {{{{- toYaml . | nindent 12 }}}}
            {{{{- end }}}}
          securityContext:
            {{{{- toYaml .Values.securityContext | nindent 12 }}}}
          resources:
            {{{{- toYaml .Values.resources | nindent 12 }}}}
          livenessProbe:
            {{{{- toYaml .Values.livenessProbe | nindent 12 }}}}
          readinessProbe:
            {{{{- toYaml .Values.readinessProbe | nindent 12 }}}}
          volumeMounts:
            - name: tmp
              mountPath: /tmp
            {{{{- with .Values.extraVolumeMounts }}}}
            {{{{- toYaml . | nindent 12 }}}}
            {{{{- end }}}}
      volumes:
        - name: tmp
          emptyDir: {{}}
        {{{{- with .Values.extraVolumes }}}}
        {{{{- toYaml . | nindent 8 }}}}
        {{{{- end }}}}
      {{{{- with .Values.nodeSelector }}}}
      nodeSelector:
        {{{{- toYaml . | nindent 8 }}}}
      {{{{- end }}}}
      {{{{- with .Values.affinity }}}}
      affinity:
        {{{{- toYaml . | nindent 8 }}}}
      {{{{- end }}}}
      {{{{- with .Values.tolerations }}}}
      tolerations:
        {{{{- toYaml . | nindent 8 }}}}
      {{{{- end }}}}
'''

def helm_service(name):
    return f'''apiVersion: v1
kind: Service
metadata:
  name: {{{{- include "{name}.fullname" . }}}}
  namespace: {{{{- .Values.namespaceOverride | default .Release.Namespace }}}}
  labels:
    {{{{- include "{name}.labels" . | nindent 4 }}}}
spec:
  type: {{{{- .Values.service.type }}}}
  selector:
    {{{{- include "{name}.selectorLabels" . | nindent 4 }}}}
  ports:
    - name: http
      port: {{{{- .Values.service.port }}}}
      targetPort: {{{{- .Values.service.targetPort }}}}
      protocol: {{{{- .Values.service.protocol }}}}
'''

# ─── Main generation loop ────────────────────────────────────────────────────

created = 0
for (name, domain, port, db, lang, desc, ns, domain_label) in SERVICES:
    src = f"{BASE}/src/{domain}/{name}"
    helm = f"{BASE}/helm/charts/{name}"

    # ── Source files ────────────────────────────────────────────────────────
    if lang == "go":
        w(f"{src}/main.go",    go_main(name, port))
        w(f"{src}/go.mod",     go_mod(name))
        w(f"{src}/Dockerfile", go_dockerfile(name, port))
        w(f"{src}/Makefile",   go_makefile(name))

    elif lang == "java":
        pkg_path = java_pkg(name)
        w(f"{src}/src/main/java/com/enterprise/{pkg_path}/Application.java",     java_main(name))
        w(f"{src}/src/main/java/com/enterprise/{pkg_path}/HealthController.java", java_health(name))
        w(f"{src}/pom.xml",     java_pom(name))
        w(f"{src}/Dockerfile",  java_dockerfile(port))
        w(f"{src}/Makefile",    java_makefile(name))

    elif lang == "kotlin":
        pkg_path = kotlin_pkg(name)
        w(f"{src}/src/main/kotlin/com/enterprise/{pkg_path}/Application.kt",     kotlin_main(name))
        w(f"{src}/src/main/kotlin/com/enterprise/{pkg_path}/HealthController.kt", kotlin_health(name))
        w(f"{src}/build.gradle.kts", kotlin_gradle(name))
        w(f"{src}/Dockerfile",       kotlin_dockerfile(port))
        w(f"{src}/Makefile",         kotlin_makefile(name))

    elif lang == "nodejs":
        w(f"{src}/index.js",    nodejs_index(name, port))
        w(f"{src}/package.json", nodejs_package(name))
        w(f"{src}/Dockerfile",  nodejs_dockerfile(name, port))
        w(f"{src}/Makefile",    nodejs_makefile(name))

    elif lang == "python":
        w(f"{src}/main.py",          python_main(name, port))
        w(f"{src}/requirements.txt", python_requirements())
        w(f"{src}/Dockerfile",       python_dockerfile(port))
        w(f"{src}/Makefile",         python_makefile(name))

    w(f"{src}/.env.example", env_example(port, db))

    # ── Helm chart ──────────────────────────────────────────────────────────
    lang_label = {"go": "Go", "java": "Java", "kotlin": "Kotlin", "nodejs": "Node.js", "python": "Python"}[lang]
    w(f"{helm}/Chart.yaml",                   helm_chart_yaml(name, desc, domain_label, lang_label))
    w(f"{helm}/values.yaml",                  helm_values(name, port, ns, domain_label))
    w(f"{helm}/templates/_helpers.tpl",       helm_helpers(name, domain_label))
    w(f"{helm}/templates/deployment.yaml",    helm_deployment(name))
    w(f"{helm}/templates/service.yaml",       helm_service(name))

    created += 1
    print(f"[{created:02d}/{len(SERVICES)}] {domain}/{name} ({lang}:{port})")

print(f"\nDone — {created} services generated.")
