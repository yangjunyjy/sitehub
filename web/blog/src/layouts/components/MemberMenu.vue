<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
import { useMember } from "@/composables/useMember";

const props = defineProps({ mobileNavOpen: Boolean });
const emit = defineEmits(["open"]);
const route = useRoute();
const { member, logout, expired } = useMember();
const open = ref(false),
  busy = ref(false),
  error = ref("");
const root = ref(null),
  trigger = ref(null),
  first = ref(null),
  signin = ref(null);
function close() {
  open.value = false;
}
function toggle() {
  open.value = !open.value;
  if (open.value) emit("open");
}
async function enter() {
  if (!open.value) toggle();
  await nextTick();
  first.value?.$el.focus();
}
function escape() {
  close();
  trigger.value?.focus();
}
function outside(e) {
  if (!root.value?.contains(e.target)) close();
}
async function signOut() {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    await logout();
  } catch (err) {
    if (err.status === 401) expired();
    else error.value = err.message;
  } finally {
    busy.value = false;
  }
  if (!member.user) {
    close();
    await nextTick();
    signin.value?.$el.focus({ preventScroll: true });
  }
}
watch(() => route.fullPath, close);
watch(
  () => props.mobileNavOpen,
  (value) => {
    if (value) close();
  },
);
onMounted(() => document.addEventListener("pointerdown", outside));
onBeforeUnmount(() => document.removeEventListener("pointerdown", outside));
</script>
<template>
  <div
    ref="root"
    class="member-menu"
    @focusout="root && !root.contains($event.relatedTarget) && close()"
    @keydown.esc.stop="escape"
  >
    <RouterLink
      v-if="!member.user"
      ref="signin"
      class="member-trigger"
      aria-label="会员登录"
      :to="{
        name: 'login',
        query: ['login', 'register'].includes(route.meta.page)
          ? route.query
          : { redirect: route.fullPath },
      }"
    >
      <BlogIcon name="user" :size="18" /><span class="trigger-label">登录</span>
    </RouterLink>
    <template v-else>
      <button
        ref="trigger"
        class="member-trigger"
        type="button"
        aria-label="账号菜单"
        :aria-expanded="open"
        aria-controls="member-options"
        @click="toggle"
        @keydown.down.prevent="enter"
      >
        <BlogIcon name="user" :size="18" /><span class="trigger-label"
          >我的账号</span
        ><BlogIcon class="caret" name="chevron" :size="12" />
      </button>
      <div v-show="open" id="member-options" class="member-options">
        <RouterLink ref="first" :to="{ name: 'account' }" @click="close"
          ><BlogIcon name="user" :size="15" />个人中心</RouterLink
        >
        <button type="button" :disabled="busy" @click="signOut">
          <BlogIcon name="logout" :size="15" />{{
            busy ? "正在退出…" : "退出登录"
          }}
        </button>
        <p v-if="error" role="alert">{{ error }}</p>
      </div>
    </template>
  </div>
</template>
<style scoped>
.member-menu {
  position: relative;
}
.member-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  min-height: 40px;
  padding: 0 4px;
  font-size: 13px;
  background: transparent;
  color: inherit;
  border: 0;
}
.caret {
  transform: rotate(90deg);
  transition: transform 0.18s;
}
[aria-expanded="true"] .caret {
  transform: rotate(-90deg);
}
.member-options {
  position: absolute;
  top: calc(100% + 5px);
  right: 0;
  width: 154px;
  padding: 6px;
  background: #fffdfa;
  color: var(--journal-ink);
  border: 1px solid var(--journal-rule);
  border-radius: 7px;
  box-shadow: 0 8px 28px #29232d20;
}
.member-options a,
.member-options button {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 9px 10px;
  color: inherit;
  background: none;
  border: none;
  font-size: 14px;
  text-align: left;
  border-radius: 3px;
}
.member-options :is(a, button):is(:hover, :focus-visible) {
  background: var(--journal-wash);
  color: var(--journal-accent);
}
.member-options p {
  font-size: 12px;
  color: #963623;
  margin: 4px;
}
@media (max-width: 730px) {
  .trigger-label {
    display: none;
  }
  .member-trigger {
    min-width: 40px;
  }
}
</style>
