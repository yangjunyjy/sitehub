<script setup>
import { computed } from "vue";
const props = defineProps({ fields: Array });
const groups = computed(() => {
  const result = new Map();
  for (const field of props.fields) {
    if (field.show && !field.show()) continue;
    const name = field.group || "基本信息";
    if (!result.has(name)) result.set(name, []);
    result.get(name).push(field);
  }
  return [...result];
});
</script>
<template>
  <component
    v-for="[name, items] in groups"
    :key="name"
    :is="name === '高级设置' ? 'details' : 'section'"
    class="field-group"
  >
    <component :is="name === '高级设置' ? 'summary' : 'h3'">{{
      name
    }}</component>
    <div class="editor-grid">
      <label v-for="f in items" :key="f.label" :class="{ full: f.multiline }">
        <span>{{ f.label }}</span>
        <select
          v-if="f.choices"
          v-model="f.value"
          :aria-label="f.label"
          :required="f.required"
        >
          <option v-for="v in f.choices" :value="v" :key="v">
            {{ v || f.emptyLabel || "无" }}
          </option>
        </select>
        <textarea
          v-else-if="f.multiline"
          v-model="f.value"
          rows="3"
          :aria-label="f.label"
          :required="f.required"
          :placeholder="f.placeholder"
        ></textarea>
        <input
          v-else
          :value="f.value"
          @input="f.value = $event.target.value"
          :aria-label="f.label"
          :type="f.type || 'text'"
          :required="f.required"
          :readonly="f.readOnly"
          :pattern="f.pattern"
          :step="f.step"
          :placeholder="f.placeholder"
        />
        <small v-if="f.hint">{{ f.hint }}</small>
      </label>
    </div>
  </component>
</template>
