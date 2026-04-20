import 'dotenv/config'
import { buildApp } from './src/app.js'

const app = await buildApp()

try {
  await app.listen({ port: Number(process.env.PORT ?? 8082), host: '0.0.0.0' })
} catch (err) {
  app.log.error(err)
  process.exit(1)
}
