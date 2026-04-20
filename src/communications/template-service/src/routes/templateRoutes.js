'use strict';

/**
 * Route registration for the template service.
 * All routes are implemented in TemplateController.
 * This module re-exports the router for explicit mounting in app.js.
 */

const templateController = require('../controllers/TemplateController');

module.exports = templateController;
