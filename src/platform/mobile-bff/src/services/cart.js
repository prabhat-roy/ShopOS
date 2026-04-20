// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/commerce/ is compiled.
export class CartService {
  constructor (addr) { this.addr = addr }

  async getCart (_userId)                              { throw notImplemented() }
  async addItem (_userId, _productId, _qty)            { throw notImplemented() }
  async updateItem (_userId, _itemId, _qty)            { throw notImplemented() }
  async removeItem (_userId, _itemId)                  { throw notImplemented() }
  async clearCart (_userId)                            { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
