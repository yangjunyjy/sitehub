import request from "@/utils/request";
import { uint64 } from "@/utils/native-api.mjs";

export const getReviewQueue = (params, config) =>
  request.get("/blog/admin/community", { ...config, params });
export const reviewItem = (kind, id, status) =>
  request.post("/blog/admin/community", { kind, id: uint64(id), status });
