<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Check, Close, DArrowRight, Lock, RefreshRight, Warning } from '@element-plus/icons-vue'
import { generateSliderCaptcha, verifySliderCaptcha } from '@/api/sliderCaptcha'

const props = defineProps({ modelValue: { type: Boolean, default: false } })
const emit = defineEmits(['update:modelValue', 'success', 'fail'])
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const status = ref('loading')
const message = ref('')
const challenge = ref(null)
const progress = ref(0)
const imageArea = ref(null)
const track = ref(null)
const thumb = ref(null)
const imageWidth = ref(0)
const trackTravel = ref(0)
let generation = 0
let controller
let successTimer
let observer
let pointer
let startedAt = 0
let moves = 0

const canMove = computed(() => status.value === 'ready' || status.value === 'dragging')
const canRefresh = computed(() => !['loading', 'verifying', 'success'].includes(status.value))
const scale = computed(() => imageWidth.value / (challenge.value?.imgWidth || 300))
const pieceSize = computed(() => challenge.value?.pieceSize || 44)
const piecePadding = computed(() => challenge.value?.piecePadding ?? 4)
const imageTravel = computed(() => (challenge.value?.imgWidth || 300) - pieceSize.value)
const pieceStyle = computed(() => ({
  width: `${(pieceSize.value + piecePadding.value * 2) * scale.value}px`,
  height: `${(pieceSize.value + piecePadding.value * 2) * scale.value}px`,
  top: `${((challenge.value?.y || 0) - piecePadding.value) * scale.value}px`,
  transform: `translate3d(${(progress.value * imageTravel.value - piecePadding.value) * scale.value}px, 0, 0)`,
}))
const trackText = computed(
  () =>
    ({
      loading: '正在准备验证',
      ready: '按住滑块，拖动到缺口',
      dragging: '对齐缺口后松开',
      verifying: '正在验证…',
      success: '验证通过',
      mismatch: '请换一张图片重试',
      error: '请重新加载验证',
    })[status.value],
)
const statusText = computed(
  () =>
    message.value ||
    {
      loading: '正在加载拼图，请稍候',
      ready: '将拼图移到图中的缺口处',
      dragging: '慢慢移动，对齐后松开即可',
      verifying: '正在确认拼图位置，请稍候',
      success: '验证通过，即将发送邮箱验证码',
      mismatch: '还差一点，换一张图片再试试',
      error: '验证暂时不可用，请重新加载',
    }[status.value],
)

function stopPointer() {
  const active = pointer
  pointer = null
  if (active?.element.hasPointerCapture(active.id)) active.element.releasePointerCapture(active.id)
}

function invalidate() {
  generation++
  controller?.abort()
  clearTimeout(successTimer)
  stopPointer()
  startedAt = 0
  moves = 0
}

function measure() {
  const width = imageArea.value?.clientWidth || 0
  if (imageWidth.value && width !== imageWidth.value && status.value === 'dragging') cancelDrag()
  imageWidth.value = width
  trackTravel.value = Math.max(
    0,
    (track.value?.clientWidth || 0) - (thumb.value?.offsetWidth || 48),
  )
}

watch(
  [imageArea, track],
  () => {
    observer?.disconnect()
    observer = new ResizeObserver(measure)
    if (imageArea.value) observer.observe(imageArea.value)
    if (track.value) observer.observe(track.value)
    measure()
  },
  { flush: 'post' },
)

async function decodeImage(src) {
  if (typeof src !== 'string' || !src.startsWith('data:image/')) throw new Error('invalid image')
  const image = new Image()
  image.src = src
  await image.decode()
  return image
}

async function loadCaptcha() {
  invalidate()
  const current = generation
  controller = new AbortController()
  status.value = 'loading'
  message.value = ''
  challenge.value = null
  progress.value = 0
  try {
    const response = await generateSliderCaptcha({ signal: controller.signal })
    const data = response.data
    if (!data?.captchaKey || !Number.isFinite(data.y)) throw new Error('invalid challenge')
    const [background, piece] = await Promise.all([
      decodeImage(data.backgroundImage),
      decodeImage(data.sliderImage),
    ])
    if (current !== generation || !visible.value) return
    const size = data.pieceSize || 44
    const padding = data.piecePadding ?? 4
    if (
      background.naturalWidth <= size ||
      data.y < 0 ||
      data.y + size > background.naturalHeight ||
      piece.naturalWidth !== size + padding * 2
    )
      throw new Error('invalid image dimensions')
    challenge.value = {
      ...data,
      imgWidth: background.naturalWidth,
      imgHeight: background.naturalHeight,
    }
    status.value = 'ready'
    await nextTick()
    measure()
    thumb.value?.focus({ preventScroll: true })
  } catch (error) {
    if (current !== generation || !visible.value) return
    status.value = 'error'
    message.value = '验证码加载失败，请重新加载后重试'
  }
}

function beginDrag(event, source = 'thumb') {
  if (!canMove.value || pointer || !event.isPrimary || event.button !== 0 || !trackTravel.value)
    return
  event.preventDefault()
  measure()
  event.currentTarget.setPointerCapture(event.pointerId)
  pointer = {
    id: event.pointerId,
    element: event.currentTarget,
    startX: event.clientX,
    startProgress: progress.value,
    travel: source === 'piece' ? imageTravel.value * scale.value : trackTravel.value,
  }
  status.value = 'dragging'
  message.value = ''
  startedAt = performance.now()
  moves = 0
  thumb.value?.focus({ preventScroll: true })
}

function moveDrag(event) {
  if (!pointer || event.pointerId !== pointer.id) return
  event.preventDefault()
  const value = Math.max(
    0,
    Math.min(1, pointer.startProgress + (event.clientX - pointer.startX) / pointer.travel),
  )
  if (value !== progress.value) moves++
  progress.value = value
}

function endDrag(event) {
  if (!pointer || event.pointerId !== pointer.id) return
  moveDrag(event)
  stopPointer()
  verify()
}

function cancelDrag() {
  stopPointer()
  if (status.value !== 'dragging') return
  progress.value = 0
  startedAt = 0
  moves = 0
  status.value = 'ready'
}

function onLostPointer(event) {
  if (pointer?.id === event.pointerId) cancelDrag()
}

function onKeydown(event) {
  if (!canMove.value || pointer) return
  if (['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) {
    event.preventDefault()
    if (!startedAt) startedAt = performance.now()
    status.value = 'dragging'
    message.value = ''
    const step = (event.shiftKey ? 5 : 1) / imageTravel.value
    const value =
      event.key === 'Home'
        ? 0
        : event.key === 'End'
          ? 1
          : progress.value + (event.key === 'ArrowRight' ? step : -step)
    const bounded = Math.max(0, Math.min(1, value))
    if (bounded !== progress.value) moves++
    progress.value = bounded
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    verify()
  }
}

async function verify() {
  if (status.value !== 'dragging' || !challenge.value) return
  const duration = Math.round(performance.now() - startedAt)
  // 未移动、过快或超时都可直接重试，不消耗当前图片。
  if (!progress.value || moves < 3 || duration < 300 || duration > 10000) {
    cancelDrag()
    message.value = duration > 10000 ? '操作时间较长，请重新拖动' : '请稍慢一些拖动，对齐缺口后松开'
    return
  }
  status.value = 'verifying'
  message.value = ''
  const current = generation
  controller = new AbortController()
  try {
    const response = await verifySliderCaptcha(
      challenge.value.captchaKey,
      Math.round(progress.value * imageTravel.value),
      { duration, moveCount: moves },
      { signal: controller.signal },
    )
    if (current !== generation || !visible.value) return
    if (!response.data?.value) {
      status.value = 'mismatch'
      emit('fail')
      return
    }
    status.value = 'success'
    successTimer = setTimeout(() => {
      if (current !== generation || !visible.value) return
      emit('success', challenge.value.captchaKey)
      visible.value = false
    }, 650)
  } catch (error) {
    if (current !== generation || !visible.value) return
    status.value = 'error'
    message.value = '验证请求未完成，请换一张图片重试'
    emit('fail')
  }
}

watch(
  visible,
  (value) => {
    if (value) loadCaptcha()
    else invalidate()
  },
  { immediate: true },
)
onBeforeUnmount(() => {
  invalidate()
  observer?.disconnect()
})
</script>

<template>
  <el-dialog
    v-model="visible"
    class="captcha-dialog"
    width="400px"
    align-center
    append-to-body
    destroy-on-close
    :show-close="false"
    :close-on-click-modal="false"
    title="安全验证"
    @opened="measure"
  >
    <template #header="{ titleId }">
      <div class="captcha-heading">
        <span class="captcha-shield"
          ><el-icon :size="20"><Lock /></el-icon
        ></span>
        <div>
          <span class="captcha-eyebrow">SITEHUB · 内容管理</span>
          <h2 :id="titleId">完成安全验证</h2>
        </div>
        <button
          class="captcha-icon-button"
          type="button"
          aria-label="关闭安全验证"
          @click="visible = false"
        >
          <el-icon :size="18"><Close /></el-icon>
        </button>
      </div>
    </template>
    <div
      class="captcha-content"
      :data-state="status"
      :aria-busy="status === 'loading' || status === 'verifying'"
    >
      <p class="captcha-instruction">拖动下方滑块，让拼图与缺口重合。</p>
      <div
        ref="imageArea"
        class="captcha-image"
        :style="{ aspectRatio: `${challenge?.imgWidth || 300} / ${challenge?.imgHeight || 150}` }"
      >
        <template v-if="challenge">
          <img
            :src="challenge.backgroundImage"
            class="captcha-background"
            alt="将拼图移入图片中的缺口"
            draggable="false"
          />
          <img
            :src="challenge.sliderImage"
            class="captcha-piece"
            :style="pieceStyle"
            alt=""
            draggable="false"
            @pointerdown="beginDrag($event, 'piece')"
            @pointermove="moveDrag"
            @pointerup="endDrag"
            @pointercancel="cancelDrag"
            @lostpointercapture="onLostPointer"
          />
        </template>
        <div v-if="status === 'loading'" class="captcha-image-overlay captcha-skeleton">
          <span class="captcha-spinner"></span><span>正在准备拼图</span>
        </div>
        <div v-else-if="status === 'error' && !challenge" class="captcha-image-overlay">
          <el-icon :size="28"><Warning /></el-icon><span>验证暂时不可用</span
          ><button class="captcha-retry" type="button" @click="loadCaptcha">重新加载</button>
        </div>
        <div v-else-if="status === 'success'" class="captcha-success-badge">
          <el-icon><Check /></el-icon> 验证通过
        </div>
      </div>
      <div ref="track" class="captcha-track" :class="`is-${status}`">
        <div
          class="captcha-track-fill"
          :style="{ width: `${progress * trackTravel + 48}px` }"
        ></div>
        <span class="captcha-track-text" :style="{ opacity: progress > 0 ? 0 : 1 }">{{
          trackText
        }}</span>
        <button
          ref="thumb"
          type="button"
          role="slider"
          class="captcha-thumb"
          :style="{ transform: `translate3d(${progress * trackTravel}px, 0, 0)` }"
          :disabled="!canMove"
          aria-label="拖动拼图滑块"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-valuenow="Math.round(progress * 100)"
          :aria-valuetext="`当前位置 ${Math.round(progress * 100)}%`"
          aria-describedby="captcha-keyboard-help"
          @pointerdown="beginDrag"
          @pointermove="moveDrag"
          @pointerup="endDrag"
          @pointercancel="cancelDrag"
          @lostpointercapture="onLostPointer"
          @keydown="onKeydown"
        >
          <span
            v-if="status === 'verifying' || status === 'loading'"
            class="captcha-spinner"
          ></span>
          <el-icon v-else-if="status === 'success'" :size="22"><Check /></el-icon>
          <el-icon v-else-if="status === 'mismatch' || status === 'error'" :size="20"
            ><RefreshRight
          /></el-icon>
          <el-icon v-else :size="22"><DArrowRight /></el-icon>
        </button>
      </div>
      <div
        class="captcha-feedback"
        :class="{
          'is-error': ['error', 'mismatch'].includes(status),
          'is-success': status === 'success',
        }"
        role="status"
        aria-live="polite"
      >
        <el-icon v-if="status === 'success'"><Check /></el-icon
        ><el-icon v-else-if="['error', 'mismatch'].includes(status)"><Warning /></el-icon
        ><span>{{ statusText }}</span>
      </div>
      <div class="captcha-actions">
        <span>通过后发送邮箱验证码</span>
        <button type="button" class="captcha-refresh" :disabled="!canRefresh" @click="loadCaptcha">
          <el-icon><RefreshRight /></el-icon
          >{{ ['mismatch', 'error'].includes(status) ? '重新验证' : '换一张' }}
        </button>
      </div>
      <p id="captcha-keyboard-help" class="captcha-keyboard-help">
        也可用 ← → 微调位置，按 Enter 确认
      </p>
    </div>
  </el-dialog>
</template>

<style lang="scss">
.el-dialog.captcha-dialog {
  box-sizing: border-box;
  width: min(400px, calc(100vw - 32px));
  padding: 24px;
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 24px 80px #0000000d;
  color: #303133;
  --el-dialog-margin-top: 10vh;
}
.captcha-dialog .el-dialog__header {
  padding: 0;
  margin: 0;
}
.captcha-dialog .el-dialog__body {
  padding: 0;
  color: inherit;
}
.captcha-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}
.captcha-shield {
  width: 44px;
  height: 44px;
  border-radius: 13px;
  background: #f5f7fa;
  color: #409eff;
  display: grid;
  place-items: center;
}
.captcha-eyebrow {
  display: block;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 1.2px;
  color: #606266;
  line-height: 1.6;
}
.captcha-heading h2 {
  margin: 2px 0 0;
  font-size: 19px;
  font-weight: 650;
  line-height: 1.4;
}
.captcha-icon-button {
  margin-left: auto;
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 0;
  background: transparent;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
}
.captcha-icon-button:hover {
  background: #f5f7fa;
  color: #303133;
}
.captcha-instruction {
  margin: 22px 0 14px;
  font-size: 13px;
  line-height: 1.6;
  color: #606266;
}
.captcha-image {
  position: relative;
  overflow: hidden;
  width: 100%;
  border-radius: 4px;
  background: #f5f7fa;
  isolation: isolate;
  user-select: none;
}
.captcha-background {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}
.captcha-piece {
  position: absolute;
  left: 0;
  will-change: transform;
  filter: drop-shadow(0 2px 4px #409eff15);
  cursor: grab;
  touch-action: none;
  user-select: none;
}
.captcha-content[data-state='dragging'] .captcha-piece {
  cursor: grabbing;
}
.captcha-content[data-state='success'] .captcha-piece {
  filter: none;
}
.captcha-image-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  color: #606266;
  font-size: 13px;
}
.captcha-skeleton {
  background: #f5f7fa;
  background-size: 200% 100%;
  animation: captcha-shimmer 2s ease-in-out infinite;
}
.captcha-spinner {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: captcha-spin 0.8s linear infinite;
  box-sizing: border-box;
}
.captcha-success-badge {
  position: absolute;
  inset: auto 0 0;
  background: #148467ed;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  padding: 9px;
}
.captcha-track {
  height: 48px;
  position: relative;
  overflow: hidden;
  box-sizing: border-box;
  border-radius: 4px;
  background: #f5f7fa;
  box-shadow: inset 0 0 0 1px #909399;
  margin-top: 18px;
  user-select: none;
}
.captcha-track-fill {
  position: absolute;
  inset: 0 auto 0 0;
  background: #f5f7fa;
  pointer-events: none;
}
.captcha-track-text {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  padding-left: 34px;
  font-size: 12px;
  letter-spacing: 0.4px;
  color: #606266;
  pointer-events: none;
}
.captcha-thumb {
  position: absolute;
  left: 0;
  top: 0;
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  padding: 0;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: white;
  color: #409eff;
  box-shadow: 2px 0 6px #0000000d;
  cursor: grab;
  touch-action: none;
  user-select: none;
  will-change: transform;
}
.captcha-thumb:hover {
  background: #f8faff;
  border-color: #ebeef5;
}
.captcha-thumb:focus-visible {
  outline: 2px solid #409eff;
  outline-offset: -3px;
}
.captcha-thumb:disabled {
  cursor: default;
  color: #909399;
}
.is-dragging .captcha-thumb {
  cursor: grabbing;
  background: #3169db;
  color: white;
  border-color: #ebeef5;
}
.is-dragging .captcha-track-text {
  opacity: 0.5;
}
.is-success .captcha-track-fill {
  background: #e4f4ed;
}
.is-success .captcha-thumb {
  color: #148467;
  border-color: #9edbc5;
  background: #f5f7fa;
}
.is-success .captcha-track-text {
  color: #148467;
}
.is-mismatch .captcha-track-fill,
.is-error .captcha-track-fill {
  background: #fff0eb;
}
.is-mismatch .captcha-thumb,
.is-error .captcha-thumb {
  color: #c46443;
  border-color: #f2cfc0;
}
.captcha-feedback {
  min-height: 40px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  line-height: 1.5;
  color: #606266;
}
.captcha-feedback.is-error {
  color: #bc603b;
}
.captcha-feedback.is-success {
  color: #148467;
}
.captcha-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
  font-size: 12px;
  color: #606266;
}
.captcha-refresh,
.captcha-retry {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  border: 0;
  background: transparent;
  padding: 6px 0 6px 8px;
  font: inherit;
  font-weight: 500;
  color: #409eff;
  cursor: pointer;
}
.captcha-refresh:hover,
.captcha-retry:hover {
  color: #409eff;
}
.captcha-refresh:disabled {
  color: #909399;
  cursor: default;
}
.captcha-keyboard-help {
  margin: 10px 0 0;
  font-size: 11px;
  color: #909399;
  line-height: 1.5;
}
.captcha-content button:focus-visible,
.captcha-heading button:focus-visible {
  outline: 2px solid #409eff;
  outline-offset: 2px;
}
@keyframes captcha-spin {
  to {
    transform: rotate(360deg);
  }
}
@keyframes captcha-shimmer {
  to {
    background-position: -200% 0;
  }
}
@media (max-width: 420px) {
  .el-dialog.captcha-dialog {
    padding: 20px;
  }
  .captcha-heading h2 {
    font-size: 18px;
  }
  .captcha-instruction {
    margin-top: 18px;
  }
  .captcha-track-text {
    font-size: 11px;
  }
  .captcha-keyboard-help {
    display: none;
  }
}
@media (prefers-reduced-motion: reduce) {
  .captcha-skeleton,
  .captcha-spinner {
    animation: none;
  }
}
</style>
