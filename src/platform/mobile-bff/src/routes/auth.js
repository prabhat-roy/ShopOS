export default async function authRoutes (fastify) {
  fastify.post('/login', {
    schema: {
      body: {
        type: 'object',
        required: ['email', 'password'],
        properties: {
          email:    { type: 'string', format: 'email' },
          password: { type: 'string', minLength: 1 },
        },
      },
    },
  }, async (request, reply) => {
    const { email, password } = request.body
    const result = await fastify.services.auth.login(email, password)
    return reply.send(result)
  })

  fastify.post('/register', {
    schema: {
      body: {
        type: 'object',
        required: ['email', 'password', 'first_name'],
        properties: {
          email:      { type: 'string', format: 'email' },
          password:   { type: 'string', minLength: 8 },
          first_name: { type: 'string', minLength: 1 },
          last_name:  { type: 'string' },
        },
      },
    },
  }, async (request, reply) => {
    const result = await fastify.services.auth.register(request.body)
    return reply.status(201).send(result)
  })

  fastify.post('/refresh', {
    schema: {
      body: {
        type: 'object',
        required: ['refresh_token'],
        properties: { refresh_token: { type: 'string' } },
      },
    },
  }, async (request, reply) => {
    const result = await fastify.services.auth.refreshToken(request.body.refresh_token)
    return reply.send(result)
  })
}
