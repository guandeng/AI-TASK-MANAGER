/**
 * 评论 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Comment API', () => {
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

  describe('GET /api/tasks/:taskId/comments', () => {
    it('应该返回任务的评论列表', async () => {
      const res = await request(app).get('/api/tasks/1/comments');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的评论应该都属于指定任务', async () => {
      const res = await request(app).get('/api/tasks/1/comments');
      expect(res.status).toBe(200);
      res.body.forEach(comment => {
        expect(comment.taskId).toBe(1);
      });
    });

    it('不存在的任务应该返回空数组', async () => {
      const res = await request(app).get('/api/tasks/99999/comments');
      expect(res.status).toBe(200);
      expect(res.body).toHaveLength(0);
    });
  });

  describe('GET /api/tasks/:taskId/comments/tree', () => {
    it('应该返回评论树结构', async () => {
      const res = await request(app).get('/api/tasks/1/comments/tree');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的评论应该是顶级评论（没有父评论）', async () => {
      const res = await request(app).get('/api/tasks/1/comments/tree');
      expect(res.status).toBe(200);
      res.body.forEach(comment => {
        expect(comment.parentId).toBeNull();
      });
    });
  });

  describe('GET /api/tasks/:taskId/comments/statistics', () => {
    it('应该返回评论统计', async () => {
      const res = await request(app).get('/api/tasks/1/comments/statistics');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('total');
      expect(res.body.total).toBeGreaterThanOrEqual(0);
    });

    it('评论数量应该正确', async () => {
      // 先创建评论
      await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '测试评论', authorId: 1 });

      const res = await request(app).get('/api/tasks/1/comments/statistics');
      expect(res.status).toBe(200);
      expect(res.body.total).toBeGreaterThan(0);
    });
  });

  describe('GET /api/tasks/:taskId/comments/:commentId', () => {
    it('应该返回评论详情', async () => {
      const res = await request(app).get('/api/tasks/1/comments/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id', 1);
      expect(res.body).toHaveProperty('content');
      expect(res.body).toHaveProperty('authorId');
    });

    it('应该返回404当评论不存在时', async () => {
      const res = await request(app).get('/api/tasks/1/comments/99999');
      expect(res.status).toBe(404);
      expect(res.body).toHaveProperty('error');
    });
  });

  describe('POST /api/tasks/:taskId/comments', () => {
    it('应该创建评论', async () => {
      const res = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '这是一条新评论', authorId: 1 });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.content).toBe('这是一条新评论');
      expect(res.body.authorId).toBe(1);
      expect(res.body.taskId).toBe(1);
    });

    it('创建的评论应该包含时间戳', async () => {
      const res = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '时间戳测试', authorId: 1 });
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('createdAt');
      expect(res.body).toHaveProperty('updatedAt');
    });

    it('创建的评论应该出现在列表中', async () => {
      const createRes = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '列表测试评论', authorId: 1 });
      expect(createRes.status).toBe(200);

      const listRes = await request(app).get('/api/tasks/1/comments');
      expect(listRes.status).toBe(200);
      const found = listRes.body.find(c => c.id === createRes.body.id);
      expect(found).toBeDefined();
    });

    it('应该创建回复评论', async () => {
      const res = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '这是一条回复', authorId: 1, parentId: 1 });
      expect(res.status).toBe(200);
      expect(res.body.parentId).toBe(1);
    });
  });

  describe('PUT /api/tasks/:taskId/comments/:commentId', () => {
    it('应该更新评论内容', async () => {
      const res = await request(app)
        .put('/api/tasks/1/comments/1')
        .send({ content: '更新后的评论内容' });
      expect(res.status).toBe(200);
      expect(res.body.content).toBe('更新后的评论内容');
    });

    it('应该返回404当评论不存在时', async () => {
      const res = await request(app)
        .put('/api/tasks/1/comments/99999')
        .send({ content: '更新内容' });
      expect(res.status).toBe(404);
    });

    it('应该更新 updatedAt 时间戳', async () => {
      const beforeRes = await request(app).get('/api/tasks/1/comments/1');
      const beforeTime = beforeRes.body.updatedAt;

      await new Promise(resolve => setTimeout(resolve, 10));

      const res = await request(app)
        .put('/api/tasks/1/comments/1')
        .send({ content: '时间戳更新测试' });
      expect(res.status).toBe(200);
      expect(res.body.updatedAt).toBeGreaterThanOrEqual(beforeTime);
    });
  });

  describe('DELETE /api/tasks/:taskId/comments/:commentId', () => {
    it('应该删除评论', async () => {
      // 先创建评论
      const createRes = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '待删除评论', authorId: 1 });
      const commentId = createRes.body.id;

      const res = await request(app).delete(`/api/tasks/1/comments/${commentId}`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当删除不存在的评论时', async () => {
      const res = await request(app).delete('/api/tasks/1/comments/99999');
      expect(res.status).toBe(404);
    });

    it('删除后评论应该不存在', async () => {
      // 先创建评论
      const createRes = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '删除测试评论', authorId: 1 });
      const commentId = createRes.body.id;

      // 删除
      await request(app).delete(`/api/tasks/1/comments/${commentId}`);

      // 验证不存在
      const res = await request(app).get(`/api/tasks/1/comments/${commentId}`);
      expect(res.status).toBe(404);
    });
  });

  describe('GET /api/tasks/:taskId/comments/:commentId/replies', () => {
    it('应该返回回复列表', async () => {
      const res = await request(app).get('/api/tasks/1/comments/1/replies');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('返回的回复应该都是指定评论的子评论', async () => {
      // 先创建回复
      await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '回复内容', authorId: 1, parentId: 1 });

      const res = await request(app).get('/api/tasks/1/comments/1/replies');
      expect(res.status).toBe(200);
      res.body.forEach(reply => {
        expect(reply.parentId).toBe(1);
      });
    });

    it('没有回复时应该返回空数组', async () => {
      // 创建一个新评论（没有回复）
      const createRes = await request(app)
        .post('/api/tasks/1/comments')
        .send({ content: '无回复评论', authorId: 1 });
      const commentId = createRes.body.id;

      const res = await request(app).get(`/api/tasks/1/comments/${commentId}/replies`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveLength(0);
    });
  });
});
