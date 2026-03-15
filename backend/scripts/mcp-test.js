#!/usr/bin/env node

/**
 * MCP 任务测评工具
 *
 * 功能：
 * 1. 从 MCP 服务器拉取任务列表
 * 2. 测试任务结构的可执行性
 * 3. 对任务进行角色匹配和打分
 * 4. 输出测评报告
 *
 * 使用方式：
 *   node scripts/mcp-test.js          # 完整测试
 *   node scripts/mcp-test.js --quick  # 快速测试（仅结构验证）
 *   node scripts/mcp-test.js --json   # JSON 输出
 */

import { spawn } from 'child_process';
import { readFile } from 'fs/promises';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const PROJECT_ROOT = join(__dirname, '..');

// ========== 配置 ==========
const MCP_SERVER_PATH = join(PROJECT_ROOT, 'backend/cmd/mcp-server/main.go');
const BACKEND_PATH = join(PROJECT_ROOT, 'backend');

// 角色库
const ROLES = {
  'backend-developer': {
    keywords: ['backend', 'go', 'mysql', 'api', '数据库', '接口', '服务端'],
    description: '后端开发工程师'
  },
  'frontend-developer': {
    keywords: ['frontend', 'vue', 'typescript', 'css', 'html', '前端', '界面', '组件'],
    description: '前端开发工程师'
  },
  'fullstack-developer': {
    keywords: ['fullstack', '前后端', '集成', '联调'],
    description: '全栈开发工程师'
  },
  'devops-engineer': {
    keywords: ['devops', 'docker', 'k8s', 'ci/cd', '部署', '运维', '服务器'],
    description: '运维工程师'
  },
  'qa-engineer': {
    keywords: ['test', 'qa', '测试', '单元测试', '集成测试'],
    description: '测试工程师'
  },
  'tech-lead': {
    keywords: ['架构', '设计', '重构', '技术选型', '方案'],
    description: '技术负责人'
  }
};

// ========== 工具函数 ==========

/**
 * 睡眠函数
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 发送 JSON-RPC 请求到 MCP 服务器
 */
function sendMcpRequest(method, params = {}) {
  return new Promise((resolve, reject) => {
    // 构建启动 MCP 服务器的命令
    const mcpProcess = spawn('go', ['run', 'cmd/mcp-server/main.go'], {
      cwd: BACKEND_PATH,
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let output = '';
    let errorOutput = '';
    let initialized = false;
    let timeout;

    // 设置超时
    const processTimeout = setTimeout(() => {
      mcpProcess.kill();
      reject(new Error('MCP 服务器启动超时'));
    }, 10000);

    mcpProcess.stderr.on('data', (data) => {
      errorOutput += data.toString();
    });

    mcpProcess.stdout.on('data', (data) => {
      output += data.toString();
      const lines = output.split('\n');

      for (const line of lines) {
        if (!line.trim()) continue;
        try {
          const response = JSON.parse(line);
          if (response.method === 'notifications/initialized') {
            initialized = true;
          }
          if (response.id !== undefined) {
            clearTimeout(processTimeout);
            mcpProcess.kill();
            resolve(response);
          }
        } catch (e) {
          // 忽略非 JSON 输出
        }
      }
    });

    mcpProcess.on('close', (code) => {
      if (!initialized && method !== 'initialize') {
        clearTimeout(processTimeout);
        // 尝试使用简化方式 - 直接查询数据库
      }
    });

    // 发送初始化请求
    const initRequest = {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: {},
        clientInfo: {
          name: 'mcp-test-tool',
          version: '1.0.0'
        }
      }
    };

    mcpProcess.stdin.write(JSON.stringify(initRequest) + '\n');

    // 如果是初始化请求，发送通知
    setTimeout(() => {
      const initializedNotification = {
        jsonrpc: '2.0',
        method: 'notifications/initialized',
        params: {}
      };
      mcpProcess.stdin.write(JSON.stringify(initializedNotification) + '\n');
    }, 100);

    // 发送实际请求
    setTimeout(() => {
      if (method !== 'initialize') {
        const request = {
          jsonrpc: '2.0',
          id: 2,
          method,
          params
        };
        mcpProcess.stdin.write(JSON.stringify(request) + '\n');
      }
    }, 200);
  });
}

/**
 * 从数据库直接获取任务（备用方案）
 */
async function getTasksFromDB() {
  const { exec } = await import('child_process');

  return new Promise((resolve, reject) => {
    // 读取配置文件获取数据库连接信息
    const configPath = join(PROJECT_ROOT, 'task.json');

    readFile(configPath, 'utf-8')
      .then(configStr => {
        const config = JSON.parse(configStr);
        const db = config.storage?.database;

        if (!db) {
          reject(new Error('数据库配置缺失'));
          return;
        }

        // 构建 mysql 命令查询任务
        const mysqlCmd = `mysql -h${db.host} -P${db.port} -u${db.username} -p${db.password} ${db.database} -e "SELECT id, title, description, status, priority, category, details, module, estimated_hours, created_at FROM task_task ORDER BY id DESC LIMIT 20" --json`;

        exec(mysqlCmd, (error, stdout, stderr) => {
          if (error) {
            // 尝试使用 go 程序查询
            const goQuery = spawn('go', ['run', 'cmd/mcp-server/main.go'], {
              cwd: BACKEND_PATH
            });
            // 简化处理，返回空数组
            resolve([]);
            return;
          }

          try {
            const tasks = JSON.parse(stdout);
            resolve(tasks);
          } catch (e) {
            resolve([]);
          }
        });
      })
      .catch(err => {
        console.error('读取配置失败:', err.message);
        resolve([]);
      });
  });
}

// ========== 评估函数 ==========

/**
 * 评估结构完整性 (20 分)
 */
function evaluateStructureCompleteness(task) {
  let score = 0;
  const feedback = [];

  // 必填字段
  const requiredFields = [
    { field: 'title', name: '标题', score: 4 },
    { field: 'description', name: '描述', score: 4 },
    { field: 'status', name: '状态', score: 3 },
    { field: 'priority', name: '优先级', score: 3 },
    { field: 'details', name: '详情', score: 3 },
    { field: 'module', name: '模块', score: 3 }
  ];

  for (const { field, name, score: fieldScore } of requiredFields) {
    if (task[field] && task[field].toString().trim() !== '') {
      score += fieldScore;
    } else {
      feedback.push(`缺少${name}`);
    }
  }

  return { score, maxScore: 20, feedback };
}

/**
 * 评估 Schema 合规性 (20 分)
 */
function evaluateSchemaCompliance(task) {
  let score = 20;
  const feedback = [];

  // 检查状态有效性
  const validStatuses = ['pending', 'processing', 'done', 'blocked', 'cancelled'];
  if (task.status && !validStatuses.includes(task.status)) {
    score -= 5;
    feedback.push(`无效的状态值：${task.status}`);
  }

  // 检查优先级有效性
  const validPriorities = ['low', 'medium', 'high'];
  if (task.priority && !validPriorities.includes(task.priority)) {
    score -= 5;
    feedback.push(`无效的优先级值：${task.priority}`);
  }

  // 检查分类有效性
  const validCategories = ['', 'frontend', 'backend', 'devops', 'qa'];
  if (task.category && !validCategories.includes(task.category)) {
    score -= 5;
    feedback.push(`无效的分类值：${task.category}`);
  }

  // 检查时间字段格式
  if (task.startDate && isNaN(Date.parse(task.startDate))) {
    score -= 2;
    feedback.push('开始日期格式错误');
  }
  if (task.dueDate && isNaN(Date.parse(task.dueDate))) {
    score -= 3;
    feedback.push('截止日期格式错误');
  }

  return { score: Math.max(0, score), maxScore: 20, feedback };
}

/**
 * 评估可执行性 (25 分)
 */
function evaluateExecutability(task) {
  let score = 25;
  const feedback = [];

  // 描述长度检查
  const descLength = task.description?.length || 0;
  if (descLength < 20) {
    score -= 10;
    feedback.push('描述过于简短');
  } else if (descLength < 50) {
    score -= 5;
    feedback.push('描述不够详细');
  }

  // 详情完整性
  const detailsLength = task.details?.length || 0;
  if (detailsLength < 50) {
    score -= 8;
    feedback.push('详情信息不足');
  }

  // 检查是否有验收标准
  if (!task.acceptanceCriteria || task.acceptanceCriteria.trim() === '') {
    score -= 4;
    feedback.push('缺少验收标准');
  }

  // 检查是否有输入/输出定义
  if (!task.input || task.input.trim() === '') {
    score -= 2;
    feedback.push('缺少输入依赖定义');
  }
  if (!task.output || task.output.trim() === '') {
    score -= 1;
    feedback.push('缺少输出交付物定义');
  }

  return { score: Math.max(0, score), maxScore: 25, feedback };
}

/**
 * 评估角色匹配度 (20 分)
 */
function evaluateRoleMatching(task) {
  const matchedRoles = [];
  const taskText = `${task.title} ${task.description} ${task.details} ${task.category}`.toLowerCase();

  for (const [role, config] of Object.entries(ROLES)) {
    let matchCount = 0;
    for (const keyword of config.keywords) {
      if (taskText.includes(keyword.toLowerCase())) {
        matchCount++;
      }
    }
    if (matchCount > 0) {
      matchedRoles.push({ role, matchCount, ...config });
    }
  }

  // 如果没有匹配到任何角色，根据分类判断
  if (matchedRoles.length === 0) {
    if (task.category === 'backend') {
      matchedRoles.push({ role: 'backend-developer', matchCount: 1, ...ROLES['backend-developer'] });
    } else if (task.category === 'frontend') {
      matchedRoles.push({ role: 'frontend-developer', matchCount: 1, ...ROLES['frontend-developer'] });
    }
  }

  // 计算分数
  let score = 10; // 基础分
  if (matchedRoles.length > 0) {
    score += Math.min(10, matchedRoles[0].matchCount * 3);
  }
  const topRole = matchedRoles.length > 0 ? matchedRoles[0] : { role: 'unassigned', description: '未分配' };

  return {
    score,
    maxScore: 20,
    feedback: matchedRoles.length > 0 ? [] : ['未匹配到合适角色'],
    topRole: topRole.role,
    roleDescription: topRole.description,
    allMatches: matchedRoles
  };
}

/**
 * 评估依赖合理性 (15 分)
 */
function evaluateDependencyReasoning(task, allTasks) {
  // 简化实现，实际需要从数据库获取依赖关系
  let score = 15;
  const feedback = [];

  // 如果任务有 input 字段，检查是否引用了其他任务
  if (task.input) {
    const inputText = task.input.toLowerCase();
    if (inputText.includes('task') || inputText.includes('任务') || inputText.includes('依赖')) {
      score += 0; // 有明确的依赖说明
    } else {
      score -= 2;
      feedback.push('依赖描述不够明确');
    }
  }

  // 检查风险评估
  if (!task.risk || task.risk.trim() === '') {
    score -= 3;
    feedback.push('缺少风险评估');
  }

  return { score: Math.max(0, score), maxScore: 15, feedback };
}

/**
 * 综合评估任务
 */
function evaluateTask(task, allTasks = []) {
  const evaluations = {
    structure: evaluateStructureCompleteness(task),
    schema: evaluateSchemaCompliance(task),
    executability: evaluateExecutability(task),
    roleMatching: evaluateRoleMatching(task),
    dependency: evaluateDependencyReasoning(task, allTasks)
  };

  const totalScore = Object.values(evaluations).reduce((sum, e) => sum + e.score, 0);
  const maxTotalScore = Object.values(evaluations).reduce((sum, e) => sum + e.maxScore, 0);

  return {
    taskId: task.id,
    title: task.title,
    evaluations,
    totalScore,
    maxTotalScore,
    topRole: evaluations.roleMatching.topRole,
    roleDescription: evaluations.roleMatching.roleDescription
  };
}

// ========== 输出格式化 ==========

/**
 * 格式化分数显示
 */
function formatScore(score, maxScore) {
  const percentage = (score / maxScore) * 100;
  let icon = '✓';
  if (percentage < 50) icon = '✗';
  else if (percentage < 70) icon = '~';
  return `${score}/${maxScore} ${icon}`;
}

/**
 * 输出文本报告
 */
function printTextReport(results) {
  const separator = '=====================================';
  const subSeparator = '─────────────────────────';

  console.log('\n' + separator);
  console.log('       MCP 任务测评报告');
  console.log(separator);
  console.log(`测评时间：${new Date().toLocaleString('zh-CN')}`);
  console.log(separator);

  if (results.length === 0) {
    console.log('\n  暂无任务数据');
    console.log(separator);
    return;
  }

  // 统计信息
  const totalTasks = results.length;
  const avgScore = Math.round(results.reduce((sum, r) => sum + r.totalScore, 0) / totalTasks);
  const avgPercentage = Math.round((avgScore / results[0].maxTotalScore) * 100);

  console.log(`\n任务总数：${totalTasks}`);
  console.log(`平均得分：${avgScore}/${results[0].maxTotalScore} (${avgPercentage}%)`);
  console.log(separator);

  // 角色分布统计
  const roleStats = {};
  for (const result of results) {
    roleStats[result.topRole] = (roleStats[result.topRole] || 0) + 1;
  }
  console.log('\n【角色分布】');
  for (const [role, count] of Object.entries(roleStats)) {
    const roleName = ROLES[role]?.description || role;
    console.log(`  ${roleName}: ${count} 个任务`);
  }
  console.log(separator);

  // 详细任务评估
  for (const result of results) {
    const percentage = Math.round((result.totalScore / result.maxTotalScore) * 100);
    let grade = 'A';
    if (percentage < 60) grade = 'C';
    else if (percentage < 80) grade = 'B';

    console.log(`\n[任务 ${result.taskId}] ${result.title}`);
    console.log(`  等级：${grade} | 得分：${result.totalScore}/${result.maxTotalScore} (${percentage}%)`);
    console.log(subSeparator);

    const e = result.evaluations;
    console.log(`  结构完整性：   ${formatScore(e.structure.score, e.structure.maxScore)}`);
    console.log(`  Schema 合规性：  ${formatScore(e.schema.score, e.schema.maxScore)}`);
    console.log(`  可执行性：     ${formatScore(e.executability.score, e.executability.maxScore)}`);
    console.log(`  角色匹配度：   ${formatScore(e.roleMatching.score, e.roleMatching.maxScore)}`);
    console.log(`  依赖合理性：   ${formatScore(e.dependency.score, e.dependency.maxScore)}`);
    console.log(subSeparator);
    console.log(`  推荐角色：${ROLES[result.topRole]?.description || result.topRole}`);

    // 收集所有反馈
    const allFeedback = [
      ...e.structure.feedback,
      ...e.schema.feedback,
      ...e.executability.feedback,
      ...e.roleMatching.feedback,
      ...e.dependency.feedback
    ];

    if (allFeedback.length > 0) {
      console.log(`  改进建议:`);
      for (const fb of allFeedback) {
        console.log(`    - ${fb}`);
      }
    }
    console.log();
  }

  console.log(separator);
  console.log('              测评完成');
  console.log(separator + '\n');
}

/**
 * 输出 JSON 报告
 */
function printJsonReport(results) {
  console.log(JSON.stringify({
    timestamp: new Date().toISOString(),
    totalTasks: results.length,
    averageScore: results.length > 0
      ? Math.round(results.reduce((sum, r) => sum + r.totalScore, 0) / results.length)
      : 0,
    results
  }, null, 2));
}

// ========== 主函数 ==========

async function main() {
  const args = process.argv.slice(2);
  const isQuickMode = args.includes('--quick');
  const isJsonMode = args.includes('--json');
  const isHelpMode = args.includes('--help') || args.includes('-h');

  if (isHelpMode) {
    console.log(`
MCP 任务测评工具

使用方式:
  node scripts/mcp-test.js [选项]

选项:
  --quick    快速模式（仅结构验证）
  --json     JSON 格式输出
  --help     显示帮助信息

示例:
  node scripts/mcp-test.js              # 完整测试
  node scripts/mcp-test.js --quick      # 快速测试
  node scripts/mcp-test.js --json       # JSON 输出
`);
    process.exit(0);
  }

  console.log('\n正在启动 MCP 任务测评...\n');

  try {
    // 获取任务数据
    let tasks = [];

    if (!isQuickMode) {
      // 尝试通过 MCP 获取
      try {
        console.log('通过 MCP 获取任务列表...');
        const response = await sendMcpRequest('list_tasks');
        if (response.result) {
          // 解析 MCP 响应
          tasks = JSON.parse(response.result?.tasks || response.result || '[]');
        }
      } catch (e) {
        console.log('MCP 方式获取失败，尝试直接查询数据库...');
      }
    }

    // 如果 MCP 方式失败或快速模式，尝试直接查询
    if (tasks.length === 0) {
      console.log('查询数据库获取任务列表...');
      tasks = await getTasksFromDB();
    }

    if (tasks.length === 0) {
      console.log('\n警告：未获取到任何任务数据');
      console.log('请确认：');
      console.log('  1. 数据库已正确配置');
      console.log('  2. 数据库中存在任务数据');

      // 显示示例数据
      console.log('\n--- 示例评估（使用模拟数据）---\n');
      tasks = [{
        id: 1,
        title: '用户登录功能',
        description: '实现用户登录接口',
        status: 'pending',
        priority: 'high',
        category: 'backend',
        details: '需要实现用户名密码验证、JWT 生成、会话管理',
        module: 'auth'
      }];
    }

    console.log(`获取到 ${tasks.length} 个任务\n`);

    // 评估每个任务
    const results = tasks.map(task => evaluateTask(task, tasks));

    // 输出报告
    if (isJsonMode) {
      printJsonReport(results);
    } else {
      printTextReport(results);
    }

  } catch (error) {
    console.error('\n测评失败:', error.message);
    console.error('\n请检查：');
    console.error('  1. Go 环境是否正确安装');
    console.error('  2. 后端配置是否正确 (backend/config/config.json)');
    console.error('  3. 数据库是否可连接');
    process.exit(1);
  }
}

// 运行
main();
