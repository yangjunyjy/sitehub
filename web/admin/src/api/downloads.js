import request from "@/utils/request";
import { uint64 } from "@/utils/native-api.mjs";

const base = "/product/admin/releases";
const write = (path, values) =>
  request
    .post(base + path, { ...values, revision: uint64(values.revision || 0) })
    .then((r) => r.data);
const mutation = (path, { id, revision }) => write(path, { id, revision });

export const getReleaseFiles = () => request.get(base + "/files");
export const listReleaseGroups = () =>
  request.get(base + "/groups").then((r) => r.data.items);
export const getReleaseSettings = () =>
  request.get(base + "/settings").then((r) => r.data);
export const saveReleaseSettings = (value) => write("/settings/update", value);
export const createReleaseGroup = (value) => write("/groups/create", value);
export const updateReleaseGroup = (value) => write("/groups/update", value);
export const deleteReleaseGroup = (row) => mutation("/groups/delete", row);
export const createReleaseFile = (value) => write("/files/create", value);
export const updateReleaseFile = (value) => write("/files/update", value);
export const publishReleaseFile = (row) => mutation("/files/publish", row);
export const withdrawReleaseFile = (row) => mutation("/files/withdraw", row);
export const deleteReleaseFile = (row) => mutation("/files/delete", row);

export const uploadDownloadPackage = (file, onProgress) => {
  const data = new FormData();
  data.append("file", file);
  data.append("kind", "release");
  return request.post("/file/upload", data, {
    timeout: 600000,
    onUploadProgress: (event) =>
      onProgress?.(
        event.total ? Math.round((event.loaded / event.total) * 100) : 0,
      ),
  });
};
