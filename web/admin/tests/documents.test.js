import test from 'node:test'
import assert from 'node:assert/strict'
import { flattenDocuments, documentPayload, validateDocument } from '../src/utils/documents.js'
const docs = {
  groups: [
    {
      title: 'Getting started',
      items: [
        { id: 'one', slug: 'overview', title: 'Overview', summary: 'Intro', content: '# Hello' },
      ],
    },
    { title: 'Guides', items: [] },
  ],
}
test('document editing excludes list metadata and validates local fields', () => {
  const row = flattenDocuments(docs)[0]
  assert.deepEqual(documentPayload(row), docs.groups[0].items[0])
  assert.equal(validateDocument({ ...row, slug: ' quick-start ' }).slug, 'quick-start')
  assert.throws(() => validateDocument({ ...row, slug: 'invalid/slug' }), /访问标识/)
  assert.throws(() => validateDocument({ ...row, content: '  ' }), /正文/)
  assert.throws(() => validateDocument({ ...row, title: '' }), /标题/)
})

test('cached metadata and internal IDs cannot block create or edit before server validation', () => {
  const row = flattenDocuments(docs)[0]
  // The server checks current ownership; cached rows may have been renamed or deleted in another tab.
  assert.equal(validateDocument({ ...row, id: 'two', slug: 'one' }, docs).slug, 'one')
  assert.equal(validateDocument({ ...row, id: 'two' }, docs).slug, 'overview')
  assert.equal(validateDocument(row, docs, 'one').id, 'one')
})
