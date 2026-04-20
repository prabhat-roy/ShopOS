// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/commerce/ is compiled.
export class OrderService {
  constructor (addr) { this.addr = addr }

  async listOrders (_userId, _page, _pageSize) { throw notImplemented() }
  async getOrder (_orderId, _userId)           { throw notImplemented() }
  async placeOrder (_userId, _req)             { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
