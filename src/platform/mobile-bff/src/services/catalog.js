// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/catalog/ is compiled.
export class CatalogService {
  constructor (addr) { this.addr = addr }

  async listProducts (_query)          { throw notImplemented() }
  async getProduct (_id)               { throw notImplemented() }
  async listCategories ()              { throw notImplemented() }
  async search (_q, _page, _pageSize) { throw notImplemented() }
  async getFeed (_cursor)              { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
