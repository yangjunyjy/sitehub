import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const [app, destination] = process.argv.slice(2)
if (!['admin', 'product', 'blog'].includes(app)) throw new Error('Unknown frontend')
if (!destination) throw new Error('Usage: node scripts/build-frontend-runtime.mjs <app> <output.mjs>')
const webRequire = createRequire(resolve(projectRoot, 'web/package.json'))
const { build } = webRequire('esbuild')
await build({
  entryPoints: [resolve(projectRoot, `web/${app}/server/start.mjs`)],
  outfile: resolve(destination),
  bundle: true,
  platform: 'node',
  format: 'esm',
  target: 'node22',
  minify: true,
  sourcemap: false,
  legalComments: 'eof',
})
