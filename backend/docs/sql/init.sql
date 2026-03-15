-- AI Task Manager - 完整数据库初始化脚本
-- Database: MySQL 8.0+
-- Created: 2026-03-15
-- Table Prefix: task_

SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 需求表
-- ----------------------------
DROP TABLE IF EXISTS `task_requirement`;
CREATE TABLE `task_requirement` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '需求 ID',
  `title` VARCHAR(500) NOT NULL COMMENT '需求标题',
  `content` LONGTEXT COMMENT '需求内容 (Markdown 格式)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'draft' COMMENT '状态：draft-草稿，active-进行中，completed-已完成，archived-已归档',
  `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级：high-高，medium-中，low-低',
  `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签 (JSON 数组格式)',
  `assignee` VARCHAR(100) DEFAULT NULL COMMENT '负责人',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_priority` (`priority`),
  INDEX `idx_assignee` (`assignee`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求表';

-- ----------------------------
-- 2. 任务表
-- ----------------------------
DROP TABLE IF EXISTS `task_task`;
CREATE TABLE `task_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务 ID',
  `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求 ID',
  `title` VARCHAR(500) NOT NULL COMMENT '任务标题 (原文)',
  `title_trans` VARCHAR(500) DEFAULT NULL COMMENT '任务标题 (翻译)',
  `description` TEXT COMMENT '任务描述 (原文)',
  `description_trans` TEXT COMMENT '任务描述 (翻译)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '任务状态：pending-待处理，in-progress-进行中，done-已完成，deferred-已延期',
  `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级：high-高，medium-中，low-低',
  `category` VARCHAR(50) DEFAULT NULL COMMENT '分类：frontend-前端，backend-后端，fullstack-全栈',
  `details` TEXT COMMENT '任务详情 (原文)',
  `details_trans` TEXT COMMENT '任务详情 (翻译)',
  `acceptance_criteria` TEXT COMMENT '验收标准',
  `input` TEXT COMMENT '输入',
  `output` TEXT COMMENT '输出',
  `risk` TEXT COMMENT '风险',
  `module` VARCHAR(100) DEFAULT NULL COMMENT '模块',
  `test_strategy` TEXT COMMENT '测试策略 (原文)',
  `test_strategy_trans` TEXT COMMENT '测试策略 (翻译)',
  `assignee` VARCHAR(100) DEFAULT NULL COMMENT '负责人',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_priority` (`priority`),
  INDEX `idx_assignee` (`assignee`),
  INDEX `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_task_requirement` FOREIGN KEY (`requirement_id`) REFERENCES `task_requirement` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- ----------------------------
-- 3. 子任务表
-- ----------------------------
DROP TABLE IF EXISTS `task_subtask`;
CREATE TABLE `task_subtask` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '子任务 ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '所属任务 ID',
  `title` VARCHAR(500) NOT NULL COMMENT '子任务标题 (原文)',
  `title_trans` VARCHAR(500) DEFAULT NULL COMMENT '子任务标题 (翻译)',
  `description` TEXT COMMENT '子任务描述 (原文)',
  `description_trans` TEXT COMMENT '子任务描述 (翻译)',
  `details` TEXT COMMENT '子任务详情 (原文)',
  `details_trans` TEXT COMMENT '子任务详情 (翻译)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '子任务状态',
  `priority` VARCHAR(20) DEFAULT 'medium' COMMENT '优先级',
  `code_interface` TEXT COMMENT '代码接口定义 (JSON)',
  `acceptance_criteria` TEXT COMMENT '验收标准 (JSON 数组)',
  `related_files` TEXT COMMENT '关联文件 (JSON 数组)',
  `code_hints` TEXT COMMENT '代码提示',
  `sort_order` INT UNSIGNED DEFAULT 0 COMMENT '排序序号',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_subtask_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子任务表';

-- ----------------------------
-- 4. 任务依赖关系表
-- ----------------------------
DROP TABLE IF EXISTS `task_dependency`;
CREATE TABLE `task_dependency` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '依赖关系 ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务 ID',
  `depends_on_task_id` BIGINT UNSIGNED NOT NULL COMMENT '依赖的任务 ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_dependency` (`task_id`, `depends_on_task_id`),
  INDEX `idx_depends_on` (`depends_on_task_id`),
  CONSTRAINT `fk_dependency_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_dependency_depends_on` FOREIGN KEY (`depends_on_task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务依赖关系表';

-- ----------------------------
-- 5. 子任务依赖关系表
-- ----------------------------
DROP TABLE IF EXISTS `task_subtask_dependency`;
CREATE TABLE `task_subtask_dependency` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '依赖关系 ID',
  `subtask_id` BIGINT UNSIGNED NOT NULL COMMENT '子任务 ID',
  `depends_on_subtask_id` BIGINT UNSIGNED NOT NULL COMMENT '依赖的子任务 ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_subtask_dependency` (`subtask_id`, `depends_on_subtask_id`),
  INDEX `idx_depends_on` (`depends_on_subtask_id`),
  CONSTRAINT `fk_subtask_dependency_subtask` FOREIGN KEY (`subtask_id`) REFERENCES `task_subtask` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_subtask_dependency_depends_on` FOREIGN KEY (`depends_on_subtask_id`) REFERENCES `task_subtask` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子任务依赖关系表';

-- ----------------------------
-- 6. 活动记录表
-- ----------------------------
DROP TABLE IF EXISTS `task_activity`;
CREATE TABLE `task_activity` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动 ID',
  `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求 ID',
  `task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联任务 ID',
  `subtask_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联子任务 ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '用户 ID',
  `action` VARCHAR(50) NOT NULL COMMENT '操作类型：create, update, delete, status_change',
  `field` VARCHAR(100) DEFAULT NULL COMMENT '变更字段',
  `old_value` TEXT COMMENT '旧值',
  `new_value` TEXT COMMENT '新值',
  `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_subtask_id` (`subtask_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动记录表';

-- ----------------------------
-- 7. 评论表
-- ----------------------------
DROP TABLE IF EXISTS `task_comment`;
CREATE TABLE `task_comment` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论 ID',
  `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求 ID',
  `task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联任务 ID',
  `subtask_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联子任务 ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '用户 ID',
  `parent_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论 ID',
  `content` TEXT NOT NULL COMMENT '评论内容',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_subtask_id` (`subtask_id`),
  INDEX `idx_parent_id` (`parent_id`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ----------------------------
-- 8. 备份表
-- ----------------------------
DROP TABLE IF EXISTS `task_backup`;
CREATE TABLE `task_backup` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '备份 ID',
  `requirement_id` BIGINT UNSIGNED NOT NULL COMMENT '关联需求 ID',
  `backup_type` VARCHAR(20) NOT NULL DEFAULT 'full' COMMENT '备份类型：full-完整，incremental-增量',
  `data_snapshot` LONGTEXT COMMENT '数据快照 (JSON)',
  `task_count` INT UNSIGNED DEFAULT 0 COMMENT '任务数量',
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending-等待，success-成功，failed-失败',
  `error_message` TEXT COMMENT '错误信息',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='备份表';

-- ----------------------------
-- 9. 备份计划表
-- ----------------------------
DROP TABLE IF EXISTS `task_backup_schedule`;
CREATE TABLE `task_backup_schedule` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '计划 ID',
  `requirement_id` BIGINT UNSIGNED NOT NULL COMMENT '关联需求 ID',
  `interval_type` VARCHAR(20) NOT NULL COMMENT '间隔类型：hourly-每小时，daily-每天，weekly-每周，monthly-每月',
  `interval_value` INT UNSIGNED NOT NULL COMMENT '间隔值',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `last_backup_at` TIMESTAMP NULL DEFAULT NULL COMMENT '上次备份时间',
  `next_backup_at` TIMESTAMP NULL DEFAULT NULL COMMENT '下次备份时间',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_requirement_id` (`requirement_id`),
  INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='备份计划表';

-- ----------------------------
-- 10. 菜单表
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
-- 11. 语言表
-- ----------------------------
DROP TABLE IF EXISTS `task_language`;
CREATE TABLE `task_language` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '语言 ID',
  `name` VARCHAR(50) NOT NULL COMMENT '语言名称',
  `code` VARCHAR(10) NOT NULL COMMENT '语言代码',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认语言',
  `sort_order` INT UNSIGNED DEFAULT 0 COMMENT '排序序号',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  INDEX `idx_enabled` (`enabled`),
  INDEX `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='语言表';

-- ----------------------------
-- 12. 项目模板表
-- ----------------------------
DROP TABLE IF EXISTS `task_project_template`;
CREATE TABLE `task_project_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `name` VARCHAR(200) NOT NULL COMMENT '模板名称',
  `description` TEXT COMMENT '模板描述',
  `field_schema` LONGTEXT COMMENT '字段定义 (JSON)',
  `created_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人',
  `is_builtin` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否内置模板',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_is_builtin` (`is_builtin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目模板表';

-- ----------------------------
-- 13. 任务模板表
-- ----------------------------
DROP TABLE IF EXISTS `task_task_template`;
CREATE TABLE `task_task_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务模板 ID',
  `project_template_id` BIGINT UNSIGNED NOT NULL COMMENT '所属项目模板 ID',
  `title` VARCHAR(500) NOT NULL COMMENT '任务标题',
  `description` TEXT COMMENT '任务描述',
  `details` TEXT COMMENT '任务详情',
  `category` VARCHAR(50) DEFAULT NULL COMMENT '分类',
  `priority` VARCHAR(20) DEFAULT 'medium' COMMENT '优先级',
  `sort_order` INT UNSIGNED DEFAULT 0 COMMENT '排序序号',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_project_template_id` (`project_template_id`),
  CONSTRAINT `fk_task_template_project` FOREIGN KEY (`project_template_id`) REFERENCES `task_project_template` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务模板表';

-- ----------------------------
-- 14. 消息表
-- ----------------------------
DROP TABLE IF EXISTS `task_message`;
CREATE TABLE `task_message` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息 ID',
  `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求 ID',
  `task_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联任务 ID',
  `type` VARCHAR(50) NOT NULL COMMENT '消息类型：split_requirement-需求拆分，task_created-任务创建',
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending-处理中，success-成功，failed-失败',
  `title` VARCHAR(200) DEFAULT NULL COMMENT '消息标题',
  `content` TEXT COMMENT '消息内容',
  `error_message` TEXT COMMENT '错误信息',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_type` (`type`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- ----------------------------
-- 15. 任务质量评分表
-- ----------------------------
DROP TABLE IF EXISTS `task_quality_score`;
CREATE TABLE `task_quality_score` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评分 ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务 ID',
  `version` INT UNSIGNED NOT NULL COMMENT '版本号',
  `clarity_score` INT DEFAULT NULL COMMENT '清晰度评分 (1-10)',
  `completeness_score` INT DEFAULT NULL COMMENT '完整性评分 (1-10)',
  `structure_score` INT DEFAULT NULL COMMENT '结构化评分 (1-10)',
  `executability_score` INT DEFAULT NULL COMMENT '可执行性评分 (1-10)',
  `consistency_score` INT DEFAULT NULL COMMENT '一致性评分 (1-10)',
  `total_score` INT DEFAULT NULL COMMENT '总分 (0-100)',
  `comment` TEXT COMMENT 'AI 评价',
  `task_snapshot` LONGTEXT COMMENT '任务快照 (JSON)',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_version` (`version`),
  CONSTRAINT `fk_quality_score_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务质量评分表';

-- ----------------------------
-- 初始化语言数据
-- ----------------------------
INSERT INTO `task_language` (`name`, `code`, `enabled`, `is_default`, `sort_order`) VALUES
('中文', 'zh-CN', 1, 1, 1),
('English', 'en', 1, 0, 2);

-- ----------------------------
-- 初始化菜单数据
-- ----------------------------
INSERT INTO `task_menu` (`key`, `parent_key`, `title`, `path`, `route_name`, `icon`, `sort`, `enabled`, `hide_in_menu`, `fixed`, `i18n_key`) VALUES
('requirement', NULL, '需求管理', '', '', 'mdi:file-document-outline', 1, 1, 0, 0, ''),
('requirement_list', 'requirement', '需求列表', '/requirement/list', 'requirement_list', 'mdi:format-list-bulleted', 1, 1, 0, 0, 'route.requirement_list'),
('requirement_task-list', 'requirement', '任务列表', '/requirement/task-list', 'requirement_task-list', 'mdi:clipboard-list', 2, 1, 0, 0, 'route.requirement_task-list'),
('manage', '', '系统管理', '', '', 'mdi:cog', 999, 1, 0, 0, '', '', 0),
('manage_config', 'manage', '系统配置', '/manage/config', 'manage_config', 'mdi:cog-outline', 1, 1, 0, 0, 'route.manage_config'),
('manage_menu', 'manage', '菜单管理', '/manage/menu', 'manage_menu', 'mdi:menu', 2, 1, 0, 0, 'route.manage_menu'),
('team', '', '团队管理', '', '', 'mdi:account-group', 3, 0, 0, 0, '', '', 0),
('team_members', 'team', '成员管理', '/team/members', 'team_members', 'mdi:account-multiple', 1, 1, 0, 0, 'route.team_members'),
('team_workload', 'team', '工作量统计', '/team/workload', 'team_workload', 'mdi:chart-bar', 2, 1, 0, 0, 'route.team_workload'),
('templates', NULL, '模板管理', '', '', 'mdi:file-document-multiple', 4, 1, 0, 0, '', '', 0),
('templates_projects', 'templates', '项目模板', '/templates/projects', 'templates_projects', 'mdi:folder-multiple', 1, 1, 0, 0, 'route.templates_projects'),
('templates_tasks', 'templates', '任务模板', '/templates/tasks', 'templates_tasks', 'mdi:clipboard-text-multiple', 2, 1, 0, 0, 'route.templates_tasks');

SET FOREIGN_KEY_CHECKS = 1;
