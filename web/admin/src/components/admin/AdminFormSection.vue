<template>
  <section class="admin-form-section">
    <header v-if="title || $slots.header" class="admin-form-heading">
      <slot name="header"
        ><div>
          <h2>{{ title }}</h2>
          <p v-if="description">{{ description }}</p>
        </div>
        <div class="admin-form-actions"><slot name="actions" /></div
      ></slot>
    </header>
    <div class="admin-form-content"><slot /></div>
    <footer v-if="$slots.footer" class="admin-form-footer"><slot name="footer" /></footer>
  </section>
</template>
<script setup>
defineProps({ title: String, description: String })
</script>
