/**
 * 菜单管理模块
 * 用于管理前端左侧菜单配置
 */
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { log } from './utils.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 菜单数据存储路径
const MENU_FILE_PATH = path.join(__dirname, '../../menu.json');

// 默认菜单配置
const DEFAULT_MENUS = [
  {
    key: 'requirement',
    label: '需求管理',
    icon: 'mdi:file-document-outline',
    order: 1,
    enabled: true,
    hideInMenu: false,
    children: [
      {
        key: 'requirement_list',
        label: '需求列表',
        icon: 'mdi:format-list-bulleted',
        path: '/requirement/list',
        routeName: 'requirement_list',
        order: 1,
        enabled: true,
        hideInMenu: false,
        parentKey: 'requirement',
        i18nKey: 'route.requirement_list'
      },
      {
        key: 'requirement_task-list',
        label: '任务列表',
        icon: 'mdi:clipboard-list',
        path: '/requirement/task-list',
        routeName: 'requirement_task-list',
        order: 2,
        enabled: true,
        hideInMenu: false,
        parentKey: 'requirement',
        i18nKey: 'route.requirement_task-list'
      }
    ]
  },
  {
    key: 'manage',
    label: '系统管理',
    icon: 'mdi:cog',
    order: 2,
    enabled: true,
    hideInMenu: false,
    children: [
      {
        key: 'manage_config',
        label: '系统配置',
        icon: 'mdi:cog-outline',
        path: '/manage/config',
        routeName: 'manage_config',
        order: 1,
        enabled: true,
        hideInMenu: false,
        parentKey: 'manage',
        i18nKey: 'route.manage_config'
      },
      {
        key: 'manage_menu',
        label: '菜单管理',
        icon: 'mdi:menu',
        path: '/manage/menu',
        routeName: 'manage_menu',
        order: 2,
        enabled: true,
        hideInMenu: false,
        parentKey: 'manage',
        i18nKey: 'route.manage_menu'
      }
    ]
  },
  {
    key: 'team',
    label: '团队管理',
    icon: 'mdi:account-group',
    order: 3,
    enabled: true,
    hideInMenu: false,
    children: [
      {
        key: 'team_members',
        label: '成员管理',
        icon: 'mdi:account-multiple',
        path: '/team/members',
        routeName: 'team_members',
        order: 1,
        enabled: true,
        hideInMenu: false,
        parentKey: 'team',
        i18nKey: 'route.team_members'
      },
      {
        key: 'team_workload',
        label: '工作量统计',
        icon: 'mdi:chart-bar',
        path: '/team/workload',
        routeName: 'team_workload',
        order: 2,
        enabled: true,
        hideInMenu: false,
        parentKey: 'team',
        i18nKey: 'route.team_workload'
      }
    ]
  },
  {
    key: 'templates',
    label: '模板管理',
    icon: 'mdi:file-document-multiple',
    order: 4,
    enabled: true,
    hideInMenu: false,
    children: [
      {
        key: 'templates_projects',
        label: '项目模板',
        icon: 'mdi:folder-multiple',
        path: '/templates/projects',
        routeName: 'templates_projects',
        order: 1,
        enabled: true,
        hideInMenu: false,
        parentKey: 'templates',
        i18nKey: 'route.templates_projects'
      },
      {
        key: 'templates_tasks',
        label: '任务模板',
        icon: 'mdi:clipboard-text-multiple',
        path: '/templates/tasks',
        routeName: 'templates_tasks',
        order: 2,
        enabled: true,
        hideInMenu: false,
        parentKey: 'templates',
        i18nKey: 'route.templates_tasks'
      }
    ]
  }
];

/**
 * 读取菜单数据
 * @returns {Promise<Array>} 菜单列表
 */
export async function readMenuData() {
  try {
    if (!fs.existsSync(MENU_FILE_PATH)) {
      // 初始化菜单数据
      await writeMenuData(flattenMenus(DEFAULT_MENUS));
      return flattenMenus(DEFAULT_MENUS);
    }

    const content = await fs.promises.readFile(MENU_FILE_PATH, 'utf-8');
    const data = JSON.parse(content);

    // 兼容两种格式：扁平列表和嵌套树形
    if (Array.isArray(data)) {
      // 检查是否是扁平结构
      if (data.length > 0 && !data[0].children) {
        return data;
      }
      // 嵌套结构，展平
      return flattenMenus(data);
    }

    return [];
  } catch (error) {
    log('error', `Error reading menu data: ${error.message}`);
    return flattenMenus(DEFAULT_MENUS);
  }
}

/**
 * 写入菜单数据
 * @param {Array} menus 菜单列表（扁平结构）
 * @returns {Promise<boolean>} 是否成功
 */
export async function writeMenuData(menus) {
  try {
    await fs.promises.writeFile(MENU_FILE_PATH, JSON.stringify(menus, null, 2), 'utf-8');
    return true;
  } catch (error) {
    log('error', `Error writing menu data: ${error.message}`);
    return false;
  }
}

/**
 * 将嵌套菜单转换为扁平结构
 * @param {Array} nestedMenus 嵌套菜单
 * @param {string} parentKey 父级 key
 * @returns {Array} 扁平菜单列表
 */
function flattenMenus(nestedMenus, parentKey = null) {
  const result = [];

  for (const menu of nestedMenus) {
    const flatMenu = { ...menu, parentKey };
    delete flatMenu.children;

    result.push(flatMenu);

    if (menu.children && menu.children.length > 0) {
      result.push(...flattenMenus(menu.children, menu.key));
    }
  }

  return result;
}

/**
 * 将扁平菜单构建为树形结构
 * @param {Array} flatMenus 扁平菜单列表
 * @returns {Array} 树形菜单
 */
export function buildMenuTree(flatMenus) {
  const map = new Map();
  const roots = [];

  // 创建所有节点
  for (const menu of flatMenus) {
    map.set(menu.key, { ...menu, children: [] });
  }

  // 构建树形结构
  for (const menu of flatMenus) {
    const node = map.get(menu.key);
    if (menu.parentKey) {
      const parent = map.get(menu.parentKey);
      if (parent) {
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    } else {
      roots.push(node);
    }
  }

  // 排序
  const sortByOrder = (a, b) => (a.order || 0) - (b.order || 0);
  const sortChildren = nodes => {
    nodes.sort(sortByOrder);
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        sortChildren(node.children);
      }
    });
  };
  sortChildren(roots);

  // 清理空的 children 数组
  const cleanEmptyChildren = nodes => {
    nodes.forEach(node => {
      if (node.children && node.children.length === 0) {
        delete node.children;
      } else if (node.children) {
        cleanEmptyChildren(node.children);
      }
    });
  };
  cleanEmptyChildren(roots);

  return roots;
}

/**
 * 获取菜单列表（扁平结构）
 * @returns {Promise<Object>} { menus: [...] }
 */
export async function listMenus() {
  const menus = await readMenuData();
  return { menus };
}

/**
 * 获取菜单树
 * @returns {Promise<Object>} { menus: [...] } 树形结构
 */
export async function getMenuTree() {
  const flatMenus = await readMenuData();
  const tree = buildMenuTree(flatMenus);
  return { menus: tree };
}

/**
 * 获取单个菜单
 * @param {string} key 菜单 key
 * @returns {Promise<Object|null>} 菜单对象
 */
export async function getMenu(key) {
  const menus = await readMenuData();
  return menus.find(m => m.key === key) || null;
}

/**
 * 创建菜单
 * @param {Object} data 菜单数据
 * @returns {Promise<Object>} { success, message, menu }
 */
export async function createMenu(data) {
  const menus = await readMenuData();

  // 检查 key 是否已存在
  if (menus.some(m => m.key === data.key)) {
    return { success: false, message: '菜单标识已存在' };
  }

  const newMenu = {
    key: data.key,
    label: data.label,
    icon: data.icon || '',
    path: data.path || '',
    routeName: data.routeName || '',
    order: data.order || 0,
    hideInMenu: data.hideInMenu || false,
    fixed: data.fixed || false,
    parentKey: data.parentKey || null,
    i18nKey: data.i18nKey || '',
    href: data.href || '',
    newWindow: data.newWindow || false,
    enabled: data.enabled !== false,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  };

  menus.push(newMenu);
  await writeMenuData(menus);

  log('info', `Created menu: ${data.key}`);
  return { success: true, message: '菜单创建成功', menu: newMenu };
}

/**
 * 更新菜单
 * @param {string} key 菜单 key
 * @param {Object} data 更新数据
 * @returns {Promise<Object>} { success, message, menu }
 */
export async function updateMenu(key, data) {
  const menus = await readMenuData();
  const index = menus.findIndex(m => m.key === key);

  if (index === -1) {
    return { success: false, message: '菜单不存在' };
  }

  // 不能修改 key
  const updateData = { ...data };
  delete updateData.key;
  delete updateData.createdAt;

  menus[index] = {
    ...menus[index],
    ...updateData,
    updatedAt: new Date().toISOString()
  };

  await writeMenuData(menus);

  log('info', `Updated menu: ${key}`);
  return { success: true, message: '菜单更新成功', menu: menus[index] };
}

/**
 * 删除菜单
 * @param {string} key 菜单 key
 * @returns {Promise<Object>} { success, message }
 */
export async function deleteMenu(key) {
  const menus = await readMenuData();
  const index = menus.findIndex(m => m.key === key);

  if (index === -1) {
    return { success: false, message: '菜单不存在' };
  }

  // 删除菜单及其子菜单
  const filteredMenus = menus.filter(m => m.key !== key && m.parentKey !== key);
  await writeMenuData(filteredMenus);

  log('info', `Deleted menu: ${key}`);
  return { success: true, message: '菜单删除成功' };
}

/**
 * 批量删除菜单
 * @param {Array<string>} keys 菜单 key 列表
 * @returns {Promise<Object>} { success, deletedKeys }
 */
export async function batchDeleteMenus(keys) {
  const menus = await readMenuData();
  const deletedKeys = [];
  const allKeysToDelete = new Set(keys);

  // 收集所有需要删除的 key（包括子菜单）
  keys.forEach(key => {
    menus.forEach(m => {
      if (m.parentKey === key) {
        allKeysToDelete.add(m.key);
      }
    });
  });

  const filteredMenus = menus.filter(m => !allKeysToDelete.has(m.key));
  await writeMenuData(filteredMenus);

  log('info', `Batch deleted menus: ${[...allKeysToDelete].join(', ')}`);
  return { success: true, deletedKeys: [...allKeysToDelete] };
}

/**
 * 更新菜单排序
 * @param {Array} orderData 排序数据 [{ key, order }, ...]
 * @returns {Promise<Object>} { success }
 */
export async function reorderMenus(orderData) {
  const menus = await readMenuData();

  orderData.forEach(({ key, order }) => {
    const menu = menus.find(m => m.key === key);
    if (menu) {
      menu.order = order;
      menu.updatedAt = new Date().toISOString();
    }
  });

  await writeMenuData(menus);

  log('info', 'Menus reordered');
  return { success: true };
}

/**
 * 移动菜单
 * @param {string} key 菜单 key
 * @param {string|null} targetParentKey 目标父级 key
 * @returns {Promise<Object>} { success, message, menu }
 */
export async function moveMenu(key, targetParentKey) {
  const menus = await readMenuData();
  const index = menus.findIndex(m => m.key === key);

  if (index === -1) {
    return { success: false, message: '菜单不存在' };
  }

  // 检查是否会形成循环
  if (targetParentKey) {
    let currentKey = targetParentKey;
    while (currentKey) {
      if (currentKey === key) {
        return { success: false, message: '不能将菜单移动到自己的子菜单下' };
      }
      const parent = menus.find(m => m.key === currentKey);
      currentKey = parent?.parentKey;
    }
  }

  menus[index].parentKey = targetParentKey || null;
  menus[index].updatedAt = new Date().toISOString();

  await writeMenuData(menus);

  log('info', `Moved menu ${key} to parent ${targetParentKey || 'root'}`);
  return { success: true, message: '菜单移动成功', menu: menus[index] };
}

/**
 * 切换菜单启用状态
 * @param {string} key 菜单 key
 * @param {boolean} enabled 是否启用
 * @returns {Promise<Object>} { success, message, menu }
 */
export async function toggleMenuEnabled(key, enabled) {
  const menus = await readMenuData();
  const index = menus.findIndex(m => m.key === key);

  if (index === -1) {
    return { success: false, message: '菜单不存在' };
  }

  menus[index].enabled = enabled;
  menus[index].updatedAt = new Date().toISOString();

  await writeMenuData(menus);

  log('info', `Menu ${key} ${enabled ? 'enabled' : 'disabled'}`);
  return { success: true, message: enabled ? '菜单已启用' : '菜单已禁用', menu: menus[index] };
}
