export async function api(path, body) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  const response = await fetch('./api/' + path, {
    method: body === undefined ? 'GET' : 'POST',
    credentials: 'same-origin',
    headers,
    body: JSON.stringify(body),
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
