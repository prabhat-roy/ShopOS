'use strict';

const { Router } = require('express');

function createConsentRouter(consentController) {
  const router = Router();

  // Grant consent
  // POST /consent
  // Body: { customerId, type, source }
  router.post('/', consentController.grantConsent);

  // Revoke consent by type
  // DELETE /consent
  // Body: { customerId, type, source }
  router.delete('/', consentController.revokeConsentByType);

  // Get all consent statuses for a customer
  // GET /consent/:customerId
  router.get('/:customerId', consentController.getAllConsents);

  // Check a single consent type for a customer
  // GET /consent/:customerId/:type
  router.get('/:customerId/:type/history', consentController.getConsentHistory);

  // Check single consent type
  router.get('/:customerId/:type', consentController.checkConsent);

  // Revoke all consents for a customer
  // DELETE /consent/:customerId
  // Body: { source }
  router.delete('/:customerId', consentController.revokeAllConsents);

  return router;
}

module.exports = createConsentRouter;
