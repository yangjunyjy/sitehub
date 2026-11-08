import http from 'node:http'
import { existsSync } from 'node:fs'
import { resolve } from 'node:path'
import sirv from 'sirv'

function json(res, status, message) {
  res.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' })
  res.end(JSON.stringify({ code: status, message }))
}

export function createFrontendServer({ app, dist }) {
  if (!existsSync(resolve(dist, 'index.html'))) throw new Error(`Build ${app} before starting its server`)
  if (!['admin', 'site'].includes(app)) throw new Error('app must be admin or site')
  const assets = sirv(dist, {
    etag: true, single: true,
    setHeaders(res, pathname) {
      res.setHeader('X-Content-Type-Options', 'nosniff')
      res.setHeader('Cache-Control', pathname.includes('/assets/') ? 'public, max-age=31536000, immutable' : 'no-cache')
    },
  })
  const server = http.createServer((req, res) => {
    if (req.url === '/healthz' && ['GET', 'HEAD'].includes(req.method)) {
      res.writeHead(200, { 'Content-Type': 'application/json', 'Cache-Control': 'no-store' })
      return res.end(JSON.stringify({ status: 'ok', app }))
    }
    if (/^\/(?:v1|api|uploads)(?:\/|$)/.test(req.url.split('?')[0])) return json(res, 404, 'API traffic must use the gateway')
    if (!['GET', 'HEAD'].includes(req.method)) return json(res, 405, 'Method not allowed')
    assets(req, res, () => json(res, 404, 'Not found'))
  })
  server.requestTimeout = 600000
  return server
}
