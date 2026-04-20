'use strict';

const { v4: uuidv4 } = require('uuid');
const ShipmentRepository = require('../repositories/ShipmentRepository');

const VALID_STATUSES = ['created', 'in_transit', 'out_for_delivery', 'delivered', 'failed', 'returned'];

class TrackingService {
  constructor(repository) {
    this.repository = repository || new ShipmentRepository();
  }

  /**
   * Create a new shipment. Generates a trackingNumber if not provided.
   * @param {Object} data
   * @returns {Promise<Object>}
   */
  async createShipment(data) {
    const shipmentData = {
      ...data,
      trackingNumber: data.trackingNumber || `TRK-${uuidv4().replace(/-/g, '').toUpperCase().slice(0, 16)}`,
      events: [
        {
          timestamp: new Date(),
          location: data.originAddress || '',
          description: 'Shipment created',
          status: 'created',
        },
      ],
    };
    return this.repository.create(shipmentData);
  }

  /**
   * Retrieve a shipment by tracking number.
   * @param {string} trackingNumber
   * @returns {Promise<Object|null>}
   */
  async getShipment(trackingNumber) {
    return this.repository.findByTrackingNumber(trackingNumber);
  }

  /**
   * List shipments with optional filters.
   * @param {Object} filters
   * @returns {Promise<{shipments: Object[], total: number}>}
   */
  async listShipments(filters) {
    return this.repository.list(filters);
  }

  /**
   * Add a manual tracking event to a shipment.
   * @param {string} trackingNumber
   * @param {Object} eventData - { location, description, status }
   * @returns {Promise<Object>}
   * @throws {Error} if shipment not found
   */
  async addTrackingEvent(trackingNumber, eventData) {
    const event = {
      timestamp: eventData.timestamp ? new Date(eventData.timestamp) : new Date(),
      location: eventData.location || '',
      description: eventData.description,
      status: eventData.status || '',
    };

    const shipment = await this.repository.addEvent(trackingNumber, event);
    if (!shipment) {
      const err = new Error(`Shipment not found: ${trackingNumber}`);
      err.statusCode = 404;
      throw err;
    }
    return shipment;
  }

  /**
   * Update the status of a shipment and record a tracking event.
   * @param {string} trackingNumber
   * @param {string} status
   * @param {string} [location]
   * @param {string} [description]
   * @returns {Promise<Object>}
   * @throws {Error} if status invalid or shipment not found
   */
  async updateStatus(trackingNumber, status, location, description) {
    if (!VALID_STATUSES.includes(status)) {
      const err = new Error(`Invalid status: ${status}. Must be one of: ${VALID_STATUSES.join(', ')}`);
      err.statusCode = 400;
      throw err;
    }

    const eventData = {
      timestamp: new Date(),
      location: location || '',
      description: description || `Status updated to ${status}`,
      status,
    };

    const shipment = await this.repository.updateStatus(trackingNumber, status, eventData);
    if (!shipment) {
      const err = new Error(`Shipment not found: ${trackingNumber}`);
      err.statusCode = 404;
      throw err;
    }
    return shipment;
  }

  /**
   * Mark a shipment as delivered.
   * @param {string} trackingNumber
   * @returns {Promise<Object>}
   * @throws {Error} if shipment not found
   */
  async markDelivered(trackingNumber) {
    const shipment = await this.repository.updateDelivered(trackingNumber);
    if (!shipment) {
      const err = new Error(`Shipment not found: ${trackingNumber}`);
      err.statusCode = 404;
      throw err;
    }
    return shipment;
  }
}

module.exports = TrackingService;
