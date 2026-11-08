import { onBeforeUnmount, onActivated, onDeactivated, ref } from 'vue'
import { createLatestRequest } from '@/utils/latest-request'

// Pages own submitted filters and API shapes; this composable owns asynchronous
// list state. Refresh and pagination never read an unsubmitted search form.
export function useAdminList({
  fetcher,
  params,
  rows,
  pagination,
  selectRows,
  errorMessage,
  onLoaded,
}) {
  const loading = ref(false)
  const loadError = ref('')
  const request = createLatestRequest()
  let activated = false
  const loadList = () =>
    request.run((signal) => fetcher(params(), { signal, silentError: true }), {
      start() {
        loading.value = true
        loadError.value = ''
      },
      success(result) {
        rows.value = selectRows(result.data || {})
        pagination.total = Math.max(0, Number(result.data?.total) || 0)
        onLoaded?.()
      },
      failure(error) {
        loadError.value = error.response?.data?.message || error.message || errorMessage
        rows.value = []
        pagination.total = 0
      },
      finish() {
        loading.value = false
      },
    })
  onBeforeUnmount(() => request.cancel())
  onDeactivated(() => {
    request.cancel()
    loading.value = false
  })
  onActivated(() => {
    if (activated) loadList()
    activated = true
  })
  return { loading, loadError, loadList }
}
