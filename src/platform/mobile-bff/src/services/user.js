// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/identity/ is compiled.
export class UserService {
  constructor (addr) { this.addr = addr }

  async getProfile (_userId)           { throw notImplemented() }
  async updateProfile (_userId, _req)  { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
