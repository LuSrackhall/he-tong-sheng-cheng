<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'
import { useEscapeKey } from '../composables/useEscapeKey'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()

const statusLabels: Record<string, string> = { active: '使用中', full: '已用完', archived: '已作废' }
useEscapeKey(() => { showCreate.value = false })

interface ReceiptBook {
  id: number
  prefix: string
  startNum: number
  currentNum: number
  totalPages: number
  status: string
  createdAt: string
}

const books = ref<ReceiptBook[]>([])
const total = ref(0)
const page = ref(0)
const pageSize = 20
const showCreate = ref(false)
const form = ref({ prefix: '', startNum: 1, totalPages: 50 })
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')

const loading = ref(false)

async function fetchBooks() {
  loading.value = true
  try {
    const { data } = await api.get('/receipt-books', { params: { offset: page.value * pageSize, limit: pageSize } })
    books.value = data.data
    total.value = data.total
  } catch {
    toast.error('加载收据本列表失败')
  } finally {
    loading.value = false
  }
}
async function createBook() {
  if (submitLock.value) return
  if (!form.value.prefix || form.value.totalPages <= 0) return
  submitLock.value = true
  saving.value = true
  try {
    await api.post('/receipt-books', form.value)
    showCreate.value = false
    errorMessage.value = ''
    form.value = { prefix: '', startNum: 1, totalPages: 50 }
    fetchBooks()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '创建失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

onMounted(fetchBooks)
</script>

<template>
  <div>
    <div class="page-header"><h2>收据本管理</h2><button class="btn btn-primary" @click="showCreate = true">+ 新建收据本</button></div>

    <div v-if="loading" class="empty-state">加载中...</div>
    <div v-else-if="books.length === 0" class="empty-state">暂无收据本，请创建新收据本</div>

    <div v-else class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>前缀</th>
            <th>起始编号</th>
            <th>当前编号</th>
            <th>总页数</th>
            <th>状态</th>
            <th>创建时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in books" :key="b.id">
            <td>#{{ b.id }}</td>
            <td>{{ b.prefix }}</td>
            <td>{{ b.startNum }}</td>
            <td>{{ b.currentNum }}</td>
            <td>{{ b.totalPages }}</td>
            <td>
              <span class="badge" :class="{
                'badge-success': b.status === 'active',
                'badge-info': b.status === 'full',
                'badge-warning': b.status === 'archived',
              }">{{ statusLabels[b.status] || b.status }}</span>
            </td>
            <td>{{ new Date(b.createdAt).toLocaleDateString('zh-CN') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchBooks()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchBooks()">下一页</button>
    </div>

    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal-content">
        <h3>新建收据本</h3>
        <div class="form-group"><label class="label">前缀（如 "SK-2026-"）</label><input class="input" v-model="form.prefix" placeholder="SK-2026-" /></div>
        <div class="form-group"><label class="label">起始编号</label><input class="input" type="number" v-model="form.startNum" /></div>
        <div class="form-group"><label class="label">总页数</label><input class="input" type="number" v-model="form.totalPages" /></div>
        <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showCreate = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="createBook">{{ saving ? '创建中...' : '创建' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
