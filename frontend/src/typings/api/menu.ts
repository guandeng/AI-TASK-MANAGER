/** 菜单项类型 */
export interface MenuItem {
  /** 菜单唯一标识 */
  key: string;
  /** 菜单名称 */
  label: string;
  /** 菜单图标 */
  icon?: string;
  /** 路由路径 */
  path?: string;
  /** 路由名称 */
  routeName?: string;
  /** 排序顺序 */
  order: number;
  /** 是否在菜单中隐藏 */
  hideInMenu?: boolean;
  /** 是否固定在标签栏 */
  fixed?: boolean;
  /** 父级菜单 key */
  parentKey?: string;
  /** 子菜单 */
  children?: MenuItem[];
  /** 国际化 key */
  i18nKey?: string;
  /** 链接地址(外链) */
  href?: string;
  /** 是否新窗口打开 */
  newWindow?: boolean;
  /** 是否启用 */
  enabled: boolean;
  /** 创建时间 */
  createdAt?: string;
  /** 更新时间 */
  updatedAt?: string;
}

/** 菜单树节点(用于前端展示) */
export interface MenuTreeNode extends MenuItem {
  /** 子节点 */
  children?: MenuTreeNode[];
}

/** 菜单表单数据 */
export interface MenuFormData {
  key: string;
  label: string;
  icon?: string;
  path?: string;
  routeName?: string;
  order: number;
  hideInMenu: boolean;
  fixed: boolean;
  parentKey?: string;
  i18nKey?: string;
  href?: string;
  newWindow: boolean;
  enabled: boolean;
}

/** 菜单列表响应 */
export interface MenuListResponse {
  menus: MenuItem[];
}

/** 菜单操作响应 */
export interface MenuOperationResponse {
  success: boolean;
  message: string;
  menu?: MenuItem;
}
