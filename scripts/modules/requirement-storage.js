import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import mysql from 'mysql2/promise';
import { exportToEnv } from './config-loader.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const REQUIREMENT_TABLE = `${TABLE_PREFIX}requirement`;
const DOCUMENT_TABLE = `${TABLE_PREFIX}requirement_document`;

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

    // 创建需求表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${REQUIREMENT_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`title\` VARCHAR(500) NOT NULL,
        \`content\` LONGTEXT,
        \`status\` VARCHAR(50) NOT NULL DEFAULT 'draft',
        \`priority\` VARCHAR(20) NOT NULL DEFAULT 'medium',
        \`tags\` VARCHAR(500) DEFAULT NULL,
        \`assignee\` VARCHAR(100) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_status\` (\`status\`),
        INDEX \`idx_priority\` (\`priority\`),
        INDEX \`idx_assignee\` (\`assignee\`),
        INDEX \`idx_created_at\` (\`created_at\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 创建文档表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${DOCUMENT_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`requirement_id\` BIGINT UNSIGNED NOT NULL,
        \`name\` VARCHAR(255) NOT NULL,
        \`path\` VARCHAR(500) NOT NULL,
        \`size\` BIGINT UNSIGNED DEFAULT 0,
        \`mime_type\` VARCHAR(100) DEFAULT NULL,
        \`description\` VARCHAR(500) DEFAULT NULL,
        \`uploaded_by\` VARCHAR(100) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_requirement_id\` (\`requirement_id\`),
        CONSTRAINT \`fk_requirement_document_requirement\`
          FOREIGN KEY (\`requirement_id\`)
          REFERENCES \`${REQUIREMENT_TABLE}\` (\`id\`)
          ON DELETE CASCADE
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 检查并添加任务表的需求关联字段
    const TASK_TABLE = `${TABLE_PREFIX}task`;
    try {
      const [columns] = await writePool.query(`
        SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = 'requirement_id'
      `, [process.env.DB_DATABASE, TASK_TABLE]);

      if (columns.length === 0) {
        await writePool.query(`
          ALTER TABLE \`${TASK_TABLE}\`
          ADD COLUMN \`requirement_id\` BIGINT UNSIGNED DEFAULT NULL AFTER \`requirement\`,
          ADD INDEX \`idx_requirement_id\` (\`requirement_id\`)
        `);
        console.log('[RequirementStorage] Added requirement_id column to task table');
      }
    } catch (err) {
      console.warn('[RequirementStorage] Could not add requirement_id to task table:', err.message);
    }
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

function toNullableString(value) {
  return value == null ? null : String(value);
}

/**
 * 获取需求列表
 */
export async function getRequirementList(filters = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${REQUIREMENT_TABLE}\` WHERE 1=1`;
  const params = [];

  if (filters.status) {
    sql += ` AND \`status\` = ?`;
    params.push(filters.status);
  }

  if (filters.priority) {
    sql += ` AND \`priority\` = ?`;
    params.push(filters.priority);
  }

  if (filters.assignee) {
    sql += ` AND \`assignee\` = ?`;
    params.push(filters.assignee);
  }

  if (filters.keyword) {
    sql += ` AND (\`title\` LIKE ? OR \`content\` LIKE ?)`;
    params.push(`%${filters.keyword}%`, `%${filters.keyword}%`);
  }

  sql += ` ORDER BY \`created_at\` DESC`;

  if (filters.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(filters.limit));
  }

  const [rows] = await getReadPool().query(sql, params);

  return rows.map(row => ({
    id: Number(row.id),
    title: row.title,
    content: row.content || '',
    status: row.status,
    priority: row.priority,
    tags: row.tags ? JSON.parse(row.tags) : [],
    assignee: row.assignee || undefined,
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  }));
}

/**
 * 获取需求详情（含文档列表）
 */
export async function getRequirementById(id) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${REQUIREMENT_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  if (rows.length === 0) {
    return null;
  }

  const row = rows[0];
  const requirement = {
    id: Number(row.id),
    title: row.title,
    content: row.content || '',
    status: row.status,
    priority: row.priority,
    tags: row.tags ? JSON.parse(row.tags) : [],
    assignee: row.assignee || undefined,
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString(),
    documents: []
  };

  // 获取关联文档
  const [docRows] = await getReadPool().query(
    `SELECT * FROM \`${DOCUMENT_TABLE}\` WHERE \`requirement_id\` = ? ORDER BY \`created_at\` DESC`,
    [Number(id)]
  );

  requirement.documents = docRows.map(doc => ({
    id: Number(doc.id),
    name: doc.name,
    path: doc.path,
    size: Number(doc.size),
    mimeType: doc.mime_type,
    description: doc.description || undefined,
    uploadedBy: doc.uploaded_by || undefined,
    createdAt: doc.created_at.toISOString()
  }));

  return requirement;
}

/**
 * 创建需求
 */
export async function createRequirement(data) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `INSERT INTO \`${REQUIREMENT_TABLE}\` (title, content, status, priority, tags, assignee)
     VALUES (?, ?, ?, ?, ?, ?)`,
    [
      data.title,
      data.content || '',
      data.status || 'draft',
      data.priority || 'medium',
      data.tags ? JSON.stringify(data.tags) : null,
      toNullableString(data.assignee)
    ]
  );

  return getRequirementById(result.insertId);
}

/**
 * 更新需求
 */
export async function updateRequirement(id, data) {
  await ensureDbReady();

  const fields = [];
  const params = [];

  if (data.title !== undefined) {
    fields.push('`title` = ?');
    params.push(data.title);
  }

  if (data.content !== undefined) {
    fields.push('`content` = ?');
    params.push(data.content);
  }

  if (data.status !== undefined) {
    fields.push('`status` = ?');
    params.push(data.status);
  }

  if (data.priority !== undefined) {
    fields.push('`priority` = ?');
    params.push(data.priority);
  }

  if (data.tags !== undefined) {
    fields.push('`tags` = ?');
    params.push(data.tags ? JSON.stringify(data.tags) : null);
  }

  if (data.assignee !== undefined) {
    fields.push('`assignee` = ?');
    params.push(toNullableString(data.assignee));
  }

  if (fields.length === 0) {
    return getRequirementById(id);
  }

  params.push(Number(id));

  await writePool.query(
    `UPDATE \`${REQUIREMENT_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
    params
  );

  return getRequirementById(id);
}

/**
 * 删除需求
 */
export async function deleteRequirement(id) {
  await ensureDbReady();

  // 删除关联文档文件
  const [docRows] = await writePool.query(
    `SELECT \`path\` FROM \`${DOCUMENT_TABLE}\` WHERE \`requirement_id\` = ?`,
    [Number(id)]
  );

  for (const doc of docRows) {
    try {
      if (fs.existsSync(doc.path)) {
        fs.unlinkSync(doc.path);
      }
    } catch (err) {
      console.warn(`[RequirementStorage] Failed to delete file ${doc.path}:`, err.message);
    }
  }

  // 数据库会自动删除文档记录（CASCADE）
  const [result] = await writePool.query(
    `DELETE FROM \`${REQUIREMENT_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  return result.affectedRows > 0;
}

/**
 * 添加文档
 */
export async function addDocument(requirementId, fileData) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `INSERT INTO \`${DOCUMENT_TABLE}\` (requirement_id, name, path, size, mime_type, description, uploaded_by)
     VALUES (?, ?, ?, ?, ?, ?, ?)`,
    [
      Number(requirementId),
      fileData.name,
      fileData.path,
      fileData.size || 0,
      toNullableString(fileData.mimeType),
      toNullableString(fileData.description),
      toNullableString(fileData.uploadedBy)
    ]
  );

  return {
    id: Number(result.insertId),
    requirementId: Number(requirementId),
    name: fileData.name,
    path: fileData.path,
    size: fileData.size || 0,
    mimeType: fileData.mimeType,
    description: fileData.description,
    uploadedBy: fileData.uploadedBy
  };
}

/**
 * 删除文档
 */
export async function deleteDocument(requirementId, documentId) {
  await ensureDbReady();

  // 获取文档信息
  const [rows] = await writePool.query(
    `SELECT \`path\` FROM \`${DOCUMENT_TABLE}\` WHERE \`id\` = ? AND \`requirement_id\` = ?`,
    [Number(documentId), Number(requirementId)]
  );

  if (rows.length === 0) {
    return false;
  }

  // 删除文件
  try {
    if (fs.existsSync(rows[0].path)) {
      fs.unlinkSync(rows[0].path);
    }
  } catch (err) {
    console.warn(`[RequirementStorage] Failed to delete file ${rows[0].path}:`, err.message);
  }

  // 删除数据库记录
  const [result] = await writePool.query(
    `DELETE FROM \`${DOCUMENT_TABLE}\` WHERE \`id\` = ? AND \`requirement_id\` = ?`,
    [Number(documentId), Number(requirementId)]
  );

  return result.affectedRows > 0;
}

/**
 * 获取需求统计
 */
export async function getRequirementStatistics() {
  await ensureDbReady();

  const [statusRows] = await getReadPool().query(`
    SELECT \`status\`, COUNT(*) AS count FROM \`${REQUIREMENT_TABLE}\` GROUP BY \`status\`
  `);

  const [priorityRows] = await getReadPool().query(`
    SELECT \`priority\`, COUNT(*) AS count FROM \`${REQUIREMENT_TABLE}\` GROUP BY \`priority\`
  `);

  const [totalRow] = await getReadPool().query(`
    SELECT COUNT(*) AS total FROM \`${REQUIREMENT_TABLE}\`
  `);

  const stats = {
    total: Number(totalRow[0]?.total || 0),
    draft: 0,
    active: 0,
    completed: 0,
    archived: 0,
    highPriority: 0,
    mediumPriority: 0,
    lowPriority: 0
  };

  statusRows.forEach(row => {
    if (row.status === 'draft') stats.draft = Number(row.count);
    if (row.status === 'active') stats.active = Number(row.count);
    if (row.status === 'completed') stats.completed = Number(row.count);
    if (row.status === 'archived') stats.archived = Number(row.count);
  });

  priorityRows.forEach(row => {
    if (row.priority === 'high') stats.highPriority = Number(row.count);
    if (row.priority === 'medium') stats.mediumPriority = Number(row.count);
    if (row.priority === 'low') stats.lowPriority = Number(row.count);
  });

  return stats;
}

/**
 * 关闭数据库连接
 */
export async function closeRequirementStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}
