import request from '@/utils/request';

export const getPublicMenus = () => request.get('/rbac/menus/public', { params: { system: 'product' } });
export const getWebsiteContent = ({ signal } = {}) => request.get('/product/content', { signal });
