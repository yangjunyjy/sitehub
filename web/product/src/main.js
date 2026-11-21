import { createApp } from 'vue';
import App from './App.vue';
import { createProductRouter } from './router';
import { getPublicMenus } from './api/site';
import { loadPublicSite } from '@shared/navigation/public-site.mjs';
import './styles/global.scss';

async function start() {
  const pages = import.meta.glob('./views/**/index.vue');
  const settings = import.meta.glob('./views/**/menu.json', { eager: true, import: 'default' });
  const site = await loadPublicSite(getPublicMenus, pages, settings);
  createApp(App).provide('navigation', site).use(createProductRouter(site.routes)).mount('#app');
}
start();
