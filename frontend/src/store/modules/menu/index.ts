import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import {
  fetchMenuList,
  createMenu as createMenuApi,
  updateMenu as updateMenuApi,
  deleteMenu as deleteMenuApi,
  batchDeleteMenus as batchDeleteMenusApi,
  reorderMenus as reorderMenusApi,
  moveMenu as moveMenuApi,
  toggleMenuEnabled as toggleMenuEnabledApi
} from '@/service/api/menu';
import type { MenuItem, MenuFormData, MenuTreeNode } from '@/typings/api/menu';

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

  // Actions
  async function loadMenus() {
    loading.value = true;
    try {
      const { data, error } = await fetchMenuList();
      if (!error && data) {
        // 后端返回格式: { code: 0, message: "success", data: { list, total, page, pageSize } }
        const responseData = (data as any).data || data;

        if (Array.isArray(responseData)) {
          menus.value = responseData;
        } else if (responseData && 'list' in responseData) {
          menus.value = responseData.list || [];
        } else if (responseData && 'menus' in responseData) {
          menus.value = responseData.menus || [];
        }
      }
    } catch (error) {
      window.$message?.error('加载菜单列表失败');
      console.error('Failed to load menus:', error);
    } finally {
      loading.value = false;
    }
  }

  async function createMenu(formData: MenuFormData) {
    loading.value = true;
    try {
      const { data, error } = await createMenuApi(formData);
      if (!error && data) {
        const responseData = (data as any).data;
        if (responseData?.success) {
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
        const responseData = (data as any).data;
        if (responseData?.success) {
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
        const responseData = (data as any).data;
        if (responseData?.success) {
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
        const responseData = (data as any).data;
        if (responseData?.success) {
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
      if (!error && data?.success) {
        await loadMenus();
        window.$message?.success('菜单移动成功');
        return true;
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
        const responseData = (data as any).data;
        if (responseData?.success) {
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
