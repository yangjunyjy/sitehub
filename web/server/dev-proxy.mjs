// Vite development only. Production Node servers serve static pages exclusively.
export function gatewayProxy(app, env = process.env) {
  if (!['admin', 'site'].includes(app)) throw new Error('Invalid frontend app')
  const target = new URL(env.GATEWAY_URL || 'http://127.0.0.1:8082')
  if (!['http:', 'https:'].includes(target.protocol) || target.username || target.password || target.pathname !== '/' || target.search || target.hash) throw new Error('Invalid GATEWAY_URL')
  const host = env.GATEWAY_HOST || (app === 'admin' ? 'admin.sitehub.local' : 'sitehub.local')
  if (!/^[a-zA-Z0-9.-]+(?::[0-9]+)?$/.test(host)) throw new Error('Invalid GATEWAY_HOST')
  return { '/v1': { target: target.origin, changeOrigin: true, headers: { host }, timeout: 600000, proxyTimeout: 600000 } }
}
