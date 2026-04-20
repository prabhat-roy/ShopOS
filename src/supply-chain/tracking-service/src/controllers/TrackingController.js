'use strict';

const TrackingService = require('../services/TrackingService');

class TrackingController {
  constructor(trackingService) {
    this.trackingService = trackingService || new TrackingService();
    // Bind all methods so they work as standalone express handlers
    this.createShipment = this.createShipment.bind(this);
    this.getShipment = this.getShipment.bind(this);
    this.listShipments = this.listShipments.bind(this);
    this.addEvent = this.addEvent.bind(this);
    this.updateStatus = this.updateStatus.bind(this);
    this.markDelivered = this.markDelivered.bind(this);
  }

  /**
   * POST /shipments
   */
  async createShipment(req, res) {
    const { carrier, originAddress, destinationAddress, estimatedDelivery, metadata, trackingNumber } = req.body;

    if (!carrier) {
      return res.status(400).json({ error: 'carrier is required' });
    }

    try {
      const shipment = await this.trackingService.createShipment({
        carrier,
        trackingNumber,
        originAddress,
        destinationAddress,
        estimatedDelivery: estimatedDelivery ? new Date(estimatedDelivery) : undefined,
        metadata,
      });
      return res.status(201).json(shipment);
    } catch (err) {
      if (err.code === 11000) {
        return res.status(409).json({ error: 'Tracking number already exists' });
      }
      return res.status(500).json({ error: err.message });
    }
  }

  /**
   * GET /shipments/:trackingNumber
   */
  async getShipment(req, res) {
    const { trackingNumber } = req.params;
    try {
      const shipment = await this.trackingService.getShipment(trackingNumber);
      if (!shipment) {
        return res.status(404).json({ error: `Shipment not found: ${trackingNumber}` });
      }
      return res.status(200).json(shipment);
    } catch (err) {
      return res.status(500).json({ error: err.message });
    }
  }

  /**
   * GET /shipments
   */
  async listShipments(req, res) {
    const { carrier, status } = req.query;
    const limit = Math.min(parseInt(req.query.limit, 10) || 20, 100);
    const offset = parseInt(req.query.offset, 10) || 0;

    try {
      const result = await this.trackingService.listShipments({ carrier, status, limit, offset });
      return res.status(200).json(result);
    } catch (err) {
      return res.status(500).json({ error: err.message });
    }
  }

  /**
   * POST /shipments/:trackingNumber/events
   */
  async addEvent(req, res) {
    const { trackingNumber } = req.params;
    const { description, location, status, timestamp } = req.body;

    if (!description) {
      return res.status(400).json({ error: 'description is required' });
    }

    try {
      const shipment = await this.trackingService.addTrackingEvent(trackingNumber, {
        description,
        location,
        status,
        timestamp,
      });
      return res.status(201).json(shipment);
    } catch (err) {
      const statusCode = err.statusCode || 500;
      return res.status(statusCode).json({ error: err.message });
    }
  }

  /**
   * PATCH /shipments/:trackingNumber/status
   */
  async updateStatus(req, res) {
    const { trackingNumber } = req.params;
    const { status, location, description } = req.body;

    if (!status) {
      return res.status(400).json({ error: 'status is required' });
    }

    try {
      const shipment = await this.trackingService.updateStatus(trackingNumber, status, location, description);
      return res.status(200).json(shipment);
    } catch (err) {
      const statusCode = err.statusCode || 500;
      return res.status(statusCode).json({ error: err.message });
    }
  }

  /**
   * POST /shipments/:trackingNumber/deliver
   */
  async markDelivered(req, res) {
    const { trackingNumber } = req.params;
    try {
      await this.trackingService.markDelivered(trackingNumber);
      return res.status(204).send();
    } catch (err) {
      const statusCode = err.statusCode || 500;
      return res.status(statusCode).json({ error: err.message });
    }
  }
}

module.exports = TrackingController;
