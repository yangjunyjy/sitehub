import test from 'node:test'
import assert from 'node:assert/strict'
import { createLatestRequest } from '../src/utils/latest-request.js'

function deferred() {
  let resolve, reject
  const promise = new Promise((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}

test('a slower request cannot overwrite the latest rows, error or loading state', async () => {
  const request = createLatestRequest()
  const old = deferred(),
    latest = deferred()
  const state = { rows: [], error: '', loading: false }
  const handlers = {
    start: () => {
      state.loading = true
      state.error = ''
    },
    success: (rows) => {
      state.rows = rows
    },
    failure: (error) => {
      state.error = error.message
    },
    finish: () => {
      state.loading = false
    },
  }
  let oldSignal
  const oldRun = request.run((signal) => {
    oldSignal = signal
    return old.promise
  }, handlers)
  const latestRun = request.run(() => latest.promise, handlers)
  assert.equal(oldSignal.aborted, true)
  old.reject(new Error('old failure'))
  assert.equal(await oldRun, false)
  assert.deepEqual(state, { rows: [], error: '', loading: true })
  latest.resolve(['latest result'])
  assert.equal(await latestRun, true)
  assert.deepEqual(state, { rows: ['latest result'], error: '', loading: false })
})

test('a superseded success is ignored even if its fetcher ignores cancellation', async () => {
  const request = createLatestRequest(),
    old = deferred()
  const accepted = []
  const oldRun = request.run(() => old.promise, { success: (value) => accepted.push(value) })
  await request.run(() => Promise.resolve('new'), { success: (value) => accepted.push(value) })
  old.resolve('old')
  await oldRun
  assert.deepEqual(accepted, ['new'])
})

test('unmount cancellation blocks all callbacks from a pending request', async () => {
  const request = createLatestRequest(),
    pending = deferred(),
    calls = []
  const run = request.run(() => pending.promise, {
    success: () => calls.push('success'),
    failure: () => calls.push('failure'),
    finish: () => calls.push('finish'),
  })
  request.cancel()
  pending.resolve('late')
  await run
  assert.deepEqual(calls, [])
})
