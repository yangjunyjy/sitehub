import test from 'node:test'
import assert from 'node:assert/strict'
import { webcrypto } from 'node:crypto'
import { createUUID } from '../src/utils/uuid.js'

const uuidV4 = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/

function withCrypto(t, value) {
  const descriptor = Object.getOwnPropertyDescriptor(globalThis, 'crypto')
  Object.defineProperty(globalThis, 'crypto', { configurable: true, value })
  t.after(() => {
    if (descriptor) Object.defineProperty(globalThis, 'crypto', descriptor)
    else delete globalThis.crypto
  })
}

test('uses the native UUID API with its Crypto receiver', (t) => {
  const crypto = {
    randomUUID() {
      assert.equal(this, crypto)
      return '6e0f0544-e700-4cba-97c5-adab5eb8e191'
    },
  }
  withCrypto(t, crypto)
  assert.equal(createUUID(), '6e0f0544-e700-4cba-97c5-adab5eb8e191')
})

test('HTTP fallback generates independent UUID v4 values using secure random bytes', (t) => {
  withCrypto(t, { getRandomValues: webcrypto.getRandomValues.bind(webcrypto) })
  const values = Array.from({ length: 1000 }, () => createUUID())
  assert(values.every((value) => uuidV4.test(value)))
  assert.equal(new Set(values).size, values.length)
})

test('fallback preserves leading zeroes and sets UUID version and variant bits', (t) => {
  const crypto = {
    getRandomValues(bytes) {
      assert.equal(this, crypto)
      bytes.fill(0)
      return bytes
    },
  }
  withCrypto(t, crypto)
  assert.equal(createUUID(), '00000000-0000-4000-8000-000000000000')
})

test('missing random APIs produce a recoverable error', (t) => {
  withCrypto(t, undefined)
  assert.throws(() => createUUID(), /无法生成操作编号/)
})
