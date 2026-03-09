-- 需求管理模块 SQL 表结构
-- 创建时间: 2026-03-09

-- ============================================
-- 1. 需求表 (task_requirement)
-- ============================================
CREATE TABLE IF NOT EXISTS `task_requirement` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '需求ID',
  `title` VARCHAR(500) NOT NULL COMMENT '需求标题',
  `content` LONGTEXT COMMENT '需求内容(Markdown格式)',
  `status` VARCHAR(50) NOT NULL DEFAULT 'draft' COMMENT '状态: draft-草稿, active-进行中, completed-已完成, archived-已归档',
  `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级: high-高, medium-中, low-低',
  `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签(JSON数组格式)',
  `assignee` VARCHAR(100) DEFAULT NULL COMMENT '负责人',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_priority` (`priority`),
  INDEX `idx_assignee` (`assignee`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='需求表';

-- ============================================
-- 2. 需求文档表 (task_requirement_document)
-- ============================================
CREATE TABLE IF NOT EXISTS `task_requirement_document` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文档ID',
  `requirement_id` BIGINT UNSIGNED NOT NULL COMMENT '关联需求ID',
  `name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `path` VARCHAR(500) NOT NULL COMMENT '存储路径',
  `size` BIGINT UNSIGNED DEFAULT 0 COMMENT '文件大小(字节)',
  `mime_type` VARCHAR(100) DEFAULT NULL COMMENT '文件MIME类型',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '文档描述',
  `uploaded_by` VARCHAR(100) DEFAULT NULL COMMENT '上传者',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  PRIMARY KEY (`id`),
  INDEX `idx_requirement_id` (`requirement_id`),
  CONSTRAINT `fk_requirement_document_requirement`
    FOREIGN KEY (`requirement_id`)
    REFERENCES `task_requirement` (`id`)
    ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='需求文档附件表';

-- ============================================
-- 3. 修改任务表 - 添加需求关联字段
-- ============================================
-- 注意: 如果 task_task 表已存在，需要单独执行 ALTER 语句
-- ALTER TABLE `task_task`
--   ADD COLUMN `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求ID' AFTER `id`,
--   ADD INDEX `idx_requirement_id` (`requirement_id`),
--   ADD CONSTRAINT `fk_task_requirement`
--     FOREIGN KEY (`requirement_id`)
--     REFERENCES `task_requirement` (`id`)
--     ON DELETE SET NULL;

-- ============================================
-- 建议的完整任务表结构 (含需求关联)
-- ============================================
-- CREATE TABLE IF NOT EXISTS `task_task` (
--   `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
--   `requirement_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联需求ID',
--   `title` VARCHAR(500) NOT NULL,
--   `title_trans` VARCHAR(500) DEFAULT NULL,
--   `description` TEXT,
--   `description_trans` TEXT,
--   `status` VARCHAR(50) NOT NULL DEFAULT 'pending',
--   `priority` VARCHAR(20) NOT NULL DEFAULT 'medium',
--   `details` TEXT,
--   `details_trans` TEXT,
--   `test_strategy` TEXT,
--   `test_strategy_trans` TEXT,
--   `assignee` VARCHAR(100) DEFAULT NULL,
--   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--   PRIMARY KEY (`id`),
--   INDEX `idx_status` (`status`),
--   INDEX `idx_priority` (`priority`),
--   INDEX `idx_requirement_id` (`requirement_id`),
--   CONSTRAINT `fk_task_requirement`
--     FOREIGN KEY (`requirement_id`)
--     REFERENCES `task_requirement` (`id`)
--     ON DELETE SET NULL
-- ) ENGINE=InnoDB
--   DEFAULT CHARSET=utf8mb4
--   COLLATE=utf8mb4_unicode_ci;
