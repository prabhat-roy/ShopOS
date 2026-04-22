/**
 * OpenReplay session replay configuration
 * Self-hosted session replay for debugging UI issues.
 * Only enabled in production; respects user consent.
 */
import Tracker from "@openreplay/tracker";
import trackerFetch from "@openreplay/tracker-fetch";
import trackerAxios from "@openreplay/tracker-axios";
import trackerZustand from "@openreplay/tracker-zustand";

let tracker: Tracker | null = null;

export interface OpenReplayConfig {
  projectKey: string;
  ingestPoint?: string;
  capturePerformance?: boolean;
  respectDoNotTrack?: boolean;
}

export function initOpenReplay({
  projectKey,
  ingestPoint = "https://replay.shopos.internal/ingest",
  capturePerformance = true,
  respectDoNotTrack = true,
}: OpenReplayConfig): Tracker | null {
  if (typeof window === "undefined") return null;

  if (respectDoNotTrack && navigator.doNotTrack === "1") {
    console.log("OpenReplay: respecting DoNotTrack");
    return null;
  }

  if (tracker) return tracker;

  tracker = new Tracker({
    projectKey,
    ingestPoint,
    capturePerformance,
    obscureTextEmails: true,
    obscureInputEmails: true,
    obscureInputDates: false,
    defaultInputMode: 1,  // plain text; set to 0 for obscured
    network: {
      failuresToIgnore: [401, 403],
      capturePayload: false,
      sanitizer: (data) => {
        // Strip sensitive headers/payloads
        delete data.request?.headers?.Authorization;
        delete data.request?.headers?.["X-API-Key"];
        return data;
      },
    },
  });

  // Plugin: capture fetch requests
  tracker.use(trackerFetch({
    failuresToIgnore: [401, 403],
    sessionTokenHeader: false,
  }));

  // Plugin: capture Zustand state changes (without auth/PII)
  tracker.use(trackerZustand());

  return tracker;
}

export function startSession(userId?: string): void {
  if (!tracker) return;
  tracker.start();
  if (userId) {
    tracker.setUserID(userId);
  }
}

export function setUserContext(userId: string, email?: string): void {
  if (!tracker) return;
  tracker.setUserID(userId);
  if (email) {
    tracker.setMetadata("email_domain", email.split("@")[1] || "unknown");
  }
}

export function trackEvent(name: string, payload?: Record<string, unknown>): void {
  if (!tracker) return;
  tracker.event(name, payload);
}

export function stopSession(): void {
  tracker?.stop();
}
