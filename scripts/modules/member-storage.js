/**
 * member-storage.js
 * 团队成员数据存储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
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

/**
 * 共享数据库连接池（与其他 storage 模块共用）
 */
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

    // 创建成员表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${MEMBER_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '成员ID',
        \`name\` VARCHAR(100) NOT NULL COMMENT '姓名',
        \`email\` VARCHAR(255) DEFAULT NULL COMMENT '邮箱',
        \`avatar\` VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
        \`role\` VARCHAR(50) NOT NULL DEFAULT 'member' COMMENT '角色: admin-管理员, leader-组长, member-成员',
        \`department\` VARCHAR(100) DEFAULT NULL COMMENT '部门',
        \`skills\` VARCHAR(500) DEFAULT NULL COMMENT '技能标签(JSON数组)',
        \`status\` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active-活跃, inactive-停用',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        PRIMARY KEY (\`id\`),
        UNIQUE KEY \`uk_email\` (\`email\`),
        INDEX \`idx_role\` (\`role\`),
        INDEX \`idx_department\` (\`department\`),
        INDEX \`idx_status\` (\`status\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    console.log('[MemberStorage] Member table initialized');
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
 * 格式化成员数据
 */
function formatMember(row) {
  return {
    id: Number(row.id),
    name: row.name,
    email: row.email || undefined,
    avatar: row.avatar || undefined,
    role: row.role,
    department: row.department || undefined,
    skills: row.skills ? JSON.parse(row.skills) : [],
    status: row.status,
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  };
}

/**
 * 获取成员列表
 * @param {Object} filters - 筛选条件
 * @param {string} filters.status - 状态筛选
 * @param {string} filters.role - 角色筛选
 * @param {string} filters.department - 部门筛选
 * @param {string} filters.keyword - 关键词搜索
 * @param {number} filters.limit - 限制数量
 * @param {number} filters.offset - 偏移量
 */
export async function getMemberList(filters = {}) {
  await ensureDbReady();

  let sql = `SELECT * FROM \`${MEMBER_TABLE}\` WHERE 1=1`;
  const params = [];

  if (filters.status) {
    sql += ` AND \`status\` = ?`;
    params.push(filters.status);
  }

  if (filters.role) {
    sql += ` AND \`role\` = ?`;
    params.push(filters.role);
  }

  if (filters.department) {
    sql += ` AND \`department\` = ?`;
    params.push(filters.department);
  }

  if (filters.keyword) {
    sql += ` AND (\`name\` LIKE ? OR \`email\` LIKE ?)`;
    params.push(`%${filters.keyword}%`, `%${filters.keyword}%`);
  }

  sql += ` ORDER BY \`created_at\` DESC`;

  if (filters.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(filters.limit));
    if (filters.offset) {
      sql += ` OFFSET ?`;
      params.push(Number(filters.offset));
    }
  }

  const [rows] = await getReadPool().query(sql, params);
  return rows.map(formatMember);
}

/**
 * 获取成员总数
 */
export async function getMemberCount(filters = {}) {
  await ensureDbReady();

  let sql = `SELECT COUNT(*) AS total FROM \`${MEMBER_TABLE}\` WHERE 1=1`;
  const params = [];

  if (filters.status) {
    sql += ` AND \`status\` = ?`;
    params.push(filters.status);
  }

  if (filters.role) {
    sql += ` AND \`role\` = ?`;
    params.push(filters.role);
  }

  if (filters.department) {
    sql += ` AND \`department\` = ?`;
    params.push(filters.department);
  }

  if (filters.keyword) {
    sql += ` AND (\`name\` LIKE ? OR \`email\` LIKE ?)`;
    params.push(`%${filters.keyword}%`, `%${filters.keyword}%`);
  }

  const [rows] = await getReadPool().query(sql, params);
  return Number(rows[0]?.total || 0);
}

/**
 * 获取成员详情
 * @param {number} id - 成员ID
 */
export async function getMemberById(id) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  if (rows.length === 0) {
    return null;
  }

  return formatMember(rows[0]);
}

/**
 * 通过邮箱获取成员
 * @param {string} email - 成员邮箱
 */
export async function getMemberByEmail(email) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MEMBER_TABLE}\` WHERE \`email\` = ?`,
    [email]
  );

  if (rows.length === 0) {
    return null;
  }

  return formatMember(rows[0]);
}

/**
 * 创建成员
 * @param {Object} data - 成员数据
 */
export async function createMember(data) {
  await ensureDbReady();

  if (!data.name || !data.name.trim()) {
    throw new Error('成员姓名不能为空');
  }

  // 检查邮箱唯一性
  if (data.email) {
    const existing = await getMemberByEmail(data.email);
    if (existing) {
      throw new Error('该邮箱已被使用');
    }
  }

  const [result] = await writePool.query(
    `INSERT INTO \`${MEMBER_TABLE}\` (name, email, avatar, role, department, skills, status)
     VALUES (?, ?, ?, ?, ?, ?, ?)`,
    [
      data.name.trim(),
      toNullableString(data.email),
      toNullableString(data.avatar),
      data.role || 'member',
      toNullableString(data.department),
      data.skills ? JSON.stringify(data.skills) : null,
      data.status || 'active'
    ]
  );

  return getMemberById(result.insertId);
}

/**
 * 更新成员
 * @param {number} id - 成员ID
 * @param {Object} data - 更新数据
 */
export async function updateMember(id, data) {
  await ensureDbReady();

  const existing = await getMemberById(id);
  if (!existing) {
    throw new Error('成员不存在');
  }

  // 检查邮箱唯一性
  if (data.email && data.email !== existing.email) {
    const emailOwner = await getMemberByEmail(data.email);
    if (emailOwner && emailOwner.id !== id) {
      throw new Error('该邮箱已被其他成员使用');
    }
  }

  const fields = [];
  const params = [];

  if (data.name !== undefined) {
    if (!data.name || !data.name.trim()) {
      throw new Error('成员姓名不能为空');
    }
    fields.push('`name` = ?');
    params.push(data.name.trim());
  }

  if (data.email !== undefined) {
    fields.push('`email` = ?');
    params.push(toNullableString(data.email));
  }

  if (data.avatar !== undefined) {
    fields.push('`avatar` = ?');
    params.push(toNullableString(data.avatar));
  }

  if (data.role !== undefined) {
    fields.push('`role` = ?');
    params.push(data.role);
  }

  if (data.department !== undefined) {
    fields.push('`department` = ?');
    params.push(toNullableString(data.department));
  }

  if (data.skills !== undefined) {
    fields.push('`skills` = ?');
    params.push(data.skills ? JSON.stringify(data.skills) : null);
  }

  if (data.status !== undefined) {
    fields.push('`status` = ?');
    params.push(data.status);
  }

  if (fields.length === 0) {
    return existing;
  }

  params.push(Number(id));

  await writePool.query(
    `UPDATE \`${MEMBER_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
    params
  );

  return getMemberById(id);
}

/**
 * 删除员
 * @param {number} id - 成员ID
 */
export async function deleteMember(id) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `DELETE FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
    [Number(id)]
  );

  return result.affectedRows > 0;
}

/**
 * 获取成员统计
 */
export async function getMemberStatistics() {
  await ensureDbReady();

  const [statusRows] = await getReadPool().query(`
    SELECT \`status\`, COUNT(*) AS count FROM \`${MEMBER_TABLE}\` GROUP BY \`status\`
  `);

  const [roleRows] = await getReadPool().query(`
    SELECT \`role\`, COUNT(*) AS count FROM \`${MEMBER_TABLE}\` GROUP BY \`role\`
  `);

  const [departmentRows] = await getReadPool().query(`
    SELECT \`department\`, COUNT(*) AS count FROM \`${MEMBER_TABLE}\` WHERE \`department\` IS NOT NULL GROUP BY \`department\`
  `);

  const [totalRow] = await getReadPool().query(`
    SELECT COUNT(*) AS total FROM \`${MEMBER_TABLE}\`
  `);

  const stats = {
    total: Number(totalRow[0]?.total || 0),
    active: 0,
    inactive: 0,
    admin: 0,
    leader: 0,
    member: 0,
    departments: {}
  };

  statusRows.forEach(row => {
    if (row.status === 'active') stats.active = Number(row.count);
    if (row.status === 'inactive') stats.inactive = Number(row.count);
  });

  roleRows.forEach(row => {
    if (row.role === 'admin') stats.admin = Number(row.count);
    if (row.role === 'leader') stats.leader = Number(row.count);
    if (row.role === 'member') stats.member = Number(row.count);
  });

  departmentRows.forEach(row => {
    if (row.department) {
      stats.departments[row.department] = Number(row.count);
    }
  });

  return stats;
}

/**
 * 获取所有部门列表
 */
export async function getDepartmentList() {
  await ensureDbReady();

  const [rows] = await getReadPool().query(`
    SELECT DISTINCT \`department\` FROM \`${MEMBER_TABLE}\` WHERE \`department\` IS NOT NULL ORDER BY \`department\`
  `);

  return rows.map(row => row.department);
}

/**
 * 批量获取成员
 * @param {number[]} ids - 成员ID列表
 */
export async function getMembersByIds(ids) {
  if (!ids || ids.length === 0) {
    return [];
  }

  await ensureDbReady();

  const placeholders = ids.map(() => '?').join(',');
  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${MEMBER_TABLE}\` WHERE \`id\` IN (${placeholders})`,
    ids.map(Number)
  );

  return rows.map(formatMember);
}

/**
 * 关闭数据库连接
 */
export async function closeMemberStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export default {
  getMemberList,
  getMemberCount,
  getMemberById,
  getMemberByEmail,
  createMember,
  updateMember,
  deleteMember,
  getMemberStatistics,
  getDepartmentList,
  getMembersByIds,
  closeMemberStorage
};
