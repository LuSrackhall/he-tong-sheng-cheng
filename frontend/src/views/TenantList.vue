<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { tenantApi, contractApi, type Tenant, type Contract } from '../api'

function useDebounce<F extends (...args: any[]) => void>(fn: F, delay: number): F {
  let timer: ReturnType<typeof setTimeout>
  return ((...args: any[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }) as F
}

const tenants = ref<Tenant[]>([])
const total = ref(0)
const search = ref('')
const page = ref(0)
const pageSize = 20
const showDetail = ref(false)
const viewing = ref<Tenant | null>(null)
const editing = ref(false)
const form = ref({ name: '', phone: '', idCard: '' })
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')

// Related contracts
const relatedContracts = ref<Contract[]>([])
const loadingContracts = ref(false)

async function fetchTenants() {
  const { data } = await tenantApi.list({ search: search.value, offset: page.value * pageSize, limit: pageSize })
  tenants.value = data.data
  total.value = data.total
}

function openDetail(t: Tenant) {
  viewing.value = t
  editing.value = false
  form.value = { name: t.name, phone: t.phone || '', idCard: t.idCard || '' }
  showDetail.value = true
  fetchRelatedContracts(t.id)
}

function startEdit() {
  editing.value = true
  errorMessage.value = ''
}

function cancelEdit() {
  editing.value = false
  if (viewing.value) {
    form.value = { name: viewing.value.name, phone: viewing.value.phone || '', idCard: viewing.value.idCard || '' }
  }
  errorMessage.value = ''
}

async function fetchRelatedContracts(tenantId: number) {
  loadingContracts.value = true
  try {
    const { data } = await contractApi.list({ tenantId, limit: 50 })
    relatedContracts.value = data.data
  } catch {
    relatedContracts.value = []
  } finally {
    loadingContracts.value = false
  }
}

async function save() {
  if (submitLock.value) return
  if (!form.value.name) return
  if (!viewing.value) return
  submitLock.value = true
  saving.value = true
  errorMessage.value = ''
  try {
    await tenantApi.update(viewing.value.id, form.value)
    editing.value = false
    // Refresh viewing data
    const { data } = await tenantApi.get(viewing.value.id)
    viewing.value = data
    fetchTenants()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '保存失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

onMounted(fetchTenants)

const onSearchInput = useDebounce(() => { page.value = 0; fetchTenants() }, 300)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>租户管理 — 查看与补全</h2>
    </div>

    <div class="card" style="margin-bottom: var(--space-md); padding: var(--space-md); font-size: 0.8125rem; color: var(--color-text-secondary);">
      租户在签合同时自动创建，此处仅用于查看和补全信息
    </div>

    <div class="form-group">
      <input class="input" v-model="search" @input="onSearchInput" placeholder="搜索姓名/电话/身份证号..." />
    </div>

    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>姓名</th>
            <th>电话</th>
            <th>身份证号</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tenants" :key="t.id" @click="openDetail(t)" style="cursor: pointer;">
            <td>{{ t.name }}</td>
            <td>{{ t.phone || '—' }}</td>
            <td>{{ t.idCard || '—' }}</td>
            <td>{{ new Date(t.createdAt).toLocaleDateString('zh-CN') }}</td>
            <td>
              <button class="btn btn-secondary btn-sm" @click.stop="openDetail(t)">查看详情</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchTenants()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchTenants()">下一页</button>
    </div>

    <!-- Detail / Edit Modal -->
    <div v-if="showDetail" class="modal-overlay" @click.self="showDetail = false">
      <div class="modal-content" style="max-width: 640px;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-md);">
          <h3 style="margin-bottom: 0;">
            {{ editing ? '补全租户信息' : '租户详情' }}
          </h3>
          <button v-if="!editing" class="btn btn-secondary btn-sm" @click="startEdit">补全信息</button>
        </div>

        <!-- View mode -->
        <template v-if="!editing">
          <div style="display: flex; gap: var(--space-lg);">
            <div class="form-group" style="flex: 1;">
              <label class="label">姓名</label>
              <div style="font-size: 0.9375rem;">{{ viewing?.name }}</div>
            </div>
            <div class="form-group" style="flex: 1;">
              <label class="label">电话</label>
              <div style="font-size: 0.875rem;">{{ viewing?.phone || '—' }}</div>
            </div>
          </div>
          <div class="form-group">
            <label class="label">身份证号</label>
            <div style="font-size: 0.875rem;">{{ viewing?.idCard || '—' }}</div>
          </div>
          <div class="form-group">
            <label class="label">创建时间</label>
            <div style="font-size: 0.875rem;">{{ viewing?.createdAt ? new Date(viewing.createdAt).toLocaleDateString('zh-CN') : '' }}</div>
          </div>

          <!-- Related Contracts -->
          <div style="margin-top: var(--space-lg);">
            <label class="label" style="font-size: 0.875rem; font-weight: 600;">历史合同记录</label>
            <div v-if="loadingContracts" style="text-align: center; padding: var(--space-lg); color: var(--color-text-tertiary);">加载中...</div>
            <div v-else-if="relatedContracts.length === 0" style="text-align: center; padding: var(--space-md); color: var(--color-text-tertiary); font-size: 0.8125rem;">暂无关联合同</div>
            <div v-else style="margin-top: var(--space-sm);">
              <div
                v-for="c in relatedContracts"
                :key="c.id"
                style="padding: var(--space-sm) var(--space-md); border: 1px solid var(--color-border); border-radius: var(--radius-md); margin-bottom: var(--space-sm); font-size: 0.8125rem; display: flex; justify-content: space-between; align-items: center;"
              >
                <div>
                  <div style="font-weight: 500;">合同 #{{ c.id }}</div>
                  <div style="color: var(--color-text-secondary);">
                    资产：{{ c.asset?.name || '—' }}
                  </div>
                  <div style="color: var(--color-text-secondary);">
                    租期：{{ new Date(c.startDate).toLocaleDateString('zh-CN') }} — {{ new Date(c.endDate).toLocaleDateString('zh-CN') }}
                  </div>
                </div>
                <div style="text-align: right;">
                  <div style="font-weight: 500; color: var(--color-primary);">¥{{ c.monthlyRent.toLocaleString() }}/月</div>
                  <span :class="{ 'badge badge-success': c.status === 'active', 'badge badge-warning': c.status === 'pending', 'badge badge-info': c.status === 'ended' }">
                    {{ c.status }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- Edit mode -->
        <template v-if="editing">
          <div class="form-group">
            <label class="label">姓名</label>
            <input class="input" v-model="form.name" />
          </div>
          <div class="form-group">
            <label class="label">电话</label>
            <input class="input" v-model="form.phone" />
          </div>
          <div class="form-group">
            <label class="label">身份证号</label>
            <input class="input" v-model="form.idCard" />
          </div>
          <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>
        </template>

        <div style="display: flex; gap: 8px; justify-content: flex-end; margin-top: var(--space-md);">
          <button v-if="editing" class="btn btn-secondary" @click="cancelEdit">取消</button>
          <button v-if="editing" class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? '保存中...' : '保存' }}
          </button>
          <button v-if="!editing" class="btn btn-secondary" @click="showDetail = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>
