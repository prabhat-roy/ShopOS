import Fastify from 'fastify'
import cors from '@fastify/cors'
import helmet from '@fastify/helmet'
import { config } from './config.js'
import { errorHandler } from './plugins/errorHandler.js'
import healthRoutes from './routes/health.js'
import authRoutes from './routes/auth.js'
import catalogRoutes from './routes/catalog.js'
import cartRoutes from './routes/cart.js'
import orderRoutes from './routes/orders.js'
import userRoutes from './routes/user.js'
import deviceRoutes from './routes/device.js'
import { AuthService } from './services/auth.js'
import { CatalogService } from './services/catalog.js'
import { CartService } from './services/cart.js'
import { OrderService } from './services/order.js'
import { UserService } from './services/user.js'
import { NotificationService } from './services/notification.js'

export async function buildApp (opts = {}) {
  const app = Fastify({
    logger: config.env === 'production'
      ? true
      : { transport: { target: 'pino-pretty' }, level: 'info' },
    ...opts,
  })

  await app.register(helmet)
  await app.register(cors, { origin: config.corsOrigins })

  app.setErrorHandler(errorHandler)

  const services = {
    auth:    new AuthService(config.authServiceAddr),
    catalog: new CatalogService(config.catalogAddr),
    cart:    new CartService(config.cartServiceAddr),
    order:   new OrderService(config.orderServiceAddr),
    user:    new UserService(config.userServiceAddr),
    notif:   new NotificationService(config.notifServiceAddr),
  }

  app.decorate('services', services)

  await app.register(healthRoutes)
  await app.register(authRoutes,    { prefix: '/auth' })
  await app.register(catalogRoutes, { prefix: '/catalog' })
  await app.register(cartRoutes,    { prefix: '/cart' })
  await app.register(orderRoutes,   { prefix: '/orders' })
  await app.register(userRoutes,    { prefix: '/users' })
  await app.register(deviceRoutes,  { prefix: '/device' })

  return app
}
