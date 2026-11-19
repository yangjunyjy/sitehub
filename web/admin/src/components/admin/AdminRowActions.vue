<template>
  <div class="admin-row-actions" role="group" aria-label="行操作"><slot /></div>
</template>
<style lang="scss" scoped>
.admin-row-actions {
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  gap: 10px;
  vertical-align: middle;
}
</style>
