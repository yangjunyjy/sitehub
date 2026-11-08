import test from 'node:test'
import assert from 'node:assert/strict'
import {
  normalizeTablePreferences,
  moveTableColumn,
  tablePreferenceStorageKey,
} from '../src/utils/admin-table.js'

const columns = [
  { key: 'selection', type: 'selection', locked: true, utility: true },
  { key: 'name' },
  { key: 'status' },
  { key: 'actions', utility: true, fixed: 'right' },
]

test('cached preferences reconcile removed/new columns without storing arbitrary fields', () => {
  const saved = {
    version: 1,
    order: ['actions', 'retired', 'status', 'status'],
    hidden: ['retired', 'status'],
    appearance: { border: true, stripe: 'false' },
    data: ['private record'],
    density: 'compact',
  }
  const result = normalizeTablePreferences(columns, saved)
  assert.deepEqual(result.order, ['selection', 'status', 'name', 'actions'])
  assert.deepEqual(result.hidden, ['status'])
  assert.deepEqual(result.appearance, { border: true, stripe: false, headerBackground: false })
  assert.equal(result.density, 'compact')
  assert.equal(result.data, undefined)
})

test('locked structural columns and at least one data column remain visible', () => {
  const result = normalizeTablePreferences(columns, {
    version: 1,
    hidden: columns.map((column) => column.key),
  })
  assert.deepEqual(result.hidden, ['status', 'actions'])
  const tree = normalizeTablePreferences([{ key: 'tree', locked: true }, { key: 'status' }], {
    version: 1,
    order: ['status', 'tree'],
    hidden: ['tree'],
  })
  assert.deepEqual(tree.order, ['tree', 'status'])
  assert.deepEqual(tree.hidden, [])
})

test('reordering stays within fixed groups and does not mutate the original order', () => {
  const order = ['selection', 'name', 'status', 'actions']
  assert.deepEqual(moveTableColumn(columns, order, 'status', 'name'), [
    'selection',
    'status',
    'name',
    'actions',
  ])
  for (const [source, target] of [
    ['name', 'actions'],
    ['selection', 'status'],
    ['status', 'selection'],
    ['missing', 'name'],
  ]) {
    assert.deepEqual(moveTableColumn(columns, order, source, target), order)
  }
  assert.deepEqual(order, ['selection', 'name', 'status', 'actions'])
})

test('invalid preference versions and field types recover to safe defaults', () => {
  for (const saved of [
    null,
    2,
    'invalid',
    { version: 99, hidden: ['name'] },
    { version: 1, order: {}, hidden: {}, density: 'huge' },
  ]) {
    assert.deepEqual(
      normalizeTablePreferences(columns, saved),
      normalizeTablePreferences(columns, null),
    )
  }
})

test('preference keys isolate accounts and pages', () => {
  assert.notEqual(tablePreferenceStorageKey('roles', 1), tablePreferenceStorageKey('roles', 2))
  assert.notEqual(tablePreferenceStorageKey('roles', 1), tablePreferenceStorageKey('users', 1))
  assert.notEqual(tablePreferenceStorageKey('a:b', 'c'), tablePreferenceStorageKey('b', 'c:a'))
  assert.equal(tablePreferenceStorageKey('', 1), '')
})
