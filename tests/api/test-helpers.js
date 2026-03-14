/**
 * 测试环境配置
 * 用于 API 单元测试的 setup 文件
 */
import express from 'express';
import cors from 'cors';
import multer from 'multer';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 测试用的数据存储路径
const testDir = path.join(__dirname, '../test-data');
const testTasksPath = path.join(testDir, 'test-tasks.json');
const testRequirementsPath = path.join(testDir, 'test-requirements.json');
const testMembersPath = path.join(testDir, 'test-members.json');
const testAssignmentsPath = path.join(testDir, 'test-assignments.json');
const testCommentsPath = path.join(testDir, 'test-comments.json');
const testActivitiesPath = path.join(testDir, 'test-activities.json');

// 确保测试数据目录存在
if (!fs.existsSync(testDir)) {
  fs.mkdirSync(testDir, { recursive: true });
}

// 初始化测试数据
function initTestData() {
  // 清理旧数据
  cleanupTestData();

  // 确保目录存在
  if (!fs.existsSync(testDir)) {
    fs.mkdirSync(testDir, { recursive: true });
  }

  // 测试任务数据
  const testTasks = {
    tasks: [
      {
        id: 1,
        title: '测试任务1',
        description: '测试任务描述1',
        status: 'pending',
        priority: 'high',
        assignee: 'user1',
        requirementId: 1,
        subtasks: [
          { id: 1, title: '子任务1', status: 'pending' },
          { id: 2, title: '子任务2', status: 'done' }
        ],
        dependencies: [],
        createdAt: Date.now(),
        updatedAt: Date.now()
      },
      {
        id: 2,
        title: '测试任务2',
        description: '测试任务描述2',
        status: 'in-progress',
        priority: 'medium',
        assignee: 'user2',
        requirementId: 1,
        subtasks: [],
        dependencies: [],
        createdAt: Date.now(),
        updatedAt: Date.now()
      }
    ]
  };

  // 测试需求数据
  const testRequirements = {
    requirements: [
      {
        id: 1,
        title: '测试需求1',
        description: '测试需求描述1',
        status: 'pending',
        priority: 'high',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        documents: []
      },
      {
        id: 2,
        title: '测试需求2',
        description: '测试需求描述2',
        status: 'done',
        priority: 'low',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        documents: []
      }
    ]
  };

  // 测试成员数据
  const testMembers = {
    members: [
      {
        id: 1,
        name: '张三',
        email: 'zhangsan@test.com',
        role: 'admin',
        department: '技术部',
        status: 'active',
        skills: ['JavaScript', 'Vue'],
        createdAt: Date.now(),
        updatedAt: Date.now()
      },
      {
        id: 2,
        name: '李四',
        email: 'lisi@test.com',
        role: 'member',
        department: '产品部',
        status: 'active',
        skills: ['产品规划'],
        createdAt: Date.now(),
        updatedAt: Date.now()
      }
    ]
  };

  // 测试分配数据
  const testAssignments = {
    assignments: [
      {
        id: 1,
        taskId: 1,
        subtaskId: null,
        memberId: 1,
        hours: 8,
        createdAt: Date.now()
      },
      {
        id: 2,
        taskId: 1,
        subtaskId: 1,
        memberId: 2,
        hours: 4,
        createdAt: Date.now()
      }
    ]
  };

  // 测试评论数据
  const testComments = {
    comments: [
      {
        id: 1,
        taskId: 1,
        content: '这是一条评论',
        authorId: 1,
        parentId: null,
        createdAt: Date.now(),
        updatedAt: Date.now()
      }
    ]
  };

  // 测试活动数据
  const testActivities = {
    activities: [
      {
        id: 1,
        taskId: 1,
        type: 'created',
        description: '任务创建',
        operatorId: 1,
        createdAt: Date.now()
      }
    ]
  };

  // 写入测试数据文件
  fs.writeFileSync(testTasksPath, JSON.stringify(testTasks, null, 2));
  fs.writeFileSync(testRequirementsPath, JSON.stringify(testRequirements, null, 2));
  fs.writeFileSync(testMembersPath, JSON.stringify(testMembers, null, 2));
  fs.writeFileSync(testAssignmentsPath, JSON.stringify(testAssignments, null, 2));
  fs.writeFileSync(testCommentsPath, JSON.stringify(testComments, null, 2));
  fs.writeFileSync(testActivitiesPath, JSON.stringify(testActivities, null, 2));
}

// 清理测试数据
function cleanupTestData() {
  const files = [
    testTasksPath,
    testRequirementsPath,
    testMembersPath,
    testAssignmentsPath,
    testCommentsPath,
    testActivitiesPath
  ];
  files.forEach(file => {
    if (fs.existsSync(file)) {
      fs.unlinkSync(file);
    }
  });
}

// 辅助函数
function readJsonFile(filePath) {
  if (!fs.existsSync(filePath)) {
    return null;
  }
  const content = fs.readFileSync(filePath, 'utf-8');
  return JSON.parse(content);
}

function writeJsonFile(filePath, data) {
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2));
}

// 创建测试用的 Express 应用
function createTestApp() {
  const app = express();
  app.use(cors());
  app.use(express.json());

  // 配置文件上传
  const uploadDir = path.join(testDir, 'uploads');
  if (!fs.existsSync(uploadDir)) {
    fs.mkdirSync(uploadDir, { recursive: true });
  }
  const upload = multer({ dest: uploadDir });

  // ==================== 任务 API ====================

  // GET /api/tasks - 获取任务列表
  app.get('/api/tasks', (req, res) => {
    const requirementId = parseInt(req.query.requirementId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    let tasks = data.tasks || [];
    if (requirementId) {
      tasks = tasks.filter(t => t.requirementId === requirementId);
    }

    res.json({ tasks });
  });

  // GET /api/tasks/:taskId - 获取任务详情
  app.get('/api/tasks/:taskId', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const task = data.tasks.find(t => t.id === taskId);
    if (!task) {
      return res.status(404).json({ error: 'Task not found' });
    }

    res.json(task);
  });

  // PUT /api/tasks/:taskId - 更新任务
  app.put('/api/tasks/:taskId', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const updates = req.body;
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    data.tasks[taskIndex] = { ...data.tasks[taskIndex], ...updates, updatedAt: Date.now() };
    writeJsonFile(testTasksPath, data);
    res.json(data.tasks[taskIndex]);
  });

  // DELETE /api/tasks/:taskId - 删除任务
  app.delete('/api/tasks/:taskId', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    data.tasks.splice(taskIndex, 1);
    writeJsonFile(testTasksPath, data);
    res.json({ success: true });
  });

  // POST /api/tasks/:taskId/copy - 复制任务
  app.post('/api/tasks/:taskId/copy', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const task = data.tasks.find(t => t.id === taskId);
    if (!task) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const newId = Math.max(...data.tasks.map(t => t.id)) + 1;
    const newTask = {
      ...task,
      id: newId,
      title: `${task.title} (副本)`,
      status: 'pending',
      createdAt: Date.now(),
      updatedAt: Date.now()
    };
    data.tasks.push(newTask);
    writeJsonFile(testTasksPath, data);
    res.json(newTask);
  });

  // POST /api/tasks/batch-delete - 批量删除任务
  app.post('/api/tasks/batch-delete', (req, res) => {
    const { taskIds } = req.body;
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const beforeCount = data.tasks.length;
    data.tasks = data.tasks.filter(t => !taskIds.includes(t.id));
    writeJsonFile(testTasksPath, data);
    res.json({ success: true, deletedCount: beforeCount - data.tasks.length });
  });

  // PUT /api/tasks/:taskId/subtasks/reorder - 子任务排序
  app.put('/api/tasks/:taskId/subtasks/reorder', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const { subtaskIds } = req.body;
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const task = data.tasks[taskIndex];
    if (!task.subtasks || task.subtasks.length === 0) {
      return res.status(400).json({ error: 'No subtasks to reorder' });
    }

    // 按照 subtaskIds 顺序重新排列子任务
    const reorderedSubtasks = [];
    subtaskIds.forEach((subtaskId, index) => {
      const subtask = task.subtasks.find(s => s.id === subtaskId);
      if (subtask) {
        reorderedSubtasks.push({ ...subtask, order: index });
      }
    });

    task.subtasks = reorderedSubtasks;
    task.updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json(task);
  });

  // PUT /api/tasks/:taskId/subtasks/:subtaskId - 更新子任务
  app.put('/api/tasks/:taskId/subtasks/:subtaskId', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const subtaskId = parseInt(req.params.subtaskId, 10);
    const updates = req.body;
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const task = data.tasks[taskIndex];
    if (!task.subtasks) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    const subtaskIndex = task.subtasks.findIndex(s => s.id === subtaskId);
    if (subtaskIndex === -1) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    task.subtasks[subtaskIndex] = { ...task.subtasks[subtaskIndex], ...updates };
    task.updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json(task.subtasks[subtaskIndex]);
  });

  // DELETE /api/tasks/:taskId/subtasks - 删除所有子任务
  app.delete('/api/tasks/:taskId/subtasks', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    data.tasks[taskIndex].subtasks = [];
    data.tasks[taskIndex].updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json({ success: true });
  });

  // DELETE /api/tasks/:taskId/subtasks/:subtaskId - 删除子任务
  app.delete('/api/tasks/:taskId/subtasks/:subtaskId', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const subtaskId = parseInt(req.params.subtaskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const task = data.tasks[taskIndex];
    if (!task.subtasks) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    const subtaskIndex = task.subtasks.findIndex(s => s.id === subtaskId);
    if (subtaskIndex === -1) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    task.subtasks.splice(subtaskIndex, 1);
    task.updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json({ success: true });
  });

  // POST /api/tasks/:taskId/subtasks/:subtaskId/regenerate - 重新生成子任务
  app.post('/api/tasks/:taskId/subtasks/:subtaskId/regenerate', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const subtaskId = parseInt(req.params.subtaskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const task = data.tasks[taskIndex];
    if (!task.subtasks) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    const subtaskIndex = task.subtasks.findIndex(s => s.id === subtaskId);
    if (subtaskIndex === -1) {
      return res.status(404).json({ error: 'Subtask not found' });
    }

    // 模拟重新生成
    task.subtasks[subtaskIndex] = {
      ...task.subtasks[subtaskIndex],
      title: `${task.subtasks[subtaskIndex].title} (重新生成)`,
      updatedAt: Date.now()
    };
    task.updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json(task.subtasks[subtaskIndex]);
  });

  // POST /api/tasks/:taskId/expand - 展开任务
  app.post('/api/tasks/:taskId/expand', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    const task = data.tasks[taskIndex];
    // 模拟展开，添加新子任务
    const newSubtask = {
      id: (task.subtasks?.length || 0) + 1,
      title: '自动生成的子任务',
      status: 'pending',
      description: '通过展开自动生成'
    };

    if (!task.subtasks) {
      task.subtasks = [];
    }
    task.subtasks.push(newSubtask);
    task.updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json(task);
  });

  // PUT /api/tasks/:taskId/time - 更新任务时间
  app.put('/api/tasks/:taskId/time', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const { estimatedHours, actualHours } = req.body;
    const data = readJsonFile(testTasksPath);
    if (!data) {
      return res.status(404).json({ error: 'Tasks file not found' });
    }

    const taskIndex = data.tasks.findIndex(t => t.id === taskId);
    if (taskIndex === -1) {
      return res.status(404).json({ error: 'Task not found' });
    }

    if (estimatedHours !== undefined) {
      data.tasks[taskIndex].estimatedHours = estimatedHours;
    }
    if (actualHours !== undefined) {
      data.tasks[taskIndex].actualHours = actualHours;
    }
    data.tasks[taskIndex].updatedAt = Date.now();
    writeJsonFile(testTasksPath, data);
    res.json(data.tasks[taskIndex]);
  });

  // ==================== 成员 API ====================

  // GET /api/members - 获取成员列表
  app.get('/api/members', (req, res) => {
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.json([]);
    }
    res.json(data.members || []);
  });

  // GET /api/members/statistics - 获取成员统计
  app.get('/api/members/statistics', (req, res) => {
    const data = readJsonFile(testMembersPath);
    const members = data?.members || [];
    res.json({
      total: members.length,
      active: members.filter(m => m.status === 'active').length,
      inactive: members.filter(m => m.status === 'inactive').length,
      byDepartment: {}
    });
  });

  // GET /api/members/departments - 获取部门列表
  app.get('/api/members/departments', (req, res) => {
    const data = readJsonFile(testMembersPath);
    const members = data?.members || [];
    const departments = [...new Set(members.map(m => m.department).filter(Boolean))];
    res.json(departments);
  });

  // GET /api/members/search - 搜索成员
  app.get('/api/members/search', (req, res) => {
    const keyword = req.query.keyword || '';
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.json([]);
    }

    const members = data.members.filter(m =>
      m.name.includes(keyword) ||
      m.email?.includes(keyword) ||
      m.department?.includes(keyword)
    );
    res.json(members);
  });

  // GET /api/members/:id - 获取成员详情
  app.get('/api/members/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.status(404).json({ error: 'Member not found' });
    }

    const member = data.members.find(m => m.id === id);
    if (!member) {
      return res.status(404).json({ error: 'Member not found' });
    }
    res.json(member);
  });

  // POST /api/members - 创建成员
  app.post('/api/members', (req, res) => {
    const memberData = req.body;
    const data = readJsonFile(testMembersPath) || { members: [] };

    const newId = data.members.length > 0
      ? Math.max(...data.members.map(m => m.id)) + 1
      : 1;

    const newMember = {
      id: newId,
      ...memberData,
      status: memberData.status || 'active',
      createdAt: Date.now(),
      updatedAt: Date.now()
    };
    data.members.push(newMember);
    writeJsonFile(testMembersPath, data);
    res.json(newMember);
  });

  // PUT /api/members/:id - 更新成员
  app.put('/api/members/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const updates = req.body;
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.status(404).json({ error: 'Member not found' });
    }

    const memberIndex = data.members.findIndex(m => m.id === id);
    if (memberIndex === -1) {
      return res.status(404).json({ error: 'Member not found' });
    }

    data.members[memberIndex] = {
      ...data.members[memberIndex],
      ...updates,
      updatedAt: Date.now()
    };
    writeJsonFile(testMembersPath, data);
    res.json(data.members[memberIndex]);
  });

  // DELETE /api/members/:id - 删除成员
  app.delete('/api/members/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.status(404).json({ error: 'Member not found' });
    }

    const memberIndex = data.members.findIndex(m => m.id === id);
    if (memberIndex === -1) {
      return res.status(404).json({ error: 'Member not found' });
    }

    data.members.splice(memberIndex, 1);
    writeJsonFile(testMembersPath, data);
    res.json({ success: true });
  });

  // POST /api/members/:id/activate - 激活成员
  app.post('/api/members/:id/activate', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.status(404).json({ error: 'Member not found' });
    }

    const memberIndex = data.members.findIndex(m => m.id === id);
    if (memberIndex === -1) {
      return res.status(404).json({ error: 'Member not found' });
    }

    data.members[memberIndex].status = 'active';
    data.members[memberIndex].updatedAt = Date.now();
    writeJsonFile(testMembersPath, data);
    res.json(data.members[memberIndex]);
  });

  // POST /api/members/:id/deactivate - 停用成员
  app.post('/api/members/:id/deactivate', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testMembersPath);
    if (!data) {
      return res.status(404).json({ error: 'Member not found' });
    }

    const memberIndex = data.members.findIndex(m => m.id === id);
    if (memberIndex === -1) {
      return res.status(404).json({ error: 'Member not found' });
    }

    data.members[memberIndex].status = 'inactive';
    data.members[memberIndex].updatedAt = Date.now();
    writeJsonFile(testMembersPath, data);
    res.json(data.members[memberIndex]);
  });

  // ==================== 分配 API ====================

  // GET /api/tasks/:taskId/assignments - 获取任务分配列表
  app.get('/api/tasks/:taskId/assignments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.json([]);
    }
    const assignments = data.assignments.filter(a => a.taskId === taskId && !a.subtaskId);
    res.json(assignments);
  });

  // GET /api/tasks/:taskId/assignments/overview - 获取分配概览
  app.get('/api/tasks/:taskId/assignments/overview', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.json({ taskAssignments: [], subtaskAssignments: {} });
    }
    const taskAssignments = data.assignments.filter(a => a.taskId === taskId && !a.subtaskId);
    const subtaskAssignments = {};
    data.assignments
      .filter(a => a.taskId === taskId && a.subtaskId)
      .forEach(a => {
        if (!subtaskAssignments[a.subtaskId]) {
          subtaskAssignments[a.subtaskId] = [];
        }
        subtaskAssignments[a.subtaskId].push(a);
      });
    res.json({ taskAssignments, subtaskAssignments });
  });

  // POST /api/tasks/:taskId/assignments - 分配任务
  app.post('/api/tasks/:taskId/assignments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const { memberId, subtaskId, hours } = req.body;
    const data = readJsonFile(testAssignmentsPath) || { assignments: [] };

    const newId = data.assignments.length > 0
      ? Math.max(...data.assignments.map(a => a.id)) + 1
      : 1;

    const newAssignment = {
      id: newId,
      taskId,
      subtaskId: subtaskId || null,
      memberId,
      hours: hours || 0,
      createdAt: Date.now()
    };
    data.assignments.push(newAssignment);
    writeJsonFile(testAssignmentsPath, data);
    res.json(newAssignment);
  });

  // DELETE /api/tasks/:taskId/assignments/:assignmentId - 取消分配
  app.delete('/api/tasks/:taskId/assignments/:assignmentId', (req, res) => {
    const assignmentId = parseInt(req.params.assignmentId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.status(404).json({ error: 'Assignment not found' });
    }

    const index = data.assignments.findIndex(a => a.id === assignmentId);
    if (index === -1) {
      return res.status(404).json({ error: 'Assignment not found' });
    }

    data.assignments.splice(index, 1);
    writeJsonFile(testAssignmentsPath, data);
    res.json({ success: true });
  });

  // GET /api/tasks/:taskId/subtasks/:subtaskId/assignments - 获取子任务分配
  app.get('/api/tasks/:taskId/subtasks/:subtaskId/assignments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const subtaskId = parseInt(req.params.subtaskId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.json([]);
    }
    const assignments = data.assignments.filter(
      a => a.taskId === taskId && a.subtaskId === subtaskId
    );
    res.json(assignments);
  });

  // POST /api/tasks/:taskId/subtasks/:subtaskId/assignments - 分配子任务
  app.post('/api/tasks/:taskId/subtasks/:subtaskId/assignments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const subtaskId = parseInt(req.params.subtaskId, 10);
    const { memberId, hours } = req.body;
    const data = readJsonFile(testAssignmentsPath) || { assignments: [] };

    const newId = data.assignments.length > 0
      ? Math.max(...data.assignments.map(a => a.id)) + 1
      : 1;

    const newAssignment = {
      id: newId,
      taskId,
      subtaskId,
      memberId,
      hours: hours || 0,
      createdAt: Date.now()
    };
    data.assignments.push(newAssignment);
    writeJsonFile(testAssignmentsPath, data);
    res.json(newAssignment);
  });

  // DELETE /api/tasks/:taskId/subtasks/:subtaskId/assignments/:assignmentId
  app.delete('/api/tasks/:taskId/subtasks/:subtaskId/assignments/:assignmentId', (req, res) => {
    const assignmentId = parseInt(req.params.assignmentId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.status(404).json({ error: 'Assignment not found' });
    }

    const index = data.assignments.findIndex(a => a.id === assignmentId);
    if (index === -1) {
      return res.status(404).json({ error: 'Assignment not found' });
    }

    data.assignments.splice(index, 1);
    writeJsonFile(testAssignmentsPath, data);
    res.json({ success: true });
  });

  // GET /api/members/:memberId/assignments - 获取成员分配
  app.get('/api/members/:memberId/assignments', (req, res) => {
    const memberId = parseInt(req.params.memberId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.json([]);
    }
    const assignments = data.assignments.filter(a => a.memberId === memberId);
    res.json(assignments);
  });

  // GET /api/members/:memberId/workload - 获取成员工作负载
  app.get('/api/members/:memberId/workload', (req, res) => {
    const memberId = parseInt(req.params.memberId, 10);
    const data = readJsonFile(testAssignmentsPath);
    if (!data) {
      return res.json({ totalHours: 0, assignmentCount: 0 });
    }
    const memberAssignments = data.assignments.filter(a => a.memberId === memberId);
    const totalHours = memberAssignments.reduce((sum, a) => sum + (a.hours || 0), 0);
    res.json({ totalHours, assignmentCount: memberAssignments.length });
  });

  // ==================== 评论 API ====================

  // GET /api/tasks/:taskId/comments - 获取评论列表
  app.get('/api/tasks/:taskId/comments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.json([]);
    }
    const comments = data.comments.filter(c => c.taskId === taskId);
    res.json(comments);
  });

  // GET /api/tasks/:taskId/comments/tree - 获取评论树
  app.get('/api/tasks/:taskId/comments/tree', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.json([]);
    }
    const comments = data.comments.filter(c => c.taskId === taskId && !c.parentId);
    res.json(comments);
  });

  // GET /api/tasks/:taskId/comments/statistics - 获取评论统计
  app.get('/api/tasks/:taskId/comments/statistics', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testCommentsPath);
    const comments = data?.comments?.filter(c => c.taskId === taskId) || [];
    res.json({ total: comments.length });
  });

  // GET /api/tasks/:taskId/comments/:commentId - 获取评论详情
  app.get('/api/tasks/:taskId/comments/:commentId', (req, res) => {
    const commentId = parseInt(req.params.commentId, 10);
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.status(404).json({ error: 'Comment not found' });
    }
    const comment = data.comments.find(c => c.id === commentId);
    if (!comment) {
      return res.status(404).json({ error: 'Comment not found' });
    }
    res.json(comment);
  });

  // POST /api/tasks/:taskId/comments - 创建评论
  app.post('/api/tasks/:taskId/comments', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const { content, authorId, parentId } = req.body;
    const data = readJsonFile(testCommentsPath) || { comments: [] };

    const newId = data.comments.length > 0
      ? Math.max(...data.comments.map(c => c.id)) + 1
      : 1;

    const newComment = {
      id: newId,
      taskId,
      content,
      authorId,
      parentId: parentId || null,
      createdAt: Date.now(),
      updatedAt: Date.now()
    };
    data.comments.push(newComment);
    writeJsonFile(testCommentsPath, data);
    res.json(newComment);
  });

  // PUT /api/tasks/:taskId/comments/:commentId - 更新评论
  app.put('/api/tasks/:taskId/comments/:commentId', (req, res) => {
    const commentId = parseInt(req.params.commentId, 10);
    const { content } = req.body;
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.status(404).json({ error: 'Comment not found' });
    }

    const index = data.comments.findIndex(c => c.id === commentId);
    if (index === -1) {
      return res.status(404).json({ error: 'Comment not found' });
    }

    data.comments[index].content = content;
    data.comments[index].updatedAt = Date.now();
    writeJsonFile(testCommentsPath, data);
    res.json(data.comments[index]);
  });

  // DELETE /api/tasks/:taskId/comments/:commentId - 删除评论
  app.delete('/api/tasks/:taskId/comments/:commentId', (req, res) => {
    const commentId = parseInt(req.params.commentId, 10);
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.status(404).json({ error: 'Comment not found' });
    }

    const index = data.comments.findIndex(c => c.id === commentId);
    if (index === -1) {
      return res.status(404).json({ error: 'Comment not found' });
    }

    data.comments.splice(index, 1);
    writeJsonFile(testCommentsPath, data);
    res.json({ success: true });
  });

  // GET /api/tasks/:taskId/comments/:commentId/replies - 获取回复列表
  app.get('/api/tasks/:taskId/comments/:commentId/replies', (req, res) => {
    const commentId = parseInt(req.params.commentId, 10);
    const data = readJsonFile(testCommentsPath);
    if (!data) {
      return res.json([]);
    }
    const replies = data.comments.filter(c => c.parentId === commentId);
    res.json(replies);
  });

  // ==================== 活动 API ====================

  // GET /api/tasks/:taskId/activities - 获取任务活动
  app.get('/api/tasks/:taskId/activities', (req, res) => {
    const taskId = parseInt(req.params.taskId, 10);
    const data = readJsonFile(testActivitiesPath);
    if (!data) {
      return res.json([]);
    }
    const activities = data.activities.filter(a => a.taskId === taskId);
    res.json(activities);
  });

  // GET /api/activities - 获取全局活动
  app.get('/api/activities', (req, res) => {
    const data = readJsonFile(testActivitiesPath);
    res.json(data?.activities || []);
  });

  // GET /api/activities/statistics - 获取活动统计
  app.get('/api/activities/statistics', (req, res) => {
    const data = readJsonFile(testActivitiesPath);
    const activities = data?.activities || [];
    res.json({ total: activities.length });
  });

  // ==================== 需求 API ====================

  // GET /api/requirements - 获取需求列表
  app.get('/api/requirements', (req, res) => {
    const data = readJsonFile(testRequirementsPath);
    res.json(data?.requirements || []);
  });

  // GET /api/requirements/statistics - 获取需求统计
  app.get('/api/requirements/statistics', (req, res) => {
    const data = readJsonFile(testRequirementsPath);
    const requirements = data?.requirements || [];
    res.json({
      total: requirements.length,
      pending: requirements.filter(r => r.status === 'pending').length,
      done: requirements.filter(r => r.status === 'done').length
    });
  });

  // GET /api/requirements/:id - 获取需求详情
  app.get('/api/requirements/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }
    const requirement = data.requirements.find(r => r.id === id);
    if (!requirement) {
      return res.status(404).json({ error: 'Requirement not found' });
    }
    res.json(requirement);
  });

  // POST /api/requirements - 创建需求
  app.post('/api/requirements', (req, res) => {
    const requirementData = req.body;
    const data = readJsonFile(testRequirementsPath) || { requirements: [] };

    const newId = data.requirements.length > 0
      ? Math.max(...data.requirements.map(r => r.id)) + 1
      : 1;

    const newRequirement = {
      id: newId,
      ...requirementData,
      status: requirementData.status || 'pending',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      documents: []
    };
    data.requirements.push(newRequirement);
    writeJsonFile(testRequirementsPath, data);
    res.json(newRequirement);
  });

  // PUT /api/requirements/:id - 更新需求
  app.put('/api/requirements/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const updates = req.body;
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const index = data.requirements.findIndex(r => r.id === id);
    if (index === -1) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    data.requirements[index] = {
      ...data.requirements[index],
      ...updates,
      updatedAt: Date.now()
    };
    writeJsonFile(testRequirementsPath, data);
    res.json(data.requirements[index]);
  });

  // DELETE /api/requirements/:id - 删除需求
  app.delete('/api/requirements/:id', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const index = data.requirements.findIndex(r => r.id === id);
    if (index === -1) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    data.requirements.splice(index, 1);
    writeJsonFile(testRequirementsPath, data);
    res.json({ success: true });
  });

  // POST /api/requirements/:id/documents - 上传文档
  app.post('/api/requirements/:id/documents', upload.single('file'), (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const index = data.requirements.findIndex(r => r.id === id);
    if (index === -1) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    if (!req.file) {
      return res.status(400).json({ error: 'No file uploaded' });
    }

    const docId = Date.now();
    const document = {
      id: docId,
      filename: req.file.originalname,
      path: req.file.path,
      createdAt: Date.now()
    };

    if (!data.requirements[index].documents) {
      data.requirements[index].documents = [];
    }
    data.requirements[index].documents.push(document);
    writeJsonFile(testRequirementsPath, data);
    res.json(document);
  });

  // DELETE /api/requirements/:id/documents/:docId - 删除文档
  app.delete('/api/requirements/:id/documents/:docId', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const docId = parseInt(req.params.docId, 10);
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const index = data.requirements.findIndex(r => r.id === id);
    if (index === -1) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const docIndex = data.requirements[index].documents?.findIndex(d => d.id === docId);
    if (docIndex === -1 || docIndex === undefined) {
      return res.status(404).json({ error: 'Document not found' });
    }

    data.requirements[index].documents.splice(docIndex, 1);
    writeJsonFile(testRequirementsPath, data);
    res.json({ success: true });
  });

  // GET /api/requirements/:id/documents/:docId/download - 下载文档
  app.get('/api/requirements/:id/documents/:docId/download', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const docId = parseInt(req.params.docId, 10);
    const data = readJsonFile(testRequirementsPath);
    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const requirement = data.requirements.find(r => r.id === id);
    if (!requirement) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const document = requirement.documents?.find(d => d.id === docId);
    if (!document) {
      return res.status(404).json({ error: 'Document not found' });
    }

    res.download(document.path, document.filename);
  });

  // POST /api/requirements/:id/split-tasks - 拆分任务
  app.post('/api/requirements/:id/split-tasks', (req, res) => {
    const id = parseInt(req.params.id, 10);
    const data = readJsonFile(testRequirementsPath);
    const taskData = readJsonFile(testTasksPath) || { tasks: [] };

    if (!data) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    const requirement = data.requirements.find(r => r.id === id);
    if (!requirement) {
      return res.status(404).json({ error: 'Requirement not found' });
    }

    // 模拟拆分，创建一个任务
    const newTaskId = taskData.tasks.length > 0
      ? Math.max(...taskData.tasks.map(t => t.id)) + 1
      : 1;

    const newTask = {
      id: newTaskId,
      title: `任务：${requirement.title}`,
      description: requirement.description,
      status: 'pending',
      priority: requirement.priority || 'medium',
      requirementId: id,
      subtasks: [],
      dependencies: [],
      createdAt: Date.now(),
      updatedAt: Date.now()
    };

    taskData.tasks.push(newTask);
    writeJsonFile(testTasksPath, taskData);
    res.json({ tasks: [newTask] });
  });

  return app;
}

export {
  createTestApp,
  initTestData,
  cleanupTestData,
  testTasksPath,
  testRequirementsPath,
  testMembersPath,
  testAssignmentsPath,
  testCommentsPath,
  testActivitiesPath
};
