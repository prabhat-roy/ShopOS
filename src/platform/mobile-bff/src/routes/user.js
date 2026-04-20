export default async function userRoutes (fastify) {
  const userId = (req) => req.headers['x-user-id'] ?? ''

  fastify.get('/me', async (request) => {
    return fastify.services.user.getProfile(userId(request))
  })

  fastify.put('/me', {
    schema: {
      body: {
        type: 'object',
        properties: {
          first_name: { type: 'string' },
          last_name:  { type: 'string' },
          phone:      { type: 'string' },
        },
      },
    },
  }, async (request) => {
    return fastify.services.user.updateProfile(userId(request), request.body)
  })
}
