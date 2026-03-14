/**
 * 任务 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Task API', () => {
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

  describe('GET /api/tasks', () => {
    it('应该返回任务列表', async () => {
      const res = await request(app).get('/api/tasks');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('tasks');
      expect(Array.isArray(res.body.tasks)).toBe(true);
    });

    it('应该按需求ID筛选任务', async () => {
      const res = await request(app).get('/api/tasks?requirementId=1');
      expect(res.status).toBe(200);
      res.body.tasks.forEach(task => {
        expect(task.requirementId).toBe(1);
      });
    });

    it('应该返回空数组当没有匹配的任务时', async () => {
      const res = await request(app).get('/api/tasks?requirementId=99999');
      expect(res.status).toBe(200);
      expect(res.body.tasks).toHaveLength(0);
    });
  });

  describe('GET /api/tasks/:taskId', () => {
    it('应该返回任务详情', async () => {
      const res = await request(app).get('/api/tasks/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id', 1);
      expect(res.body).toHaveProperty('title');
      expect(res.body).toHaveProperty('status');
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app).get('/api/tasks/99999');
      expect(res.status).toBe(404);
      expect(res.body).toHaveProperty('error');
    });

    it('应该包含子任务列表', async () => {
      const res = await request(app).get('/api/tasks/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('subtasks');
      expect(Array.isArray(res.body.subtasks)).toBe(true);
    });
  });

  describe('PUT /api/tasks/:taskId', () => {
    it('应该更新任务状态', async () => {
      const res = await request(app)
        .put('/api/tasks/1')
        .send({ status: 'done' });
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('done');
    });

    it('应该更新任务标题', async () => {
      const res = await request(app)
        .put('/api/tasks/1')
        .send({ title: '更新后的标题' });
      expect(res.status).toBe(200);
      expect(res.body.title).toBe('更新后的标题');
    });

    it('应该更新任务优先级', async () => {
      const res = await request(app)
        .put('/api/tasks/1')
        .send({ priority: 'low' });
      expect(res.status).toBe(200);
      expect(res.body.priority).toBe('low');
    });

    it('应该更新任务负责人', async () => {
      const res = await request(app)
        .put('/api/tasks/1')
        .send({ assignee: 'newUser' });
      expect(res.status).toBe(200);
      expect(res.body.assignee).toBe('newUser');
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/99999')
        .send({ status: 'done' });
      expect(res.status).toBe(404);
    });

    it('应该更新 updatedAt 时间戳', async () => {
      const beforeRes = await request(app).get('/api/tasks/1');
      const beforeTime = beforeRes.body.updatedAt;

      // 等待一小段时间确保时间戳不同
      await new Promise(resolve => setTimeout(resolve, 10));

      const res = await request(app)
        .put('/api/tasks/1')
        .send({ status: 'pending' });
      expect(res.status).toBe(200);
      expect(res.body.updatedAt).toBeGreaterThanOrEqual(beforeTime);
    });
  });

  describe('DELETE /api/tasks/:taskId', () => {
    it('应该删除已存在的任务', async () => {
      const res = await request(app).delete('/api/tasks/2');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当删除不存在的任务时', async () => {
      const res = await request(app).delete('/api/tasks/99999');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/tasks/:taskId/copy', () => {
    it('应该复制任务', async () => {
      const res = await request(app).post('/api/tasks/1/copy');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.id).not.toBe(1);
      expect(res.body.title).toContain('副本');
    });

    it('复制后状态应该重置为pending', async () => {
      const res = await request(app).post('/api/tasks/1/copy');
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('pending');
    });

    it('应该返回404当复制不存在的任务时', async () => {
      const res = await request(app).post('/api/tasks/99999/copy');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/tasks/batch-delete', () => {
    it('应该批量删除任务', async () => {
      const res = await request(app)
        .post('/api/tasks/batch-delete')
        .send({ taskIds: [1, 2] });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该处理空数组', async () => {
      const res = await request(app)
        .post('/api/tasks/batch-delete')
        .send({ taskIds: [] });
      expect(res.status).toBe(200);
      expect(res.body.deletedCount).toBe(0);
    });
  });

  describe('PUT /api/tasks/:taskId/time', () => {
    it('应该更新预估时间', async () => {
      const res = await request(app)
        .put('/api/tasks/1/time')
        .send({ estimatedHours: 8 });
      expect(res.status).toBe(200);
      expect(res.body.estimatedHours).toBe(8);
    });

    it('应该更新实际时间', async () => {
      const res = await request(app)
        .put('/api/tasks/1/time')
        .send({ actualHours: 6 });
      expect(res.status).toBe(200);
      expect(res.body.actualHours).toBe(6);
    });

    it('应该同时更新预估和实际时间', async () => {
      const res = await request(app)
        .put('/api/tasks/1/time')
        .send({ estimatedHours: 10, actualHours: 12 });
      expect(res.status).toBe(200);
      expect(res.body.estimatedHours).toBe(10);
      expect(res.body.actualHours).toBe(12);
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/99999/time')
        .send({ estimatedHours: 8 });
      expect(res.status).toBe(404);
    });
  });
});
