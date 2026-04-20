'use strict';

const express = require('express');

const PORT         = process.env.PORT || 50131;
const SERVICE_NAME = 'template-service';

// --- Express health endpoint ---
const app = express();
app.use(express.json());

app.get('/healthz', (_req, res) => {
  res.json({ status: 'ok' });
});

const server = app.listen(PORT, () => {
  console.log(`${SERVICE_NAME} HTTP/healthz listening on port ${PORT}`);
});

// TODO: Wire up gRPC server for template rendering using generated proto client
//       from proto/communications/template.proto once proto files are compiled.
//       Example:
//         const grpc        = require('@grpc/grpc-js');
//         const protoLoader  = require('@grpc/proto-loader');
//         const packageDef   = protoLoader.loadSync('../../../../../proto/communications/template.proto');
//         const proto        = grpc.loadPackageDefinition(packageDef).communications;
//         const grpcServer   = new grpc.Server();
//         grpcServer.addService(proto.TemplateService.service, {
//           renderTemplate: (call, cb) => { /* Handlebars / Mustache rendering logic */ cb(null, result); },
//           getTemplate:    (call, cb) => { /* Fetch from MongoDB */ cb(null, template); },
//         });
//         grpcServer.bindAsync(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure(), () => grpcServer.start());

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received — shutting down gracefully');
  server.close(() => {
    console.log('HTTP server closed');
    process.exit(0);
  });
});
