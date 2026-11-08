import test from 'node:test'
import assert from 'node:assert/strict'
import {
  normalizePackage,
  filterPackages,
  packageError,
  formatFileSize,
} from '../src/utils/downloads.js'
import { normalizeDownloadURL, downloadHref, isLocalReleaseURL } from '../../shared/downloads.mjs'

test('new download records default to draft and omit UI metadata', () => {
  const item = normalizePackage(
    { id: 'a', platform: 'Linux', url: '/v1/file/local/releases/a.tgz', selected: true },
    { version: 'v1', product: 'Deploy' },
  )
  assert.equal(item.status, 'draft')
  assert.equal(item.version, 'v1')
  assert.equal('selected' in item, false)
  assert.equal(normalizePackage({ platform: 'Windows' }).status, 'draft')
})
test('release filtering applies product platform status and searchable metadata together', () => {
  const items = [
    normalizePackage({
      id: 'a',
      product: 'Deploy',
      platform: 'Linux',
      version: 'v2',
      fileName: 'deploy-arm.tgz',
      status: 'published',
    }),
    normalizePackage({
      id: 'b',
      product: 'Deploy',
      platform: 'Windows',
      version: 'v1',
      status: 'draft',
    }),
  ]
  assert.deepEqual(
    filterPackages(items, {
      product: 'Deploy',
      platform: 'Linux',
      status: 'published',
      keyword: 'ARM',
    }).map((item) => item.id),
    ['a'],
  )
  assert.deepEqual(filterPackages(items, { status: 'archived' }), [])
})
test('published files need a valid link while drafts can be saved without one', () => {
  const draft = normalizePackage({
    product: 'Deploy',
    platform: 'Linux',
    version: 'v1',
    status: 'draft',
  })
  assert.equal(packageError(draft), '')
  assert.notEqual(packageError({ ...draft, status: 'published' }), '')
  assert.notEqual(packageError({ ...draft, url: 'javascript:alert(1)' }), '')
  assert.equal(packageError({ ...draft, status: 'published', url: '/v1/file/local/releases/test.zip' }), '')
  assert.notEqual(packageError({ ...draft, checksum: 'bad' }), '')
  assert.equal(formatFileSize(2 * 1024 * 1024), '2 MB')
  assert.equal(formatFileSize(0), '—')
})

test('native downloads use package authorization and preserve external storage locations', () => {
  const item = normalizePackage({ id: 'arm', product: 'Deploy', platform: 'Linux', version: '1.0', status: 'published', url: '/v1/file/local/releases/stored-id.tgz', fileName: 'deploy_arm.tgz' })
  assert.equal(packageError(item), '')
  assert.equal(isLocalReleaseURL(item.url), true)
  assert.equal(downloadHref(item), '/v1/file/download?id=arm')
  assert.equal(normalizeDownloadURL('https://cdn.example.test/a.zip?signature=xyz'), 'https://cdn.example.test/a.zip?signature=xyz')
  assert.equal(downloadHref({ ...item, url: 'https://cdn.example.test/a.tgz' }), '/v1/file/download?id=arm')
  for (const fileName of ['../a.zip', 'a:stream', '..']) assert.notEqual(packageError({ ...item, fileName }), '')
})
