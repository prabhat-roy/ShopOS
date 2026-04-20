// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/identity/ is compiled.
export class AuthService {
  constructor (addr) { this.addr = addr }

  async login (_email, _password) { throw notImplemented() }
  async register (_data)          { throw notImplemented() }
  async refreshToken (_token)     { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
