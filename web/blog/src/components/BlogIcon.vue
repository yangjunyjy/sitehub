<script setup>
defineProps({ name: String, size: { type: [String, Number], default: 20 } });
const paths = {
  search: "m21 21-5-5M19 10a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z",
  menu: "M3 6h18M3 12h18M3 18h18",
  close: "m6 6 12 12M6 18 18 6",
  arrow: "M12 4v16m-6-6 6 6 6-6",
  chevron: "m9 5 7 7-7 7",
  heart:
    "M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z",
  github:
    "M9 19c-4 1-4-2-6-2m12 5v-4c0-1 .1-1.8-.5-2.5 3-.3 6-1.5 6-6.5 0-1.4-.5-2.6-1.3-3.5.1-.3.6-1.7-.2-3.5 0 0-1.1-.3-3.6 1.3a12 12 0 0 0-6.8 0C6.1 1.7 5 2 5 2c-.8 1.8-.3 3.2-.2 3.5C4 6.4 3.5 7.6 3.5 9c0 5 3 6.2 6 6.5-.5.5-.6 1.1-.5 2.5v4",
  mail: "M3 5h18v14H3V5Zm0 1 9 7 9-7",
  clock: "M12 8v5l3 2M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0Z",
  folder: "M3 6h6l2 2h10v12H3V6Z",
  tag: "M3 3h8l10 10-8 8L3 11V3Zm4 4h.01",
  user: "M20 21v-2a7 7 0 0 0-14 0v2M17 7a5 5 0 1 1-10 0 5 5 0 0 1 10 0Z",
  link: "m10 13 4-4m-6 6-2 2a4 4 0 0 1-6-6l5-5a4 4 0 0 1 6 0m2 3 2-2a4 4 0 0 1 6 6l-5 5a4 4 0 0 1-6 0",
  feather: "M20 4c-4-4-10 0-12 3l-3 9 5 1 7-3c4-3 6-7 3-10ZM3 22 16 9",
  moon: "M21 13A9 9 0 0 1 11 3a9 9 0 1 0 10 10Z",
};
</script>
<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="paths[name] || paths.feather" />
  </svg>
</template>
