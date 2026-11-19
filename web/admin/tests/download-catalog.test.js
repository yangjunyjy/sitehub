import test from 'node:test'
import assert from 'node:assert/strict'
import { filterDownloadCatalog, downloadGroups } from '../src/utils/download-catalog.mjs'
import { normalizePackage, filterPackages } from '../src/utils/downloads.js'

const groups = [{ id: 'tools', name: '工具' }, { id: 'deploy', name: 'Deploy' }]
const packages = [
  { id: 'cluster', groupId: 'deploy', product: 'Deploy', edition: '分布式版', version: '2.1', platform: 'Linux', arch: 'ARM64', sortOrder: 20 },
  { id: 'single', groupId: 'deploy', product: 'Deploy', edition: '单节点 API 版', version: '2.0', platform: 'Linux', arch: 'AMD64', sortOrder: 10 },
  { id: 'tool', groupId: 'tools', product: '诊断工具', platform: '跨平台', fileName: 'CHECK.zip' },
  { id: 'legacy', product: '旧文件' },
]

test('catalog combines deployment edition, platform and keyword without mixing products', () => {
  assert.deepEqual(filterDownloadCatalog(packages, { group: 'deploy', edition: '单节点 API 版', platform: 'Linux' }).map(x => x.id), ['single'])
  assert.deepEqual(filterDownloadCatalog(packages, { q: 'check.ZIP' }).map(x => x.id), ['tool'])
  assert.deepEqual(filterDownloadCatalog(packages, { group: 'deploy', arch: 'x86' }), [])
  assert.deepEqual(filterDownloadCatalog(packages, { group: 'deploy' }).map(x => x.id), ['single', 'cluster'])
})

test('group ordering and legacy files survive grouping and admin normalization', () => {
  assert.deepEqual(downloadGroups(groups, packages).map(x => x.id), ['tools', 'deploy', 'ungrouped'])
  assert.equal(downloadGroups([], [packages[3]])[0].packages[0].id, 'legacy')
  const normalized = packages.map(x => normalizePackage(x))
  assert.equal(normalized[0].edition, '分布式版')
  assert.equal(normalized[0].sortOrder, 20)
  assert.deepEqual(filterPackages(normalized, { groupId: 'deploy', edition: '分布式版' }).map(x => x.id), ['cluster'])
})
