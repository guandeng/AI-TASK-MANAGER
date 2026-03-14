/**
 * menu-storage.js
 * 菜单数据存储模块 - 基于文件存储
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 菜单数据文件路径
const MENU_FILE = path.join(__dirname, '../../data/menus.json');

// 默认菜单配置
const DEFAULT_MENUS = [
  {
    key: 'requirement',
    label: '需求任务管理',
    icon: 'carbon:task',
    order: 1,
    enabled: true,
    i18nKey: 'route.requirement',
    children: [
      {
        key: 'requirement_list',
        label: '需求列表',
        icon: 'carbon:list',
        path: '/requirement/list',
        routeName: 'requirement_list',
        order: 1,
        enabled: true,
        parentKey: 'requirement',
        i18nKey: 'route.requirement_list'
      },
      {
        key: 'requirement_task-list',
        label: '任务列表',
        icon: 'carbon:task-view',
        path: '/requirement/task-list',
        routeName: 'requirement_task-list',
        order: 2,
        enabled: true,
        parentKey: 'requirement',
        i18nKey: 'route.requirement_task-list'
      }
    ]
  },
  {
    key: 'manage',
    label: '系统管理',
    icon: 'carbon:settings',
    order: 2,
    enabled: true,
    i18nKey: 'route.manage',
    children: [
      {
        key: 'manage_config',
        label: '配置管理',
        icon: 'carbon:settings-adjust',
        path: '/manage/config',
        routeName: 'manage_config',
        order: 1,
        enabled: true,
        parentKey: 'manage',
        i18nKey: 'route.manage_config'
      },
      {
        key: 'manage_menu',
        label: '菜单管理',
        icon: 'carbon:menu',
        path: '/manage/menu',
        routeName: 'manage_menu',
        order: 2,
        enabled: true,
        parentKey: 'manage',
        i18nKey: 'route.manage_menu'
      }
    ]
  }
];

/**
 * 确保数据目录存在
 */
function ensureDataDir() {
  const dataDir = path.dirname(MENU_FILE);
  if (!fs.existsSync(dataDir)) {
    fs.mkdirSync(dataDir, { recursive: true });
  }
}

/**
 * 读取菜单数据
 */
export async function readMenuData() {
  try {
    ensureDataDir();

    if (!fs.existsSync(MENU_FILE)) {
      // 创建默认菜单文件
      await writeMenuData(DEFAULT_MENUS);
      return DEFAULT_MENUS;
    }

    const data = fs.readFileSync(MENU_FILE, 'utf-8');
    return JSON.parse(data);
  } catch (error) {
    console.error('[MenuStorage] Error reading menu data:', error.message);
    return DEFAULT_MENUS;
  }
}

/**
 * 写入菜单数据
 */
export async function writeMenuData(menus) {
  try {
    ensureDataDir();
    fs.writeFileSync(MENU_FILE, JSON.stringify(menus, null, 2), 'utf-8');
    return true;
  } catch (error) {
    console.error('[MenuStorage] Error writing menu data:', error.message);
    return false;
  }
}

/**
 * 获取菜单列表（扁平化）
 */
export async function getMenuList() {
  const menus = await readMenuData();
  return flattenMenus(menus);
}

/**
 * 获取菜单树
 */
export async function getMenuTree() {
  return readMenuData();
}

/**
 * 扁平化菜单
 */
function flattenMenus(menus, result = []) {
  for (const menu of menus) {
    const { children, ...menuWithoutChildren } = menu;
    result.push(menuWithoutChildren);
    if (children && children.length > 0) {
      flattenMenus(children, result);
    }
  }
  return result;
}

/**
 * 获取单个菜单
 */
export async function getMenuByKey(key) {
  const menus = await getMenuList();
  return menus.find(m => m.key === key);
}

/**
 * 创建菜单
 */
export async function createMenu(menuData) {
  const menus = await readMenuData();

  if (menuData.parentKey) {
    // 添加到父菜单的 children
    const parent = findMenuByKey(menus, menuData.parentKey);
    if (parent) {
      if (!parent.children) parent.children = [];
      parent.children.push({ ...menuData, children: [] });
    }
  } else {
    // 添加到顶级菜单
    menus.push({ ...menuData, children: [] });
  }

  await writeMenuData(menus);
  return menuData;
}

/**
 * 更新菜单
 */
export async function updateMenuByKey(key, updates) {
  const menus = await readMenuData();
  const menu = findMenuByKey(menus, key);

  if (!menu) return null;

  Object.assign(menu, updates, { updatedAt: new Date().toISOString() });
  await writeMenuData(menus);
  return menu;
}

/**
 * 删除菜单
 */
export async function deleteMenuByKey(key) {
  const menus = await readMenuData();
  const result = removeMenuByKey(menus, key);

  if (result.removed) {
    await writeMenuData(menus);
  }

  return result.removed;
}

/**
 * 批量删除菜单
 */
export async function batchDeleteMenus(keys) {
  const menus = await readMenuData();
  let removedCount = 0;

  for (const key of keys) {
    const result = removeMenuByKey(menus, key);
    if (result.removed) removedCount++;
  }

  if (removedCount > 0) {
    await writeMenuData(menus);
  }

  return { success: true, deletedKeys: keys.filter(k => removedCount > 0) };
}

/**
 * 查找菜单
 */
function findMenuByKey(menus, key) {
  for (const menu of menus) {
    if (menu.key === key) return menu;
    if (menu.children && menu.children.length > 0) {
      const found = findMenuByKey(menu.children, key);
      if (found) return found;
    }
  }
  return null;
}

/**
 * 删除菜单
 */
function removeMenuByKey(menus, key) {
  const index = menus.findIndex(m => m.key === key);
  if (index !== -1) {
    menus.splice(index, 1);
    return { removed: true };
  }

  for (const menu of menus) {
    if (menu.children && menu.children.length > 0) {
      const result = removeMenuByKey(menu.children, key);
      if (result.removed) return result;
    }
  }

  return { removed: false };
}

/**
 * 更新菜单排序
 */
export async function reorderMenus(orderData) {
  const menus = await readMenuData();
  const menuList = flattenMenus(menus);

  for (const { key, order } of orderData) {
    const menu = menuList.find(m => m.key === key);
    if (menu) menu.order = order;
  }

  // 重新排序
  sortMenus(menus);
  await writeMenuData(menus);
  return true;
}

/**
 * 排序菜单
 */
function sortMenus(menus) {
  menus.sort((a, b) => (a.order || 0) - (b.order || 0));
  for (const menu of menus) {
    if (menu.children && menu.children.length > 0) {
      sortMenus(menu.children);
    }
  }
}

/**
 * 移动菜单
 */
export async function moveMenu(key, targetParentKey) {
  const menus = await readMenuData();
  const menuData = extractMenu(menus, key);

  if (!menuData) return null;

  if (targetParentKey) {
    const targetParent = findMenuByKey(menus, targetParentKey);
    if (targetParent) {
      if (!targetParent.children) targetParent.children = [];
      menuData.parentKey = targetParentKey;
      targetParent.children.push(menuData);
    }
  } else {
    menuData.parentKey = undefined;
    menus.push(menuData);
  }

  await writeMenuData(menus);
  return menuData;
}

/**
 * 提取菜单（从树中移除并返回）
 */
function extractMenu(menus, key) {
  const index = menus.findIndex(m => m.key === key);
  if (index !== -1) {
    return menus.splice(index, 1)[0];
  }

  for (const menu of menus) {
    if (menu.children && menu.children.length > 0) {
      const extracted = extractMenu(menu.children, key);
      if (extracted) return extracted;
    }
  }

  return null;
}

/**
 * 切换菜单启用状态
 */
export async function toggleMenuEnabled(key, enabled) {
  return updateMenuByKey(key, { enabled });
}

export default {
  getMenuList,
  getMenuTree,
  getMenuByKey,
  createMenu,
  updateMenuByKey,
  deleteMenuByKey,
  batchDeleteMenus,
  reorderMenus,
  moveMenu,
  toggleMenuEnabled
};
