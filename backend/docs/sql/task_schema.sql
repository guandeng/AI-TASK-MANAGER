-- 任务管理系统数据库表结构
-- Database: MySQL 8.0+
-- Host: 192.168.8.167
-- Port: 3306
-- Database: task
-- Created: 2026-03-08
-- Table Prefix: task_

-- ----------------------------
-- 1. 任务表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `task_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求ID',
  `title` VARCHAR(500) NOT NULL COMMENT '任务标题(原文)',
  `title_trans` VARCHAR(500) DEFAULT NULL COMMENT '任务标题(翻译)',
  `description` TEXT COMMENT '任务描述(原文)',
  `description_trans` TEXT COMMENT '任务描述(翻译)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '任务状态: pending-待处理, in-progress-进行中, done-已完成, deferred-已延期',
  `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级: high-高, medium-中, low-低',
  `details` TEXT COMMENT '任务详情(原文)',
  `details_trans` TEXT COMMENT '任务详情(翻译)',
  `test_strategy` TEXT COMMENT '测试策略(原文)',
  `test_strategy_trans` TEXT COMMENT '测试策略(翻译)',
  `module` VARCHAR(100) DEFAULT NULL COMMENT '模块归属',
  `input` TEXT DEFAULT NULL COMMENT '输入依赖',
  `output` TEXT DEFAULT NULL COMMENT '输出交付物',
  `risk` TEXT DEFAULT NULL COMMENT '风险点',
  `acceptance_criteria` TEXT DEFAULT NULL COMMENT '验收标准(JSON)',
  `assignee` VARCHAR(100) DEFAULT NULL COMMENT '负责人',
  `start_date` DATE DEFAULT NULL COMMENT '开始日期',
  `due_date` DATE DEFAULT NULL COMMENT '截止日期',
  `estimated_hours` DECIMAL(10,2) DEFAULT NULL COMMENT '预估工时',
  `actual_hours` DECIMAL(10,2) DEFAULT NULL COMMENT '实际工时',
  `completed_at` TIMESTAMP NULL DEFAULT NULL COMMENT '完成时间',
  `tags` JSON DEFAULT NULL COMMENT '标签',
  `custom_fields` LONGTEXT DEFAULT NULL COMMENT '自定义字段(JSON)',
  `is_expanding` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否正在拆分',
  `expand_message_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '拆分消息ID',
  `expand_started_at` TIMESTAMP NULL DEFAULT NULL COMMENT '拆分开始时间',
  `category` VARCHAR(50) DEFAULT NULL COMMENT '分类: frontend-前端, backend-后端',
  `language_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '编程语言ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_priority` (`priority`),
  INDEX `idx_assignee` (`assignee`),
  INDEX `idx_created_at` (`created_at`),
  INDEX `idx_category` (`category`),
  INDEX `idx_language_id` (`language_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- ----------------------------
-- 2. 子任务表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `task_subtask` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '子任务ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '所属任务ID',
  `title` VARCHAR(500) NOT NULL COMMENT '子任务标题(原文)',
  `title_trans` VARCHAR(500) DEFAULT NULL COMMENT '子任务标题(翻译)',
  `description` TEXT COMMENT '子任务描述(原文)',
  `description_trans` TEXT COMMENT '子任务描述(翻译)',
  `details` TEXT COMMENT '子任务详情(原文)',
  `details_trans` TEXT COMMENT '子任务详情(翻译)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '子任务状态',
  `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级',
  `sort_order` INT UNSIGNED DEFAULT 0 COMMENT '排序序号',
  `estimated_hours` DECIMAL(10,2) DEFAULT NULL COMMENT '预估工时',
  `actual_hours` DECIMAL(10,2) DEFAULT NULL COMMENT '实际工时',
  `code_interface` TEXT DEFAULT NULL COMMENT '代码接口(JSON)',
  `acceptance_criteria` TEXT DEFAULT NULL COMMENT '验收标准(JSON)',
  `related_files` TEXT DEFAULT NULL COMMENT '关联文件(JSON)',
  `code_hints` TEXT DEFAULT NULL COMMENT '代码提示',
  `risk` TEXT DEFAULT NULL COMMENT '风险点',
  `custom_fields` LONGTEXT DEFAULT NULL COMMENT '自定义字段(JSON)',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_status` (`status`),
  CONSTRAINT `fk_subtask_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子任务表';

-- ----------------------------
-- 3. 任务依赖关系表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `task_dependency` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '依赖关系ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
  `depends_on_task_id` BIGINT UNSIGNED NOT NULL COMMENT '依赖的任务ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_dependency` (`task_id`, `depends_on_task_id`),
  INDEX `idx_depends_on` (`depends_on_task_id`),
  CONSTRAINT `fk_dependency_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_dependency_depends_on` FOREIGN KEY (`depends_on_task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务依赖关系表';

-- ----------------------------
-- 4. 子任务依赖关系表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `task_subtask_dependency` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '依赖关系ID',
  `subtask_id` BIGINT UNSIGNED NOT NULL COMMENT '子任务ID',
  `depends_on_subtask_id` BIGINT UNSIGNED NOT NULL COMMENT '依赖的子任务ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_subtask_dependency` (`subtask_id`, `depends_on_subtask_id`),
  INDEX `idx_depends_on` (`depends_on_subtask_id`),
  CONSTRAINT `fk_subtask_dependency_subtask` FOREIGN KEY (`subtask_id`) REFERENCES `task_subtask` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_subtask_dependency_depends_on` FOREIGN KEY (`depends_on_subtask_id`) REFERENCES `task_subtask` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子任务依赖关系表';

-- ----------------------------
-- 5. 任务历史记录表 (可选 - 用于追踪状态变更)
-- ----------------------------
CREATE TABLE IF NOT EXISTS `task_history` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
  `subtask_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '子任务ID(如果变更的是子任务)',
  `field_name` VARCHAR(50) NOT NULL COMMENT '变更字段名',
  `old_value` VARCHAR(500) DEFAULT NULL COMMENT '旧值',
  `new_value` VARCHAR(500) DEFAULT NULL COMMENT '新值',
  `changed_by` VARCHAR(100) DEFAULT NULL COMMENT '变更人',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '变更时间',
  PRIMARY KEY (`id`),
  INDEX `idx_task_id` (`task_id`),
  INDEX `idx_subtask_id` (`subtask_id`),
  INDEX `idx_created_at` (`created_at`),
  CONSTRAINT `fk_history_task` FOREIGN KEY (`task_id`) REFERENCES `task_task` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_history_subtask` FOREIGN KEY (`subtask_id`) REFERENCES `task_subtask` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务历史记录表';

-- ----------------------------
-- 示例数据
-- ----------------------------
-- INSERT INTO `task_task` (`title`, `title_trans`, `description`, `status`, `priority`, `assignee`) VALUES
-- ('Implement user authentication', '实现用户认证', 'Add login and registration functionality', 'pending', 'high', '张三'),
-- ('Create dashboard', '创建仪表盘', 'Build the main dashboard page', 'in-progress', 'medium', '李四');

-- INSERT INTO `task_subtask` (`task_id`, `title`, `title_trans`, `status`) VALUES
-- (1, 'Design login form', '设计登录表单', 'done'),
-- (1, 'Implement JWT token', '实现JWT令牌', 'in-progress');
