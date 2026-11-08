import test from 'node:test'
import assert from 'node:assert/strict'
import { uint64, stringifyJSON, queryString, viewValues, allPages } from './native-api.mjs'

test('native requests preserve uint64 precision and collection query fields', () => {
  assert.equal(stringifyJSON({ id: uint64('18446744073709551615'), enabled: false }), '{"id":18446744073709551615,"enabled":false}')
  assert.throws(() => uint64(Number.MAX_SAFE_INTEGER + 1))
  assert.throws(() => uint64('-1'))
  assert.throws(() => uint64('18446744073709551616'))
  assert.equal(queryString({ tagIDs: ['1', '2'], status: 0, trash: false }), 'tagIDs=1&tagIDs=2&status=0&trash=false')
  assert.deepEqual(viewValues({ id: '18446744073709551615', version: '3', total: '4', status: '0' }), { id: '18446744073709551615', version: '3', total: 4, status: 0 })
})

test('catalog aggregation requests all pages and propagates failures', async () => {
  const result = await allPages(async ({ page }) => ({ code: 0, data: { items: [page], total: '3' } }))
  assert.deepEqual(result.data, [1, 2, 3])
  await assert.rejects(allPages(async () => { throw new Error('offline') }), /offline/)
})
