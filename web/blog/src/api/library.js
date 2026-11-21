import request from "@/utils/request";
import { uint64 } from "@/utils/native-api.mjs";

export const getInteraction = (articleId) =>
  request
    .get("/blog/articles/interaction", { params: { articleId } })
    .then((r) => r.data);
export const setLike = (articleId, liked) =>
  request
    .post("/blog/articles/like", { articleId: uint64(articleId), liked })
    .then((r) => r.data);
export const setFavorite = (articleId, favorited, folderId = "0") =>
  request
    .post("/blog/articles/favorite", {
      articleId: uint64(articleId),
      favorited,
      folderId: uint64(folderId),
    })
    .then((r) => r.data);
export const getFolders = () =>
  request.get("/blog/favorites/folders").then((r) => r.data.items);
export const createFolder = (name) =>
  request.post("/blog/favorites/folders", { name }).then((r) => r.data);
export const getFavorites = (folderId, page, pageSize) =>
  request
    .get("/blog/favorites/articles", { params: { folderId, page, pageSize } })
    .then((r) => r.data);
