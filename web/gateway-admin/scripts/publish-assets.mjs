// Keep the last successful Vue build available to go build and gopls.
import { copyFile, mkdir, readdir, rename, rm } from 'node:fs/promises'
import { randomUUID } from 'node:crypto'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const source = fileURLToPath(new URL('../dist/', import.meta.url))
const target = fileURLToPath(new URL('../static/', import.meta.url))

async function publish(file) {
  const destination = path.join(target, file)
  const directory = path.dirname(destination)
  await mkdir(directory, { recursive: true })
  // Go embed skips underscore-prefixed files inside the embedded directory.
  const temporary = path.join(directory, '_publish-' + randomUUID())
  try {
    await copyFile(path.join(source, file), temporary)
    await rename(temporary, destination)
  } finally {
    await rm(temporary, { force: true })
  }
}

async function filesIn(directory, prefix = '') {
  const files = []
  for (const entry of await readdir(path.join(directory, prefix), { withFileTypes: true })) {
    const file = path.join(prefix, entry.name)
    if (entry.isDirectory()) files.push(...await filesIn(directory, file))
    else if (entry.isFile()) files.push(file)
  }
  return files
}

const next = new Set(await filesIn(source))
if (!next.has('index.html')) throw new Error('Vue build is missing index.html')
await mkdir(target, { recursive: true })
const previous = await filesIn(target)
for (const file of next) {
  if (file !== 'index.html') await publish(file)
}
// Make the new entry point visible only after its hashed dependencies are in place.
await publish('index.html')
for (const file of previous) {
  if (!next.has(file) && !path.basename(file).startsWith('_publish-')) {
    await rm(path.join(target, file), { force: true })
  }
}
