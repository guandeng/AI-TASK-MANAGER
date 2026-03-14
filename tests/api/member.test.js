/**
 * 成员 API 单元测试
 */
import request from 'supertest';
import { createTestApp, initTestData, cleanupTestData } from './test-helpers.js';

let app;

describe('Member API', () => {
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

  describe('GET /api/members', () => {
    it('应该返回成员列表', async () => {
      const res = await request(app).get('/api/members');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('成员应该包含必要字段', async () => {
      const res = await request(app).get('/api/members');
      expect(res.status).toBe(200);
      if (res.body.length > 0) {
        const member = res.body[0];
        expect(member).toHaveProperty('id');
        expect(member).toHaveProperty('name');
        expect(member).toHaveProperty('role');
        expect(member).toHaveProperty('status');
      }
    });
  });

  describe('GET /api/members/statistics', () => {
    it('应该返回成员统计', async () => {
      const res = await request(app).get('/api/members/statistics');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('total');
      expect(res.body).toHaveProperty('active');
      expect(res.body).toHaveProperty('inactive');
      expect(res.body).toHaveProperty('byDepartment');
    });

    it('统计数字应该正确', async () => {
      const res = await request(app).get('/api/members/statistics');
      expect(res.status).toBe(200);
      expect(res.body.total).toBeGreaterThanOrEqual(0);
      expect(res.body.active).toBeGreaterThanOrEqual(0);
      expect(res.body.inactive).toBeGreaterThanOrEqual(0);
    });
  });

  describe('GET /api/members/departments', () => {
    it('应该返回部门列表', async () => {
      const res = await request(app).get('/api/members/departments');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });

    it('部门列表不应该包含重复项', async () => {
      const res = await request(app).get('/api/members/departments');
      expect(res.status).toBe(200);
      const uniqueDepartments = [...new Set(res.body)];
      expect(res.body.length).toBe(uniqueDepartments.length);
    });
  });

  describe('GET /api/members/search', () => {
    it('应该按关键字搜索成员', async () => {
      const res = await request(app).get('/api/members/search?keyword=张');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
      res.body.forEach(member => {
        expect(
          member.name.includes('张') ||
          member.email?.includes('张') ||
          member.department?.includes('张')
        ).toBe(true);
      });
    });

    it('空关键字应该返回所有成员', async () => {
      const res = await request(app).get('/api/members/search?keyword=');
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body)).toBe(true);
    });
  });

  describe('GET /api/members/:id', () => {
    it('应该返回成员详情', async () => {
      const res = await request(app).get('/api/members/1');
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id', 1);
      expect(res.body).toHaveProperty('name');
    });

    it('应该返回404当成员不存在时', async () => {
      const res = await request(app).get('/api/members/99999');
      expect(res.status).toBe(404);
      expect(res.body).toHaveProperty('error');
    });
  });

  describe('POST /api/members', () => {
    it('应该创建新成员', async () => {
      const newMember = {
        name: '测试成员',
        email: 'test@example.com',
        role: 'member',
        department: '测试部'
      };
      const res = await request(app)
        .post('/api/members')
        .send(newMember);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('id');
      expect(res.body.name).toBe(newMember.name);
      expect(res.body.email).toBe(newMember.email);
    });

    it('新成员默认状态应该是active', async () => {
      const res = await request(app)
        .post('/api/members')
        .send({ name: '新成员' });
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('active');
    });

    it('创建的成员应该出现在列表中', async () => {
      const newMember = { name: '列表测试成员' };
      const createRes = await request(app)
        .post('/api/members')
        .send(newMember);
      expect(createRes.status).toBe(200);

      const listRes = await request(app).get('/api/members');
      expect(listRes.status).toBe(200);
      const found = listRes.body.find(m => m.name === newMember.name);
      expect(found).toBeDefined();
    });
  });

  describe('PUT /api/members/:id', () => {
    it('应该更新成员信息', async () => {
      const res = await request(app)
        .put('/api/members/1')
        .send({ name: '更新后的名字' });
      expect(res.status).toBe(200);
      expect(res.body.name).toBe('更新后的名字');
    });

    it('应该更新成员角色', async () => {
      const res = await request(app)
        .put('/api/members/1')
        .send({ role: 'admin' });
      expect(res.status).toBe(200);
      expect(res.body.role).toBe('admin');
    });

    it('应该更新成员部门', async () => {
      const res = await request(app)
        .put('/api/members/1')
        .send({ department: '新部门' });
      expect(res.status).toBe(200);
      expect(res.body.department).toBe('新部门');
    });

    it('应该返回404当成员不存在时', async () => {
      const res = await request(app)
        .put('/api/members/99999')
        .send({ name: '不存在' });
      expect(res.status).toBe(404);
    });

    it('应该更新 updatedAt 时间戳', async () => {
      const beforeRes = await request(app).get('/api/members/1');
      const beforeTime = beforeRes.body.updatedAt;

      await new Promise(resolve => setTimeout(resolve, 10));

      const res = await request(app)
        .put('/api/members/1')
        .send({ name: '时间戳测试' });
      expect(res.status).toBe(200);
      expect(res.body.updatedAt).toBeGreaterThanOrEqual(beforeTime);
    });
  });

  describe('DELETE /api/members/:id', () => {
    it('应该删除成员', async () => {
      // 先创建一个成员用于删除
      const createRes = await request(app)
        .post('/api/members')
        .send({ name: '待删除成员' });
      const memberId = createRes.body.id;

      const res = await request(app).delete(`/api/members/${memberId}`);
      expect(res.status).toBe(200);
      expect(res.body).toHaveProperty('success', true);
    });

    it('应该返回404当删除不存在的成员时', async () => {
      const res = await request(app).delete('/api/members/99999');
      expect(res.status).toBe(404);
    });

    it('删除后成员应该不存在', async () => {
      // 先创建一个成员
      const createRes = await request(app)
        .post('/api/members')
        .send({ name: '删除测试成员' });
      const memberId = createRes.body.id;

      // 删除
      await request(app).delete(`/api/members/${memberId}`);

      // 验证不存在
      const res = await request(app).get(`/api/members/${memberId}`);
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/members/:id/activate', () => {
    it('应该激活成员', async () => {
      const res = await request(app).post('/api/members/1/activate');
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('active');
    });

    it('应该返回404当成员不存在时', async () => {
      const res = await request(app).post('/api/members/99999/activate');
      expect(res.status).toBe(404);
    });
  });

  describe('POST /api/members/:id/deactivate', () => {
    it('应该停用成员', async () => {
      const res = await request(app).post('/api/members/1/deactivate');
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('inactive');
    });

    it('应该返回404当成员不存在时', async () => {
      const res = await request(app).post('/api/members/99999/deactivate');
      expect(res.status).toBe(404);
    });

    it('停用后再激活应该成功', async () => {
      await request(app).post('/api/members/1/deactivate');
      const res = await request(app).post('/api/members/1/activate');
      expect(res.status).toBe(200);
      expect(res.body.status).toBe('active');
    });
  });
});
