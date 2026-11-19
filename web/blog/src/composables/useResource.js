import { onBeforeUnmount, ref } from "vue";
export function useResource(fetcher, initial = null) {
  const data = ref(initial),
    loading = ref(false),
    error = ref("");
  let generation = 0,
    disposed = false;
  async function load(...args) {
    const id = ++generation;
    loading.value = true;
    error.value = "";
    try {
      const result = await fetcher(...args);
      if (!disposed && generation === id) data.value = result.data;
    } catch (failure) {
      if (!disposed && generation === id)
        error.value = failure.message || "暂时无法加载，请重试";
    } finally {
      if (!disposed && generation === id) loading.value = false;
    }
  }
  onBeforeUnmount(() => {
    disposed = true;
    generation++;
  });
  return { data, loading, error, load };
}
