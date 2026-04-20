'use strict';

const express = require('express');
const ctrl = require('../controllers/seoController');

const router = express.Router();

// Health check
router.get('/healthz', (req, res) => {
  res.json({ status: 'ok' });
});

// Generate SEO metadata for a product page
router.post('/seo/product', ctrl.productMeta);

// Generate SEO metadata for a category page
router.post('/seo/category', ctrl.categoryMeta);

// Generate SEO metadata for a brand page
router.post('/seo/brand', ctrl.brandMeta);

// Generate an XML sitemap fragment for a list of URLs
router.get('/seo/sitemap', ctrl.sitemap);

module.exports = router;
