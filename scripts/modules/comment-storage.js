/**
 * comment-storage.js
 * 评论数据存储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const COMMENT_TABLE = `${TABLE_PREFIX}comment`;
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

    // 创建评论表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${COMMENT_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
        \`task_id\` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
        \`subtask_id\` BIGINT UNSIGNED DEFAULT NULL COMMENT '子任务ID(可选)',
        \`member_id\` BIGINT UNSIGNED NOT NULL COMMENT '评论人ID',
        \`parent_id\` BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论ID(用于回复)',
        \`content\` TEXT NOT NULL COMMENT '评论内容',
        \`mentions\` VARCHAR(500) DEFAULT NULL COMMENT '@提及的成员ID列表(JSON数组)',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        PRIMARY KEY (\`id\`),
        INDEX \`idx_task_id\` (\`task_id\`),
        INDEX \`idx_subtask_id\` (\`subtask_id\`),
        INDEX \`idx_member_id\` (\`member_id\`),
        INDEX \`idx_parent_id\` (\`parent_id\`),
        INDEX \`idx_created_at\` (\`created_at\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    console.log('[CommentStorage] Comment table initialized');
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
 * 格式化评论数据
 */
async function formatComment(row) {
  // 获取评论人信息
  const [memberRows] = await getReadPool().query(
    `SELECT id, name, email, avatar FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
    [row.member_id]
  );

  const member = memberRows[0] ? {
    id: Number(memberRows[0].id),
    name: memberRows[0].name,
    email: memberRows[0].email,
    avatar: memberRows[0].avatar
  } : null;

  // 获取回复数量
  const [replyRows] = await getReadPool().query(
    `SELECT COUNT(*) as count FROM \`${COMMENT_TABLE}\` WHERE \`parent_id\` = ?`,
    [row.id]
  );
  const replyCount = replyRows[0] ? Number(replyRows[0].count) : 0;

  return {
    id: Number(row.id),
    taskId: Number(row.task_id),
    subtaskId: row.subtask_id ? Number(row.subtask_id) : undefined,
    memberId: Number(row.member_id),
    member,
    parentId: row.parent_id ? Number(row.parent_id) : undefined,
    content: row.content,
    mentions: row.mentions ? JSON.parse(row.mentions) : [],
    replyCount,
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  };
}

/**
 * 获取任务的评论列表
 */
export async function getTaskComments(taskId, options = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${COMMENT_TABLE}\` WHERE \`task_id\` = ?`;
  const params = [Number(taskId)];

  if (options.subtaskId !== undefined) {
    if (options.subtaskId === null) {
      sql += ` AND \`subtask_id\` IS NULL`;
    } else {
      sql += ` AND \`subtask_id\` = ?`;
      params.push(Number(options.subtaskId));
    }
  }

  sql += ` ORDER BY \`created_at\` DESC`;

  if (options.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(options.limit));
  }

  const [rows] = await getReadPool().query(sql, params);
  return Promise.all(rows.map(formatComment));
}

/**
 * 获取评论详情
 */
export async function getCommentById(id) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${COMMENT_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  if (rows.length === 0) {
    return null;
  }

  return formatComment(rows[0]);
}

/**
 * 创建评论
 */
export async function createComment(data) {
  await ensureDbReady();

  if (!data.content || !data.content.trim()) {
    throw new Error('评论内容不能为空');
  }

  if (!data.taskId) {
    throw new Error('任务ID不能为空');
  }

  if (!data.memberId) {
    throw new Error('评论人ID不能为空');
  }

  const [result] = await writePool.query(
    `INSERT INTO \`${COMMENT_TABLE}\` (task_id, subtask_id, member_id, parent_id, content, mentions)
     VALUES (?, ?, ?, ?, ?, ?)`,
    [
      Number(data.taskId),
      data.subtaskId ? Number(data.subtaskId) : null,
      Number(data.memberId),
      data.parentId ? Number(data.parentId) : null,
      data.content.trim(),
      data.mentions ? JSON.stringify(data.mentions) : null
    ]
  );

  return getCommentById(result.insertId);
}

/**
 * 更新评论
 */
export async function updateComment(id, data) {
  await ensureDbReady();

  const existing = await getCommentById(id);
  if (!existing) {
    throw new Error('评论不存在');
  }

  const fields = [];
  const params = [];

  if (data.content !== undefined) {
    if (!data.content || !data.content.trim()) {
      throw new Error('评论内容不能为空');
    }
    fields.push('`content` = ?');
    params.push(data.content.trim());
  }

  if (data.mentions !== undefined) {
    fields.push('`mentions` = ?');
    params.push(data.mentions ? JSON.stringify(data.mentions) : null);
  }

  if (fields.length === 0) {
    return existing;
  }

  params.push(Number(id));

  await writePool.query(
    `UPDATE \`${COMMENT_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
    params
  );

  return getCommentById(id);
}

/**
 * 删除评论
 */
export async function deleteComment(id) {
  await ensureDbReady();

  // 先删除子评论
  await writePool.query(
    `DELETE FROM \`${COMMENT_TABLE}\` WHERE \`parent_id\` = ?`,
    [Number(id)]
  );

  // 删除评论本身
  const [result] = await writePool.query(
    `DELETE FROM \`${COMMENT_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  return result.affectedRows > 0;
}

/**
 * 获取评论的回复列表
 */
export async function getCommentReplies(parentId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${COMMENT_TABLE}\` WHERE \`parent_id\` = ? ORDER BY \`created_at\` ASC`,
    [Number(parentId)]
  );

  return Promise.all(rows.map(formatComment));
}

/**
 * 获取评论统计
 */
export async function getCommentStatistics(taskId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(`
    SELECT COUNT(*) as total,
           COUNT(DISTINCT member_id) as unique_authors
    FROM \`${COMMENT_TABLE}\`
    WHERE \`task_id\` = ? AND \`subtask_id\` IS NULL
  `, [Number(taskId)]);

  const stats = rows[0] || { total: 0, unique_authors: 0 };

  return {
    total: Number(stats.total),
    uniqueAuthors: Number(stats.unique_authors)
  };
}

/**
 * 关闭数据库连接
 */
export async function closeCommentStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export default {
  getTaskComments,
  getCommentById,
  createComment,
  updateComment,
  deleteComment,
  getCommentReplies,
  getCommentStatistics,
  closeCommentStorage
};
