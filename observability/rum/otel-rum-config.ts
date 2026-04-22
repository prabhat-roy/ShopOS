/**
 * OpenTelemetry RUM (Real User Monitoring) configuration
 * Shared across all frontend apps. Import and call initRum() in app entry point.
 *
 * Sends traces/metrics to OTel Collector which routes to Jaeger + Prometheus.
 */
import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { BatchSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { ZoneContextManager } from "@opentelemetry/context-zone";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { getWebAutoInstrumentations } from "@opentelemetry/auto-instrumentations-web";
import { Resource } from "@opentelemetry/resources";
import {
  SEMRESATTRS_SERVICE_NAME,
  SEMRESATTRS_SERVICE_VERSION,
  SEMRESATTRS_DEPLOYMENT_ENVIRONMENT,
} from "@opentelemetry/semantic-conventions";
import { MeterProvider, PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-http";

const OTEL_ENDPOINT =
  process.env.NEXT_PUBLIC_OTEL_ENDPOINT ||
  (typeof window !== "undefined" && (window as any).__OTEL_ENDPOINT) ||
  "https://otel.shopos.internal";

export interface RumConfig {
  serviceName: string;
  serviceVersion?: string;
  environment?: string;
  sampleRate?: number;
}

export function initRum({
  serviceName,
  serviceVersion = "1.0.0",
  environment = "production",
  sampleRate = 0.1,
}: RumConfig): void {
  if (typeof window === "undefined") return;

  const resource = new Resource({
    [SEMRESATTRS_SERVICE_NAME]: serviceName,
    [SEMRESATTRS_SERVICE_VERSION]: serviceVersion,
    [SEMRESATTRS_DEPLOYMENT_ENVIRONMENT]: environment,
  });

  const traceExporter = new OTLPTraceExporter({
    url: `${OTEL_ENDPOINT}/v1/traces`,
    headers: { "Content-Type": "application/json" },
  });

  const provider = new WebTracerProvider({
    resource,
    sampler: {
      shouldSample: () => ({
        decision: Math.random() < sampleRate ? 1 : 0,
        attributes: {},
        traceState: undefined,
      }),
      toString: () => `ProbabilitySampler{${sampleRate}}`,
    },
  });

  provider.addSpanProcessor(new BatchSpanProcessor(traceExporter));
  provider.register({ contextManager: new ZoneContextManager() });

  // Auto-instrument fetch, XHR, document load, user interaction
  registerInstrumentations({
    instrumentations: [
      getWebAutoInstrumentations({
        "@opentelemetry/instrumentation-fetch": {
          propagateTraceHeaderCorsUrls: [/shopos\.com/, /shopos\.internal/],
          clearTimingResources: true,
        },
        "@opentelemetry/instrumentation-xml-http-request": {
          propagateTraceHeaderCorsUrls: [/shopos\.com/, /shopos\.internal/],
        },
        "@opentelemetry/instrumentation-document-load": {},
        "@opentelemetry/instrumentation-user-interaction": {
          eventNames: ["click", "submit", "change"],
        },
      }),
    ],
  });

  // Web Vitals metrics
  const metricExporter = new OTLPMetricExporter({
    url: `${OTEL_ENDPOINT}/v1/metrics`,
  });

  const meterProvider = new MeterProvider({
    resource,
    readers: [
      new PeriodicExportingMetricReader({
        exporter: metricExporter,
        exportIntervalMillis: 30_000,
      }),
    ],
  });

  const meter = meterProvider.getMeter(serviceName);

  // Core Web Vitals
  const lcp = meter.createHistogram("web.lcp", {
    description: "Largest Contentful Paint",
    unit: "ms",
  });
  const fid = meter.createHistogram("web.fid", {
    description: "First Input Delay",
    unit: "ms",
  });
  const cls = meter.createHistogram("web.cls", {
    description: "Cumulative Layout Shift",
    unit: "{score}",
  });
  const ttfb = meter.createHistogram("web.ttfb", {
    description: "Time to First Byte",
    unit: "ms",
  });

  // Observe using web-vitals
  import("web-vitals").then(({ onLCP, onFID, onCLS, onTTFB }) => {
    onLCP((m) => lcp.record(m.value, { page: window.location.pathname }));
    onFID((m) => fid.record(m.value, { page: window.location.pathname }));
    onCLS((m) => cls.record(m.value, { page: window.location.pathname }));
    onTTFB((m) => ttfb.record(m.value, { page: window.location.pathname }));
  });

  // Track unhandled errors
  const errorCounter = meter.createCounter("web.errors", {
    description: "Unhandled JavaScript errors",
  });

  window.addEventListener("error", (e) => {
    errorCounter.add(1, {
      page: window.location.pathname,
      error_type: e.error?.name || "Error",
    });
  });

  window.addEventListener("unhandledrejection", (e) => {
    errorCounter.add(1, {
      page: window.location.pathname,
      error_type: "UnhandledPromiseRejection",
    });
  });
}
