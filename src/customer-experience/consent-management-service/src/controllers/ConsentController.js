'use strict';

class ConsentController {
  constructor(consentService) {
    this.consentService = consentService;
    this.grantConsent = this.grantConsent.bind(this);
    this.revokeConsentByType = this.revokeConsentByType.bind(this);
    this.getAllConsents = this.getAllConsents.bind(this);
    this.checkConsent = this.checkConsent.bind(this);
    this.revokeAllConsents = this.revokeAllConsents.bind(this);
    this.getConsentHistory = this.getConsentHistory.bind(this);
  }

  async grantConsent(req, res) {
    const { customerId, type, source } = req.body;
    const ipAddress = req.ip || req.headers['x-forwarded-for'] || null;

    if (!customerId || !type || !source) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'customerId, type, and source are required',
      });
    }

    try {
      const record = await this.consentService.grantConsent(customerId, type, source, ipAddress);
      return res.status(200).json({
        customerId: record.customer_id,
        type: record.type,
        granted: record.granted,
        source: record.source,
        updatedAt: record.updated_at,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async revokeConsentByType(req, res) {
    const { customerId, type, source } = req.body;
    const ipAddress = req.ip || req.headers['x-forwarded-for'] || null;

    if (!customerId || !type || !source) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'customerId, type, and source are required',
      });
    }

    try {
      const record = await this.consentService.revokeConsent(customerId, type, source, ipAddress);
      return res.status(200).json({
        customerId: record.customer_id,
        type: record.type,
        granted: record.granted,
        source: record.source,
        updatedAt: record.updated_at,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getAllConsents(req, res) {
    const { customerId } = req.params;

    try {
      const statusMap = await this.consentService.getConsentStatus(customerId);
      return res.status(200).json({
        customerId,
        consents: statusMap,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async checkConsent(req, res) {
    const { customerId, type } = req.params;

    try {
      const granted = await this.consentService.checkConsent(customerId, type);
      return res.status(200).json({
        customerId,
        type,
        granted,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async revokeAllConsents(req, res) {
    const { customerId } = req.params;
    const { source } = req.body;
    const ipAddress = req.ip || req.headers['x-forwarded-for'] || null;

    if (!source) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'source is required in request body',
      });
    }

    try {
      const revoked = await this.consentService.revokeAllConsents(customerId, source, ipAddress);
      return res.status(200).json({
        customerId,
        revokedCount: revoked.length,
        message: `Revoked ${revoked.length} consent(s) for customer ${customerId}`,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getConsentHistory(req, res) {
    const { customerId, type } = req.params;
    const { limit } = req.query;

    try {
      const history = await this.consentService.getConsentHistory(customerId, type, limit);
      return res.status(200).json({
        customerId,
        type,
        history,
      });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  _handleError(res, err) {
    const statusCode = err.statusCode || 500;
    const message = err.message || 'Internal server error';

    if (statusCode === 500) {
      console.error('[ConsentController] Internal error:', err);
    }

    return res.status(statusCode).json({
      error: statusCode === 400 ? 'Bad Request'
        : statusCode === 404 ? 'Not Found'
        : statusCode === 422 ? 'Unprocessable Entity'
        : 'Internal Server Error',
      message,
    });
  }
}

module.exports = ConsentController;
