import { reactive, readonly } from "vue";
import { loginMember, memberProfile, logoutMember } from "@/api/member";
const state = reactive({ user: null, ready: false });
let pending;
export function useMember() {
  return {
    member: readonly(state),
    async restore(force = false) {
      if (state.ready && !force) return;
      if (!pending)
        pending = memberProfile()
          .then((user) => {
            state.user = user;
            state.ready = true;
          })
          .catch((error) => {
            if (error.status === 401) { state.user = null; state.ready = true; }
            else throw error;
          })
          .finally(() => {
            pending = null;
          });
      return pending;
    },
    async login(values) {
      state.user = await loginMember(values);
      state.ready = true;
    },
    async logout() {
      await logoutMember();
      state.user = null;
      state.ready = true;
    },
  };
}
export function memberDestination(value) {
  return typeof value === "string" &&
    /^\/(?!\/)/.test(value) &&
    !/[\\\s]/.test(value)
    ? value
    : "/account";
}
