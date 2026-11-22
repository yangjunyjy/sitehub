<script setup>
import MemberAvatar from "@/components/member/MemberAvatar.vue";
import ProfileEditor from "./ProfileEditor.vue";
import { computed, ref } from "vue";
import { displayDate } from "@/utils/format";
const props = defineProps({ user: { type: Object, required: true } });
const editing = ref(false);
const editButton = ref(null);
const saved = ref(false);
function finish(updated) {
  editing.value = false;
  saved.value = updated;
  editButton.value?.focus();
}
const name = computed(
  () =>
    props.user.nickname ||
    props.user.username ||
    props.user.email?.split("@")[0] ||
    "读者",
);
const joined = computed(() => displayDate(props.user.createdAt));
</script>
<template>
  <header class="member-card">
    <div>
      <h1>个人中心</h1>
      <div class="identity">
        <MemberAvatar :user="user" style="--avatar-size: 48px" />
        <div>
          <strong>{{ name }}</strong>
          <p v-if="user.email">{{ user.email }}</p>
        </div>
        <button
          ref="editButton"
          class="edit-profile"
          type="button"
          :aria-expanded="editing"
          @click="
            saved = false;
            editing = !editing;
          "
        >
          编辑资料
        </button>
      </div>
    </div>
    <div class="bookplate" aria-hidden="true"></div>
    <small v-if="joined !== '日期待补充'"
      >加入于 <time :datetime="joined">{{ joined }}</time></small
    >
  </header>
  <ProfileEditor v-if="editing" :user="user" @finish="finish" />
  <p v-if="saved" role="status" class="saved">个人资料已保存</p>
</template>
<style scoped>
.member-card {
  position: relative;
  display: flex;
  align-items: end;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 22px;
  border-bottom: 2px solid var(--journal-ink);
}
.bookplate {
  margin-left: auto;
  width: clamp(140px, 23vw, 260px);
  height: 86px;
  background: url("../../../assets/skies/archive.webp") center 60% / cover;
  border: 5px solid #fff;
  outline: 1px solid var(--journal-rule);
  transform: rotate(-2deg);
}
.member-card > small {
  position: absolute;
  right: 0;
  top: 0;
}
@media (max-width: 650px) {
  .bookplate {
    display: none;
  }
  .member-card > small {
    position: static;
  }
}
h1 {
  margin: 0 0 16px;
  font: 600 clamp(30px, 4vw, 48px)/1.2 var(--journal-serif);
}
.identity {
  display: flex;
  align-items: center;
  gap: 12px;
  overflow-wrap: anywhere;
}
.identity {
  flex-wrap: wrap;
}
.edit-profile {
  padding: 6px 12px;
  border: 1px solid var(--journal-rule);
  border-radius: 20px;
  background: #ffffffa6;
  color: var(--journal-accent);
  cursor: pointer;
}
.saved {
  margin-top: 12px;
}
.identity strong {
  font-size: 15px;
}
p,
small {
  margin: 0;
  font-size: 12px;
  color: var(--journal-muted);
}
</style>
