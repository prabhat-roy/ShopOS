'use strict';

const { Router } = require('express');
const ctrl = require('../controllers/CmsController');

const router = Router();

// Search must come before /:id to avoid conflicts
router.get('/search', ctrl.searchContent);

// Slug lookup
router.get('/slug/:slug', ctrl.getBySlug);

// CRUD
router.post('/', ctrl.createContent);
router.get('/', ctrl.listContent);
router.get('/:id', ctrl.getById);
router.patch('/:id', ctrl.updateContent);

// State transitions
router.post('/:id/publish', ctrl.publishContent);
router.post('/:id/archive', ctrl.archiveContent);

router.delete('/:id', ctrl.deleteContent);

module.exports = router;
