import { cpSync, mkdirSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
const require = createRequire(import.meta.url)
const target = fileURLToPath(new URL('../public/vendor/vditor/dist', import.meta.url))
mkdirSync(target, { recursive: true })
cpSync(join(dirname(require.resolve('vditor/package.json')), 'dist'), target, { recursive: true })
