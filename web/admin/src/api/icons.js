import request from "@/utils/request";
import { uint64 } from "@/utils/native-api.mjs";
const path = "/rbac/system/icons";

export const listIcons = () =>
  request.get(path).then(({ data }) => data.items || []);
export const validateIcon = (input) => request.post(path + "/validate", input);
export const saveIcon = (input, id) =>
  id
    ? request.post(path + "/update", { ...input, id: uint64(id) })
    : request.post(path + "/create", input);
export const removeIcon = (id) =>
  request.post(path + "/delete", { id: uint64(id) });
export const readIcon = async (name) => {
  const blob = await request.get(path + "/content", {
    params: { name },
    responseType: "blob",
  });
  return blob.text();
};
