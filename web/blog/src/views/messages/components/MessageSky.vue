<script setup>
import Meteors from "@/components/inspira/Meteors.vue";
import LiquidGlass from "@/components/inspira/LiquidGlass.vue";
import { computed, onMounted, reactive, ref } from "vue";
import { messages, sendMessage } from "@/api";
import { useResource } from "@/composables/useResource";
import BlogIcon from "@/components/BlogIcon.vue";
import EarthGlobe from "./EarthGlobe.vue";
const { data, loading, error, load } = useResource(() =>
  messages({ page: 1, pageSize: 40 }),
);
const form = reactive({ nickname: "", content: "" }),
  submitting = ref(false),
  success = ref(""),
  submitError = ref("");
const visibleMessages = computed(() => (data.value?.items || []).slice(0, 18));
function position(index) {
  return {
    "--x": `${8 + ((index * 29) % 82)}%`,
    "--y": `${15 + ((index * 19) % 68)}%`,
    "--delay": `${-(index % 7)}s`,
    "--duration": `${9 + (index % 5)}s`,
  };
}
async function submit() {
  if (submitting.value) return;
  submitting.value = true;
  submitError.value = "";
  success.value = "";
  try {
    const { data: result } = await sendMessage({
      nickname: form.nickname.trim(),
      content: form.content.trim(),
    });
    if (result?.status !== "pending")
      throw new Error("留言状态尚未确认，请稍后再查看");
    success.value = "你的星光已送达，审核通过后会出现在这片星空。";
    form.content = "";
  } catch (failure) {
    submitError.value = failure.message;
  } finally {
    submitting.value = false;
  }
}
onMounted(load);
</script>
<template>
  <section class="message-universe">
    <div class="star-field" aria-hidden="true"></div>
    <Meteors :count="14" />
    <div class="earth" aria-hidden="true">
      <EarthGlobe />
    </div>
    <div class="message-bubbles" aria-label="访客留言">
      <div
        v-for="(message, index) in visibleMessages"
        :key="message.id"
        class="message-bubble"
        :style="position(index)"
      >
        <span>{{ message.nickname || "路过的星星" }}</span>
        <p>{{ message.content }}</p>
      </div>
    </div>
    <div class="message-center">
      <h1>留言板</h1>
      <form class="message-form" @submit.prevent="submit">
        <LiquidGlass :radius="20" />
        <label class="nickname-field"
          ><BlogIcon name="user" :size="15" /><input
            v-model="form.nickname"
            maxlength="40"
            aria-label="你的昵称（选填）"
            placeholder="怎么称呼你？（选填）" /></label
        ><label class="message-field">
          <textarea
            v-model="form.content"
            required
            minlength="1"
            maxlength="1000"
            aria-label="留言内容"
            placeholder="此刻，有什么想说的…"
          ></textarea
          ><button :disabled="submitting" aria-label="发送留言" type="submit">
            <BlogIcon name="arrow" :size="22" /></button
        ></label>
        <div class="message-form__footer">
          <span>{{ form.content.length }} / 1000</span
          ><span>{{ submitting ? "正在发送…" : "留言审核后公开展示" }}</span>
        </div>
      </form>
      <p
        v-if="success"
        class="message-feedback message-feedback--success"
        role="status"
      >
        {{ success }}
      </p>
      <p v-if="submitError" class="message-feedback" role="alert">
        {{ submitError }}
      </p>
      <p v-if="error" class="message-load-state">
        星空留言暂时无法加载 <button @click="load">重试</button>
      </p>
      <p v-else-if="loading" class="message-load-state" role="status">
        正在收集远方的星光…
      </p>
      <p v-else-if="!visibleMessages.length" class="message-load-state">
        星空安静，等待你留下第一束光。
      </p>
    </div>
    <div class="universe-bottom">
      <small>© {{ new Date().getFullYear() }} 云间小记</small>
    </div>
  </section>
</template>
<style lang="scss" scoped>
.message-universe {
  position: relative;
  min-height: 100svh;
  overflow: hidden;
  background: radial-gradient(
    ellipse at 50% 75%,
    #152b4b 0,
    #0b1426 38%,
    #030611 85%
  );
  color: white;
  isolation: isolate;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  place-items: center;
}
.star-field {
  position: absolute;
  inset: 0;
  z-index: -1;
  opacity: 0.8;
  background-image:
    radial-gradient(1px 1px at 20px 30px, #fff 90%, transparent),
    radial-gradient(1px 1px at 90px 100px, #c5d2ff 90%, transparent),
    radial-gradient(1.5px 1.5px at 150px 70px, #fff9 90%, transparent),
    radial-gradient(1px 1px at 50px 180px, #fff8 90%, transparent);
  background-size:
    210px 230px,
    300px 270px,
    390px 340px,
    170px 200px;
  transform: rotate(-7deg) scale(1.2);
}
.earth {
  position: absolute;
  z-index: -1;
  left: 50%;
  bottom: -66%;
  transform: translateX(-50%) rotate(-17deg);
  width: min(1100px, 130vw);
  aspect-ratio: 1;
  border-radius: 50%;
  overflow: hidden;
  background: radial-gradient(
    circle at 38% 19%,
    #85cbd4 0,
    #2474a0 17%,
    #123753 43%,
    #041326 69%
  );
  box-shadow:
    0 -7px 8px #82c5f67d,
    0 -13px 45px #4ba2e844,
    0 0 130px #396db933;
  opacity: 0.8;
}
.message-center {
  position: relative;
  z-index: 4;
  text-align: center;
  width: min(520px, calc(100% - 40px));
  min-width: 0;
  padding: 110px 0 130px;
  pointer-events: auto;
  .universe-eyebrow {
    font-size: 8px;
    letter-spacing: 4px;
    color: #aab7d1;
  }
  h1 {
    font-size: clamp(28px, 4vw, 42px);
    font-weight: 400;
    letter-spacing: 5px;
    margin: 20px 0;
  }
  .universe-subtitle {
    font-size: 12px;
    color: #9da9c0;
    letter-spacing: 2px;
    margin-bottom: 38px;
  }
}
.message-form {
  padding: 17px 22px 12px;
  isolation: isolate;
  position: relative;
  --glass-tint: rgb(14 25 46 / 72%);
  text-align: left;
  input,
  textarea {
    background: none;
    border: 0;
    outline: 0;
    color: #e3eafa;
    font-size: 12px;
    &::placeholder {
      color: #8090ac;
    }
  }
  .nickname-field {
    display: flex;
    align-items: center;
    gap: 10px;
    color: #9dacc7;
    padding: 2px 0 14px;
    border-bottom: 1px solid #bdc8e41a;
    input {
      flex: 1;
      min-width: 0;
    }
  }
  .message-field {
    display: flex;
    gap: 15px;
    align-items: center;
    padding-top: 14px;
    textarea {
      flex: 1;
      resize: vertical;
      min-height: 75px;
      width: 100%;
      line-height: 2;
    }
    button {
      width: 37px;
      height: 37px;
      border-radius: 50%;
      background: #8197bf;
      color: #fff;
      border: 0;
      flex-shrink: 0;
      svg {
        transform: rotate(-90deg);
      }
    }
  }
  &__footer {
    display: flex;
    justify-content: space-between;
    font-size: 8px;
    color: #8090ac;
    margin-top: 10px;
  }
}
.message-feedback {
  font-size: 11px;
  line-height: 2;
  color: #efadbd;
  &--success {
    color: #b4d7c8;
  }
}
.message-load-state {
  font-size: 10px;
  color: #7b89a5;
  margin: 24px 0 0;
  button {
    background: none;
    border: 0;
    color: #b7c6e1;
    margin-left: 8px;
  }
}
.message-bubbles {
  position: absolute;
  inset: 65px 0 80px;
  z-index: 2;
  pointer-events: none;
}
.message-bubble {
  position: absolute;
  left: var(--x);
  top: var(--y);
  max-width: 200px;
  min-width: 110px;
  padding: 12px 17px;
  background: linear-gradient(135deg, #8794bd18, #b4a3d00e);
  border: 1px solid #acc0e51c;
  border-radius: 8px;
  backdrop-filter: blur(5px);
  animation: drift var(--duration) ease-in-out var(--delay) infinite;
  transform: translate(-50%, -50%);
  color: #c4cce0;
  span {
    font-size: 8px;
    color: #8e9bb5;
  }
  p {
    margin: 5px 0 0;
    font-size: 10px;
    line-height: 1.8;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}
.universe-bottom {
  position: absolute;
  bottom: 27px;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  flex-direction: column;
  gap: 12px;
  font-size: 9px;
  letter-spacing: 2px;
  color: #77859f;
  small {
    font-size: 8px;
    letter-spacing: 1px;
  }
}
@keyframes drift {
  50% {
    transform: translate(-50%, calc(-50% - 18px));
  }
}
@media (max-width: 700px) {
  .earth {
    bottom: -20%;
    width: 180vw;
    opacity: 0.6;
  }
  .message-bubbles {
    opacity: 0.45;
  }
  .message-bubble {
    max-width: 130px;
    min-width: 80px;
    padding: 9px 12px;
    &:nth-child(n + 9) {
      display: none;
    }
  }
  .message-center {
    padding: 130px 0 165px;
  }
  .message-center h1 {
    letter-spacing: 2px;
    font-size: 29px;
  }
  .message-form {
    --glass-tint: rgb(14 25 46 / 85%);
  }
  .universe-bottom {
    letter-spacing: 0.5px;
    font-size: 8px;
  }
}
</style>
