export const config = {
  port:             Number(process.env.PORT ?? 8082),
  env:              process.env.ENV ?? 'development',
  grpcTimeout:      Number(process.env.GRPC_TIMEOUT_MS ?? 10_000),
  authServiceAddr:  process.env.AUTH_SERVICE_ADDR  ?? 'auth-service:50060',
  userServiceAddr:  process.env.USER_SERVICE_ADDR  ?? 'user-service:50061',
  catalogAddr:      process.env.CATALOG_SERVICE_ADDR ?? 'product-catalog-service:50070',
  cartServiceAddr:  process.env.CART_SERVICE_ADDR  ?? 'cart-service:50080',
  orderServiceAddr: process.env.ORDER_SERVICE_ADDR ?? 'order-service:50082',
  notifServiceAddr: process.env.NOTIFICATION_SERVICE_ADDR ?? 'push-notification-service:4000',
  corsOrigins:      (process.env.CORS_ALLOWED_ORIGINS ?? '*').split(','),
}
