import request from '@/utils/request';

export async function readCatalog({ mode, version, signal } = {}) {
  const { data } = await request.get('/product/manuals/catalog', { params: { mode, version }, signal });
  if (!Array.isArray(data?.modes) || (data.navigation && !Array.isArray(data.navigation.groups)))
    throw new Error('文档目录格式不正确');
  return data;
}

export async function readDocument({ id, version, mode, signal }) {
  const { data } = await request.get('/product/manuals/document', { params: { id, version, mode }, signal });
  if (typeof data?.content !== 'string') throw new Error('文档正文格式不正确');
  return data;
}
