<script setup>
import SliderCaptcha from '@shared/auth/SliderCaptcha.vue'
import { generateSliderCaptcha, verifySliderCaptcha } from '@/api/sliderCaptcha'
</script>
<template>
  <SliderCaptcha :generate="generateSliderCaptcha" :verify-challenge="verifySliderCaptcha" />
</template>
