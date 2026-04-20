'use strict';

const FeedbackType = Object.freeze({
  NPS: 'NPS',
  FEATURE_REQUEST: 'FEATURE_REQUEST',
  BUG_REPORT: 'BUG_REPORT',
  GENERAL: 'GENERAL',
  COMPLAINT: 'COMPLAINT',
});

const FeedbackStatus = Object.freeze({
  NEW: 'NEW',
  REVIEWED: 'REVIEWED',
  IN_PROGRESS: 'IN_PROGRESS',
  RESOLVED: 'RESOLVED',
  CLOSED: 'CLOSED',
});

const ALL_FEEDBACK_TYPES = Object.values(FeedbackType);
const ALL_FEEDBACK_STATUSES = Object.values(FeedbackStatus);

function isValidFeedbackType(type) {
  return ALL_FEEDBACK_TYPES.includes(type);
}

function isValidFeedbackStatus(status) {
  return ALL_FEEDBACK_STATUSES.includes(status);
}

/**
 * Validate NPS score is in range 0-10.
 */
function isValidNpsScore(score) {
  const n = parseInt(score, 10);
  return Number.isInteger(n) && n >= 0 && n <= 10;
}

/**
 * Classify NPS score bucket.
 */
function npsCategory(score) {
  if (score >= 9) return 'promoter';
  if (score >= 7) return 'passive';
  return 'detractor';
}

module.exports = {
  FeedbackType,
  FeedbackStatus,
  ALL_FEEDBACK_TYPES,
  ALL_FEEDBACK_STATUSES,
  isValidFeedbackType,
  isValidFeedbackStatus,
  isValidNpsScore,
  npsCategory,
};
