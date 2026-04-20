export default async function healthRoutes (fastify) {
  fastify.get('/healthz', async () => ({ status: 'ok' }))
}
