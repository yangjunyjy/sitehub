import request from "@/utils/request";
import { uint64, allPages } from "@/utils/native-api.mjs";

export const publicMenus = () =>
  request.get("/rbac/menus/public", { params: { system: "blog" } });

export const articles = (params = {}, options = {}) =>
  request.get("/blog/articles", { params, signal: options.signal });
export const searchArticles = (params, options = {}) =>
  request.get("/blog/articles/search", { params, signal: options.signal });
export const articleDetail = (id, options = {}) =>
  request.get("/blog/articles/detail", {
    params: { id },
    signal: options.signal,
  });
export const recordView = (id) =>
  request.post("/blog/articles/view", { id: uint64(id) });
export const categories = () =>
  allPages((params) => request.get("/blog/categories", { params }));
export const tags = () =>
  allPages((params) => request.get("/blog/tags", { params }));
export const friends = () => request.get("/blog/friends");
export const applyFriend = (data) => request.post("/blog/friends", data);
export const messages = (params = {}) =>
  request.get("/blog/messages", { params });
export const sendMessage = (data) => request.post("/blog/messages", data);
export const about = () => request.get("/blog/about");
