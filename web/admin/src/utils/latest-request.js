// A request owns its result, error and loading state until it is superseded.
// Abort is an optimization; the sequence also protects fetchers that ignore it.
export function createLatestRequest() {
  let sequence = 0
  let controller
  return {
    cancel() {
      sequence++
      controller?.abort()
      controller = undefined
    },
    async run(fetcher, { start, success, failure, finish } = {}) {
      const current = ++sequence
      controller?.abort()
      const request = new AbortController()
      controller = request
      start?.()
      try {
        const result = await fetcher(request.signal)
        if (current !== sequence || request.signal.aborted) return false
        success?.(result)
        return true
      } catch (error) {
        if (current !== sequence || request.signal.aborted) return false
        failure?.(error)
        return false
      } finally {
        if (current === sequence) {
          controller = undefined
          finish?.()
        }
      }
    },
  }
}
