export default async function catalogRoutes (fastify) {
  fastify.get('/products', {
    schema: {
      querystring: {
        type: 'object',
        properties: {
          page:        { type: 'integer', minimum: 1, default: 1 },
          page_size:   { type: 'integer', minimum: 1, maximum: 50, default: 20 },
          category_id: { type: 'string' },
          sort:        { type: 'string' },
        },
      },
    },
  }, async (request) => {
    return fastify.services.catalog.listProducts(request.query)
  })

  fastify.get('/products/:id', async (request) => {
    return fastify.services.catalog.getProduct(request.params.id)
  })

  fastify.get('/categories', async () => {
    const items = await fastify.services.catalog.listCategories()
    return { items }
  })

  fastify.get('/search', {
    schema: {
      querystring: {
        type: 'object',
        required: ['q'],
        properties: {
          q:         { type: 'string', minLength: 1 },
          page:      { type: 'integer', minimum: 1, default: 1 },
          page_size: { type: 'integer', minimum: 1, maximum: 50, default: 20 },
        },
      },
    },
  }, async (request) => {
    const { q, page, page_size } = request.query
    return fastify.services.catalog.search(q, page, page_size)
  })

  // Mobile-specific: aggregated home feed (featured + promotions)
  fastify.get('/feed', {
    schema: {
      querystring: {
        type: 'object',
        properties: { cursor: { type: 'string' } },
      },
    },
  }, async (request) => {
    return fastify.services.catalog.getFeed(request.query.cursor)
  })
}
