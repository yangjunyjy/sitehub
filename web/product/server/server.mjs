import http from 'node:http';
import { existsSync } from 'node:fs';
import { resolve } from 'node:path';
import sirv from 'sirv';

function origin(value, name) {
  const text = value.trim();
  if (!text) return '';
  let parsed;
  try { parsed = new URL(text); } catch { throw new Error(`Invalid ${name}`); }
  if (!['http:', 'https:'].includes(parsed.protocol) || parsed.username || parsed.password) throw new Error(`Invalid ${name}`);
  return text;
}

export function createFrontendServer({ dist, productURL = process.env.WEB_PRODUCT_URL || '', blogURL = process.env.WEB_BLOG_URL || '' }) {
  if (!existsSync(resolve(dist, 'index.html'))) throw new Error('Build product before starting its server');
  const config = { productURL: origin(productURL, 'WEB_PRODUCT_URL'), blogURL: origin(blogURL, 'WEB_BLOG_URL') };
  const runtimeScript = `window.__SITEHUB_CONFIG__ = ${JSON.stringify(config).replace(/</g, '\\u003c').replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029')};\n`;
  const assets = sirv(dist, { etag: true, setHeaders(res, pathname) {
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('Cache-Control', pathname.includes('/assets/') ? 'public, max-age=31536000, immutable' : 'no-cache');
  } });
  const json = (res, status, value) => { res.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' }); res.end(JSON.stringify(value)); };
  const server = http.createServer((req, res) => {
    const pathname = req.url.split('?')[0];
    if (pathname === '/sitehub-runtime.js' && ['GET', 'HEAD'].includes(req.method)) {
      res.writeHead(200, { 'Content-Type': 'application/javascript; charset=utf-8', 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' });
      return res.end(req.method === 'HEAD' ? undefined : runtimeScript);
    }
    if (pathname === '/healthz' && ['GET', 'HEAD'].includes(req.method)) return json(res, 200, { status: 'ok', app: 'product' });
    if (/^\/(?:v1|api|uploads)(?:\/|$)/.test(pathname)) return json(res, 404, { code: 404, message: 'API traffic must use the gateway' });
    if (!['GET', 'HEAD'].includes(req.method)) return json(res, 405, { code: 405, message: 'Method not allowed' });
    assets(req, res, () => {
      if (!req.headers.accept?.includes('text/html') || pathname.startsWith('/assets/')) {
        return json(res, 404, { code: 404, message: 'Not found' });
      }
      // Browser deep links can include a period, e.g. a document's versioned slug.
      req.url = '/index.html';
      assets(req, res);
    });
  });
  server.requestTimeout = 600000;
  return server;
}
