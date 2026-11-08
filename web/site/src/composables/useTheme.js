import { ref } from "vue";

// The document is initialized before paint in index.html; all controls share it.
const dark = ref(document.documentElement.classList.contains("dark"));
export function useTheme() {
  function toggleTheme() {
    dark.value = !dark.value;
    document.documentElement.classList.toggle("dark", dark.value);
    document.documentElement.style.colorScheme = dark.value ? "dark" : "light";
    try {
      localStorage.setItem("deploy-site-theme", dark.value ? "dark" : "light");
    } catch {}
  }
  return { dark, toggleTheme };
}
