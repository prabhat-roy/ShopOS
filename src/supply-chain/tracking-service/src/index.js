'use strict';

const express = require('express');

const PORT         = process.env.PORT || 50104;
const MONGO_URI    = process.env.MONGO_URI || 'mongodb://mongodb:27017/tracking';
const SERVICE_NAME = 'tracking-service';

// --- Express health endpoint ---
const app = express();
app.use(express.json());

app.get('/healthz', (_req, res) => {
  res.json({ status: 'ok' });
});

const server = app.listen(PORT, () => {
  console.log(`${SERVICE_NAME} HTTP/healthz listening on port ${PORT}`);
});

// TODO: Connect to MongoDB using Mongoose
//       const mongoose = require('mongoose');
//       mongoose.connect(MONGO_URI).then(() => console.log('MongoDB connected'));
//       Define Shipment schema: { shipmentId, orderId, carrier, status, events: [...], updatedAt }

// TODO: Wire up gRPC server for tracking handlers using generated proto client
//       from proto/supply-chain/tracking.proto once proto files are compiled.
//       Example:
//         const grpc       = require('@grpc/grpc-js');
//         const protoLoader = require('@grpc/proto-loader');
//         const packageDef  = protoLoader.loadSync('../../../../../proto/supply-chain/tracking.proto');
//         const proto       = grpc.loadPackageDefinition(packageDef).supplychain;
//         const grpcServer  = new grpc.Server();
//         grpcServer.addService(proto.TrackingService.service, {
//           getShipment:    (call, cb) => { /* query MongoDB */ },
//           updateTracking: (call, cb) => { /* write tracking event */ },
//         });
//         grpcServer.bindAsync(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure(), () => grpcServer.start());

console.log(`${SERVICE_NAME} starting — MongoDB URI: ${MONGO_URI}`);

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received — shutting down gracefully');
  server.close(() => {
    console.log('HTTP server closed');
    process.exit(0);
  });
});
