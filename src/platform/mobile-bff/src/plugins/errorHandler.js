export function errorHandler (error, request, reply) {
  const statusMap = {
    NOT_FOUND:       404,
    UNAUTHORIZED:    401,
    INVALID_INPUT:   400,
    NOT_IMPLEMENTED: 501,
  }

  const statusCode = statusMap[error.code] ?? error.statusCode ?? 500
  const message = statusCode === 500 ? 'internal server error' : error.message

  request.log.error({ err: error, path: request.url }, 'request error')

  reply.status(statusCode).send({ error: message })
}
