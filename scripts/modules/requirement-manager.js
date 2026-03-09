import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import {
  getRequirementList,
  getRequirementById,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  addDocument,
  deleteDocument,
  getRequirementStatistics
} from './requirement-storage.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 上传目录
const UPLOAD_DIR = path.resolve(process.env.UPLOAD_DIR || path.join(__dirname, '../../uploads'));

// 确保上传目录存在
function ensureUploadDir() {
  if (!fs.existsSync(UPLOAD_DIR)) {
    fs.mkdirSync(UPLOAD_DIR, { recursive: true });
  }
  return UPLOAD_DIR;
}

/**
 * 获取需求列表
 */
export async function listRequirements(filters = {}) {
  return getRequirementList(filters);
}

/**
 * 获取需求详情
 */
export async function getRequirement(id) {
  return getRequirementById(id);
}

/**
 * 创建新需求
 */
export async function createNewRequirement(data) {
  if (!data.title || !data.title.trim()) {
    throw new Error('需求标题不能为空');
  }

  return createRequirement({
    title: data.title.trim(),
    content: data.content || '',
    status: data.status || 'draft',
    priority: data.priority || 'medium',
    tags: data.tags || [],
    assignee: data.assignee || null
  });
}

/**
 * 更新需求
 */
export async function updateRequirementById(id, data) {
  const existing = await getRequirementById(id);
  if (!existing) {
    throw new Error('需求不存在');
  }

  return updateRequirement(id, data);
}

/**
 * 删除需求
 */
export async function deleteRequirementById(id) {
  const existing = await getRequirementById(id);
  if (!existing) {
    throw new Error('需求不存在');
  }

  return deleteRequirement(id);
}

/**
 * 上传文档
 */
export async function uploadDocument(requirementId, file) {
  const existing = await getRequirementById(requirementId);
  if (!existing) {
    throw new Error('需求不存在');
  }

  ensureUploadDir();

  // 生成唯一文件名
  const ext = path.extname(file.originalname || file.name);
  const timestamp = Date.now();
  const randomStr = Math.random().toString(36).substring(2, 8);
  const filename = `req_${requirementId}_${timestamp}_${randomStr}${ext}`;
  const filepath = path.join(UPLOAD_DIR, filename);

  // 写入文件
  if (file.buffer) {
    fs.writeFileSync(filepath, file.buffer);
  } else if (file.path) {
    fs.copyFileSync(file.path, filepath);
  } else {
    throw new Error('无效的文件数据');
  }

  // 保存文档记录
  const doc = await addDocument(requirementId, {
    name: file.originalname || file.name,
    path: filepath,
    size: file.size || 0,
    mimeType: file.mimetype || 'application/octet-stream',
    uploadedBy: file.uploadedBy || null
  });

  return doc;
}

/**
 * 删除文档
 */
export async function removeDocument(requirementId, documentId) {
  return deleteDocument(requirementId, documentId);
}

/**
 * 获取统计数据
 */
export async function getStatistics() {
  return getRequirementStatistics();
}

export {
  UPLOAD_DIR
};
