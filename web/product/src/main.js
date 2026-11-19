import { createApp } from 'vue';
import App from './App.vue';
import { createProductRouter } from './router';
import { getPublicMenus } from './api';
import { publicNavigation } from './router/pages';
import './styles/global.scss';

async function start() {
  let menus = null;
  try {
    const { data } = await getPublicMenus();
    menus = publicNavigation(data.items || []);
  } catch {
    // Keep the product pages available while the menu service is offline.
  }
  createApp(App)
    .provide('productMenus', menus)
    .use(createProductRouter(menus || []))
    .mount('#app');
}
start();
