export function externalAddress(value) {
  if (typeof value !== 'string' || !/^https?:\/\//i.test(value) || /[\s\\]/.test(value)) return '';
  try {
    const address = new URL(value);
    return address.hostname && !address.username && !address.password ? address.href : '';
  } catch { return ''; }
}
