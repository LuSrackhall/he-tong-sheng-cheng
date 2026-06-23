<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'

interface Receipt {
  id: number
  receiptBookId: number
  paymentId: number
  sequenceNum: number
  amount: number
  voided?: boolean
  printedAt: string
}

const receipts = ref<Receipt[]>([])
const total = ref(0)
const page = ref(0)
const pageSize = 50
const loading = ref(true)

async function fetchReceipts() {
  loading.value = true
  try {
    const { data } = await api.get('/receipts', { params: { offset: page.value * pageSize, limit: pageSize } })
    receipts.value = data.data || []
    total.value = data.total || 0
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function printReceipt(paymentId: number) {
  window.open(`/api/print/receipt/${paymentId}`, '_blank')
}

onMounted(fetchReceipts)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>收据记录</h2>
      <span style="font-size: 0.875rem; color: var(--color-text-secondary);">共 {{ total }} 条</span>
    </div>

    <div v-if="loading" class="empty-state">加载中...</div>
    <div v-else-if="receipts.length === 0" class="empty-state">暂无收据记录</div>

    <div v-else class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>收据编号</th>
            <th>收据本</th>
            <th>序号</th>
            <th>金额</th>
            <th>打印时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in receipts" :key="r.id">
            <td>#{{ r.id }}</td>
            <td>收据本 #{{ r.receiptBookId }}</td>
            <td>{{ r.sequenceNum }}</td>
            <td>¥{{ r.amount.toLocaleString() }}</td>
            <td style="font-size: 0.8125rem; color: var(--color-text-secondary);">{{ new Date(r.printedAt).toLocaleString('zh-CN') }}</td>
            <td>
              <button class="btn btn-secondary btn-sm" @click="printReceipt(r.paymentId)">打印收据</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchReceipts()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchReceipts()">下一页</button>
    </div>
  </div>
</template>
