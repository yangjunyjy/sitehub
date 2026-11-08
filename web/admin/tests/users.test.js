import test from 'node:test'
import assert from 'node:assert/strict'
import { normalizeUserRow } from '../src/utils/users.js'

test('legacy duplicate relations cannot inflate role cells or assignment IDs', () => {
  const original = {
    id: 1,
    nickName: '测试用户',
    password: 'legacy-hash',
    roles: [
      ...Array(64).fill({ id: 1, roleName: '管理员' }),
      { id: '1' },
      { id: 2, roleName: '编辑' },
      null,
    ],
  }
  const row = normalizeUserRow(original)
  assert.deepEqual(row.roleIds, [1, 2])
  assert.equal(row.roles.length, 2)
  assert.equal(row.nickname, '测试用户')
  assert(!Object.hasOwn(row, 'password'))
  assert.equal(original.roles.length, 67)
  assert.deepEqual(normalizeUserRow({ roles: null }).roles, [])
})
