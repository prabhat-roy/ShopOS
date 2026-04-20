export default async function cartRoutes (fastify) {
  const userId = (req) => req.headers['x-user-id'] ?? ''

  fastify.get('/', async (request) => {
    return fastify.services.cart.getCart(userId(request))
  })

  fastify.post('/items', {
    schema: {
      body: {
        type: 'object',
        required: ['product_id', 'quantity'],
        properties: {
          product_id: { type: 'string' },
          quantity:   { type: 'integer', minimum: 1 },
        },
      },
    },
  }, async (request) => {
    const { product_id, quantity } = request.body
    return fastify.services.cart.addItem(userId(request), product_id, quantity)
  })

  fastify.put('/items/:itemId', {
    schema: {
      body: {
        type: 'object',
        required: ['quantity'],
        properties: { quantity: { type: 'integer', minimum: 1 } },
      },
    },
  }, async (request) => {
    return fastify.services.cart.updateItem(userId(request), request.params.itemId, request.body.quantity)
  })

  fastify.delete('/items/:itemId', async (request) => {
    return fastify.services.cart.removeItem(userId(request), request.params.itemId)
  })

  fastify.delete('/', async (request) => {
    await fastify.services.cart.clearCart(userId(request))
    return { message: 'cart cleared' }
  })
}
