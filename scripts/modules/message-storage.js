/**
 * message-storage.js
 * 消息数据存储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const MESSAGE_TABLE = `${TABLE_PREFIX}message`;
const TASK_TABLE = `${TABLE_PREFIX}task`;

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

    // 创建消息表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${MESSAGE_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
        \`task_id\` BIGINT UNSIGNED NOT NULL COMMENT '关联的任务ID',
        \`type\` VARCHAR(50) NOT NULL COMMENT '消息类型: expand_task, regenerate_subtask',
        \`status\` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/processing/success/failed',
        \`title\` VARCHAR(255) NOT NULL COMMENT '消息标题',
        \`content\` TEXT DEFAULT NULL COMMENT '消息内容',
        \`error_message\` TEXT DEFAULT NULL COMMENT '错误信息',
        \`result_summary\` VARCHAR(500) DEFAULT NULL COMMENT '结果摘要',
        \`is_read\` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        PRIMARY KEY (\`id\`),
        INDEX \`idx_task_id\` (\`task_id\`),
        INDEX \`idx_type\` (\`type\`),
        INDEX \`idx_status\` (\`status\`),
        INDEX \`idx_is_read\` (\`is_read\`),
        INDEX \`idx_created_at\` (\`created_at\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    console.log('[MessageStorage] Message table initialized');
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

/**
 * 格式化消息对象
 */
function formatMessage(row) {
  return {
    id: row.id,
    taskId: row.task_id,
    type: row.type,
    status: row.status,
    title: row.title,
    content: row.content || undefined,
    errorMessage: row.error_message || undefined,
    resultSummary: row.result_summary || undefined,
    isRead: Boolean(row.is_read),
    createdAt: row.created_at,
    updatedAt: row.updated_at
  };
}

/**
 * 创建消息
 * @param {number} taskId - 任务ID
 * @param {string} type - 消息类型 'expand_task' | 'regenerate_subtask'
 * @param {string} title - 消息标题
 * @returns {Promise<{id: number}>} 创建的消息对象
 */
export async function createMessage(taskId, type, title) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `INSERT INTO \`${MESSAGE_TABLE}\` (task_id, type, status, title) VALUES (?, ?, 'pending', ?)`,
    [taskId, type, title]
  );

  return { id: result.insertId };
}

/**
 * 更新消息状态
 * @param {number} messageId - 消息ID
 * @param {string} status - 状态 'pending' | 'processing' | 'success' | 'failed'
 * @param {object} options - 可选参数
 * @param {string} [options.content] - 消息内容
 * @param {string} [options.errorMessage] - 错误信息
 * @param {string} [options.resultSummary] - 结果摘要
 */
export async function updateMessageStatus(messageId, status, options = {}) {
  await ensureDbReady();

  const { content, errorMessage, resultSummary } = options;

  await writePool.query(
    `UPDATE \`${MESSAGE_TABLE}\` SET status = ?, content = ?, error_message = ?, result_summary = ? WHERE id = ?`,
    [status, content || null, errorMessage || null, resultSummary || null, messageId]
  );
}

/**
 * 获取消息列表
 * @param {object} filters - 筛选条件
 * @param {number} [filters.taskId] - 任务ID
 * @param {string} [filters.status] - 状态
 * @param {boolean} [filters.unreadOnly] - 只获取未读
 * @param {number} [filters.limit] - 限制数量
 * @param {number} [filters.offset] - 偏移量
 * @returns {Promise<{messages: Array, total: number}>}
 */
export async function getMessages(filters = {}) {
  await ensureDbReady();

  const { taskId, status, unreadOnly, limit = 50, offset = 0 } = filters;

  const conditions = [];
  const params = [];

  if (taskId) {
    conditions.push('task_id = ?');
    params.push(taskId);
  }
  if (status) {
    conditions.push('status = ?');
    params.push(status);
  }
  if (unreadOnly) {
    conditions.push('is_read = 0');
  }

  const whereClause = conditions.length > 0 ? `WHERE ${conditions.join(' AND ')}` : '';

  // 获取总数
  const [countRows] = await getReadPool().query(
    `SELECT COUNT(*) as total FROM \`${MESSAGE_TABLE}\` ${whereClause}`,
    params
  );
  const total = countRows[0].total;

  // 获取列表
  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MESSAGE_TABLE}\` ${whereClause} ORDER BY created_at DESC LIMIT ? OFFSET ?`,
    [...params, limit, offset]
  );

  return {
    messages: rows.map(formatMessage),
    total
  };
}

/**
 * 获取单个消息
 * @param {number} messageId - 消息ID
 * @returns {Promise<object|null>}
 */
export async function getMessageById(messageId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MESSAGE_TABLE}\` WHERE id = ?`,
    [messageId]
  );

  return rows.length > 0 ? formatMessage(rows[0]) : null;
}

/**
 * 标记消息为已读
 * @param {number} messageId - 消息ID
 */
export async function markAsRead(messageId) {
  await ensureDbReady();

  await writePool.query(
    `UPDATE \`${MESSAGE_TABLE}\` SET is_read = 1 WHERE id = ?`,
    [messageId]
  );
}

/**
 * 标记所有消息为已读
 */
export async function markAllAsRead() {
  await ensureDbReady();

  await writePool.query(
    `UPDATE \`${MESSAGE_TABLE}\` SET is_read = 1 WHERE is_read = 0`
  );
}

/**
 * 删除消息
 * @param {number} messageId - 消息ID
 */
export async function deleteMessage(messageId) {
  await ensureDbReady();

  await writePool.query(
    `DELETE FROM \`${MESSAGE_TABLE}\` WHERE id = ?`,
    [messageId]
  );
}

/**
 * 获取未读消息数量
 * @returns {Promise<number>}
 */
export async function getUnreadCount() {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT COUNT(*) as count FROM \`${MESSAGE_TABLE}\` WHERE is_read = 0`
  );

  return rows[0].count;
}

/**
 * 获取任务关联的最新消息
 * @param {number} taskId - 任务ID
 * @param {string} type - 消息类型
 * @returns {Promise<object|null>}
 */
export async function getLatestMessageByTask(taskId, type) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MESSAGE_TABLE}\` WHERE task_id = ? AND type = ? ORDER BY created_at DESC LIMIT 1`,
    [taskId, type]
  );

  return rows.length > 0 ? formatMessage(rows[0]) : null;
}
