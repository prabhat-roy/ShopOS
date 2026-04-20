// TODO: Replace stub methods with @grpc/grpc-js client calls once proto/communications/ is compiled.
export class NotificationService {
  constructor (addr) { this.addr = addr }

  async registerDevice (_userId, _token, _platform, _deviceId) { throw notImplemented() }
  async unregisterDevice (_userId, _deviceId)                   { throw notImplemented() }
}

function notImplemented () {
  const err = new Error('gRPC client not yet implemented — pending proto compilation')
  err.code = 'NOT_IMPLEMENTED'
  return err
}
