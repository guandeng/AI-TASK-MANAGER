/**
 * 需求 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';
import fs from 'fs';
import path from 'path';

let app;

describe('Requirement API', () => {
  beforeAll(() => {
    app = createTestApp();
  });

  afterAll(() => {
    cleanupTestData();
  });

  // 每个测试前重置数据
  beforeEach(() => {
    initTestData();
  });

  describe('GET /api/requirements', () => {
    it('应该返回需求列表', async () => {
      const res = await request(app).get('/api/requirements');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('需求应该包含必要字段', async () => {
      const res = await request(app).get('/api/requirements');
      expect(res.status).toBe(200);
      if (res.body.length > 0) {
        const requirement = res.body[0];
        expect(requirement).toHaveProperty('id');
        expect(requirement).toHaveProperty('title');
        expect(requirement).toHaveProperty('status');
      }
    });
  });

  describe('GET /api/requirements/statistics', () => {
    it('应该返回需求统计', async () => {
      const res = await request(app).get('/api/requirements/statistics');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('total');
      expect(res.body).toHaveProperty('pending');
      expect(res.body).toHaveProperty('done');
    });

    it('统计数字应该正确', async () => {
      const res = await request(app).get('/api/requirements/statistics');
      expect(res.status).toBe(200);
      expect(res.body.total).toBeGreaterThanOrEqual(0);
      expect(res.body.pending).toBeGreaterThanOrEqual(0);
      expect(res.body.done).toBeGreaterThanOrEqual(0);
    });
  });

  describe('GET /api/requirements/:id', () => {
    it('应该返回需求详情', async () => {
      const res = await request(app).get('/api/requirements/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id', 1);
      expect(res.body).toHaveProperty('title');
      expect(res.body).toHaveProperty('description');
    });

    it('应该返回404当需求不存在时', async () => {
      const res = await request(app).get('/api/requirements/99999');
      expect(res.status).toBe(404);
      expect(res.body).toHaveProperty('error');
    });

    it('需求详情应该包含文档列表', async () => {
      const res = await request(app).get('/api/requirements/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('documents');
      expect(Array.isArray(res.body.documents)).toBe(true);
    });
  });

  describe('POST /api/requirements', () => {
    it('应该创建新需求', async () => {
      const newRequirement = {
        title: '新测试需求',
        description: '新测试需求描述',
        priority: 'high'
      };
      const res = await request(app)
        .post('/api/requirements')
        .send(newRequirement);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.title).toBe(newRequirement.title);
      expect(res.body.description).toBe(newRequirement.description);
    });

    it('新需求默认状态应该是pending', async () => {
      const res = await request(app)
        .post('/api/requirements')
        .send({ title: '状态测试需求' });
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('pending');
    });

    it('创建的需求应该包含时间戳', async () => {
      const res = await request(app)
        .post('/api/requirements')
        .send({ title: '时间戳测试' });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('createdAt');
      expect(res.body).toHaveProperty('updatedAt');
    });

    it('创建的需求应该出现在列表中', async () => {
      const newReq = { title: '列表验证需求' };
      const createRes = await request(app)
        .post('/api/requirements')
        .send(newReq);
      expect(createRes.status).toBe(200);

      const listRes = await request(app).get('/api/requirements');
      expect(listRes.status).toBe(200);
      const found = listRes.body.find(r => r.id === createRes.body.id);
      expect(found).toBeDefined();
    });
  });

  describe('PUT /api/requirements/:id', () => {
    it('应该更新需求标题', async () => {
      const res = await request(app)
        .put('/api/requirements/1')
        .send({ title: '更新后的标题' });
      expect(res.status).toBe(200);
      expect(res.body.title).toBe('更新后的标题');
    });

    it('应该更新需求状态', async () => {
      const res = await request(app)
        .put('/api/requirements/1')
        .send({ status: 'done' });
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('done');
    });

    it('应该更新需求描述', async () => {
      const res = await request(app)
        .put('/api/requirements/1')
        .send({ description: '更新后的描述' });
      expect(res.status).toBe(200);
      expect(res.body.description).toBe('更新后的描述');
    });

    it('应该更新需求优先级', async () => {
      const res = await request(app)
        .put('/api/requirements/1')
        .send({ priority: 'high' });
      expect(res.status).toBe(200);
      expect(res.body.priority).toBe('high');
    });

    it('应该返回404当需求不存在时', async () => {
      const res = await request(app)
        .put('/api/requirements/99999')
        .send({ title: '不存在' });
      expect(res.status).toBe(404);
    });

    it('应该更新 updatedAt 时间戳', async () => {
      const beforeRes = await request(app).get('/api/requirements/1');
      const beforeTime = beforeRes.body.updatedAt;

      await new Promise(resolve => setTimeout(resolve, 10));

      const res = await request(app)
        .put('/api/requirements/1')
        .send({ title: '时间戳测试' });
      expect(res.status).toBe(200);
      expect(res.body.updatedAt).toBeGreaterThanOrEqual(beforeTime);
    });
  });

  describe('DELETE /api/requirements/:id', () => {
    it('应该删除需求', async () => {
      // 先创建需求用于删除
      const createRes = await request(app)
        .post('/api/requirements')
        .send({ title: '待删除需求' });
      const reqId = createRes.body.id;

      const res = await request(app).delete(`/api/requirements/${reqId}`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当删除不存在的需求时', async () => {
      const res = await request(app).delete('/api/requirements/99999');
      expect(res.status).toBe(404);
    });

    it('删除后需求应该不存在', async () => {
      // 先创建需求
      const createRes = await request(app)
        .post('/api/requirements')
        .send({ title: '删除验证需求' });
      const reqId = createRes.body.id;

      // 删除
      await request(app).delete(`/api/requirements/${reqId}`);

      // 验证不存在
      const res = await request(app).get(`/api/requirements/${reqId}`);
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/requirements/:id/documents', () => {
    it('应该上传文档', async () => {
      // 创建测试文件
      const testDir = path.join(process.cwd(), 'tests/test-data/uploads');
      if (!fs.existsSync(testDir)) {
        fs.mkdirSync(testDir, { recursive: true });
      }
      const testFile = path.join(testDir, 'test-doc.txt');
      fs.writeFileSync(testFile, '测试文档内容');

      const res = await request(app)
        .post('/api/requirements/1/documents')
        .attach('file', testFile);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body).toHaveProperty('filename');
    });

    it('没有文件应该返回错误', async () => {
      const res = await request(app)
        .post('/api/requirements/1/documents');
      expect(res.status).toBe(400);
    });

    it('应该返回404当需求不存在时', async () => {
      const testFile = path.join(process.cwd(), 'tests/test-data/test-tasks.json');
      const res = await request(app)
        .post('/api/requirements/99999/documents')
        .attach('file', testFile);
      expect(res.status).toBe(404);
    });
  });

  describe('DELETE /api/requirements/:id/documents/:docId', () => {
    it('应该删除文档', async () => {
      // 先上传文档
      const testFile = path.join(process.cwd(), 'tests/test-data/test-tasks.json');
      const uploadRes = await request(app)
        .post('/api/requirements/1/documents')
        .attach('file', testFile);
      const docId = uploadRes.body.id;

      const res = await request(app).delete(`/api/requirements/1/documents/${docId}`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当文档不存在时', async () => {
      const res = await request(app).delete('/api/requirements/1/documents/99999');
      expect(res.status).toBe(404);
    });

    it('应该返回404当需求不存在时', async () => {
      const res = await request(app).delete('/api/requirements/99999/documents/1');
      expect(res.status).toBe(404);
    });
  });

  describe('GET /api/requirements/:id/documents/:docId/download', () => {
    it('应该返回404当需求不存在时', async () => {
      const res = await request(app).get('/api/requirements/99999/documents/1/download');
      expect(res.status).toBe(404);
    });

    it('应该返回404当文档不存在时', async () => {
      const res = await request(app).get('/api/requirements/1/documents/99999/download');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/requirements/:id/split-tasks', () => {
    it('应该拆分需求为任务', async () => {
      const res = await request(app).post('/api/requirements/1/split-tasks');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('tasks');
      expect(Array.isArray(res.body.tasks)).toBe(true);
      expect(res.body.tasks.length).toBeGreaterThan(0);
    });

    it('拆分后的任务应该关联到需求', async () => {
      const res = await request(app).post('/api/requirements/1/split-tasks');
      expect(res.status).toBe(200);
      res.body.tasks.forEach(task => {
        expect(task.requirementId).toBe(1);
      });
    });

    it('应该返回404当需求不存在时', async () => {
      const res = await request(app).post('/api/requirements/99999/split-tasks');
      expect(res.status).toBe(404);
    });

    it('拆分后的任务应该包含必要字段', async () => {
      const res = await request(app).post('/api/requirements/1/split-tasks');
      expect(res.status).toBe(200);
      const task = res.body.tasks[0];
      expect(task).toHaveProperty('id');
      expect(task).toHaveProperty('title');
      expect(task).toHaveProperty('status');
      expect(task).toHaveProperty('requirementId');
    });
  });
});
