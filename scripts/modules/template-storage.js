/**
 * template-storage.js
 * 模板数据存储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const PROJECT_TEMPLATE_TABLE = `${TABLE_PREFIX}project_template`;
const TEMPLATE_TASK_TABLE = `${TABLE_PREFIX}template_task`;
const TEMPLATE_SUBTASK_TABLE = `${TABLE_PREFIX}template_subtask`;
const TASK_TEMPLATE_TABLE = `${TABLE_PREFIX}task_template`;
const MEMBER_TABLE = `${TABLE_PREFIX}member`;

let writePool = null;
let readPools = [];
let initPromise = null;
let readPoolIndex = 0;

function getDbConfig(host) {
  return {
    host,
    port: Number(process.env.DB_PORT || 3306),
    user: process.env.DB_USERNAME,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_DATABASE,
    waitForConnections: true,
    connectionLimit: Number(process.env.DB_POOL_SIZE || 10),
    queueLimit: 0,
    charset: process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'
  };
}

function parseSlaveHosts() {
  return (process.env.DB_HOST_SLAVE || '')
    .split(',')
    .map(host => host.trim())
    .filter(Boolean);
}

async function ensureDbReady() {
  if (initPromise) {
    await initPromise;
    return;
  }

  initPromise = (async () => {
    if (!process.env.DB_HOST || !process.env.DB_DATABASE || !process.env.DB_USERNAME) {
      throw new Error('Missing MySQL configuration. Please set DB_HOST, DB_DATABASE and DB_USERNAME.');
    }

    writePool = mysql.createPool(getDbConfig(process.env.DB_HOST));
    readPools = parseSlaveHosts().map(host => mysql.createPool(getDbConfig(host)));

    // 创建项目模板表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${PROJECT_TEMPLATE_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`name\` VARCHAR(200) NOT NULL,
        \`description\` TEXT,
        \`category\` VARCHAR(50) DEFAULT NULL,
        \`icon\` VARCHAR(100) DEFAULT NULL,
        \`is_public\` TINYINT(1) NOT NULL DEFAULT 0,
        \`created_by\` BIGINT UNSIGNED DEFAULT NULL,
        \`usage_count\` INT UNSIGNED DEFAULT 0,
        \`tags\` VARCHAR(500) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_category\` (\`category\`),
        INDEX \`idx_is_public\` (\`is_public\`),
        INDEX \`idx_usage_count\` (\`usage_count\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 创建模板任务表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${TEMPLATE_TASK_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`template_id\` BIGINT UNSIGNED NOT NULL,
        \`title\` VARCHAR(500) NOT NULL,
        \`description\` TEXT,
        \`details\` TEXT,
        \`test_strategy\` TEXT,
        \`priority\` VARCHAR(20) NOT NULL DEFAULT 'medium',
        \`estimated_hours\` DECIMAL(10, 2) DEFAULT NULL,
        \`sort_order\` INT UNSIGNED DEFAULT 0,
        \`dependencies\` VARCHAR(500) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_template_id\` (\`template_id\`),
        INDEX \`idx_sort_order\` (\`sort_order\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 创建模板子任务表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${TEMPLATE_SUBTASK_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`template_task_id\` BIGINT UNSIGNED NOT NULL,
        \`title\` VARCHAR(500) NOT NULL,
        \`description\` TEXT,
        \`details\` TEXT,
        \`sort_order\` INT UNSIGNED DEFAULT 0,
        \`dependencies\` VARCHAR(500) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_template_task_id\` (\`template_task_id\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 创建独立任务模板表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${TASK_TEMPLATE_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`name\` VARCHAR(200) NOT NULL,
        \`title\` VARCHAR(500) NOT NULL,
        \`description\` TEXT,
        \`details\` TEXT,
        \`test_strategy\` TEXT,
        \`priority\` VARCHAR(20) NOT NULL DEFAULT 'medium',
        \`category\` VARCHAR(50) DEFAULT NULL,
        \`tags\` VARCHAR(500) DEFAULT NULL,
        \`is_public\` TINYINT(1) NOT NULL DEFAULT 0,
        \`created_by\` BIGINT UNSIGNED DEFAULT NULL,
        \`usage_count\` INT UNSIGNED DEFAULT 0,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_category\` (\`category\`),
        INDEX \`idx_is_public\` (\`is_public\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    console.log('[TemplateStorage] Template tables initialized');
  })();

  await initPromise;
}

function getReadPool() {
  if (!readPools.length) {
    return writePool;
  }
  const pool = readPools[readPoolIndex % readPools.length];
  readPoolIndex += 1;
  return pool;
}

// ============================================
// 项目模板相关函数
// ============================================

/**
 * 获取项目模板列表
 */
export async function getProjectTemplateList(filters = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${PROJECT_TEMPLATE_TABLE}\` WHERE 1=1`;
  const params = [];

  if (filters.category) {
    sql += ` AND \`category\` = ?`;
    params.push(filters.category);
  }

  if (filters.isPublic !== undefined) {
    sql += ` AND \`is_public\` = ?`;
    params.push(filters.isPublic ? 1 : 0);
  }

  if (filters.keyword) {
    sql += ` AND (\`name\` LIKE ? OR \`description\` LIKE ?)`;
    params.push(`%${filters.keyword}%`, `%${filters.keyword}%`);
  }

  sql += ` ORDER BY \`usage_count\` DESC, \`created_at\` DESC`;

  if (filters.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(filters.limit));
  }

  const [rows] = await getReadPool().query(sql, params);

  return rows.map(row => ({
    id: Number(row.id),
    name: row.name,
    description: row.description || '',
    category: row.category,
    icon: row.icon,
    isPublic: Boolean(row.is_public),
    createdBy: row.created_by ? Number(row.created_by) : undefined,
    usageCount: Number(row.usage_count),
    tags: row.tags ? JSON.parse(row.tags) : [],
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  }));
}

/**
 * 获取项目模板详情
 */
export async function getProjectTemplateById(id) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${PROJECT_TEMPLATE_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  if (rows.length === 0) {
    return null;
  }

  const row = rows[0];
  const template = {
    id: Number(row.id),
    name: row.name,
    description: row.description || '',
    category: row.category,
    icon: row.icon,
    isPublic: Boolean(row.is_public),
    createdBy: row.created_by ? Number(row.created_by) : undefined,
    usageCount: Number(row.usage_count),
    tags: row.tags ? JSON.parse(row.tags) : [],
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString(),
    tasks: []
  };

  // 获取模板任务
  const [taskRows] = await getReadPool().query(
    `SELECT * FROM \`${TEMPLATE_TASK_TABLE}\` WHERE \`template_id\` = ? ORDER BY \`sort_order\``,
    [Number(id)]
  );

  for (const taskRow of taskRows) {
    const task = {
      id: Number(taskRow.id),
      title: taskRow.title,
      description: taskRow.description || '',
      details: taskRow.details || '',
      testStrategy: taskRow.test_strategy || '',
      priority: taskRow.priority,
      estimatedHours: taskRow.estimated_hours ? Number(taskRow.estimated_hours) : undefined,
      sortOrder: Number(taskRow.sort_order),
      dependencies: taskRow.dependencies ? JSON.parse(taskRow.dependencies) : [],
      subtasks: []
    };

    // 获取子任务
    const [subtaskRows] = await getReadPool().query(
      `SELECT * FROM \`${TEMPLATE_SUBTASK_TABLE}\` WHERE \`template_task_id\` = ? ORDER BY \`sort_order\``,
      [taskRow.id]
    );

    task.subtasks = subtaskRows.map(st => ({
      id: Number(st.id),
      title: st.title,
      description: st.description || '',
      details: st.details || '',
      sortOrder: Number(st.sort_order),
      dependencies: st.dependencies ? JSON.parse(st.dependencies) : []
    }));

    template.tasks.push(task);
  }

  return template;
}

/**
 * 创建项目模板
 */
export async function createProjectTemplate(data) {
  await ensureDbReady();

  if (!data.name || !data.name.trim()) {
    throw new Error('模板名称不能为空');
  }

  const [result] = await writePool.query(
    `INSERT INTO \`${PROJECT_TEMPLATE_TABLE}\` (name, description, category, icon, is_public, created_by, tags)
     VALUES (?, ?, ?, ?, ?, ?, ?)`,
    [
      data.name.trim(),
      data.description || '',
      data.category || null,
      data.icon || null,
      data.isPublic ? 1 : 0,
      data.createdBy || null,
      data.tags ? JSON.stringify(data.tags) : null
    ]
  );

  const templateId = result.insertId;

  // 创建模板任务
  if (data.tasks && data.tasks.length > 0) {
    for (let i = 0; i < data.tasks.length; i++) {
      const task = data.tasks[i];
      const [taskResult] = await writePool.query(
        `INSERT INTO \`${TEMPLATE_TASK_TABLE}\` (template_id, title, description, details, test_strategy, priority, estimated_hours, sort_order, dependencies)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          templateId,
          task.title,
          task.description || '',
          task.details || '',
          task.testStrategy || '',
          task.priority || 'medium',
          task.estimatedHours || null,
          i,
          task.dependencies ? JSON.stringify(task.dependencies) : null
        ]
      );

      // 创建子任务
      if (task.subtasks && task.subtasks.length > 0) {
        for (let j = 0; j < task.subtasks.length; j++) {
          const subtask = task.subtasks[j];
          await writePool.query(
            `INSERT INTO \`${TEMPLATE_SUBTASK_TABLE}\` (template_task_id, title, description, details, sort_order, dependencies)
             VALUES (?, ?, ?, ?, ?, ?)`,
            [
              taskResult.insertId,
              subtask.title,
              subtask.description || '',
              subtask.details || '',
              j,
              subtask.dependencies ? JSON.stringify(subtask.dependencies) : null
            ]
          );
        }
      }
    }
  }

  return getProjectTemplateById(templateId);
}

/**
 * 更新项目模板
 */
export async function updateProjectTemplate(id, data) {
  await ensureDbReady();

  const existing = await getProjectTemplateById(id);
  if (!existing) {
    throw new Error('模板不存在');
  }

  const fields = [];
  const params = [];

  if (data.name !== undefined) {
    fields.push('`name` = ?');
    params.push(data.name.trim());
  }

  if (data.description !== undefined) {
    fields.push('`description` = ?');
    params.push(data.description);
  }

  if (data.category !== undefined) {
    fields.push('`category` = ?');
    params.push(data.category);
  }

  if (data.icon !== undefined) {
    fields.push('`icon` = ?');
    params.push(data.icon);
  }

  if (data.isPublic !== undefined) {
    fields.push('`is_public` = ?');
    params.push(data.isPublic ? 1 : 0);
  }

  if (data.tags !== undefined) {
    fields.push('`tags` = ?');
    params.push(data.tags ? JSON.stringify(data.tags) : null);
  }

  if (fields.length > 0) {
    params.push(Number(id));
    await writePool.query(
      `UPDATE \`${PROJECT_TEMPLATE_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
      params
    );
  }

  return getProjectTemplateById(id);
}

/**
 * 删除项目模板
 */
export async function deleteProjectTemplate(id) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `DELETE FROM \`${PROJECT_TEMPLATE_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  return result.affectedRows > 0;
}

/**
 * 增加模板使用次数
 */
export async function incrementTemplateUsage(id) {
  await ensureDbReady();

  await writePool.query(
    `UPDATE \`${PROJECT_TEMPLATE_TABLE}\` SET \`usage_count\` = \`usage_count\` + 1 WHERE \`id\` = ?`,
    [Number(id)]
  );
}

// ============================================
// 独立任务模板相关函数
// ============================================

/**
 * 获取任务模板列表
 */
export async function getTaskTemplateList(filters = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${TASK_TEMPLATE_TABLE}\` WHERE 1=1`;
  const params = [];

  if (filters.category) {
    sql += ` AND \`category\` = ?`;
    params.push(filters.category);
  }

  if (filters.isPublic !== undefined) {
    sql += ` AND \`is_public\` = ?`;
    params.push(filters.isPublic ? 1 : 0);
  }

  if (filters.keyword) {
    sql += ` AND (\`name\` LIKE ? OR \`title\` LIKE ? OR \`description\` LIKE ?)`;
    params.push(`%${filters.keyword}%`, `%${filters.keyword}%`, `%${filters.keyword}%`);
  }

  sql += ` ORDER BY \`usage_count\` DESC, \`created_at\` DESC`;

  if (filters.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(filters.limit));
  }

  const [rows] = await getReadPool().query(sql, params);

  return rows.map(row => ({
    id: Number(row.id),
    name: row.name,
    title: row.title,
    description: row.description || '',
    details: row.details || '',
    testStrategy: row.test_strategy || '',
    priority: row.priority,
    category: row.category,
    tags: row.tags ? JSON.parse(row.tags) : [],
    isPublic: Boolean(row.is_public),
    createdBy: row.created_by ? Number(row.created_by) : undefined,
    usageCount: Number(row.usage_count),
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  }));
}

/**
 * 获取任务模板详情
 */
export async function getTaskTemplateById(id) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${TASK_TEMPLATE_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  if (rows.length === 0) {
    return null;
  }

  const row = rows[0];
  return {
    id: Number(row.id),
    name: row.name,
    title: row.title,
    description: row.description || '',
    details: row.details || '',
    testStrategy: row.test_strategy || '',
    priority: row.priority,
    category: row.category,
    tags: row.tags ? JSON.parse(row.tags) : [],
    isPublic: Boolean(row.is_public),
    createdBy: row.created_by ? Number(row.created_by) : undefined,
    usageCount: Number(row.usage_count),
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  };
}

/**
 * 创建任务模板
 */
export async function createTaskTemplate(data) {
  await ensureDbReady();

  if (!data.name || !data.name.trim()) {
    throw new Error('模板名称不能为空');
  }

  if (!data.title || !data.title.trim()) {
    throw new Error('任务标题不能为空');
  }

  const [result] = await writePool.query(
    `INSERT INTO \`${TASK_TEMPLATE_TABLE}\` (name, title, description, details, test_strategy, priority, category, tags, is_public, created_by)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      data.name.trim(),
      data.title.trim(),
      data.description || '',
      data.details || '',
      data.testStrategy || '',
      data.priority || 'medium',
      data.category || null,
      data.tags ? JSON.stringify(data.tags) : null,
      data.isPublic ? 1 : 0,
      data.createdBy || null
    ]
  );

  return getTaskTemplateById(result.insertId);
}

/**
 * 更新任务模板
 */
export async function updateTaskTemplate(id, data) {
  await ensureDbReady();

  const existing = await getTaskTemplateById(id);
  if (!existing) {
    throw new Error('模板不存在');
  }

  const fields = [];
  const params = [];

  if (data.name !== undefined) {
    fields.push('`name` = ?');
    params.push(data.name.trim());
  }

  if (data.title !== undefined) {
    fields.push('`title` = ?');
    params.push(data.title.trim());
  }

  if (data.description !== undefined) {
    fields.push('`description` = ?');
    params.push(data.description);
  }

  if (data.details !== undefined) {
    fields.push('`details` = ?');
    params.push(data.details);
  }

  if (data.testStrategy !== undefined) {
    fields.push('`test_strategy` = ?');
    params.push(data.testStrategy);
  }

  if (data.priority !== undefined) {
    fields.push('`priority` = ?');
    params.push(data.priority);
  }

  if (data.category !== undefined) {
    fields.push('`category` = ?');
    params.push(data.category);
  }

  if (data.tags !== undefined) {
    fields.push('`tags` = ?');
    params.push(data.tags ? JSON.stringify(data.tags) : null);
  }

  if (data.isPublic !== undefined) {
    fields.push('`is_public` = ?');
    params.push(data.isPublic ? 1 : 0);
  }

  if (fields.length === 0) {
    return existing;
  }

  params.push(Number(id));
  await writePool.query(
    `UPDATE \`${TASK_TEMPLATE_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
    params
  );

  return getTaskTemplateById(id);
}

/**
 * 删除任务模板
 */
export async function deleteTaskTemplate(id) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `DELETE FROM \`${TASK_TEMPLATE_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  return result.affectedRows > 0;
}

/**
 * 增加任务模板使用次数
 */
export async function incrementTaskTemplateUsage(id) {
  await ensureDbReady();

  await writePool.query(
    `UPDATE \`${TASK_TEMPLATE_TABLE}\` SET \`usage_count\` = \`usage_count\` + 1 WHERE \`id\` = ?`,
    [Number(id)]
  );
}

/**
 * 获取模板分类列表
 */
export async function getTemplateCategories() {
  await ensureDbReady();

  const [projectCategories] = await getReadPool().query(`
    SELECT DISTINCT category FROM \`${PROJECT_TEMPLATE_TABLE}\` WHERE category IS NOT NULL ORDER BY category
  `);

  const [taskCategories] = await getReadPool().query(`
    SELECT DISTINCT category FROM \`${TASK_TEMPLATE_TABLE}\` WHERE category IS NOT NULL ORDER BY category
  `);

  const allCategories = new Set([
    ...projectCategories.map(r => r.category),
    ...taskCategories.map(r => r.category)
  ]);

  return [...allCategories];
}

/**
 * 关闭数据库连接
 */
export async function closeTemplateStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export default {
  getProjectTemplateList,
  getProjectTemplateById,
  createProjectTemplate,
  updateProjectTemplate,
  deleteProjectTemplate,
  incrementTemplateUsage,
  getTaskTemplateList,
  getTaskTemplateById,
  createTaskTemplate,
  updateTaskTemplate,
  deleteTaskTemplate,
  incrementTaskTemplateUsage,
  getTemplateCategories,
  closeTemplateStorage
};
