import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import {
  batchDeleteMenus as batchDeleteMenusApi,
  createMenu as createMenuApi,
  deleteMenu as deleteMenuApi,
  fetchMenuList,
  moveMenu as moveMenuApi,
  reorderMenus as reorderMenusApi,
  toggleMenuEnabled as toggleMenuEnabledApi,
  updateMenu as updateMenuApi
} from '@/service/api/menu';
import type { MenuFormData, MenuItem, MenuTreeNode } from '@/typings/api/menu';

export const useMenuStore = defineStore('menu-store', () => {
  // 状态
  const menus = ref<MenuItem[]>([]);
  const loading = ref(false);
  const currentMenu = ref<MenuItem | null>(null);

  // 将扁平结构转换为树形结构
  function buildMenuTree(menuList: MenuItem[]): MenuTreeNode[] {
    const map = new Map<string, MenuTreeNode>();
    const roots: MenuTreeNode[] = [];

    // 先创建所有节点
    menuList.forEach(menu => {
      map.set(menu.key, { ...menu, children: [] });
    });

    // 构建树形结构
    menuList.forEach(menu => {
      const node = map.get(menu.key)!;
      if (menu.parentKey) {
        const parent = map.get(menu.parentKey);
        if (parent) {
          if (!parent.children) {
            parent.children = [];
          }
          parent.children.push(node);
        } else {
          roots.push(node);
        }
      } else {
        roots.push(node);
      }
    });

    // 排序
    const sortByOrder = (a: MenuTreeNode, b: MenuTreeNode) => a.order - b.order;
    const sortChildren = (nodes: MenuTreeNode[]) => {
      nodes.sort(sortByOrder);
      nodes.forEach(node => {
        if (node.children && node.children.length > 0) {
          sortChildren(node.children);
        }
      });
    };
    sortChildren(roots);

    return roots;
  }

  // 计算属性：菜单树
  const menuTree = computed(() => buildMenuTree(menus.value));

  // 计算属性：启用的菜单树
  const enabledMenuTree = computed(() => {
    const filterEnabled = (nodes: MenuTreeNode[]): MenuTreeNode[] => {
      return nodes
        .filter(node => node.enabled)
        .map(node => ({
          ...node,
          children: node.children ? filterEnabled(node.children) : []
        }))
        .filter(node => !node.hideInMenu);
    };
    return filterEnabled(menuTree.value);
  });

  // 系统管理模块的默认菜单
  const defaultSystemMenus: MenuItem[] = [
    { key: 'manage', label: '系统管理', order: 1, enabled: true },
    {
      key: 'manage_user',
      label: '用户管理',
      parentKey: 'manage',
      path: '/manage/user',
      routeName: 'manage_user',
      order: 1,
      enabled: true
    },
    {
      key: 'manage_role',
      label: '角色管理',
      parentKey: 'manage',
      path: '/manage/role',
      routeName: 'manage_role',
      order: 2,
      enabled: true
    },
    {
      key: 'manage_menu',
      label: '菜单管理',
      parentKey: 'manage',
      path: '/manage/menu',
      routeName: 'manage_menu',
      order: 3,
      enabled: true
    }
  ];

  // Actions
  async function loadMenus() {
    loading.value = true;
    try {
      const { data, error } = await fetchMenuList();
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "success", data: { list, total, page, pageSize } }
        const responseData = (data as any).data || data;

        let menuList: MenuItem[] = [];
        if (Array.isArray(responseData)) {
          menuList = responseData;
        } else if (responseData && 'list' in responseData) {
          menuList = responseData.list || [];
        } else if (responseData && 'menus' in responseData) {
          menuList = responseData.menus || [];
        }

        // 如果菜单数量为 500 或获取不到菜单，只保留系统管理模块
        if (menuList.length === 500 || menuList.length === 0) {
          menus.value = defaultSystemMenus;
        } else {
          menus.value = menuList;
        }
      } else {
        // 获取失败时使用默认系统管理菜单
        menus.value = defaultSystemMenus;
      }
    } catch (error) {
      window.$message?.error('加载菜单列表失败');
      console.error('Failed to load menus:', error);
      // 异常时使用默认系统管理菜单
      menus.value = defaultSystemMenus;
    } finally {
      loading.value = false;
    }
  }

  async function createMenu(formData: MenuFormData) {
    loading.value = true;
    try {
      const { data, error } = await createMenuApi(formData);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "success", data: { menu: {...} } }
        const responseData = (data as any).data;
        if (responseData?.menu) {
          await loadMenus();
          window.$message?.success('菜单创建成功');
          return responseData.menu;
        }
      }
      return null;
    } catch (error) {
      window.$message?.error('创建菜单失败');
      console.error('Failed to create menu:', error);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function updateMenu(key: string, formData: Partial<MenuFormData>) {
    loading.value = true;
    try {
      const { data, error } = await updateMenuApi(key, formData);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "success", data: { menu: {...} } }
        const responseData = (data as any).data;
        if (responseData?.menu) {
          await loadMenus();
          window.$message?.success('菜单更新成功');
          return responseData.menu;
        }
      }
      return null;
    } catch (error) {
      window.$message?.error('更新菜单失败');
      console.error('Failed to update menu:', error);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function deleteMenu(key: string) {
    loading.value = true;
    try {
      const { data, error } = await deleteMenuApi(key);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "删除菜单成功" }
        const response = data as any;
        if (response?.code === 0) {
          await loadMenus();
          window.$message?.success('菜单删除成功');
          return true;
        }
      }
      return false;
    } catch (error) {
      window.$message?.error('删除菜单失败');
      console.error('Failed to delete menu:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function batchDeleteMenus(keys: string[]) {
    if (!keys.length) {
      return { successKeys: [], failedKeys: [] };
    }

    loading.value = true;
    try {
      const { data, error } = await batchDeleteMenusApi(keys);
      if (!error && data) {
        const responseData = (data as any).data;
        const successKeys = responseData?.deletedKeys || [];
        const failedKeys = keys.filter(k => !successKeys.includes(k));

        if (successKeys.length) {
          await loadMenus();
        }

        return { successKeys, failedKeys };
      }
      return { successKeys: [], failedKeys: keys };
    } finally {
      loading.value = false;
    }
  }

  async function reorderMenus(orderData: { key: string; order: number }[]) {
    try {
      const { data, error } = await reorderMenusApi(orderData);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "菜单排序成功" }
        const response = data as any;
        if (response?.code === 0) {
          await loadMenus();
          window.$message?.success('菜单排序更新成功');
          return true;
        }
      }
      return false;
    } catch (error) {
      window.$message?.error('更新菜单排序失败');
      console.error('Failed to reorder menus:', error);
      return false;
    }
  }

  async function moveMenu(key: string, targetParentKey: string | null) {
    loading.value = true;
    try {
      const { data, error } = await moveMenuApi(key, targetParentKey);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "移动菜单成功" }
        const response = data as any;
        if (response?.code === 0) {
          await loadMenus();
          window.$message?.success('菜单移动成功');
          return true;
        }
      }
      return false;
    } catch (error) {
      window.$message?.error('移动菜单失败');
      console.error('Failed to move menu:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function toggleMenuEnabled(key: string, enabled: boolean) {
    try {
      const { data, error } = await toggleMenuEnabledApi(key, enabled);
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "菜单状态已更新" }
        const response = data as any;
        if (response?.code === 0) {
          const menu = menus.value.find(m => m.key === key);
          if (menu) {
            menu.enabled = enabled;
          }
          window.$message?.success(enabled ? '菜单已启用' : '菜单已禁用');
          return true;
        }
      }
      return false;
    } catch (error) {
      window.$message?.error('切换菜单状态失败');
      console.error('Failed to toggle menu:', error);
      return false;
    }
  }

  function setCurrentMenu(menu: MenuItem | null) {
    currentMenu.value = menu;
  }

  function clearCurrentMenu() {
    currentMenu.value = null;
  }

  return {
    // 状态
    menus,
    loading,
    currentMenu,
    // 计算属性
    menuTree,
    enabledMenuTree,
    // Actions
    loadMenus,
    createMenu,
    updateMenu,
    deleteMenu,
    batchDeleteMenus,
    reorderMenus,
    moveMenu,
    toggleMenuEnabled,
    setCurrentMenu,
    clearCurrentMenu
  };
});
