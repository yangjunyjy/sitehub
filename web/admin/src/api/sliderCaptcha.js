import request from '@/utils/request'

/**
 * 生成滑块拼图验证码
 * @returns {Promise} { captchaKey, backgroundImage, sliderImage, y }
 */
export function generateSliderCaptcha({ signal } = {}) {
  return request({
    url: '/rbac/captcha',
    method: 'get',
    signal,
    silentError: true,
  })
}

/**
 * 校验滑块位置
 * @param {string} captchaKey - 验证码key
 * @param {number} sliderX - 滑块X坐标
 * @param {object} behavior - 行为数据
 * @param {number} behavior.duration - 拖动时长(ms)
 * @param {number} behavior.moveCount - 移动轨迹点数
 * @returns {Promise}
 */
export function verifySliderCaptcha(captchaKey, sliderX, behavior, { signal } = {}) {
  return request({
    url: '/rbac/captcha/verify',
    method: 'post',
    data: { captchaKey, sliderX, ...behavior },
    signal,
    silentError: true,
  })
}
