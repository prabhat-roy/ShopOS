'use strict';

const express = require('express');

const PORT = process.env.PORT || 50079;
const app  = express();

app.use(express.json());

// Health check
app.get('/healthz', (_req, res) => {
  res.json({ status: 'ok' });
});

// TODO: Wire up gRPC server using generated proto client from
//       proto/catalog/seo.proto once proto files are compiled.
//       Example:
//         const grpc       = require('@grpc/grpc-js');
//         const protoLoader = require('@grpc/proto-loader');
//         const packageDef  = protoLoader.loadSync('../../../../../proto/catalog/seo.proto');
//         const proto       = grpc.loadPackageDefinition(packageDef).catalog;
//         const grpcServer  = new grpc.Server();
//         grpcServer.addService(proto.SeoService.service, handlers);
//         grpcServer.bindAsync(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure(), () => grpcServer.start());

const server = app.listen(PORT, () => {
  console.log(`seo-service HTTP/healthz listening on port ${PORT}`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received — shutting down gracefully');
  server.close(() => {
    console.log('HTTP server closed');
    process.exit(0);
  });
});
