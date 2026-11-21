import { createApp } from 'vue';
import App from './App.vue';
import { createBlogRouter } from './router';
import { publicMenus } from './api';
import { loadPublicSite } from '@shared/navigation/public-site.mjs';
import './styles/global.scss';

async function start() {
  const pages = import.meta.glob('./views/**/index.vue');
  const settings = import.meta.glob('./views/**/menu.json', { eager: true, import: 'default' });
  const site = await loadPublicSite(publicMenus, pages, settings);
  createApp(App).provide('navigation', site).use(createBlogRouter(site.routes)).mount('#app');
}
start();
