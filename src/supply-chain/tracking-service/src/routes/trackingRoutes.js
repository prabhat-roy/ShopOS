'use strict';

const { Router } = require('express');
const TrackingController = require('../controllers/TrackingController');

/**
 * Build and return an Express router for tracking endpoints.
 * Accepts an optional pre-built controller (for testing).
 * @param {TrackingController} [controller]
 * @returns {Router}
 */
function createTrackingRouter(controller) {
  const router = Router();
  const ctrl = controller || new TrackingController();

  // POST /shipments — create a shipment
  router.post('/', ctrl.createShipment);

  // GET /shipments — list shipments with optional filters
  router.get('/', ctrl.listShipments);

  // GET /shipments/:trackingNumber — get a single shipment
  router.get('/:trackingNumber', ctrl.getShipment);

  // POST /shipments/:trackingNumber/events — add a tracking event
  router.post('/:trackingNumber/events', ctrl.addEvent);

  // PATCH /shipments/:trackingNumber/status — update shipment status
  router.patch('/:trackingNumber/status', ctrl.updateStatus);

  // POST /shipments/:trackingNumber/deliver — mark delivered (204)
  router.post('/:trackingNumber/deliver', ctrl.markDelivered);

  return router;
}

module.exports = createTrackingRouter;
