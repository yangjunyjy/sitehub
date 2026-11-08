import request from "@/utils/request";
import { uint64, allPages } from "../../../shared/native-api.mjs";

const API_PREFIX = "";

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

export const getArticles = (params = {}, options = {}) =>
  request({
    url: `${API_PREFIX}/blog/articles`,
    method: "get",
    params: { ...params, categoryID: params.categoryId, tagIDs: params.tagIds ? String(params.tagIds).split(",") : undefined },
    signal: options.signal,
  });

export const getArticleDetail = (id, options = {}) =>
  request({
    url: `${API_PREFIX}/blog/articles/detail`,
    params: { id },
    method: "get",
    signal: options.signal,
  });

export const getCategories = (params = {}) =>
  request({
    url: `${API_PREFIX}/blog/categories`,
    method: "get",
    params: { page: 1, pageSize: 100, ...params },
  });

export const getTags = () =>
  allPages(params => request.get("/blog/tags", { params }));

export const getLatestArticles = (options = {}) =>
  request({
    url: `${API_PREFIX}/blog/articles/recent`,
    method: "get",
    signal: options.signal,
  });

export const getPopularArticles = (options = {}) =>
  request({
    url: `${API_PREFIX}/blog/articles/hot`,
    method: "get",
    params: { limit: 5 },
    signal: options.signal,
  });

export const incrementArticleView = (id) =>
  request({
    url: `${API_PREFIX}/blog/articles/view`,
    method: "post",
    data: { id: uint64(id) },
  });
