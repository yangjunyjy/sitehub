import test from 'node:test'
import assert from 'node:assert/strict'
import { websiteURL, previewWebsite } from '../src/utils/website.js'
import { websiteURL as downloadWebsiteURL } from '../src/utils/downloads.js'

test('all previews resolve against runtime configuration or the current public origin', () => {
  const previous = globalThis.window
  try {
    assert.equal(downloadWebsiteURL, websiteURL)
    for (const { origin, configured, path, expected } of [
      { origin: 'http://203.0.113.20:8082', path: '/download', expected: 'http://203.0.113.20:8082/download' },
      { origin: 'http://203.0.113.20:8082', configured: '', path: '/#capabilities', expected: 'http://203.0.113.20:8082/#capabilities' },
      { origin: 'https://site.example', path: '/blog/article?q=preview#details', expected: 'https://site.example/blog/article?q=preview#details' },
      { origin: 'https://admin.example', configured: 'https://site.example', path: '/download?product=deploy#linux', expected: 'https://site.example/download?product=deploy#linux' },
      { origin: 'https://admin.example', configured: 'https://site.example/base/', path: 'manual/install?q=test#step-2', expected: 'https://site.example/base/manual/install?q=test#step-2' },
      { origin: 'https://admin.example', configured: 'https://site.example/base/?lang=zh#start', path: '?q=test#step-2', expected: 'https://site.example/base/?q=test#step-2' },
    ]) {
      globalThis.window = { location: { origin }, ...(configured === undefined ? {} : { __SITEHUB_CONFIG__: { websiteURL: configured } }) }
      assert.equal(websiteURL(path), expected)
      assert.equal(downloadWebsiteURL(path), expected)
    }
    let opened
    globalThis.window = { location: { origin: 'https://site.example' }, open: (...args) => { opened = args } }
    assert.equal(websiteURL(), 'https://site.example/')
    previewWebsite('/#capabilities')
    assert.deepEqual(opened, ['https://site.example/#capabilities', '_blank', 'noopener,noreferrer'])
  } finally {
    if (previous === undefined) delete globalThis.window
    else globalThis.window = previous
  }
})
