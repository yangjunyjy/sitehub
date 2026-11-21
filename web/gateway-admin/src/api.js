export async function api(path, body, options = {}) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (options.totp) headers['X-TOTP-Token'] = options.totp
  const response = await fetch('./api/' + path, {
    method: body === undefined ? 'GET' : 'POST',
    credentials: 'same-origin',
    headers,
    body: JSON.stringify(body),
    signal: options.signal,
  })
  if (response.status === 204) return
  const data = await response.json()
  if (!response.ok) {
    if (response.status === 401 && !path.startsWith('auth/'))
      window.dispatchEvent(new Event('session-expired'))
    throw Object.assign(new Error(data.message || '请求失败，请重试'), { status: response.status })
  }
  return data
}
export const generate = async ({ signal }) => ({
  data: await api('auth/captcha', undefined, { signal }),
})
export const verifyChallenge = async (captchaKey, sliderX, behavior, { signal }) => ({
  data: await api('auth/captcha/verify', { captchaKey, sliderX, ...behavior }, { signal }),
})
