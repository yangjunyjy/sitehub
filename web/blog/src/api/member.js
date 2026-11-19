import request from "@/utils/request";
export const loginMember = async (values) => {
  const { data } = await request.post("/rbac/members/login", values);
  if (!data?.user) throw new Error("登录响应无效，请重试");
  return data.user;
};
export const memberProfile = () =>
  request.get("/rbac/members/me").then((response) => response.data);
export const logoutMember = async () => {
  await request.post("/rbac/members/logout", {});
};
