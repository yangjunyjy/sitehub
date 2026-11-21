import request from '@/utils/request';

export async function readDownloads({ signal } = {}) {
  const { data } = await request.get('/product/releases/catalog', { signal });
  if (!Array.isArray(data?.packages)) throw new Error('发行目录格式不正确');
  return data;
}
