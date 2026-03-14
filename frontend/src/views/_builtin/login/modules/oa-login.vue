<script setup lang="ts">
import { ref } from 'vue';
import { useRouterPush } from '@/hooks/common/router';
import { localStg } from '@/utils/storage';

defineOptions({
  name: 'OaLogin'
});

const { redirectFromLogin } = useRouterPush(false);

const loading = ref(false);

// OA 登录地址（模拟）
const OA_LOGIN_URL = 'https://oa.example.com/login?redirect=';

// 模拟用户信息
const mockUserInfo = {
  userId: 'oa-user-' + Date.now(),
  userName: 'OA 用户',
  roles: ['R_SUPER'],
  buttons: []
};

async function handleOaLogin() {
  loading.value = true;

  try {
    // 模拟 OA 登录过程
    // 在实际场景中，这里会跳转到 OA 系统进行认证
    // 然后 OA 系统会回调并带上授权码

    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 500));

    // 模拟设置 token
    const mockToken = 'mock-oa-token-' + Date.now();
    localStg.set('token', mockToken);
    localStg.set('refreshToken', 'mock-oa-refresh-token');

    // 存储用户信息
    localStg.set('userInfo', mockUserInfo);

    // 显示成功提示
    window.$notification?.success({
      title: '登录成功',
      content: '欢迎回来！',
      duration: 2000
    });

    // 使用页面刷新跳转，确保路由守卫重新执行
    setTimeout(() => {
      window.location.href = '/';
    }, 500);
  } catch (error) {
    window.$message?.error('登录失败，请重试');
    loading.value = false;
  }
}

// 打开 OA 登录页面（模拟跳转）
function openOaPage() {
  // 在新窗口打开 OA 系统
  window.open(OA_LOGIN_URL + encodeURIComponent(window.location.origin), '_blank');
}
</script>

<template>
  <div class="oa-login-container">
    <div class="oa-login-content">
      <div class="oa-icon">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="size-80px text-primary">
          <path d="M4.5 6.375a4.125 4.125 0 1 1 8.25 0 4.125 4.125 0 0 1-8.25 0ZM14.25 8.625a3.375 3.375 0 1 1 6.75 0 3.375 3.375 0 0 1-6.75 0ZM1.5 19.125a7.125 7.125 0 0 1 14.25 0v.003l-.001.119a.75.75 0 0 1-.363.63 13.067 13.067 0 0 1-6.761 1.873c-2.472 0-4.786-.684-6.76-1.873a.75.75 0 0 1-.364-.63l-.001-.122ZM17.25 19.128l-.001.144a2.25 2.25 0 0 1-.233.96 10.088 10.088 0 0 0 5.06-1.01.75.75 0 0 0 .42-.643 4.875 4.875 0 0 0-6.957-4.611 8.586 8.586 0 0 1 1.71 5.157v.003Z" />
        </svg>
      </div>

      <div class="oa-title">
        <h2>AI 任务管理系统</h2>
        <p class="text-gray-500 mt-2">使用 OA 账号快速登录</p>
      </div>

      <NSpace vertical :size="24" class="w-full">
        <NButton
          type="primary"
          size="large"
          round
          block
          :loading="loading"
          @click="handleOaLogin"
        >
          <template #icon>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="size-20px">
              <path fill-rule="evenodd" d="M8.25 6.75a3.75 3.75 0 1 1 7.5 0 3.75 3.75 0 0 1-7.5 0ZM15.75 9.75a3 3 0 1 1 6 0 3 3 0 0 1-6 0ZM2.25 9.75a3 3 0 1 1 6 0 3 3 0 0 1-6 0ZM6.31 15.117A6.745 6.745 0 0 1 12 12a6.745 6.745 0 0 1 6.709 7.498.75.75 0 0 1-.372.568A12.696 12.696 0 0 1 12 21.75c-2.305 0-4.47-.612-6.337-1.684a.75.75 0 0 1-.372-.568 6.787 6.787 0 0 1 1.019-4.38Z" clip-rule="evenodd" />
              <path d="M5.082 14.254a8.287 8.287 0 0 0-1.308 5.135 9.687 9.687 0 0 1-1.764-.44l-.115-.04a.563.563 0 0 1-.373-.487l-.01-.121a3.75 3.75 0 0 1 3.57-4.047ZM20.226 19.389a8.287 8.287 0 0 0-1.308-5.135 3.75 3.75 0 0 1 3.57 4.047l-.01.121a.563.563 0 0 1-.373.486l-.115.04c-.567.2-1.156.349-1.764.441Z" />
            </svg>
          </template>
          OA 登录
        </NButton>

        <div class="text-center">
          <NButton text type="primary" @click="openOaPage">
            打开 OA 系统
          </NButton>
        </div>
      </NSpace>
    </div>
  </div>
</template>

<style scoped>
.oa-login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.oa-login-content {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.oa-icon {
  margin-bottom: 24px;
  display: flex;
  justify-content: center;
}

.oa-title {
  text-align: center;
  margin-bottom: 32px;
}

.oa-title h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}
</style>
