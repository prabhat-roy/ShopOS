'use strict';

const Shipment = require('../models/Shipment');

class ShipmentRepository {
  /**
   * Create a new shipment document.
   * @param {Object} data
   * @returns {Promise<Object>}
   */
  async create(data) {
    const shipment = new Shipment(data);
    await shipment.save();
    return shipment.toObject();
  }

  /**
   * Find a shipment by its tracking number.
   * @param {string} trackingNumber
   * @returns {Promise<Object|null>}
   */
  async findByTrackingNumber(trackingNumber) {
    const shipment = await Shipment.findOne({ trackingNumber }).lean();
    return shipment || null;
  }

  /**
   * Find a shipment by its MongoDB _id.
   * @param {string} id
   * @returns {Promise<Object|null>}
   */
  async findById(id) {
    const shipment = await Shipment.findById(id).lean();
    return shipment || null;
  }

  /**
   * List shipments with optional filters and pagination.
   * @param {Object} opts
   * @param {string} [opts.carrier]
   * @param {string} [opts.status]
   * @param {number} [opts.limit=20]
   * @param {number} [opts.offset=0]
   * @returns {Promise<{shipments: Object[], total: number}>}
   */
  async list({ carrier, status, limit = 20, offset = 0 } = {}) {
    const query = {};
    if (carrier) query.carrier = carrier;
    if (status) query.status = status;

    const [shipments, total] = await Promise.all([
      Shipment.find(query)
        .sort({ createdAt: -1 })
        .skip(offset)
        .limit(limit)
        .lean(),
      Shipment.countDocuments(query),
    ]);

    return { shipments, total };
  }

  /**
   * Append a tracking event to a shipment's events array.
   * @param {string} trackingNumber
   * @param {Object} event
   * @returns {Promise<Object|null>}
   */
  async addEvent(trackingNumber, event) {
    const shipment = await Shipment.findOneAndUpdate(
      { trackingNumber },
      { $push: { events: event } },
      { new: true }
    ).lean();
    return shipment || null;
  }

  /**
   * Update the status of a shipment and optionally push a tracking event.
   * @param {string} trackingNumber
   * @param {string} status
   * @param {Object} [eventData]
   * @returns {Promise<Object|null>}
   */
  async updateStatus(trackingNumber, status, eventData) {
    const update = { status };
    if (eventData) {
      update.$push = { events: eventData };
    }
    const shipment = await Shipment.findOneAndUpdate(
      { trackingNumber },
      eventData ? { $set: { status }, $push: { events: eventData } } : { $set: { status } },
      { new: true }
    ).lean();
    return shipment || null;
  }

  /**
   * Mark a shipment as delivered, set actualDelivery timestamp, and push a delivery event.
   * @param {string} trackingNumber
   * @returns {Promise<Object|null>}
   */
  async updateDelivered(trackingNumber) {
    const now = new Date();
    const deliveryEvent = {
      timestamp: now,
      location: '',
      description: 'Package delivered',
      status: 'delivered',
    };

    const shipment = await Shipment.findOneAndUpdate(
      { trackingNumber },
      {
        $set: { status: 'delivered', actualDelivery: now },
        $push: { events: deliveryEvent },
      },
      { new: true }
    ).lean();
    return shipment || null;
  }
}

module.exports = ShipmentRepository;
