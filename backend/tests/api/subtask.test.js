/**
 * 子任务 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Subtask API', () => {
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

  describe('PUT /api/tasks/:taskId/subtasks/:subtaskId', () => {
    it('应该更新子任务状态', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/1')
        .send({ status: 'done' });
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('done');
    });

    it('应该更新子任务标题', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/1')
        .send({ title: '更新后的子任务标题' });
      expect(res.status).toBe(200);
      expect(res.body.title).toBe('更新后的子任务标题');
    });

    it('应该更新子任务描述', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/1')
        .send({ description: '新的描述' });
      expect(res.status).toBe(200);
      expect(res.body.description).toBe('新的描述');
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/99999/subtasks/1')
        .send({ status: 'done' });
      expect(res.status).toBe(404);
    });

    it('应该返回404当子任务不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/99999')
        .send({ status: 'done' });
      expect(res.status).toBe(404);
    });
  });

  describe('PUT /api/tasks/:taskId/subtasks/reorder', () => {
    it('应该重新排序子任务', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/reorder')
        .send({ subtaskIds: [2, 1] });
      expect(res.status).toBe(200);
      expect(res.body.subtasks).toBeDefined();
      expect(res.body.subtasks[0].id).toBe(2);
      expect(res.body.subtasks[1].id).toBe(1);
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/99999/subtasks/reorder')
        .send({ subtaskIds: [1, 2] });
      expect(res.status).toBe(404);
    });

    it('应该返回400当任务没有子任务时', async () => {
      // 任务2没有子任务
      const res = await request(app)
        .put('/api/tasks/2/subtasks/reorder')
        .send({ subtaskIds: [1, 2] });
      expect(res.status).toBe(400);
    });

    it('应该正确设置子任务顺序', async () => {
      const res = await request(app)
        .put('/api/tasks/1/subtasks/reorder')
        .send({ subtaskIds: [1, 2] });
      expect(res.status).toBe(200);
      expect(res.body.subtasks[0].order).toBe(0);
      expect(res.body.subtasks[1].order).toBe(1);
    });
  });

  describe('DELETE /api/tasks/:taskId/subtasks/:subtaskId', () => {
    it('应该删除子任务', async () => {
      const res = await request(app).delete('/api/tasks/1/subtasks/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app).delete('/api/tasks/99999/subtasks/1');
      expect(res.status).toBe(404);
    });

    it('应该返回404当子任务不存在时', async () => {
      const res = await request(app).delete('/api/tasks/1/subtasks/99999');
      expect(res.status).toBe(404);
    });
  });

  describe('DELETE /api/tasks/:taskId/subtasks', () => {
    it('应该删除任务的所有子任务', async () => {
      const res = await request(app).delete('/api/tasks/1/subtasks');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app).delete('/api/tasks/99999/subtasks');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/tasks/:taskId/subtasks/:subtaskId/regenerate', () => {
    it('应该重新生成子任务', async () => {
      const res = await request(app).post('/api/tasks/1/subtasks/1/regenerate');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id', 1);
      expect(res.body.title).toContain('重新生成');
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app).post('/api/tasks/99999/subtasks/1/regenerate');
      expect(res.status).toBe(404);
    });

    it('应该返回404当子任务不存在时', async () => {
      const res = await request(app).post('/api/tasks/1/subtasks/99999/regenerate');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/tasks/:taskId/expand', () => {
    it('应该展开任务并生成子任务', async () => {
      const res = await request(app).post('/api/tasks/1/expand');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('subtasks');
      expect(res.body.subtasks.length).toBeGreaterThan(0);
    });

    it('应该返回404当任务不存在时', async () => {
      const res = await request(app).post('/api/tasks/99999/expand');
      expect(res.status).toBe(404);
    });
  });
});
