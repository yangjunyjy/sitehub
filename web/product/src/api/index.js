import request from "@/utils/request";


const API_PREFIX = "";

export const getPublicMenus = () => request.get('/rbac/menus/public', { params: { system: 'product' } });

export const getWebsiteContent = (options = {}) =>
  request({
    url: `${API_PREFIX}/product/content`,
    method: "get",
    signal: options.signal,
  });

export const getDownloads = () =>
  request({
    url: `${API_PREFIX}/product/downloads`,
    method: "get",
  });

export const getDocumentModes = async (options = {}) => {
  const response = await request({
    url: API_PREFIX + "/product/docs/modes",
    method: "get",
    signal: options.signal,
  });
  if (!Array.isArray(response.data?.modes) || !response.data.defaultMode)
    throw new Error("文档模式接口格式不匹配，请联系管理员完成迁移");
  return response;
};
export const getDocsNavigation = async (options = {}) => {
  const response = await request({
    url: `${API_PREFIX}/product/docs/navigation`,
    method: "get",
    params: { version: options.version, mode: options.mode },
    signal: options.signal,
  });
  if (
    !response.data?.versionId ||
    !Array.isArray(response.data.versions) ||
    !Array.isArray(response.data.groups)
  )
    throw new Error("文档接口格式不匹配，请联系管理员完成迁移");
  return response;
};

export const getDocument = (id, version, mode, options = {}) =>
  request({
    url: `${API_PREFIX}/product/docs/document`,
    method: "get",
    params: { id, version, mode },
    signal: options.signal,
  });

