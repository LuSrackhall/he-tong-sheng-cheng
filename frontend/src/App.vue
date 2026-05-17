<script setup lang="ts">
import { useAuthStore } from './stores/auth'
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

watch(() => auth.token, (val) => {
  if (val) {
    auth.fetchMe()
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
        <router-link to="/new-contract">
          <span>📝</span> 签新合同
        </router-link>
        <router-link to="/collect-rent">
          <span>💰</span> 收租金
        </router-link>
        <router-link to="/arrears">
          <span>📊</span> 催缴清单
        </router-link>
        <div style="height: 1px; background: var(--color-border); margin: var(--space-sm) 0;" />
        <router-link to="/assets">
          <span>🏠</span> 资产管理
        </router-link>
        <router-link to="/tenants">
          <span>👤</span> 租户管理
        </router-link>
        <router-link to="/contracts">
          <span>📄</span> 合同管理
        </router-link>
        <router-link to="/receipt-books">
          <span>🧾</span> 收据本管理
        </router-link>
        <router-link v-if="auth.isAdmin" to="/users">
          <span>⚙️</span> 用户管理
        </router-link>
        <router-link to="/settings">
          <span>🔧</span> 系统设置
        </router-link>
      </nav>
      <div style="display: flex; flex-direction: column; gap: 8px;">
        <div style="font-size: 0.8125rem; color: var(--color-text-secondary);">
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
