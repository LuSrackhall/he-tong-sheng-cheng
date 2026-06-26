<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()

const arrearsContracts = ref<any[]>([])
const loading = ref(true)
const activeTab = ref(1)

const tabs = [
  { level: 1, name: '应缴预警', color: 'var(--color-warning)' },
  { level: 2, name: '近期应缴提醒', color: 'var(--color-warning)' },
  { level: 3, name: '逾期未缴催收', color: 'var(--color-danger)' },
  { level: 4, name: '到期预警', color: 'var(--color-warning)' },
  { level: 5, name: '已到期欠费追缴', color: 'var(--color-danger)' },
]

const suggestedActions: Record<number, string> = {
  1: '列入观察，心中有数',
  2: '主动联系，提醒缴纳',
  3: '上门催收，限期缴纳',
  4: '即将到期，清算账款',
  5: '进入追讨，法律途径',
}

const allEmpty = computed(() => arrearsContracts.value.length === 0 && !loading.value)

const filteredContracts = computed(() =>
  arrearsContracts.value.filter((c: any) => c.arrearsLevel === activeTab.value)
)

const levelCounts = computed(() => {
  const counts: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 }
  for (const c of arrearsContracts.value) {
    if (counts[c.arrearsLevel] !== undefined) counts[c.arrearsLevel]++
  }
  return counts
})

async function fetchArrears() {
  loading.value = true
  try {
    const { data } = await api.get('/arrears')
    arrearsContracts.value = Array.isArray(data) ? data : (data as any).data || []
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(fetchArrears)

function goToCollect(c: any) {
  router.push({ path: '/collect-rent', query: { contractId: String(c.id) } })
}
</script>

<template>
  <div>
    <div class="page-header"><h2>催缴清单</h2></div>

    <div style="display: flex; gap: 4px; margin-bottom: var(--space-lg); flex-wrap: wrap;">
      <button
        v-for="tab in tabs" :key="tab.level"
        class="btn"
        :class="activeTab === tab.level ? 'btn-primary' : 'btn-secondary'"
        @click="activeTab = tab.level"
        :style="activeTab === tab.level ? { background: tab.color, borderColor: tab.color } : {}"
      >
        {{ tab.name }}
        <span v-if="levelCounts[tab.level] > 0" class="tab-badge" :style="activeTab === tab.level ? { background: 'rgba(255,255,255,0.3)' } : {}">{{ levelCounts[tab.level] }}</span>
      </button>
    </div>

    <div v-if="loading" class="empty-state">加载中...</div>
    <div v-else-if="allEmpty" class="empty-state-all">
      <div style="font-size: 2rem; margin-bottom: var(--space-sm);">✅</div>
      <div style="font-weight: 600; margin-bottom: var(--space-xs);">暂无需要催缴的合同</div>
      <div style="color: var(--color-text-tertiary); font-size: 0.875rem;">所有合同收款状态正常</div>
    </div>
    <div v-else-if="filteredContracts.length === 0" class="empty-state">
      暂无{{ tabs.find(t => t.level === activeTab)?.name }}的合同
    </div>

    <div class="table-wrapper">
      <table v-if="filteredContracts.length > 0">
        <thead>
          <tr>
            <th>合同编号</th>
            <th>资产</th>
            <th>租户</th>
            <th>已收 / 应收</th>
            <th>欠费</th>
            <th>钱用到</th>
            <th>到期日</th>
            <th>建议操作</th>
            <th>收款</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filteredContracts" :key="c.id">
            <td>#{{ c.id }}</td>
            <td>{{ c.asset?.name || '-' }}</td>
            <td>{{ c.tenant?.name || '-' }}</td>
            <td>¥{{ c.totalReceived?.toLocaleString() }} / ¥{{ c.totalReceivable?.toLocaleString() }}</td>
            <td style="color: var(--color-danger); font-weight: 500;">¥{{ ((c.totalReceivable || 0) - (c.totalReceived || 0)).toLocaleString() }}</td>
            <td>{{ c.usedUpDate?.substring(0, 10) || '-' }}</td>
            <td>{{ c.endDate?.substring(0, 10) || '-' }}</td>
            <td style="font-size: 0.8125rem; color: var(--color-text-secondary);">{{ suggestedActions[activeTab] }}</td>
            <td><button class="btn btn-primary btn-sm" @click="goToCollect(c)">去收款</button></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  font-size: 0.6875rem;
  font-weight: 600;
  background: var(--color-border);
  color: var(--color-text-secondary);
  margin-left: 6px;
}

.empty-state-all {
  text-align: center;
  padding: var(--space-2xl) var(--space-lg);
  margin-top: var(--space-lg);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}
</style>
