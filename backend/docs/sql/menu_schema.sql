-- 菜单管理系统数据库表结构
-- Database: MySQL 8.0+
-- Host: 192.168.8.167
-- Port: 3306
-- Database: task
-- Created: 2026-03-14
-- Table Prefix: task_

-- ----------------------------
-- 菜单表
-- ----------------------------
DROP TABLE IF EXISTS `task_menu`;
CREATE TABLE `task_menu` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `key` VARCHAR(100) NOT NULL COMMENT '菜单唯一标识',
  `parent_key` VARCHAR(100) DEFAULT NULL COMMENT '父级菜单 key',
  `title` VARCHAR(100) NOT NULL COMMENT '菜单名称',
  `path` VARCHAR(200) DEFAULT NULL COMMENT '路由路径',
  `route_name` VARCHAR(100) DEFAULT NULL COMMENT '路由名称',
  `icon` VARCHAR(50) DEFAULT NULL COMMENT '菜单图标',
  `sort` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '排序序号',
  `enabled` TINYINT(1) NOT NULL DEFAULT '1' COMMENT '是否启用：1-启用，0-禁用',
  `hide_in_menu` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否在菜单中隐藏：1-隐藏，0-显示',
  `fixed` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否固定在标签栏：1-固定，0-不固定',
  `i18n_key` VARCHAR(100) DEFAULT NULL COMMENT '国际化 key',
  `href` VARCHAR(200) DEFAULT NULL COMMENT '外链地址',
  `new_window` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否新窗口打开：1-是，0-否',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`key`),
  INDEX `idx_parent_key` (`parent_key`),
  INDEX `idx_sort` (`sort`),
  INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

-- ----------------------------
-- 初始化菜单数据
-- ----------------------------
INSERT INTO `task_menu` (`key`, `parent_key`, `title`, `path`, `route_name`, `icon`, `sort`, `enabled`, `hide_in_menu`, `fixed`, `i18n_key`, `href`, `new_window`) VALUES
-- 需求管理模块
('requirement', NULL, '需求管理', '', '', 'mdi:file-document-outline', 1, 1, 0, 0, '', '', 0),
('requirement_list', 'requirement', '需求列表', '/requirement/list', 'requirement_list', 'mdi:format-list-bulleted', 1, 1, 0, 0, 'route.requirement_list', '', 0),
('requirement_task-list', 'requirement', '任务列表', '/requirement/task-list', 'requirement_task-list', 'mdi:clipboard-list', 2, 1, 0, 0, 'route.requirement_task-list', '', 0),

-- 系统管理模块
('manage', '', '系统管理', '', '', 'mdi:cog', 999, 1, 0, 0, '', '', 0),
('manage_config', 'manage', '系统配置', '/manage/config', 'manage_config', 'mdi:cog-outline', 1, 1, 0, 0, 'route.manage_config', '', 0),
('manage_menu', 'manage', '菜单管理', '/manage/menu', 'manage_menu', 'mdi:menu', 2, 1, 0, 0, 'route.manage_menu', '', 0),

-- 团队管理模块（默认禁用）
('team', '', '团队管理', '', '', 'mdi:account-group', 3, 0, 0, 0, '', '', 0),
('team_members', 'team', '成员管理', '/team/members', 'team_members', 'mdi:account-multiple', 1, 1, 0, 0, 'route.team_members', '', 0),
('team_workload', 'team', '工作量统计', '/team/workload', 'team_workload', 'mdi:chart-bar', 2, 1, 0, 0, 'route.team_workload', '', 0),

-- 模板管理模块
('templates', NULL, '模板管理', '', '', 'mdi:file-document-multiple', 4, 1, 0, 0, '', '', 0),
('templates_projects', 'templates', '项目模板', '/templates/projects', 'templates_projects', 'mdi:folder-multiple', 1, 1, 0, 0, 'route.templates_projects', '', 0),
('templates_tasks', 'templates', '任务模板', '/templates/tasks', 'templates_tasks', 'mdi:clipboard-text-multiple', 2, 1, 0, 0, 'route.templates_tasks', '', 0);
