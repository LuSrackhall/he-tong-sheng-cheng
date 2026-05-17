<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')

async function submit() {
  error.value = ''
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  const ok = await auth.login(username.value, password.value)
  if (ok) {
    router.push('/')
  } else {
    error.value = '用户名或密码错误'
  }
}
</script>

<template>
  <div style="display: flex; align-items: center; justify-content: center; min-height: 100vh; background: var(--color-bg);">
    <div class="card" style="width: 380px;">
      <h2 style="text-align: center; margin-bottom: 24px; font-size: 1.5rem;">租赁管理系统</h2>
      <form @submit.prevent="submit">
        <div class="form-group">
          <label class="label">用户名</label>
          <input class="input" v-model="username" placeholder="请输入用户名" />
        </div>
        <div class="form-group">
          <label class="label">密码</label>
          <input class="input" type="password" v-model="password" placeholder="请输入密码" />
        </div>
        <div v-if="error" style="color: var(--color-danger); font-size: 0.875rem; margin-bottom: 16px;">{{ error }}</div>
        <button class="btn btn-primary" style="width: 100%;" :disabled="auth.loading">
          {{ auth.loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>
