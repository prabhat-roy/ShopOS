// Mobile-specific: push notification device token management
export default async function deviceRoutes (fastify) {
  const userId = (req) => req.headers['x-user-id'] ?? ''

  fastify.post('/register', {
    schema: {
      body: {
        type: 'object',
        required: ['token', 'platform'],
        properties: {
          token:    { type: 'string' },
          platform: { type: 'string', enum: ['ios', 'android'] },
          device_id: { type: 'string' },
        },
      },
    },
  }, async (request, reply) => {
    const { token, platform, device_id } = request.body
    await fastify.services.notif.registerDevice(userId(request), token, platform, device_id)
    return reply.status(201).send({ message: 'device registered' })
  })

  fastify.delete('/:deviceId', async (request, reply) => {
    await fastify.services.notif.unregisterDevice(userId(request), request.params.deviceId)
    return reply.status(204).send()
  })
}
