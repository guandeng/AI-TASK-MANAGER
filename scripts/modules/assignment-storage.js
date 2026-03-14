/**
 * assignment-storage.js
 * 任务分配数据储模块
 */

import mysql from 'mysql2/promise';

const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const ASSIGNMENT_TABLE = `${TABLE_PREFIX}assignment`;
const SUBTASK_ASSIGNMENT_TABLE = `${TABLE_PREFIX}subtask_assignment`;
const TASK_TABLE = `${TABLE_PREFIX}task`;
const SUBTASK_TABLE = `${TABLE_PREFIX}subtask`;
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

    // 创建任务分配表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${ASSIGNMENT_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分配ID',
        \`task_id\` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
        \`member_id\` BIGINT UNSIGNED NOT NULL COMMENT '成员ID',
        \`role\` VARCHAR(50) NOT NULL DEFAULT 'assignee' COMMENT '角色: assignee-负责人, reviewer-审核人, collaborator-协作者',
        \`assigned_by\` BIGINT UNSIGNED DEFAULT NULL COMMENT '分配人ID',
        \`estimated_hours\` DECIMAL(10, 2) DEFAULT NULL COMMENT '预估工时(小时)',
        \`actual_hours\` DECIMAL(10, 2) DEFAULT NULL COMMENT '实际工时(小时)',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        PRIMARY KEY (\`id\`),
        UNIQUE KEY \`uk_task_member\` (\`task_id\`, \`member_id\`),
        INDEX \`idx_member_id\` (\`member_id\`),
        INDEX \`idx_role\` (\`role\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 创建子任务分配表
    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${SUBTASK_ASSIGNMENT_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分配ID',
        \`subtask_id\` BIGINT UNSIGNED NOT NULL COMMENT '子任务ID',
        \`member_id\` BIGINT UNSIGNED NOT NULL COMMENT '成员ID',
        \`role\` VARCHAR(50) NOT NULL DEFAULT 'assignee' COMMENT '角色',
        \`assigned_by\` BIGINT UNSIGNED DEFAULT NULL COMMENT '分配人ID',
        \`estimated_hours\` DECIMAL(10, 2) DEFAULT NULL COMMENT '预估工时',
        \`actual_hours\` DECIMAL(10, 2) DEFAULT NULL COMMENT '实际工时',
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        PRIMARY KEY (\`id\`),
        UNIQUE KEY \`uk_subtask_member\` (\`subtask_id\`, \`member_id\`),
        INDEX \`idx_member_id\` (\`member_id\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    // 扩展任务表添加时间字段
    await addTaskTimeFields();

    console.log('[AssignmentStorage] Assignment tables initialized');
  })();

  await initPromise;
}

/**
 * 扩展任务表添加时间字段
 */
async function addTaskTimeFields() {
  const fieldsToAdd = [
    { name: 'start_date', def: '`start_date` DATE DEFAULT NULL COMMENT \'开始日期\'' },
    { name: 'due_date', def: '`due_date` DATE DEFAULT NULL COMMENT \'截止日期\'' },
    { name: 'completed_at', def: '`completed_at` TIMESTAMP NULL DEFAULT NULL COMMENT \'完成时间\'' },
    { name: 'estimated_hours', def: '`estimated_hours` DECIMAL(10, 2) DEFAULT NULL COMMENT \'预估工时\'' },
    { name: 'actual_hours', def: '`actual_hours` DECIMAL(10, 2) DEFAULT NULL COMMENT \'实际工时\'' }
  ];

  for (const field of fieldsToAdd) {
    try {
      const [columns] = await writePool.query(`
        SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
      `, [process.env.DB_DATABASE, TASK_TABLE, field.name]);

      if (columns.length === 0) {
        await writePool.query(`ALTER TABLE \`${TASK_TABLE}\` ADD COLUMN ${field.def}`);
        console.log(`[AssignmentStorage] Added ${field.name} column to task table`);
      }
    } catch (err) {
      console.warn(`[AssignmentStorage] Could not add ${field.name} to task table:`, err.message);
    }
  }
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
 * 格式化分配数据
 */
async function formatAssignment(row) {
  // 获取成员信息
  const [memberRows] = await getReadPool().query(
    `SELECT id, name, email, avatar, role FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
    [row.member_id]
  );

  const member = memberRows[0] ? {
    id: Number(memberRows[0].id),
    name: memberRows[0].name,
    email: memberRows[0].email,
    avatar: memberRows[0].avatar,
    role: memberRows[0].role
  } : null;

  return {
    id: Number(row.id),
    taskId: Number(row.task_id),
    memberId: Number(row.member_id),
    member,
    role: row.role,
    assignedBy: row.assigned_by ? Number(row.assigned_by) : undefined,
    estimatedHours: row.estimated_hours ? Number(row.estimated_hours) : undefined,
    actualHours: row.actual_hours ? Number(row.actual_hours) : undefined,
    createdAt: row.created_at.toISOString(),
    updatedAt: row.updated_at.toISOString()
  };
}

/**
 * 获取任务的所有分配
 */
export async function getTaskAssignments(taskId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${ASSIGNMENT_TABLE}\` WHERE \`task_id\` = ? ORDER BY \`role\`, \`created_at\``,
    [Number(taskId)]
  );

  return Promise.all(rows.map(formatAssignment));
}

/**
 * 获取子任务的所有分配
 */
export async function getSubtaskAssignments(subtaskId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${SUBTASK_ASSIGNMENT_TABLE}\` WHERE \`subtask_id\` = ? ORDER BY \`role\`, \`created_at\``,
    [Number(subtaskId)]
  );

  const results = [];
  for (const row of rows) {
    const [memberRows] = await getReadPool().query(
      `SELECT id, name, email, avatar, role FROM \`${MEMBER_TABLE}\` WHERE \`id\` = ?`,
      [row.member_id]
    );

    const member = memberRows[0] ? {
      id: Number(memberRows[0].id),
      name: memberRows[0].name,
      email: memberRows[0].email,
      avatar: memberRows[0].avatar,
      role: memberRows[0].role
    } : null;

    results.push({
      id: Number(row.id),
      subtaskId: Number(row.subtask_id),
      memberId: Number(row.member_id),
      member,
      role: row.role,
      assignedBy: row.assigned_by ? Number(row.assigned_by) : undefined,
      estimatedHours: row.estimated_hours ? Number(row.estimated_hours) : undefined,
      actualHours: row.actual_hours ? Number(row.actual_hours) : undefined,
      createdAt: row.created_at.toISOString(),
      updatedAt: row.updated_at.toISOString()
    });
  }

  return results;
}

/**
 * 分配任务给成员
 */
export async function assignTaskToMember(taskId, memberId, data = {}) {
  await ensureDbReady();

  // 验证角色
  const validRoles = ['assignee', 'reviewer', 'collaborator'];
  const role = data.role || 'assignee';
  if (!validRoles.includes(role)) {
    throw new Error(`无效的角色: ${role}`);
  }

  // 检查是否已存在
  const [existing] = await writePool.query(
    `SELECT id FROM \`${ASSIGNMENT_TABLE}\` WHERE \`task_id\` = ? AND \`member_id\` = ?`,
    [Number(taskId), Number(memberId)]
  );

  if (existing.length > 0) {
    // 更新现有分配
    await writePool.query(
      `UPDATE \`${ASSIGNMENT_TABLE}\` SET \`role\` = ?, \`assigned_by\` = ?, \`estimated_hours\` = ?, \`actual_hours\` = ? WHERE \`id\` = ?`,
      [role, data.assignedBy || null, data.estimatedHours || null, data.actualHours || null, existing[0].id]
    );
  } else {
    // 创建新分配
    await writePool.query(
      `INSERT INTO \`${ASSIGNMENT_TABLE}\` (task_id, member_id, role, assigned_by, estimated_hours, actual_hours)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [Number(taskId), Number(memberId), role, data.assignedBy || null, data.estimatedHours || null, data.actualHours || null]
    );
  }

  return getTaskAssignments(taskId);
}

/**
 * 分配子任务给成员
 */
export async function assignSubtaskToMember(subtaskId, memberId, data = {}) {
  await ensureDbReady();

  const role = data.role || 'assignee';

  // 检查是否已存在
  const [existing] = await writePool.query(
    `SELECT id FROM \`${SUBTASK_ASSIGNMENT_TABLE}\` WHERE \`subtask_id\` = ? AND \`member_id\` = ?`,
    [Number(subtaskId), Number(memberId)]
  );

  if (existing.length > 0) {
    await writePool.query(
      `UPDATE \`${SUBTASK_ASSIGNMENT_TABLE}\` SET \`role\` = ?, \`assigned_by\` = ?, \`estimated_hours\` = ?, \`actual_hours\` = ? WHERE \`id\` = ?`,
      [role, data.assignedBy || null, data.estimatedHours || null, data.actualHours || null, existing[0].id]
    );
  } else {
    await writePool.query(
      `INSERT INTO \`${SUBTASK_ASSIGNMENT_TABLE}\` (subtask_id, member_id, role, assigned_by, estimated_hours, actual_hours)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [Number(subtaskId), Number(memberId), role, data.assignedBy || null, data.estimatedHours || null, data.actualHours || null]
    );
  }

  return getSubtaskAssignments(subtaskId);
}

/**
 * 移除任务分配
 */
export async function removeTaskAssignment(taskId, assignmentId) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `DELETE FROM \`${ASSIGNMENT_TABLE}\` WHERE \`id\` = ? AND \`task_id\` = ?`,
    [Number(assignmentId), Number(taskId)]
  );

  return result.affectedRows > 0;
}

/**
 * 移除子任务分配
 */
export async function removeSubtaskAssignment(subtaskId, assignmentId) {
  await ensureDbReady();

  const [result] = await writePool.query(
    `DELETE FROM \`${SUBTASK_ASSIGNMENT_TABLE}\` WHERE \`id\` = ? AND \`subtask_id\` = ?`,
    [Number(assignmentId), Number(subtaskId)]
  );

  return result.affectedRows > 0;
}

/**
 * 获取成员的所有任务分配
 */
export async function getMemberAssignments(memberId, filters = {}) {
  await ensureDbReady();

  let sql = `SELECT a.*, t.title, t.title_trans, t.status as task_status, t.priority
             FROM \`${ASSIGNMENT_TABLE}\` a
             LEFT JOIN \`${TASK_TABLE}\` t ON a.task_id = t.id
             WHERE a.member_id = ?`;
  const params = [Number(memberId)];

  if (filters.role) {
    sql += ` AND a.role = ?`;
    params.push(filters.role);
  }

  if (filters.status) {
    sql += ` AND t.status = ?`;
    params.push(filters.status);
  }

  sql += ` ORDER BY a.created_at DESC`;

  if (filters.limit) {
    sql += ` LIMIT ?`;
    params.push(Number(filters.limit));
  }

  const [rows] = await getReadPool().query(sql, params);

  return rows.map(row => ({
    id: Number(row.id),
    taskId: Number(row.task_id),
    memberId: Number(row.member_id),
    role: row.role,
    taskTitle: row.title_trans || row.title,
    taskStatus: row.task_status,
    taskPriority: row.priority,
    estimatedHours: row.estimated_hours ? Number(row.estimated_hours) : undefined,
    actualHours: row.actual_hours ? Number(row.actual_hours) : undefined,
    createdAt: row.created_at.toISOString()
  }));
}

/**
 * 获取成员工作量统计
 */
export async function getMemberWorkload(memberId) {
  await ensureDbReady();

  // 获取任务分配统计
  const [taskStats] = await getReadPool().query(`
    SELECT
      COUNT(*) as total_tasks,
      SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) as completed_tasks,
      SUM(CASE WHEN t.status = 'in-progress' THEN 1 ELSE 0 END) as in_progress_tasks,
      SUM(CASE WHEN t.status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
      SUM(a.estimated_hours) as total_estimated_hours,
      SUM(a.actual_hours) as total_actual_hours
    FROM \`${ASSIGNMENT_TABLE}\` a
    LEFT JOIN \`${TASK_TABLE}\` t ON a.task_id = t.id
    WHERE a.member_id = ?
  `, [Number(memberId)]);

  // 获取子任务分配统计
  const [subtaskStats] = await getReadPool().query(`
    SELECT COUNT(*) as total_subtasks
    FROM \`${SUBTASK_ASSIGNMENT_TABLE}\`
    WHERE member_id = ?
  `, [Number(memberId)]);

  const stats = taskStats[0] || {};

  return {
    totalTasks: Number(stats.total_tasks || 0),
    completedTasks: Number(stats.completed_tasks || 0),
    inProgressTasks: Number(stats.in_progress_tasks || 0),
    pendingTasks: Number(stats.pending_tasks || 0),
    totalSubtasks: Number(subtaskStats[0]?.total_subtasks || 0),
    totalEstimatedHours: stats.total_estimated_hours ? Number(stats.total_estimated_hours) : 0,
    totalActualHours: stats.total_actual_hours ? Number(stats.total_actual_hours) : 0
  };
}

/**
 * 更新任务时间字段
 */
export async function updateTaskTimeFields(taskId, data) {
  await ensureDbReady();

  const fields = [];
  const params = [];

  if (data.startDate !== undefined) {
    fields.push('`start_date` = ?');
    params.push(data.startDate || null);
  }

  if (data.dueDate !== undefined) {
    fields.push('`due_date` = ?');
    params.push(data.dueDate || null);
  }

  if (data.completedAt !== undefined) {
    fields.push('`completed_at` = ?');
    params.push(data.completedAt || null);
  }

  if (data.estimatedHours !== undefined) {
    fields.push('`estimated_hours` = ?');
    params.push(data.estimatedHours || null);
  }

  if (data.actualHours !== undefined) {
    fields.push('`actual_hours` = ?');
    params.push(data.actualHours || null);
  }

  if (fields.length === 0) {
    return false;
  }

  params.push(Number(taskId));

  const [result] = await writePool.query(
    `UPDATE \`${TASK_TABLE}\` SET ${fields.join(', ')} WHERE \`id\` = ?`,
    params
  );

  return result.affectedRows > 0;
}

/**
 * 关闭数据库连接
 */
export async function closeAssignmentStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export default {
  getTaskAssignments,
  getSubtaskAssignments,
  assignTaskToMember,
  assignSubtaskToMember,
  removeTaskAssignment,
  removeSubtaskAssignment,
  getMemberAssignments,
  getMemberWorkload,
  updateTaskTimeFields,
  closeAssignmentStorage
};
