import { computed, ref, watch } from 'vue'
import {
  columnGroup,
  defaultTableAppearance,
  moveTableColumn,
  normalizeTablePreferences,
  tablePreferenceStorageKey,
} from '@/utils/admin-table'

export function useAdminTablePreferences(props) {
  let userID
  try {
    userID = JSON.parse(sessionStorage.getItem('userInfo') || '{}').id
  } catch {
    /* Anonymous session or unavailable storage. */
  }
  const storageKey = computed(() => tablePreferenceStorageKey(props.preferenceKey, userID))
  const preferences = ref(normalizeTablePreferences(props.columns, null))
  watch(
    [storageKey, () => props.columns],
    () => {
      let stored
      try {
        if (storageKey.value) stored = JSON.parse(localStorage.getItem(storageKey.value) || 'null')
      } catch {
        /* Broken or unavailable storage uses defaults. */
      }
      preferences.value = normalizeTablePreferences(props.columns, stored)
    },
    { immediate: true, deep: true },
  )

  const orderedColumns = computed(() =>
    preferences.value.order
      .map((key) => props.columns.find((column) => column.key === key))
      .filter(Boolean),
  )
  const visibleColumns = computed(() =>
    orderedColumns.value.filter((column) => !preferences.value.hidden.includes(column.key)),
  )
  const tableProps = computed(() => ({
    columns: visibleColumns.value,
    ...preferences.value.appearance,
  }))

  function commit(value) {
    preferences.value = normalizeTablePreferences(props.columns, value)
    try {
      if (storageKey.value)
        localStorage.setItem(storageKey.value, JSON.stringify(preferences.value))
    } catch {
      /* Changes still work for the current page. */
    }
  }
  function canHide(column) {
    if (column.locked) return false
    if (preferences.value.hidden.includes(column.key) || column.utility) return true
    return (
      visibleColumns.value.filter(
        (item) => !item.utility && !['selection', 'index', 'expand'].includes(item.type),
      ).length > 1
    )
  }
  function setVisible(key, visible) {
    const column = props.columns.find((item) => item.key === key)
    if (!column || (!visible && !canHide(column))) return
    const hidden = preferences.value.hidden.filter((item) => item !== key)
    if (!visible) hidden.push(key)
    commit({ ...preferences.value, hidden })
  }
  function move(source, target) {
    commit({
      ...preferences.value,
      order: moveTableColumn(props.columns, preferences.value.order, source, target),
    })
  }
  function neighbor(key, step) {
    const column = props.columns.find((item) => item.key === key)
    if (!column || column.locked) return null
    const peers = orderedColumns.value.filter(
      (item) => !item.locked && columnGroup(item) === columnGroup(column),
    )
    return peers[peers.findIndex((item) => item.key === key) + step]?.key
  }
  function resetColumns() {
    commit({ ...preferences.value, order: [], hidden: [] })
  }
  function setAppearance(key, value) {
    if (key in defaultTableAppearance)
      commit({
        ...preferences.value,
        appearance: { ...preferences.value.appearance, [key]: value },
      })
  }
  function resetAppearance() {
    commit({ ...preferences.value, appearance: { ...defaultTableAppearance } })
  }
  function setDensity(value) {
    commit({ ...preferences.value, density: value })
  }

  return {
    preferences,
    orderedColumns,
    tableProps,
    canHide,
    setVisible,
    move,
    neighbor,
    resetColumns,
    setAppearance,
    resetAppearance,
    setDensity,
  }
}
