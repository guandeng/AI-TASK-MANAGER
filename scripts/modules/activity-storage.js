/**
 * activity-storage.js
 * 活动日志数据存储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const ACTIVITY_TABLE = `${TABLE_PREFIX}activity_log`;
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

    // 创建活动日志表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${ACTIVITY_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
        \`task_id\` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
        \`subtask_id\` BIGINT UNSIGNED DEFAULT NULL COMMENT '子任务ID(可选)',
        \`member_id\` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作人ID',
        \`action\` VARCHAR(50) NOT NULL COMMENT '操作类型',
        \`field_name\` VARCHAR(50) DEFAULT NULL COMMENT '变更字段名',
        \`old_value\` TEXT DEFAULT NULL COMMENT '旧值',
        \`new_value\` TEXT DEFAULT NULL COMMENT '新值',
        \`description\` VARCHAR(500) DEFAULT NULL COMMENT '操作描述',
        \`metadata\` TEXT DEFAULT NULL COMMENT '额外元数据(JSON)',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        PRIMARY KEY (\`id\`),
        INDEX \`idx_task_id\` (\`task_id\`),
        INDEX \`idx_subtask_id\` (\`subtask_id\`),
        INDEX \`idx_member_id\` (\`member_id\`),
        INDEX \`idx_action\` (\`action\`),
        INDEX \`idx_created_at\` (\`created_at\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    console.log('[ActivityStorage] Activity log table initialized');
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
 * 格式化活动日志
 */
async function formatActivity(row) {
  let member = null;
  if (row.member_id) {
    const [memberRows] = await getReadPool().query(
      `SELECT id, name, email, avatar FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
      [row.member_id]
    );
    if (memberRows[0]) {
      member = {
        id: Number(memberRows[0].id),
        name: memberRows[0].name,
        email: memberRows[0].email,
        avatar: memberRows[0].avatar
      };
    }
  }

  return {
    id: Number(row.id),
    taskId: Number(row.task_id),
    subtaskId: row.subtask_id ? Number(row.subtask_id) : undefined,
    memberId: row.member_id ? Number(row.member_id) : undefined,
    member,
    action: row.action,
    fieldName: row.field_name || undefined,
    oldValue: row.old_value || undefined,
    newValue: row.new_value || undefined,
    description: row.description || undefined,
    metadata: row.metadata ? JSON.parse(row.metadata) : undefined,
    createdAt: row.created_at.toISOString()
  };
}

/**
 * 记录活动日志
 */
export async function logActivity(data) {
  await ensureDbReady();

  await writePool.query(
    `INSERT INTO \`${ACTIVITY_TABLE}\` (task_id, subtask_id, member_id, action, field_name, old_value, new_value, description, metadata)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      Number(data.taskId),
      data.subtaskId ? Number(data.subtaskId) : null,
      data.memberId ? Number(data.memberId) : null,
      data.action,
      data.fieldName || null,
      data.oldValue || null,
      data.newValue || null,
      data.description || null,
      data.metadata ? JSON.stringify(data.metadata) : null
    ]
  );
}

/**
 * 批量记录活动日志
 */
export async function logActivities(activities) {
  await ensureDbReady();

  for (const activity of activities) {
    await logActivity(activity);
  }
}

/**
 * 获取任务的活动日志
 */
export async function getTaskActivities(taskId, options = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${ACTIVITY_TABLE}\` WHERE \`task_id\` = ?`;
  const params = [Number(taskId)];

  if (options.subtaskId !== undefined) {
    if (options.subtaskId === null) {
      sql += ` AND \`subtask_id\` IS NULL`;
    } else {
      sql += ` AND \`subtask_id\` = ?`;
      params.push(Number(options.subtaskId));
    }
  }

  if (options.action) {
    sql += ` AND \`action\` = ?`;
    params.push(options.action);
  }

  sql += ` ORDER BY \`created_at\` DESC`;

  if (options.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(options.limit));
  }

  const [rows] = await getReadPool().query(sql, params);
  return Promise.all(rows.map(formatActivity));
}

/**
 * 获取全局活动日志
 */
export async function getGlobalActivities(options = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${ACTIVITY_TABLE}\` WHERE 1=1`;
  const params = [];

  if (options.memberId) {
    sql += ` AND \`member_id\` = ?`;
    params.push(Number(options.memberId));
  }

  if (options.action) {
    sql += ` AND \`action\` = ?`;
    params.push(options.action);
  }

  if (options.startDate) {
    sql += ` AND \`created_at\` >= ?`;
    params.push(options.startDate);
  }

  if (options.endDate) {
    sql += ` AND \`created_at\` <= ?`;
    params.push(options.endDate);
  }

  sql += ` ORDER BY \`created_at\` DESC`;

  const limit = options.limit || 50;
  sql += ` LIMIT ?`;
  params.push(Number(limit));

  const offset = options.offset || 0;
  if (offset > 0) {
    sql += ` OFFSET ?`;
    params.push(Number(offset));
  }

  const [rows] = await getReadPool().query(sql, params);
  return Promise.all(rows.map(formatActivity));
}

/**
 * 获取活动统计
 */
export async function getActivityStatistics(options = {}) {
  await ensureDbReady();

  const startDate = options.startDate || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  const endDate = options.endDate || new Date().toISOString().split('T')[0];

  // 按操作类型统计
  const [actionStats] = await getReadPool().query(`
    SELECT action, COUNT(*) as count
    FROM \`${ACTIVITY_TABLE}\`
    WHERE \`created_at\` BETWEEN ? AND ?
    GROUP BY action
    ORDER BY count DESC
  `, [startDate, endDate]);

  // 按日期统计
  const [dailyStats] = await getReadPool().query(`
    SELECT DATE(\`created_at\`) as date, COUNT(*) as count
    FROM \`${ACTIVITY_TABLE}\`
    WHERE \`created_at\` BETWEEN ? AND ?
    GROUP BY DATE(\`created_at\`)
    ORDER BY date DESC
  `, [startDate, endDate]);

  // 按成员统计
  const [memberStats] = await getReadPool().query(`
    SELECT member_id, COUNT(*) as count
    FROM \`${ACTIVITY_TABLE}\`
    WHERE \`created_at\` BETWEEN ? AND ? AND \`member_id\` IS NOT NULL
    GROUP BY member_id
    ORDER BY count DESC
    LIMIT 10
  `, [startDate, endDate]);

  // 获取成员信息
  const memberIds = memberStats.map(s => s.member_id).filter(Boolean);
  const memberMap = new Map();

  if (memberIds.length > 0) {
    const placeholders = memberIds.map(() => '?').join(',');
    const [members] = await getReadPool().query(
      `SELECT id, name FROM \`${MEMBER_TABLE}\` WHERE \`id\` IN (${placeholders})`,
      memberIds
    );
    members.forEach(m => memberMap.set(Number(m.id), m.name));
  }

  return {
    byAction: actionStats.map(s => ({
      action: s.action,
      count: Number(s.count)
    })),
    byDate: dailyStats.map(s => ({
      date: s.date,
      count: Number(s.count)
    })),
    byMember: memberStats.map(s => ({
      memberId: Number(s.member_id),
      memberName: memberMap.get(Number(s.member_id)) || 'Unknown',
      count: Number(s.count)
    })),
    period: { startDate, endDate }
  };
}

/**
 * 清理旧日志
 */
export async function cleanOldActivities(daysToKeep = 90) {
  await ensureDbReady();

  const cutoffDate = new Date(Date.now() - daysToKeep * 24 * 60 * 60 * 1000);

  const [result] = await writePool.query(
    `DELETE FROM \`${ACTIVITY_TABLE}\` WHERE \`created_at\` < ?`,
    [cutoffDate]
  );

  return result.affectedRows;
}

/**
 * 关闭数据库连接
 */
export async function closeActivityStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export default {
  logActivity,
  logActivities,
  getTaskActivities,
  getGlobalActivities,
  getActivityStatistics,
  cleanOldActivities,
  closeActivityStorage
};
