import type { CustomRoute } from '@elegant-router/types';
import { layouts, views } from '../elegant/imports';
import { getRoutePath, transformElegantRoutesToVueRoutes } from '../elegant/transform';

const homePath = getRoutePath(import.meta.env.VITE_ROUTE_HOME) || '/requirement/list';

export const ROOT_ROUTE: CustomRoute = {
  name: 'root',
  path: '/',
  redirect: homePath,
  meta: {
    title: 'root',
    constant: true
  }
};

const HOME_REDIRECT_ROUTE: CustomRoute = {
  name: 'home-redirect',
  path: '/home',
  redirect: homePath,
  meta: {
    title: 'home',
    constant: true,
    hideInMenu: true
  }
};

const NOT_FOUND_ROUTE: CustomRoute = {
  name: 'not-found',
  path: '/:pathMatch(.*)*',
  component: 'layout.blank$view.404',
  meta: {
    title: 'not-found',
    constant: true
  }
};

const builtinRoutes: CustomRoute[] = [ROOT_ROUTE, HOME_REDIRECT_ROUTE, NOT_FOUND_ROUTE];

/** create builtin vue routes */
export function createBuiltinVueRoutes() {
  return transformElegantRoutesToVueRoutes(builtinRoutes, layouts, views);
}
