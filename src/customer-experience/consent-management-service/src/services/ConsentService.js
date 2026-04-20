'use strict';

const { ConsentType, ALL_CONSENT_TYPES, OPTIONAL_CONSENT_TYPES, isValidConsentType, isEssential } = require('../models/consent');

class ConsentService {
  constructor(consentRepository) {
    this.consentRepository = consentRepository;
  }

  async grantConsent(customerId, type, source, ipAddress) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }
    if (!isValidConsentType(type)) {
      throw Object.assign(
        new Error(`Invalid consent type: ${type}. Valid types: ${ALL_CONSENT_TYPES.join(', ')}`),
        { statusCode: 400 }
      );
    }
    if (!source) {
      throw Object.assign(new Error('source is required'), { statusCode: 400 });
    }

    const record = await this.consentRepository.upsertConsent(
      customerId,
      type,
      true,
      source,
      ipAddress || null
    );

    return record;
  }

  async revokeConsent(customerId, type, source, ipAddress) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }
    if (!isValidConsentType(type)) {
      throw Object.assign(
        new Error(`Invalid consent type: ${type}. Valid types: ${ALL_CONSENT_TYPES.join(', ')}`),
        { statusCode: 400 }
      );
    }
    if (isEssential(type)) {
      throw Object.assign(
        new Error('ESSENTIAL consent cannot be revoked'),
        { statusCode: 422 }
      );
    }
    if (!source) {
      throw Object.assign(new Error('source is required'), { statusCode: 400 });
    }

    const record = await this.consentRepository.upsertConsent(
      customerId,
      type,
      false,
      source,
      ipAddress || null
    );

    return record;
  }

  async getConsentStatus(customerId) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }

    const records = await this.consentRepository.getAllConsents(customerId);
    const statusMap = {};

    for (const type of ALL_CONSENT_TYPES) {
      if (type === ConsentType.ESSENTIAL) {
        statusMap[type] = true;
        continue;
      }
      statusMap[type] = false;
    }

    for (const record of records) {
      if (record.type === ConsentType.ESSENTIAL) {
        statusMap[record.type] = true;
      } else {
        statusMap[record.type] = record.granted === true;
      }
    }

    return statusMap;
  }

  async checkConsent(customerId, type) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }
    if (!isValidConsentType(type)) {
      throw Object.assign(
        new Error(`Invalid consent type: ${type}`),
        { statusCode: 400 }
      );
    }

    if (isEssential(type)) {
      return true;
    }

    return this.consentRepository.hasConsent(customerId, type);
  }

  async revokeAllConsents(customerId, source, ipAddress) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }
    if (!source) {
      throw Object.assign(new Error('source is required'), { statusCode: 400 });
    }

    const revoked = await this.consentRepository.revokeAll(customerId, source, ipAddress || null);
    return revoked;
  }

  async getConsentHistory(customerId, type, limit = 50) {
    if (!customerId) {
      throw Object.assign(new Error('customerId is required'), { statusCode: 400 });
    }
    if (type && !isValidConsentType(type)) {
      throw Object.assign(
        new Error(`Invalid consent type: ${type}`),
        { statusCode: 400 }
      );
    }

    const parsedLimit = Math.min(parseInt(limit, 10) || 50, 200);
    return this.consentRepository.getHistory(customerId, type || null, parsedLimit);
  }
}

module.exports = ConsentService;
