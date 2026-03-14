/**
 * 分配 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Assignment API', () => {
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

  describe('GET /api/tasks/:taskId/assignments', () => {
    it('应该返回任务的分配列表', async () => {
      const res = await request(app).get('/api/tasks/1/assignments');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的分配应该只包含任务级别的分配', async () => {
      const res = await request(app).get('/api/tasks/1/assignments');
      expect(res.status).toBe(200);
      res.body.forEach(assignment => {
        expect(assignment.taskId).toBe(1);
        expect(assignment.subtaskId).toBeNull();
      });
    });

    it('不存在的任务应该返回空数组', async () => {
      const res = await request(app).get('/api/tasks/99999/assignments');
      expect(res.status).toBe(200);
      expect(res.body).toHaveLength(0);
    });
  });

  describe('GET /api/tasks/:taskId/assignments/overview', () => {
    it('应该返回分配概览', async () => {
      const res = await request(app).get('/api/tasks/1/assignments/overview');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('taskAssignments');
      expect(res.body).toHaveProperty('subtaskAssignments');
      expect(Array.isArray(res.body.taskAssignments)).toBe(true);
    });

    it('taskAssignments 应该包含任务级别的分配', async () => {
      const res = await request(app).get('/api/tasks/1/assignments/overview');
      expect(res.status).toBe(200);
      res.body.taskAssignments.forEach(assignment => {
        expect(assignment.subtaskId).toBeNull();
      });
    });
  });

  describe('POST /api/tasks/:taskId/assignments', () => {
    it('应该分配任务给成员', async () => {
      const res = await request(app)
        .post('/api/tasks/1/assignments')
        .send({ memberId: 1 });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.taskId).toBe(1);
      expect(res.body.memberId).toBe(1);
      expect(res.body.subtaskId).toBeNull();
    });

    it('应该分配任务并指定工时', async () => {
      const res = await request(app)
        .post('/api/tasks/1/assignments')
        .send({ memberId: 2, hours: 8 });
      expect(res.status).toBe(200);
      expect(res.body.hours).toBe(8);
    });

    it('分配后应该出现在分配列表中', async () => {
      const createRes = await request(app)
        .post('/api/tasks/1/assignments')
        .send({ memberId: 1 });
      expect(createRes.status).toBe(200);

      const listRes = await request(app).get('/api/tasks/1/assignments');
      expect(listRes.status).toBe(200);
      const found = listRes.body.find(a => a.id === createRes.body.id);
      expect(found).toBeDefined();
    });
  });

  describe('DELETE /api/tasks/:taskId/assignments/:assignmentId', () => {
    it('应该取消分配', async () => {
      // 先创建分配
      const createRes = await request(app)
        .post('/api/tasks/1/assignments')
        .send({ memberId: 1 });
      const assignmentId = createRes.body.id;

      const res = await request(app).delete(`/api/tasks/1/assignments/${assignmentId}`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当分配不存在时', async () => {
      const res = await request(app).delete('/api/tasks/1/assignments/99999');
      expect(res.status).toBe(404);
    });
  });

  describe('GET /api/tasks/:taskId/subtasks/:subtaskId/assignments', () => {
    it('应该返回子任务的分配列表', async () => {
      const res = await request(app).get('/api/tasks/1/subtasks/1/assignments');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的分配应该只包含指定子任务的分配', async () => {
      const res = await request(app).get('/api/tasks/1/subtasks/1/assignments');
      expect(res.status).toBe(200);
      res.body.forEach(assignment => {
        expect(assignment.taskId).toBe(1);
        expect(assignment.subtaskId).toBe(1);
      });
    });
  });

  describe('POST /api/tasks/:taskId/subtasks/:subtaskId/assignments', () => {
    it('应该分配子任务给成员', async () => {
      const res = await request(app)
        .post('/api/tasks/1/subtasks/1/assignments')
        .send({ memberId: 1 });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.taskId).toBe(1);
      expect(res.body.subtaskId).toBe(1);
      expect(res.body.memberId).toBe(1);
    });

    it('应该分配子任务并指定工时', async () => {
      const res = await request(app)
        .post('/api/tasks/1/subtasks/1/assignments')
        .send({ memberId: 2, hours: 4 });
      expect(res.status).toBe(200);
      expect(res.body.hours).toBe(4);
    });
  });

  describe('DELETE /api/tasks/:taskId/subtasks/:subtaskId/assignments/:assignmentId', () => {
    it('应该取消子任务分配', async () => {
      // 先创建分配
      const createRes = await request(app)
        .post('/api/tasks/1/subtasks/1/assignments')
        .send({ memberId: 1 });
      const assignmentId = createRes.body.id;

      const res = await request(app).delete(
        `/api/tasks/1/subtasks/1/assignments/${assignmentId}`
      );
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当分配不存在时', async () => {
      const res = await request(app).delete(
        '/api/tasks/1/subtasks/1/assignments/99999'
      );
      expect(res.status).toBe(404);
    });
  });

  describe('GET /api/members/:memberId/assignments', () => {
    it('应该返回成员的分配列表', async () => {
      const res = await request(app).get('/api/members/1/assignments');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的分配应该都属于指定成员', async () => {
      const res = await request(app).get('/api/members/1/assignments');
      expect(res.status).toBe(200);
      res.body.forEach(assignment => {
        expect(assignment.memberId).toBe(1);
      });
    });
  });

  describe('GET /api/members/:memberId/workload', () => {
    it('应该返回成员工作负载', async () => {
      const res = await request(app).get('/api/members/1/workload');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('totalHours');
      expect(res.body).toHaveProperty('assignmentCount');
    });

    it('工作负载应该是非负数', async () => {
      const res = await request(app).get('/api/members/1/workload');
      expect(res.status).toBe(200);
      expect(res.body.totalHours).toBeGreaterThanOrEqual(0);
      expect(res.body.assignmentCount).toBeGreaterThanOrEqual(0);
    });

    it('分配后工作负载应该增加', async () => {
      const beforeRes = await request(app).get('/api/members/1/workload');
      const beforeHours = beforeRes.body.totalHours;

      await request(app)
        .post('/api/tasks/1/assignments')
        .send({ memberId: 1, hours: 8 });

      const afterRes = await request(app).get('/api/members/1/workload');
      expect(afterRes.body.totalHours).toBe(beforeHours + 8);
    });
  });
});
