'use strict';

const express = require('express');
const mongoose = require('mongoose');
const config = require('./config');
const cmsRoutes = require('./routes/cmsRoutes');

const app = express();

app.use(express.json({ limit: '5mb' }));
app.use(express.urlencoded({ extended: true }));

// Request logging
app.use((req, _res, next) => {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
  next();
});

// Health check
app.get('/healthz', (_req, res) => {
  const dbState = mongoose.connection.readyState;
  const dbStatus = dbState === 1 ? 'connected' : 'disconnected';
  res.json({ status: 'ok', db: dbStatus });
});

// Routes
app.use('/content', cmsRoutes);

// 404 handler
app.use((_req, res) => {
  res.status(404).json({ error: 'Route not found' });
});

// Error handler
app.use((err, _req, res, _next) => {
  console.error('[app] Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

async function connectDb() {
  await mongoose.connect(config.mongoUri, {
    serverSelectionTimeoutMS: 5000,
  });
  console.log(`[app] Connected to MongoDB: ${config.mongoUri}`);
}

module.exports = { app, connectDb };
