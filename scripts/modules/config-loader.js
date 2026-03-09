/**
 * config-loader.js
 * 配置加载模块 - 参考 OpenClaw 配置方式
 * 支持从 task.json 或环境变量加载配置
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import Ajv from 'ajv';
import dotenv from 'dotenv';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 默认配置
const DEFAULT_CONFIG = {
  version: '1.0.0',
  ai: {
    provider: 'qwen',
    providers: {
      qwen: {
        enabled: true,
        apiKey: '',
        model: 'qwen3.5-plus',
        baseUrl: 'https://coding.dashscope.aliyuncs.com/v1'
      },
      gemini: {
        enabled: false,
        apiKey: '',
        model: 'gemini-2.5-flash',
        baseUrl: ''
      },
      perplexity: {
        enabled: false,
        apiKey: '',
        model: 'sonar-pro'
      }
    },
    parameters: {
      maxTokens: 8192,
      temperature: 0.7
    }
  },
  general: {
    debug: false,
    logLevel: 'info',
    defaultSubtasks: 3,
    defaultPriority: 'medium',
    projectName: 'AI Task Manager',
    useChinese: true
  },
  storage: {
    type: 'db',
    database: {
      connection: 'mysql',
      host: 'localhost',
      port: 3306,
      database: 'ai_task',
      username: 'root',
      password: '',
      charset: 'utf8mb4',
      collation: 'utf8mb4_unicode_ci'
    }
  }
};

// 配置缓存
let cachedConfig = null;
let configFilePath = null;
let lastLoadTime = 0;
const CONFIG_TTL = 100; // 配置缓存时间 100ms，用于防止频繁读取文件

/**
 * 清除配置缓存（用于热更新）
 */
export function clearConfigCache() {
  cachedConfig = null;
  lastLoadTime = 0;
}

/**
 * 查找配置文件
 * @returns {string|null} 配置文件路径
 */
function findConfigFile() {
  const possiblePaths = [
    path.resolve(process.cwd(), 'task.json'),
    path.resolve(__dirname, '../../task.json'),
    path.resolve(__dirname, '../../../task.json')
  ];

  for (const p of possiblePaths) {
    if (fs.existsSync(p)) {
      return p;
    }
  }
  return null;
}

/**
 * 加载 JSON Schema
 * @returns {Object|null} Schema 对象
 */
function loadSchema() {
  const schemaPaths = [
    path.resolve(process.cwd(), 'task.schema.json'),
    path.resolve(__dirname, '../../task.schema.json')
  ];

  for (const p of schemaPaths) {
    if (fs.existsSync(p)) {
      try {
        const schemaContent = fs.readFileSync(p, 'utf8');
        return JSON.parse(schemaContent);
      } catch (error) {
        console.error(`Error loading schema from ${p}:`, error.message);
      }
    }
  }
  return null;
}

/**
 * 验证配置
 * @param {Object} config 配置对象
 * @returns {{valid: boolean, errors: Array}} 验证结果
 */
function validateConfig(config) {
  const schema = loadSchema();

  if (!schema) {
    console.warn('Config schema not found, skipping validation');
    return { valid: true, errors: [] };
  }

  try {
    const ajv = new Ajv({ allErrors: true, useDefaults: true, strict: false });
    const validate = ajv.compile(schema);
    const valid = validate(config);

    if (!valid) {
      console.error('Config validation failed:', JSON.stringify(validate.errors, null, 2));
    }

    return {
      valid,
      errors: validate.errors || []
    };
  } catch (error) {
    console.error('Error validating config:', error.message);
    // 如果验证过程出错，仍然允许保存（宽松模式）
    return { valid: true, errors: [], warning: error.message };
  }
}

/**
 * 从环境变量加载配置（兼容旧版 .env）
 * @returns {Object} 配置对象
 */
function loadFromEnv() {
  // 尝试加载 .env 文件
  const envPaths = [
    path.resolve(process.cwd(), '.env'),
    path.resolve(__dirname, '../../.env'),
    path.resolve(__dirname, '../../../.env')
  ];

  for (const envPath of envPaths) {
    if (fs.existsSync(envPath)) {
      dotenv.config({ path: envPath });
      break;
    }
  }

  // 判断 AI 提供商
  let provider = process.env.AI_PROVIDER || 'qwen';
  if (!process.env.AI_PROVIDER) {
    if (process.env.QWEN_API_KEY || process.env.DASHSCOPE_API_KEY) {
      provider = 'qwen';
    } else if (process.env.GOOGLE_API_KEY) {
      provider = 'gemini';
    }
  }

  // 从环境变量构建配置
  return {
    version: '1.0.0',
    ai: {
      provider,
      providers: {
        qwen: {
          enabled: true,
          apiKey: process.env.QWEN_API_KEY || process.env.DASHSCOPE_API_KEY || process.env.OPENAI_API_KEY || '',
          model: process.env.QWEN_MODEL || 'qwen3.5-plus',
          baseUrl: process.env.QWEN_BASE_URL || 'https://dashscope.aliyuncs.com/compatible-mode/v1'
        },
        gemini: {
          enabled: !!process.env.GOOGLE_API_KEY,
          apiKey: process.env.GOOGLE_API_KEY || '',
          model: process.env.GEMINI_MODEL || 'gemini-2.5-flash',
          baseUrl: process.env.GEMINI_BASE_URL || ''
        },
        perplexity: {
          enabled: !!process.env.PERPLEXITY_API_KEY,
          apiKey: process.env.PERPLEXITY_API_KEY || '',
          model: process.env.PERPLEXITY_MODEL || 'sonar-pro'
        }
      },
      parameters: {
        maxTokens: parseInt(process.env.MAX_TOKENS || '8192', 10),
        temperature: parseFloat(process.env.TEMPERATURE || '0.7')
      }
    },
    general: {
      debug: process.env.DEBUG === 'true',
      logLevel: process.env.LOG_LEVEL || 'info',
      defaultSubtasks: parseInt(process.env.DEFAULT_SUBTASKS || '3', 10),
      defaultPriority: process.env.DEFAULT_PRIORITY || 'medium',
      projectName: process.env.PROJECT_NAME || 'AI Task Manager',
      useChinese: process.env.USE_CHINESE === 'true'
    },
    storage: {
      type: process.env.TASK_STORAGE || 'db',
      database: {
        connection: process.env.DB_CONNECTION || 'mysql',
        host: process.env.DB_HOST || 'localhost',
        hostSlave: process.env.DB_HOST_SLAVE || process.env.DB_HOST || 'localhost',
        port: parseInt(process.env.DB_PORT || '3306', 10),
        database: process.env.DB_DATABASE || 'ai_task',
        username: process.env.DB_USERNAME || 'root',
        password: process.env.DB_PASSWORD || '',
        charset: process.env.DB_UTF8MB4_CHARSET || 'utf8mb4',
        collation: process.env.DB_UTF8MB4_COLLATION || 'utf8mb4_unicode_ci'
      }
    }
  };
}

/**
 * 深度合并配置对象
 * @param {Object} target 目标对象
 * @param {Object} source 源对象
 * @returns {Object} 合并后的对象
 */
function deepMerge(target, source) {
  const result = { ...target };

  for (const key of Object.keys(source)) {
    if (source[key] instanceof Object && key in target && target[key] instanceof Object) {
      result[key] = deepMerge(target[key], source[key]);
    } else {
      result[key] = source[key];
    }
  }

  return result;
}

/**
 * 加载配置
 * @param {boolean} forceReload 是否强制重新加载
 * @returns {Object} 配置对象
 */
export function loadConfig(forceReload = false) {
  const now = Date.now();

  // 如果有缓存且未过期且不强制刷新，直接返回缓存
  if (cachedConfig && !forceReload && (now - lastLoadTime) < CONFIG_TTL) {
    return cachedConfig;
  }

  // 先尝试加载环境变量作为基础
  const envConfig = loadFromEnv();

  // 然后尝试加载 task.json
  const configFile = findConfigFile();

  if (configFile) {
    configFilePath = configFile;
    try {
      const fileContent = fs.readFileSync(configFile, 'utf8');
      const fileConfig = JSON.parse(fileContent);

      // 合并配置：fileConfig 覆盖 envConfig
      cachedConfig = deepMerge(envConfig, fileConfig);
      lastLoadTime = now;

      // 验证配置
      const validation = validateConfig(cachedConfig);
      if (!validation.valid) {
        console.error('Config validation errors:', validation.errors);
        console.warn('Using config with validation errors');
      }

      if (forceReload) {
        console.log(`Configuration reloaded from: ${configFile}`);
      }
    } catch (error) {
      console.error(`Error loading config from ${configFile}:`, error.message);
      console.warn('Falling back to environment variables');
      cachedConfig = envConfig;
      lastLoadTime = now;
    }
  } else {
    console.log('No task.json found, using environment variables');
    cachedConfig = envConfig;
    lastLoadTime = now;
  }

  return cachedConfig;
}

/**
 * 获取当前配置文件路径
 * @returns {string|null} 配置文件路径
 */
export function getConfigPath() {
  return configFilePath;
}

/**
 * 保存配置到文件
 * @param {Object} config 配置对象
 * @param {string} filePath 可选的文件路径
 * @returns {boolean} 是否保存成功
 */
export function saveConfig(config, filePath = null) {
  const targetPath = filePath || configFilePath || path.resolve(process.cwd(), 'task.json');

  try {
    // 验证配置
    const validation = validateConfig(config);
    if (!validation.valid) {
      console.error('Config validation errors:', validation.errors);
      return false;
    }

    fs.writeFileSync(targetPath, JSON.stringify(config, null, 2));

    // 更新缓存
    cachedConfig = config;
    lastLoadTime = Date.now();
    configFilePath = targetPath;

    console.log(`Configuration saved to: ${targetPath}`);
    return true;
  } catch (error) {
    console.error(`Error saving config to ${targetPath}:`, error.message);
    return false;
  }
}

/**
 * 更新配置的部分字段
 * @param {Object} updates 要更新的配置字段
 * @returns {boolean} 是否更新成功
 */
export function updateConfig(updates) {
  const currentConfig = loadConfig();
  const newConfig = deepMerge(currentConfig, updates);
  return saveConfig(newConfig);
}

/**
 * 获取 AI 提供商配置
 * @param {string} providerName 提供商名称，不传则使用当前选中的
 * @returns {Object} 提供商配置
 */
export function getAIProviderConfig(providerName = null) {
  const config = loadConfig();
  const provider = providerName || config.ai.provider;
  return config.ai.providers[provider] || null;
}

/**
 * 获取数据库配置
 * @returns {Object} 数据库配置
 */
export function getDatabaseConfig() {
  const config = loadConfig();
  return config.storage.database || null;
}

/**
 * 获取通用配置
 * @returns {Object} 通用配置
 */
export function getGeneralConfig() {
  const config = loadConfig();
  return config.general || {};
}

/**
 * 重置配置为默认值
 * @returns {boolean} 是否重置成功
 */
export function resetConfig() {
  cachedConfig = null;
  return saveConfig(DEFAULT_CONFIG);
}

/**
 * 导出配置为环境变量格式
 * @returns {Object} 环境变量对象
 */
export function exportToEnv() {
  const config = loadConfig();
  const provider = config.ai.provider;
  const providerConfig = config.ai.providers[provider];

  return {
    // AI Provider
    AI_PROVIDER: provider,

    // Qwen
    QWEN_API_KEY: config.ai.providers.qwen.apiKey,
    QWEN_MODEL: config.ai.providers.qwen.model,
    QWEN_BASE_URL: config.ai.providers.qwen.baseUrl,

    // Gemini
    GOOGLE_API_KEY: config.ai.providers.gemini.apiKey,
    GEMINI_MODEL: config.ai.providers.gemini.model,
    GEMINI_BASE_URL: config.ai.providers.gemini.baseUrl,

    // Perplexity
    PERPLEXITY_API_KEY: config.ai.providers.perplexity.apiKey,
    PERPLEXITY_MODEL: config.ai.providers.perplexity.model,

    // AI Parameters
    MAX_TOKENS: String(config.ai.parameters.maxTokens),
    TEMPERATURE: String(config.ai.parameters.temperature),

    // General
    DEBUG: config.general.debug ? 'true' : 'false',
    LOG_LEVEL: config.general.logLevel,
    DEFAULT_SUBTASKS: String(config.general.defaultSubtasks),
    DEFAULT_PRIORITY: config.general.defaultPriority,
    PROJECT_NAME: config.general.projectName,
    USE_CHINESE: config.general.useChinese ? 'true' : 'false',

    // Storage
    TASK_STORAGE: config.storage.type,

    // Database
    DB_CONNECTION: config.storage.database.connection,
    DB_HOST: config.storage.database.host,
    DB_HOST_SLAVE: config.storage.database.hostSlave || config.storage.database.host,
    DB_PORT: String(config.storage.database.port),
    DB_DATABASE: config.storage.database.database,
    DB_USERNAME: config.storage.database.username,
    DB_PASSWORD: config.storage.database.password,
    DB_UTF8MB4_CHARSET: config.storage.database.charset,
    DB_UTF8MB4_COLLATION: config.storage.database.collation
  };
}

// 兼容旧版 CONFIG 对象 - 支持热更新
// 使用 Proxy 实现自动刷新配置
export function createLegacyConfig() {
  return new Proxy({}, {
    get(_target, prop) {
      // 每次访问属性时都重新加载配置（带缓存）
      const config = loadConfig();
      const provider = config.ai.provider;
      const providerConfig = config.ai.providers[provider];

      const legacyConfig = {
        // AI Model Configuration
        model: providerConfig.model,
        geminiModel: config.ai.providers.gemini.model,
        perplexityModel: config.ai.providers.perplexity.model,

        // Qwen (阿里云千问) Configuration
        qwenApiKey: config.ai.providers.qwen.apiKey,
        qwenModel: config.ai.providers.qwen.model,
        qwenBaseUrl: config.ai.providers.qwen.baseUrl,

        // AI Parameters
        maxTokens: config.ai.parameters.maxTokens,
        temperature: config.ai.parameters.temperature,

        // General Configuration
        debug: config.general.debug,
        logLevel: config.general.logLevel,
        defaultSubtasks: config.general.defaultSubtasks,
        defaultPriority: config.general.defaultPriority,
        projectName: config.general.projectName,
        projectVersion: '1.5.0',

        // API Configuration
        geminiBaseUrl: config.ai.providers.gemini.baseUrl,
        useChinese: config.general.useChinese,

        // Provider selection
        aiProvider: provider
      };

      return legacyConfig[prop];
    }
  });
}

export default {
  loadConfig,
  saveConfig,
  updateConfig,
  getConfigPath,
  getAIProviderConfig,
  getDatabaseConfig,
  getGeneralConfig,
  resetConfig,
  exportToEnv,
  createLegacyConfig,
  DEFAULT_CONFIG
};
