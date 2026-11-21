import test from 'node:test'
import assert from 'node:assert/strict'
import { documentPayload, validateDocument } from '../src/utils/documents.js'

const draft = { id: 'one', slug: 'overview', title: 'Overview', summary: 'Intro', content: '' }
test('drafts may have an empty body and exclude directory and publication metadata', () => {
  assert.deepEqual(documentPayload({ ...draft, revision: '22', publishedRevision: '21', groupId: '3' }), draft)
  assert.deepEqual(validateDocument(draft), draft)
})
test('document drafts validate their title and public slug before saving', () => {
  assert.equal(validateDocument({ ...draft, slug: ' quick-start ' }).slug, 'quick-start')
  assert.throws(() => validateDocument({ ...draft, slug: 'invalid/slug' }), /访问标识/)
  assert.throws(() => validateDocument({ ...draft, title: '' }), /标题/)
})
