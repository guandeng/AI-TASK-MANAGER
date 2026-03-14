import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import dotenv from 'dotenv';
import mysql from 'mysql2/promise';
import { exportToEnv } from './config-loader.js';

// 尝试从多个位置加载 .env 文件
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 按优先级尝试加载 .env 文件
const envPaths = [
  path.resolve(process.cwd(), '.env'),           // 当前工目录
  path.resolve(__dirname, '../../.env'),         // scripts/modules/../../.env = 项目根目录
  path.resolve(__dirname, '../../../.env')       // 备选路径
];

let envLoaded = false;
for (const envPath of envPaths) {
  if (fs.existsSync(envPath)) {
    dotenv.config({ path: envPath });
    envLoaded = true;
    break;
  }
}

if (!envLoaded) {
  // 最后尝试默认加载
  dotenv.config();
}

const mergedEnvConfig = exportToEnv();
for (const [key, value] of Object.entries(mergedEnvConfig)) {
  if (value != null) {
    process.env[key] = String(value);
  }
}

const DB_CONNECTION = (process.env.DB_CONNECTION || 'mysql').toLowerCase();
const TABLE_PREFIX = (process.env.DB_TABLE_PREFIX || 'task_').replace(/[^a-zA-Z0-9_]/g, '') || 'task_';
const TASK_TABLE = `${TABLE_PREFIX}task`;
const SUBTASK_TABLE = `${TABLE_PREFIX}subtask`;
const TASK_DEPENDENCY_TABLE = `${TABLE_PREFIX}dependency`;
const SUBTASK_DEPENDENCY_TABLE = `${TABLE_PREFIX}subtask_dependency`;
const META_TABLE = `${TABLE_PREFIX}meta`;

let writePool = null;
let readPools = [];
let initPromise = null;
let readPoolIndex = 0;

function parseSlaveHosts() {
  return (process.env.DB_HOST_SLAVE || '')
    .split(',')
    .map(host => host.trim())
    .filter(Boolean);
}

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

function normalizeStorageKey(tasksPath) {
  return path.resolve(tasksPath || 'tasks/tasks.json');
}

function toNullableString(value) {
  return value == null ? null : String(value);
}

function normalizeTaskDependency(depId) {
  if (typeof depId === 'number') {
    return depId;
  }

  if (typeof depId === 'string' && !depId.includes('.')) {
    const parsed = parseInt(depId, 10);
    return Number.isNaN(parsed) ? null : parsed;
  }

  return null;
}

function normalizeSubtaskDependency(taskId, depId) {
  if (typeof depId === 'number') {
    return `${taskId}.${depId}`;
  }

  if (typeof depId === 'string' && depId.includes('.')) {
    return depId;
  }

  if (typeof depId === 'string') {
    const parsed = parseInt(depId, 10);
    return Number.isNaN(parsed) ? null : `${taskId}.${parsed}`;
  }

  return null;
}

async function ensureDbReady() {
  if (initPromise) {
    await initPromise;
    return;
  }

  initPromise = (async () => {
    if (DB_CONNECTION !== 'mysql') {
      throw new Error(`Unsupported DB_CONNECTION: ${DB_CONNECTION}`);
    }

    if (!process.env.DB_HOST || !process.env.DB_DATABASE || !process.env.DB_USERNAME) {
      console.error(`
╔════════════════════════════════════════════════════════════════╗
║                    数据库未配置                                  ║
╠════════════════════════════════════════════════════════════════╣
║  请在 .env 文件中配置以下数据库连接信息：                          ║
║                                                                ║
║  DB_HOST=localhost         # 数据库主机地址                      ║
║  DB_PORT=3306              # 数据库端口                          ║
║  DB_DATABASE=task_manager  # 数据库名称                          ║
║  DB_USERNAME=root          # 数据库用户名                        ║
║  DB_PASSWORD=password      # 数据库密码                          ║
║                                                                ║
║  可选配置：                                                     ║
║  DB_HOST_SLAVE=slave1,slave2  # 从库地址（读写分离）              ║
║  DB_TABLE_PREFIX=task_        # 表名前缀                         ║
╚════════════════════════════════════════════════════════════════╝
      `);
      throw new Error('数据库未配置，请先在 .env 文件中设置 DB_HOST, DB_DATABASE, DB_USERNAME 等环境变量');
    }

    writePool = mysql.createPool(getDbConfig(process.env.DB_HOST));
    readPools = parseSlaveHosts().map(host => mysql.createPool(getDbConfig(host)));

    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${TASK_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`title\` VARCHAR(500) NOT NULL,
        \`title_trans\` VARCHAR(500) DEFAULT NULL,
        \`description\` TEXT,
        \`description_trans\` TEXT,
        \`status\` VARCHAR(50) NOT NULL DEFAULT 'pending',
        \`priority\` VARCHAR(20) NOT NULL DEFAULT 'medium',
        \`details\` TEXT,
        \`details_trans\` TEXT,
        \`test_strategy\` TEXT,
        \`test_strategy_trans\` TEXT,
        \`assignee\` VARCHAR(100) DEFAULT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_status\` (\`status\`),
        INDEX \`idx_priority\` (\`priority\`)
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    const [requirementIdColumns] = await writePool.query(`
      SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = 'requirement_id'
    `, [process.env.DB_DATABASE, TASK_TABLE]);

    if (requirementIdColumns.length === 0) {
      await writePool.query(`
        ALTER TABLE \`${TASK_TABLE}\`
        ADD COLUMN \`requirement_id\` BIGINT UNSIGNED DEFAULT NULL AFTER \`id\`,
        ADD INDEX \`idx_requirement_id\` (\`requirement_id\`)
      `);
    }

    const [legacyRequirementColumns] = await writePool.query(`
      SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = 'requirement'
    `, [process.env.DB_DATABASE, TASK_TABLE]);

    if (legacyRequirementColumns.length > 0) {
      await writePool.query(`
        ALTER TABLE \`${TASK_TABLE}\`
        DROP COLUMN \`requirement\`
      `);
    }

    // 添加 is_expanding 和 expand_message_id 字段
    const [isExpandingColumns] = await writePool.query(`
      SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = 'is_expanding'
    `, [process.env.DB_DATABASE, TASK_TABLE]);

    if (isExpandingColumns.length === 0) {
      await writePool.query(`
        ALTER TABLE \`${TASK_TABLE}\`
        ADD COLUMN \`is_expanding\` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否正在AI拆分子任务中',
        ADD COLUMN \`expand_message_id\` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的消息ID'
      `);
    }

    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${SUBTASK_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`task_id\` BIGINT UNSIGNED NOT NULL,
        \`title\` VARCHAR(500) NOT NULL,
        \`title_trans\` VARCHAR(500) DEFAULT NULL,
        \`description\` TEXT,
        \`description_trans\` TEXT,
        \`details\` TEXT,
        \`details_trans\` TEXT,
        \`status\` VARCHAR(50) NOT NULL DEFAULT 'pending',
        \`sort_order\` INT UNSIGNED DEFAULT 0,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        INDEX \`idx_task_id\` (\`task_id\`),
        INDEX \`idx_status\` (\`status\`),
        CONSTRAINT \`fk_${TABLE_PREFIX}subtask_task\` FOREIGN KEY (\`task_id\`) REFERENCES \`${TASK_TABLE}\` (\`id\`) ON DELETE CASCADE
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${TASK_DEPENDENCY_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`task_id\` BIGINT UNSIGNED NOT NULL,
        \`depends_on_task_id\` BIGINT UNSIGNED NOT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        UNIQUE KEY \`uk_${TABLE_PREFIX}task_dependency\` (\`task_id\`, \`depends_on_task_id\`),
        INDEX \`idx_depends_on\` (\`depends_on_task_id\`),
        CONSTRAINT \`fk_${TABLE_PREFIX}dependency_task\` FOREIGN KEY (\`task_id\`) REFERENCES \`${TASK_TABLE}\` (\`id\`) ON DELETE CASCADE,
        CONSTRAINT \`fk_${TABLE_PREFIX}dependency_depends_on\` FOREIGN KEY (\`depends_on_task_id\`) REFERENCES \`${TASK_TABLE}\` (\`id\`) ON DELETE CASCADE
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${SUBTASK_DEPENDENCY_TABLE}\` (
        \`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
        \`subtask_id\` BIGINT UNSIGNED NOT NULL,
        \`depends_on_subtask_id\` BIGINT UNSIGNED NOT NULL,
        \`created_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (\`id\`),
        UNIQUE KEY \`uk_${TABLE_PREFIX}subtask_dependency\` (\`subtask_id\`, \`depends_on_subtask_id\`),
        INDEX \`idx_depends_on\` (\`depends_on_subtask_id\`),
        CONSTRAINT \`fk_${TABLE_PREFIX}subtask_dependency_subtask\` FOREIGN KEY (\`subtask_id\`) REFERENCES \`${SUBTASK_TABLE}\` (\`id\`) ON DELETE CASCADE,
        CONSTRAINT \`fk_${TABLE_PREFIX}subtask_dependency_depends_on\` FOREIGN KEY (\`depends_on_subtask_id\`) REFERENCES \`${SUBTASK_TABLE}\` (\`id\`) ON DELETE CASCADE
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);

    await writePool.query(`
      CREATE TABLE IF NOT EXISTS \`${META_TABLE}\` (
        \`meta_key\` VARCHAR(100) NOT NULL PRIMARY KEY,
        \`meta_value\` LONGTEXT DEFAULT NULL,
        \`updated_at\` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      ) CHARACTER SET ${process.env.DB_UTF8MB4_CHARSET || 'utf8mb4'}
        COLLATE ${process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'}
    `);
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

async function upsertMeta(connection, metadata, tasksPath) {
  const meta = {
    storageKey: normalizeStorageKey(tasksPath),
    projectName: metadata.projectName || null,
    projectVersion: metadata.projectVersion || null,
    metadataJson: JSON.stringify(metadata)
  };

  for (const [metaKey, metaValue] of Object.entries(meta)) {
    await connection.query(
      `INSERT INTO \`${META_TABLE}\` (meta_key, meta_value) VALUES (?, ?)
       ON DUPLICATE KEY UPDATE meta_value = VALUES(meta_value), updated_at = CURRENT_TIMESTAMP`,
      [metaKey, metaValue]
    );
  }
}

async function readMeta() {
  const [rows] = await getReadPool().query(
    `SELECT meta_key, meta_value FROM \`${META_TABLE}\``
  );

  const metaMap = Object.fromEntries(rows.map(row => [row.meta_key, row.meta_value]));
  let metadata = {};

  if (metaMap.metadataJson) {
    try {
      metadata = JSON.parse(metaMap.metadataJson);
    } catch {
      metadata = {};
    }
  }

  if (metaMap.projectName && !metadata.projectName) {
    metadata.projectName = metaMap.projectName;
  }

  if (metaMap.projectVersion && !metadata.projectVersion) {
    metadata.projectVersion = metaMap.projectVersion;
  }

  if (metaMap.storageKey && !metadata.storageKey) {
    metadata.storageKey = metaMap.storageKey;
  }

  return metadata;
}

async function importFileDataToDb(tasksPath) {
  if (!fs.existsSync(tasksPath)) {
    return null;
  }

  const fileData = JSON.parse(fs.readFileSync(tasksPath, 'utf8'));
  await writeTaskData(tasksPath, fileData);
  return await readTaskData(tasksPath);
}

async function readTaskDataFromDb(tasksPath) {
  const [taskRows] = await getReadPool().query(
    `SELECT * FROM \`${TASK_TABLE}\` ORDER BY id ASC`
  );

  if (!taskRows.length) {
    return await importFileDataToDb(tasksPath);
  }

  const [taskDependencyRows] = await getReadPool().query(
    `SELECT task_id, depends_on_task_id FROM \`${TASK_DEPENDENCY_TABLE}\` ORDER BY task_id ASC, depends_on_task_id ASC`
  );
  const [subtaskRows] = await getReadPool().query(
    `SELECT * FROM \`${SUBTASK_TABLE}\` ORDER BY task_id ASC, sort_order ASC, id ASC`
  );
  const [subtaskDependencyRows] = await getReadPool().query(
    `SELECT subtask_id, depends_on_subtask_id FROM \`${SUBTASK_DEPENDENCY_TABLE}\``
  );
  const metadata = await readMeta();

  const tasks = taskRows.map(row => ({
    id: Number(row.id),
    requirementId: row.requirement_id == null ? undefined : Number(row.requirement_id),
    title: row.title,
    titleTrans: row.title_trans || undefined,
    description: row.description || '',
    descriptionTrans: row.description_trans || undefined,
    status: row.status,
    priority: row.priority,
    details: row.details || '',
    detailsTrans: row.details_trans || undefined,
    testStrategy: row.test_strategy || '',
    testStrategyTrans: row.test_strategy_trans || undefined,
    assignee: row.assignee || undefined,
    isExpanding: Boolean(row.is_expanding),
    expandMessageId: row.expand_message_id == null ? undefined : Number(row.expand_message_id),
    dependencies: [],
    subtasks: []
  }));

  const taskMap = new Map(tasks.map(task => [task.id, task]));

  taskDependencyRows.forEach(row => {
    const task = taskMap.get(Number(row.task_id));
    if (task) {
      task.dependencies.push(Number(row.depends_on_task_id));
    }
  });

  const subtaskRowMap = new Map();
  subtaskRows.forEach(row => {
    const task = taskMap.get(Number(row.task_id));
    if (!task) {
      return;
    }

    const exposedSubtaskId = Number(row.sort_order) || Number(row.id);
    const subtask = {
      id: exposedSubtaskId,
      title: row.title,
      titleTrans: row.title_trans || undefined,
      description: row.description || '',
      descriptionTrans: row.description_trans || undefined,
      details: row.details || '',
      detailsTrans: row.details_trans || undefined,
      status: row.status,
      dependencies: []
    };

    task.subtasks.push(subtask);
    subtaskRowMap.set(Number(row.id), {
      taskId: Number(row.task_id),
      subtaskId: exposedSubtaskId,
      subtask
    });
  });

  subtaskDependencyRows.forEach(row => {
    const source = subtaskRowMap.get(Number(row.subtask_id));
    const target = subtaskRowMap.get(Number(row.depends_on_subtask_id));

    if (!source || !target) {
      return;
    }

    source.subtask.dependencies ||= [];
    if (source.taskId === target.taskId) {
      source.subtask.dependencies.push(target.subtaskId);
    } else {
      source.subtask.dependencies.push(`${target.taskId}.${target.subtaskId}`);
    }
  });

  tasks.forEach(task => {
    if (!task.subtasks.length) {
      delete task.subtasks;
    }
  });

  return {
    tasks,
    metadata,
    projectName: metadata.projectName || process.env.PROJECT_NAME || 'Task Master',
    projectVersion: metadata.projectVersion || process.env.PROJECT_VERSION || '1.5.0'
  };
}

export async function readTaskData(tasksPath = 'tasks/tasks.json') {
  await ensureDbReady();
  return await readTaskDataFromDb(tasksPath);
}

export async function writeTaskData(tasksPath = 'tasks/tasks.json', data) {
  await ensureDbReady();

  const connection = await writePool.getConnection();

  try {
    await connection.beginTransaction();

    const tasks = Array.isArray(data?.tasks) ? data.tasks : [];
    const metadata = {
      ...(data?.metadata || {}),
      projectName: data?.projectName || data?.metadata?.projectName || process.env.PROJECT_NAME || 'Task Master',
      projectVersion: data?.projectVersion || data?.metadata?.projectVersion || process.env.PROJECT_VERSION || '1.5.0'
    };

    await upsertMeta(connection, metadata, tasksPath);

    if (tasks.length) {
      for (const task of tasks) {
        await connection.query(
          `INSERT INTO \`${TASK_TABLE}\`
          (id, requirement_id, title, title_trans, description, description_trans, status, priority, details, details_trans, test_strategy, test_strategy_trans, assignee)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
          ON DUPLICATE KEY UPDATE
            requirement_id = VALUES(requirement_id),
            title = VALUES(title),
            title_trans = VALUES(title_trans),
            description = VALUES(description),
            description_trans = VALUES(description_trans),
            status = VALUES(status),
            priority = VALUES(priority),
            details = VALUES(details),
            details_trans = VALUES(details_trans),
            test_strategy = VALUES(test_strategy),
            test_strategy_trans = VALUES(test_strategy_trans),
            assignee = VALUES(assignee),
            updated_at = CURRENT_TIMESTAMP`,
          [
            Number(task.id),
            task.requirementId == null ? null : Number(task.requirementId),
            task.title,
            toNullableString(task.titleTrans),
            toNullableString(task.description || ''),
            toNullableString(task.descriptionTrans),
            task.status || 'pending',
            task.priority || 'medium',
            toNullableString(task.details || ''),
            toNullableString(task.detailsTrans),
            toNullableString(task.testStrategy || ''),
            toNullableString(task.testStrategyTrans),
            toNullableString(task.assignee)
          ]
        );
      }

      const placeholders = tasks.map(() => '?').join(', ');
      await connection.query(
        `DELETE FROM \`${TASK_TABLE}\` WHERE id NOT IN (${placeholders})`,
        tasks.map(task => Number(task.id))
      );
    } else {
      await connection.query(`DELETE FROM \`${TASK_TABLE}\``);
    }

    await connection.query(`DELETE FROM \`${SUBTASK_DEPENDENCY_TABLE}\``);
    await connection.query(`DELETE FROM \`${TASK_DEPENDENCY_TABLE}\``);
    await connection.query(`DELETE FROM \`${SUBTASK_TABLE}\``);

    for (const task of tasks) {
      for (const depId of task.dependencies || []) {
        const normalizedDepId = normalizeTaskDependency(depId);
        if (!normalizedDepId) {
          continue;
        }

        await connection.query(
          `INSERT IGNORE INTO \`${TASK_DEPENDENCY_TABLE}\` (task_id, depends_on_task_id) VALUES (?, ?)`,
          [Number(task.id), normalizedDepId]
        );
      }
    }

    const subtaskDbIdMap = new Map();

    for (const task of tasks) {
      for (const subtask of task.subtasks || []) {
        const [result] = await connection.query(
          `INSERT INTO \`${SUBTASK_TABLE}\`
          (task_id, title, title_trans, description, description_trans, details, details_trans, status, sort_order)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
          [
            Number(task.id),
            subtask.title,
            toNullableString(subtask.titleTrans),
            toNullableString(subtask.description || ''),
            toNullableString(subtask.descriptionTrans),
            toNullableString(subtask.details || ''),
            toNullableString(subtask.detailsTrans),
            subtask.status || 'pending',
            Number(subtask.id)
          ]
        );

        subtaskDbIdMap.set(`${task.id}.${subtask.id}`, Number(result.insertId));
      }
    }

    for (const task of tasks) {
      for (const subtask of task.subtasks || []) {
        const sourceDbId = subtaskDbIdMap.get(`${task.id}.${subtask.id}`);
        if (!sourceDbId) {
          continue;
        }

        for (const depId of subtask.dependencies || []) {
          const normalizedDepKey = normalizeSubtaskDependency(task.id, depId);
          if (!normalizedDepKey) {
            continue;
          }

          const targetDbId = subtaskDbIdMap.get(normalizedDepKey);
          if (!targetDbId) {
            continue;
          }

          await connection.query(
            `INSERT IGNORE INTO \`${SUBTASK_DEPENDENCY_TABLE}\` (subtask_id, depends_on_subtask_id) VALUES (?, ?)`,
            [sourceDbId, targetDbId]
          );
        }
      }
    }

    await connection.commit();
  } catch (error) {
    await connection.rollback();
    throw error;
  } finally {
    connection.release();
  }
}

export async function taskDataExists(tasksPath = 'tasks/tasks.json') {
  try {
    await ensureDbReady();
    // 表存在即可，不需要有数据
    return true;
  } catch (error) {
    return false;
  }
}

export async function closeTaskStorage() {
  const pools = [writePool, ...readPools].filter(Boolean);
  await Promise.all(pools.map(pool => pool.end()));
  writePool = null;
  readPools = [];
  initPromise = null;
  readPoolIndex = 0;
}

export function getTaskStorageMode() {
  return 'db';
}

/**
 * 设置任务拆分状态
 * @param {number} taskId - 任务ID
 * @param {boolean} isExpanding - 是否正在拆分
 * @param {number|null} messageId - 关联的消息ID
 */
export async function setTaskExpanding(taskId, isExpanding, messageId = null) {
  await ensureDbReady();

  await writePool.query(
    `UPDATE \`${TASK_TABLE}\` SET is_expanding = ?, expand_message_id = ? WHERE id = ?`,
    [isExpanding ? 1 : 0, messageId, taskId]
  );
}

/**
 * 获取单个任务的详情
 * @param {number} taskId - 任务ID
 * @returns {Promise<object|null>}
 */
export async function getTaskById(taskId) {
  await ensureDbReady();

  const [rows] = await getReadPool().query(
    `SELECT * FROM \`${TASK_TABLE}\` WHERE id = ?`,
    [taskId]
  );

  if (rows.length === 0) {
    return null;
  }

  const row = rows[0];
  const task = {
    id: Number(row.id),
    requirementId: row.requirement_id == null ? undefined : Number(row.requirement_id),
    title: row.title,
    titleTrans: row.title_trans || undefined,
    description: row.description || '',
    descriptionTrans: row.description_trans || undefined,
    status: row.status,
    priority: row.priority,
    details: row.details || '',
    detailsTrans: row.details_trans || undefined,
    testStrategy: row.test_strategy || '',
    testStrategyTrans: row.test_strategy_trans || undefined,
    assignee: row.assignee || undefined,
    isExpanding: Boolean(row.is_expanding),
    expandMessageId: row.expand_message_id == null ? undefined : Number(row.expand_message_id),
    dependencies: [],
    subtasks: []
  };

  // 获取依赖
  const [depRows] = await getReadPool().query(
    `SELECT depends_on_task_id FROM \`${TASK_DEPENDENCY_TABLE}\` WHERE task_id = ? ORDER BY depends_on_task_id ASC`,
    [taskId]
  );
  task.dependencies = depRows.map(r => Number(r.depends_on_task_id));

  // 获取子任务
  const [subtaskRows] = await getReadPool().query(
    `SELECT * FROM \`${SUBTASK_TABLE}\` WHERE task_id = ? ORDER BY sort_order ASC, id ASC`,
    [taskId]
  );

  const subtaskDbIds = subtaskRows.map(r => Number(r.id));

  for (const stRow of subtaskRows) {
    const exposedSubtaskId = Number(stRow.sort_order) || Number(stRow.id);
    const subtask = {
      id: exposedSubtaskId,
      title: stRow.title,
      titleTrans: stRow.title_trans || undefined,
      description: stRow.description || '',
      descriptionTrans: stRow.description_trans || undefined,
      details: stRow.details || '',
      detailsTrans: stRow.details_trans || undefined,
      status: stRow.status,
      dependencies: []
    };

    // 获取子任务依赖
    if (subtaskDbIds.length > 0) {
      const [subDepRows] = await getReadPool().query(
        `SELECT sd.subtask_id, sd.depends_on_subtask_id
         FROM \`${SUBTASK_DEPENDENCY_TABLE}\` sd
         WHERE sd.subtask_id = ?
         ORDER BY sd.depends_on_subtask_id ASC`,
        [stRow.id]
      );

      for (const depRow of subDepRows) {
        // 找到依赖的子任务对应的 exposed id
        const depSubtaskRow = subtaskRows.find(r => Number(r.id) === Number(depRow.depends_on_subtask_id));
        if (depSubtaskRow) {
          const depExposedId = Number(depSubtaskRow.sort_order) || Number(depSubtaskRow.id);
          subtask.dependencies.push(depExposedId);
        }
      }
    }

    task.subtasks.push(subtask);
  }

  return task;
}
