import request from "@/utils/request";
import { uint64 } from "../utils/native-api.mjs";
export const getSystems = (config = {}) => request.get("/rbac/systems", config);
export const saveSystem = (data) =>
  request.post("/rbac/systems/save", {
    ...data,
    ...(data.id ? { id: uint64(data.id) } : {}),
  });
export const deleteSystem = (id) =>
  request.post("/rbac/systems/delete", { id: uint64(id) });
