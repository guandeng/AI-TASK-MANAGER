<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NSelect,
  NButton,
  NSpace,
  NDivider,
  NCollapse,
  NCollapseItem,
  useMessage
} from 'naive-ui';
import {
  fetchConfig,
  updateConfig,
  switchAIProvider,
  updateProviderConfig,
  resetConfig,
  type AppConfig,
  type AIProviderConfig
} from '@/service/api/config';

const message = useMessage();
const loading = ref(false);
const config = ref<AppConfig | null>(null);
const configPath = ref<string | null>(null);

// AI 配置表单
const aiForm = ref({
  provider: '',
  maxTokens: 4096,
  temperature: 0.7
});

// 通用配置表单
const generalForm = ref({
  debug: false,
  logLevel: 'info',
  defaultSubtasks: 3,
  defaultPriority: 'medium',
  projectName: '',
  useChinese: true
});

// AI 提供商配置
const providerForms = ref({
  qwen: { enabled: false, apiKey: '', model: '', baseUrl: '' },
  gemini: { enabled: false, apiKey: '', model: '', baseUrl: '' },
  perplexity: { enabled: false, apiKey: '', model: '', baseUrl: '' }
});

// 存储配置表单
const storageForm = ref({
  type: 'json',
  database: {
    connection: 'mysql',
    host: '127.0.0.1',
    hostSlave: '127.0.0.1',
    port: 3306,
    database: '',
    username: 'root',
    password: '',
    charset: 'utf8mb4',
    collation: 'utf8mb4_unicode_ci'
  }
});

// 提供商选项
const providerOptions = [
  { label: '通义千问 (Qwen)', value: 'qwen' },
  { label: 'Google Gemini', value: 'gemini' },
  { label: 'Perplexity', value: 'perplexity' }
];

// 日志级别选项
const logLevelOptions = [
  { label: '调试 (debug)', value: 'debug' },
  { label: '信息 (info)', value: 'info' },
  { label: '警告 (warn)', value: 'warn' },
  { label: '错误 (error)', value: 'error' }
];

// 优先级选项
const priorityOptions = [
  { label: '高', value: 'high' },
  { label: '中', value: 'medium' },
  { label: '低', value: 'low' }
];

// 存储类型选项
const storageTypeOptions = [
  { label: 'JSON 文件', value: 'json' },
  { label: '数据库', value: 'db' }
];

// 数据库连接类型选项
const dbConnectionOptions = [
  { label: 'MySQL', value: 'mysql' },
  { label: 'PostgreSQL', value: 'pgsql' },
  { label: 'SQLite', value: 'sqlite' }
];

// 加载配置
async function loadConfig() {
  loading.value = true;
  try {
    const result = await fetchConfig();
    const data = (result as any)?.config ? result : (result as any)?.data ? (result as any).data : null;
    const cfg = data?.config;
    if (cfg) {
      config.value = cfg;
      configPath.value = data.configPath;

      aiForm.value.provider = cfg.ai.provider;
      aiForm.value.maxTokens = cfg.ai.parameters.maxTokens;
      aiForm.value.temperature = cfg.ai.parameters.temperature;

      generalForm.value = { ...cfg.general };

      if (cfg.ai.providers.qwen) {
        providerForms.value.qwen = { ...cfg.ai.providers.qwen };
      }
      if (cfg.ai.providers.gemini) {
        providerForms.value.gemini = { ...cfg.ai.providers.gemini };
      }
      if (cfg.ai.providers.perplexity) {
        providerForms.value.perplexity = { ...cfg.ai.providers.perplexity };
      }

      if (cfg.storage) {
        storageForm.value.type = cfg.storage.type || 'json';
        if (cfg.storage.database) {
          storageForm.value.database = { ...storageForm.value.database, ...cfg.storage.database };
        }
      }
    }
  } catch (err) {
    message.error('加载配置失败');
  } finally {
    loading.value = false;
  }
}

// 保存 AI 配置
async function saveAIConfig() {
  loading.value = true;
  try {
    await updateConfig({
      ai: {
        provider: aiForm.value.provider,
        providers: {
          qwen: providerForms.value.qwen as AIProviderConfig,
          gemini: providerForms.value.gemini as AIProviderConfig,
          perplexity: providerForms.value.perplexity as AIProviderConfig
        },
        parameters: {
          maxTokens: aiForm.value.maxTokens,
          temperature: aiForm.value.temperature
        }
      }
    } as Partial<AppConfig>);
    message.success('AI 配置保存成功');
  } catch {
    message.error('保存失败');
  } finally {
    loading.value = false;
  }
}

// 保存通用配置
async function saveGeneralConfig() {
  loading.value = true;
  try {
    await updateConfig({
      general: generalForm.value
    });
    message.success('通用配置保存成功');
  } catch {
    message.error('保存失败');
  } finally {
    loading.value = false;
  }
}

// 保存存储配置
async function saveStorageConfig() {
  loading.value = true;
  try {
    await updateConfig({
      storage: storageForm.value
    } as Partial<AppConfig>);
    message.success('存储配置保存成功');
  } catch {
    message.error('保存失败');
  } finally {
    loading.value = false;
  }
}

// 重置配置
async function handleReset() {
  if (!window.confirm('确认重置所有配置吗？这将恢复默认设置。')) {
    return;
  }
  loading.value = true;
  try {
    await resetConfig();
    message.success('配置已重置');
    await loadConfig();
  } catch {
    message.error('重置失败');
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadConfig();
});
</script>

<template>
  <div class="config-page">
    <NCard title="系统配置">
      <template #header-extra>
        <NButton type="warning" :loading="loading" @click="handleReset">
          重置配置
        </NButton>
      </template>

      <NTabs type="line">
        <!-- AI 配置 -->
        <NTabPane name="ai" tab="AI 配置">
          <NForm label-placement="left" label-width="120px">
            <NFormItem label="当前提供商">
              <NSelect
                v-model:value="aiForm.provider"
                :options="providerOptions"
                style="width: 200px"
              />
            </NFormItem>
            <NFormItem label="最大 Token">
              <NInputNumber
                v-model:value="aiForm.maxTokens"
                :min="100"
                :max="128000"
                style="width: 200px"
              />
            </NFormItem>
            <NFormItem label="温度">
              <NInputNumber
                v-model:value="aiForm.temperature"
                :min="0"
                :max="2"
                :step="0.1"
                style="width: 200px"
              />
            </NFormItem>
          </NForm>

          <NDivider>提供商配置</NDivider>

          <NCollapse>
            <NCollapseItem title="通义千问 (Qwen)" name="qwen">
              <NForm label-placement="left" label-width="100px">
                <NFormItem label="启用">
                  <NSwitch v-model:value="providerForms.qwen.enabled" />
                </NFormItem>
                <NFormItem label="API Key">
                  <NInput
                    v-model:value="providerForms.qwen.apiKey"
                    type="password"
                    placeholder="请输入 API Key"
                    style="width: 300px"
                  />
                </NFormItem>
                <NFormItem label="模型">
                  <NInput
                    v-model:value="providerForms.qwen.model"
                    placeholder="如: qwen-max"
                    style="width: 200px"
                  />
                </NFormItem>
                <NFormItem label="Base URL">
                  <NInput
                    v-model:value="providerForms.qwen.baseUrl"
                    placeholder="可选"
                    style="width: 300px"
                  />
                </NFormItem>
              </NForm>
            </NCollapseItem>

            <NCollapseItem title="Google Gemini" name="gemini">
              <NForm label-placement="left" label-width="100px">
                <NFormItem label="启用">
                  <NSwitch v-model:value="providerForms.gemini.enabled" />
                </NFormItem>
                <NFormItem label="API Key">
                  <NInput
                    v-model:value="providerForms.gemini.apiKey"
                    type="password"
                    placeholder="请输入 API Key"
                    style="width: 300px"
                  />
                </NFormItem>
                <NFormItem label="模型">
                  <NInput
                    v-model:value="providerForms.gemini.model"
                    placeholder="如: gemini-pro"
                    style="width: 200px"
                  />
                </NFormItem>
                <NFormItem label="Base URL">
                  <NInput
                    v-model:value="providerForms.gemini.baseUrl"
                    placeholder="可选"
                    style="width: 300px"
                  />
                </NFormItem>
              </NForm>
            </NCollapseItem>

            <NCollapseItem title="Perplexity" name="perplexity">
              <NForm label-placement="left" label-width="100px">
                <NFormItem label="启用">
                  <NSwitch v-model:value="providerForms.perplexity.enabled" />
                </NFormItem>
                <NFormItem label="API Key">
                  <NInput
                    v-model:value="providerForms.perplexity.apiKey"
                    type="password"
                    placeholder="请输入 API Key"
                    style="width: 300px"
                  />
                </NFormItem>
                <NFormItem label="模型">
                  <NInput
                    v-model:value="providerForms.perplexity.model"
                    placeholder="如: llama-3.1-sonar-large-128k-online"
                    style="width: 200px"
                  />
                </NFormItem>
                <NFormItem label="Base URL">
                  <NInput
                    v-model:value="providerForms.perplexity.baseUrl"
                    placeholder="可选"
                    style="width: 300px"
                  />
                </NFormItem>
              </NForm>
            </NCollapseItem>
          </NCollapse>

          <NSpace class="mt-16px">
            <NButton type="primary" :loading="loading" @click="saveAIConfig">
              保存 AI 配置
            </NButton>
          </NSpace>
        </NTabPane>

        <!-- 存储配置 -->
        <NTabPane name="storage" tab="存储配置">
          <NForm label-placement="left" label-width="120px">
            <NFormItem label="存储类型">
              <NSelect
                v-model:value="storageForm.type"
                :options="storageTypeOptions"
                style="width: 200px"
              />
            </NFormItem>
          </NForm>

          <template v-if="storageForm.type === 'db'">
            <NDivider>数据库配置</NDivider>
            <NForm label-placement="left" label-width="120px">
              <NFormItem label="数据库类型">
                <NSelect
                  v-model:value="storageForm.database.connection"
                  :options="dbConnectionOptions"
                  style="width: 200px"
                />
              </NFormItem>
              <NFormItem label="主库地址">
                <NInput
                  v-model:value="storageForm.database.host"
                  placeholder="127.0.0.1"
                  style="width: 300px"
                />
              </NFormItem>
              <NFormItem label="从库地址">
                <NInput
                  v-model:value="storageForm.database.hostSlave"
                  placeholder="127.0.0.1"
                  style="width: 300px"
                />
              </NFormItem>
              <NFormItem label="端口">
                <NInputNumber
                  v-model:value="storageForm.database.port"
                  :min="1"
                  :max="65535"
                  style="width: 150px"
                />
              </NFormItem>
              <NFormItem label="数据库名">
                <NInput
                  v-model:value="storageForm.database.database"
                  placeholder="ai_task"
                  style="width: 300px"
                />
              </NFormItem>
              <NFormItem label="用户名">
                <NInput
                  v-model:value="storageForm.database.username"
                  placeholder="root"
                  style="width: 300px"
                />
              </NFormItem>
              <NFormItem label="密码">
                <NInput
                  v-model:value="storageForm.database.password"
                  type="password"
                  show-password-on="click"
                  placeholder="请输入密码"
                  style="width: 300px"
                />
              </NFormItem>
              <NFormItem label="字符集">
                <NInput
                  v-model:value="storageForm.database.charset"
                  placeholder="utf8mb4"
                  style="width: 200px"
                />
              </NFormItem>
              <NFormItem label="排序规则">
                <NInput
                  v-model:value="storageForm.database.collation"
                  placeholder="utf8mb4_unicode_ci"
                  style="width: 300px"
                />
              </NFormItem>
            </NForm>
          </template>

          <NSpace class="mt-16px">
            <NButton type="primary" :loading="loading" @click="saveStorageConfig">
              保存存储配置
            </NButton>
          </NSpace>
        </NTabPane>

        <!-- 通用配置 -->
        <NTabPane name="general" tab="通用配置">
          <NForm label-placement="left" label-width="120px">
            <NFormItem label="项目名称">
              <NInput
                v-model:value="generalForm.projectName"
                placeholder="请输入项目名称"
                style="width: 300px"
              />
            </NFormItem>
            <NFormItem label="调试模式">
              <NSwitch v-model:value="generalForm.debug" />
            </NFormItem>
            <NFormItem label="日志级别">
              <NSelect
                v-model:value="generalForm.logLevel"
                :options="logLevelOptions"
                style="width: 200px"
              />
            </NFormItem>
            <NFormItem label="默认子任务数">
              <NInputNumber
                v-model:value="generalForm.defaultSubtasks"
                :min="1"
                :max="10"
                style="width: 150px"
              />
            </NFormItem>
            <NFormItem label="默认优先级">
              <NSelect
                v-model:value="generalForm.defaultPriority"
                :options="priorityOptions"
                style="width: 150px"
              />
            </NFormItem>
            <NFormItem label="使用中文输出">
              <NSwitch v-model:value="generalForm.useChinese" />
            </NFormItem>
          </NForm>

          <NSpace class="mt-16px">
            <NButton type="primary" :loading="loading" @click="saveGeneralConfig">
              保存通用配置
            </NButton>
          </NSpace>
        </NTabPane>
      </NTabs>
    </NCard>
  </div>
</template>

<style scoped>
.config-page {
  padding: 16px;
}

.mt-16px {
  margin-top: 16px;
}
</style>
