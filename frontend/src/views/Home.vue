<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'

interface DashboardStats {
  activeContracts: number
  monthlyRevenue: number
  overdueContracts: number
  newContractsThisMonth: number
}

const stats = ref<DashboardStats | null>(null)
const loading = ref(true)
const error = ref('')

async function fetchStats() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/dashboard/stats')
    stats.value = data
  } catch {
    error.value = '加载统计数据失败'
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>概览</h2>
    </div>

    <div v-if="loading" class="loading-state">加载中...</div>

    <div v-else-if="error" class="error-state">{{ error }}</div>

    <div v-else-if="stats" class="dashboard-grid">
      <div class="stat-card">
        <div class="stat-icon">📄</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeContracts }}</div>
          <div class="stat-label">活跃合同</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💰</div>
        <div class="stat-content">
          <div class="stat-value">¥{{ stats.monthlyRevenue.toLocaleString() }}</div>
          <div class="stat-label">本月收款</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">⚠️</div>
        <div class="stat-content">
          <div class="stat-value" :class="{ 'text-danger': stats.overdueContracts > 0 }">{{ stats.overdueContracts }}</div>
          <div class="stat-label">逾期合同</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">✨</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.newContractsThisMonth }}</div>
          <div class="stat-label">本月新增</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-md);
  margin-top: var(--space-md);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-lg);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.stat-icon {
  font-size: 2rem;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.text-danger {
  color: var(--color-danger);
}

.loading-state,
.error-state {
  text-align: center;
  padding: var(--space-2xl);
  color: var(--color-text-secondary);
}
</style>
