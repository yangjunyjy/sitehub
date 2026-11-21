import { reactive, readonly } from "vue";
import { getWebsiteContent } from "@/api/site";
import {
  createContentResource,
  emptyWebsiteContent,
  normalizeWebsiteContent,
} from "@/utils/content-resource";

const state = reactive({
  loading: false,
  loaded: false,
  error: "",
  content: emptyWebsiteContent(),
});
export const loadWebsiteContent = createContentResource(
  state,
  getWebsiteContent,
  normalizeWebsiteContent,
);
export const useWebsiteContent = () => ({
  website: readonly(state),
  loadWebsiteContent,
});
