<script setup lang="ts">
import { useAuthStore } from './stores/auth'
import { useToastStore } from './stores/toast'
import { watch, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from './api'

const auth = useAuthStore()
const toast = useToastStore()
const route = useRoute()
const router = useRouter()

const templateCount = ref<number | null>(null)
const sidebarOpen = ref(false)

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
    <!-- 移动端顶部栏 -->
    <div class="mobile-header">
      <button class="hamburger" @click="sidebarOpen = !sidebarOpen">
        <span></span><span></span><span></span>
      </button>
      <span class="mobile-title">租赁管家</span>
    </div>
    <!-- 侧边栏遮罩 -->
    <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>
    <aside class="sidebar" :class="{ open: sidebarOpen }">
      <div class="sidebar-logo">租赁管家</div>
      <nav class="sidebar-nav" @click="sidebarOpen = false">
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

  <!-- 全局 Toast 通知 -->
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div v-for="t in toast.toasts" :key="t.id" :class="['toast-item', `toast-${t.type}`]">
          {{ t.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
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

<style>
/* 全局 Toast 样式 */
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast-item {
  padding: 12px 20px;
  border-radius: 10px;
  font-size: 0.875rem;
  font-weight: 500;
  color: #fff;
  pointer-events: auto;
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  max-width: 360px;
}

.toast-success { background: rgba(52, 199, 89, 0.92); }
.toast-error { background: rgba(255, 59, 48, 0.92); }
.toast-info { background: rgba(0, 122, 255, 0.92); }

.toast-enter-active { transition: all 0.3s cubic-bezier(0.25, 0.1, 0.25, 1); }
.toast-leave-active { transition: all 0.25s ease-in; }
.toast-enter-from { opacity: 0; transform: translateX(40px); }
.toast-leave-to { opacity: 0; transform: translateY(-10px); }
</style>
