<script setup lang="ts">
import { useAuthStore } from './stores/auth'
import { watch, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from './api'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const templateCount = ref<number | null>(null)

watch(() => auth.token, (val) => {
  if (val) {
    auth.fetchMe()
  }
}, { immediate: true })

watch(() => auth.isLoggedIn, async (val) => {
  if (val) {
    try {
      const { data } = await api.get('/templates')
      const templates = (data as any).data || data
      templateCount.value = Array.isArray(templates) ? templates.length : 0
    } catch {
      templateCount.value = 0
    }
  }
}, { immediate: true })
</script>

<template>
  <div v-if="!auth.isLoggedIn && route.meta.guest" class="guest-layout">
    <router-view />
  </div>
  <div v-else-if="auth.isLoggedIn" class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-logo">租赁管家</div>
      <nav class="sidebar-nav">
        <div class="nav-group-label">日常操作</div>
        <router-link to="/new-contract" class="nav-primary-action">
          <span>📝</span> 签新合同
        </router-link>
        <router-link to="/collect-rent">
          <span>💰</span> 收租金
        </router-link>
        <router-link to="/arrears">
          <span>📊</span> 催缴清单
        </router-link>
        <div class="nav-separator" />
        <div class="nav-group-label">数据管理</div>
        <router-link to="/assets">
          <span>🏠</span> 资产管理
        </router-link>
        <router-link to="/tenants">
          <span>👤</span> 租户管理
        </router-link>
        <router-link to="/contracts">
          <span>📄</span> 合同管理
        </router-link>
        <div class="nav-separator" />
        <div class="nav-group-label">系统设置</div>
        <router-link to="/settings">
          <span>🔧</span> 模板设置
          <span v-if="templateCount === 0" class="nav-warning-dot"></span>
        </router-link>
        <router-link to="/receipt-books">
          <span>🧾</span> 收据本
        </router-link>
        <router-link v-if="auth.isAdmin" to="/users">
          <span>⚙️</span> 用户管理
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <div class="sidebar-user">
          {{ auth.user?.username }} ({{ auth.user?.role }})
        </div>
        <button class="btn btn-secondary btn-sm" @click="auth.logout(); router.push('/login')">退出登录</button>
      </div>
    </aside>
    <main class="main-content">
      <router-view />
    </main>
  </div>
  <div v-else class="guest-layout">
    <router-view />
  </div>
</template>

<style scoped>
.nav-group-label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: var(--space-sm) 12px 2px;
  margin-top: var(--space-xs);
  user-select: none;
}

.nav-group-label:first-child {
  margin-top: 0;
}

.nav-separator {
  height: 1px;
  background: var(--color-border);
  margin: var(--space-sm) 12px;
}

.nav-primary-action {
  background: var(--color-primary) !important;
  color: #fff !important;
  font-weight: 600 !important;
}

.nav-primary-action:hover {
  background: var(--color-primary-hover) !important;
  color: #fff !important;
}

.nav-primary-action.router-link-active {
  background: var(--color-primary-hover) !important;
  color: #fff !important;
}

.nav-warning-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-warning);
  margin-left: auto;
  flex-shrink: 0;
}

.sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-user {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}
</style>
