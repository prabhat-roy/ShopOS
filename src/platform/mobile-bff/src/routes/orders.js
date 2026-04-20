export default async function orderRoutes (fastify) {
  const userId = (req) => req.headers['x-user-id'] ?? ''

  fastify.get('/', {
    schema: {
      querystring: {
        type: 'object',
        properties: {
          page:      { type: 'integer', minimum: 1, default: 1 },
          page_size: { type: 'integer', minimum: 1, maximum: 20, default: 10 },
        },
      },
    },
  }, async (request) => {
    const { page, page_size } = request.query
    const items = await fastify.services.order.listOrders(userId(request), page, page_size)
    return { items }
  })

  fastify.get('/:id', async (request) => {
    return fastify.services.order.getOrder(request.params.id, userId(request))
  })

  fastify.post('/', {
    schema: {
      body: {
        type: 'object',
        required: ['address_id', 'payment_id'],
        properties: {
          address_id: { type: 'string' },
          payment_id: { type: 'string' },
        },
      },
    },
  }, async (request, reply) => {
    const order = await fastify.services.order.placeOrder(userId(request), request.body)
    return reply.status(201).send(order)
  })
}
