import axios from "axios";
import { queryString, stringifyJSON, viewValues } from "./native-api.mjs";

const request = axios.create({
  baseURL: "/v1",
  timeout: 30000,
  paramsSerializer: { serialize: queryString },
});
request.interceptors.request.use((config) => {
  if (
    config.data &&
    typeof config.data === "object" &&
    !(config.data instanceof FormData)
  ) {
    config.data = stringifyJSON(config.data);
    config.headers["Content-Type"] = "application/json";
  }
  return config;
});
request.interceptors.response.use(
  ({ data }) => {
    if (data.code !== 0)
      throw new Error(data.message || "请求未能完成，请稍后重试");
    return { ...data, data: viewValues(data.data) };
  },
  (error) => {
    if (axios.isCancel(error)) return Promise.reject(error);
    error.status = error.response?.status;
    error.message =
      error.status >= 500
        ? "服务暂时不可用，请稍后重试"
        : error.response?.data?.message || "连接未能完成，请检查网络后重试";
    return Promise.reject(error);
  },
);
export default request;
