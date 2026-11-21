import { reactive, readonly } from "vue";
import { loginMember, memberProfile, logoutMember } from "@/api/member";
const state = reactive({ user: null, ready: false });
let pending;
let version = 0;
function assign(user) {
  version++;
  state.user = user;
  state.ready = true;
}

export function useMember() {
  return {
    member: readonly(state),
    async restore(force = false) {
      if (state.ready && !force) return;
      if (!pending) {
        const requestVersion = version;
        pending = memberProfile()
          .then((user) => {
            if (requestVersion === version) assign(user);
          })
          .catch((error) => {
            if (requestVersion !== version) return;
            if (error.status === 401) assign(null);
            else throw error;
          })
          .finally(() => {
            pending = null;
          });
      }
      return pending;
    },
    expired() {
      assign(null);
    },
    async login(values) {
      assign(await loginMember(values));
    },
    async logout() {
      await logoutMember();
      assign(null);
    },
  };
}
export function memberDestination(value, router) {
  const account = { name: "account" };
  if (typeof value !== "string" || !/^\/(?!\/)/.test(value) || /[\\\s]/.test(value)) return account;
  const destination = router.resolve(value);
  return destination.meta.notFound || ["login","register"].includes(destination.meta.page) ? account : destination.fullPath;
}
