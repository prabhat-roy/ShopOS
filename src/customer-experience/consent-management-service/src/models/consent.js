'use strict';

const ConsentType = Object.freeze({
  MARKETING_EMAIL: 'MARKETING_EMAIL',
  MARKETING_SMS: 'MARKETING_SMS',
  ANALYTICS: 'ANALYTICS',
  PERSONALIZATION: 'PERSONALIZATION',
  THIRD_PARTY_SHARING: 'THIRD_PARTY_SHARING',
  ESSENTIAL: 'ESSENTIAL',
});

const ConsentAction = Object.freeze({
  GRANT: 'grant',
  REVOKE: 'revoke',
});

const ALL_CONSENT_TYPES = Object.values(ConsentType);

const OPTIONAL_CONSENT_TYPES = ALL_CONSENT_TYPES.filter(
  (t) => t !== ConsentType.ESSENTIAL
);

function isValidConsentType(type) {
  return ALL_CONSENT_TYPES.includes(type);
}

function isEssential(type) {
  return type === ConsentType.ESSENTIAL;
}

module.exports = {
  ConsentType,
  ConsentAction,
  ALL_CONSENT_TYPES,
  OPTIONAL_CONSENT_TYPES,
  isValidConsentType,
  isEssential,
};
