import { parseArgs } from 'node:util'
import { fileURLToPath } from 'node:url'
import { createFrontendServer } from './server.mjs'

const { values } = parseArgs({ options: { app: { type: 'string' } }, allowPositionals: false })
const app = values.app || process.env.FRONTEND_APP
if (!['admin', 'site'].includes(app)) throw new Error('Use --app admin or --app site')
const port = Number(process.env.PORT || (app === 'admin' ? 4311 : 4310))
if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error('Invalid PORT')
const server = createFrontendServer({ app, dist: fileURLToPath(new URL(`../${app}/dist/`, import.meta.url)) })
server.on('error', error => { console.error(error.message); process.exitCode = 1 })
server.listen(port, process.env.HOST || '127.0.0.1', () => console.log(`${app} listening on ${process.env.HOST || '127.0.0.1'}:${port}`))
for (const signal of ['SIGINT', 'SIGTERM']) process.once(signal, () => {
  server.close(() => process.exit(0))
  setTimeout(() => { server.closeAllConnections(); process.exit(0) }, 10000).unref()
})
