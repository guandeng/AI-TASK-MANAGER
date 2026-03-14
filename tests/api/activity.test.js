/**
 * 活动 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Activity API', () => {
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

  describe('GET /api/tasks/:taskId/activities', () => {
    it('应该返回任务的活动列表', async () => {
      const res = await request(app).get('/api/tasks/1/activities');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的活动应该都属于指定任务', async () => {
      const res = await request(app).get('/api/tasks/1/activities');
      expect(res.status).toBe(200);
      res.body.forEach(activity => {
        expect(activity.taskId).toBe(1);
      });
    });

    it('不存在的任务应该返回空数组', async () => {
      const res = await request(app).get('/api/tasks/99999/activities');
      expect(res.status).toBe(200);
      expect(res.body).toHaveLength(0);
    });
  });

  describe('GET /api/activities', () => {
    it('应该返回全局活动列表', async () => {
      const res = await request(app).get('/api/activities');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('活动应该包含必要字段', async () => {
      const res = await request(app).get('/api/activities');
      expect(res.status).toBe(200);
      if (res.body.length > 0) {
        const activity = res.body[0];
        expect(activity).toHaveProperty('id');
        expect(activity).toHaveProperty('taskId');
        expect(activity).toHaveProperty('type');
        expect(activity).toHaveProperty('description');
        expect(activity).toHaveProperty('createdAt');
      }
    });
  });

  describe('GET /api/activities/statistics', () => {
    it('应该返回活动统计', async () => {
      const res = await request(app).get('/api/activities/statistics');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('total');
      expect(res.body.total).toBeGreaterThanOrEqual(0);
    });

    it('统计数据应该是有效的', async () => {
      const res = await request(app).get('/api/activities/statistics');
      expect(res.status).toBe(200);
      expect(typeof res.body.total).toBe('number');
    });
  });
});
