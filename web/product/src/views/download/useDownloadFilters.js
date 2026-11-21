import { computed, ref, watch } from "vue";
import {
  filterDownloadCatalog,
  downloadGroups,
} from "@/utils/download-catalog.mjs";

export function useDownloadFilters(downloads, route, router) {
  const keyword = ref(typeof route.query.q === "string" ? route.query.q : "");
  watch(
    () => route.query.q,
    (value) => {
      keyword.value = typeof value === "string" ? value : "";
    },
  );
  const filterFields = [
    { key: "product", label: "产品 / 工具" },
    { key: "edition", label: "版本类型" },
    { key: "platform", label: "运行平台" },
    { key: "arch", label: "处理器架构" },
    { key: "version", label: "版本号" },
  ];
  const filters = computed(() =>
    Object.fromEntries(
      ["group", "q", ...filterFields.map((field) => field.key)].map((key) => [
        key,
        typeof route.query[key] === "string" ? route.query[key] : "",
      ]),
    ),
  );
  const choices = (key) => [
    ...new Set(
      downloads.value.packages.map((item) => item[key]).filter(Boolean),
    ),
  ];
  const availableGroups = computed(() =>
    downloadGroups(downloads.value.groups || [], downloads.value.packages),
  );
  const filteredPackages = computed(() =>
    filterDownloadCatalog(downloads.value.packages, filters.value),
  );
  const catalog = computed(() =>
    downloadGroups(downloads.value.groups || [], filteredPackages.value),
  );
  const filterActive = computed(() =>
    Object.values(filters.value).some(Boolean),
  );
  const setFilter = (key, value) => {
    const query = { ...route.query };
    if (value) query[key] = value;
    else delete query[key];
    router.replace({ path: route.path, query });
  };
  const clearFilters = () => {
    keyword.value = "";
    router.replace({ path: route.path });
  };

  return {
    keyword,
    filterFields,
    filters,
    choices,
    availableGroups,
    filteredPackages,
    catalog,
    filterActive,
    setFilter,
    clearFilters,
  };
}
