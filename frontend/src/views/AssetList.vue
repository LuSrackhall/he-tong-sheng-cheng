<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { assetApi, contractApi, type Asset, type Contract } from '../api'

function useDebounce<F extends (...args: any[]) => void>(fn: F, delay: number): F {
  let timer: ReturnType<typeof setTimeout>
  return ((...args: any[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }) as F
}

const assets = ref<Asset[]>([])
const total = ref(0)
const search = ref('')
const page = ref(0)
const pageSize = 20
const showDetail = ref(false)
const viewing = ref<Asset | null>(null)
const editing = ref(false)
const form = ref({ name: '', assetType: 'shop', description: '' })
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')

// Related contracts
const relatedContracts = ref<Contract[]>([])
const loadingContracts = ref(false)

async function fetchAssets() {
  const { data } = await assetApi.list({ search: search.value, offset: page.value * pageSize, limit: pageSize })
  assets.value = data.data
  total.value = data.total
}

function openDetail(a: Asset) {
  viewing.value = a
  editing.value = false
  form.value = { name: a.name, assetType: a.assetType, description: a.description || '' }
  showDetail.value = true
  fetchRelatedContracts(a.id)
}

function startEdit() {
  editing.value = true
  errorMessage.value = ''
}

function cancelEdit() {
  editing.value = false
  if (viewing.value) {
    form.value = { name: viewing.value.name, assetType: viewing.value.assetType, description: viewing.value.description || '' }
  }
  errorMessage.value = ''
}

async function fetchRelatedContracts(assetId: number) {
  loadingContracts.value = true
  try {
    const { data } = await contractApi.list({ assetId, limit: 50 })
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
    await assetApi.update(viewing.value.id, form.value)
    editing.value = false
    // Refresh viewing data
    const { data } = await assetApi.get(viewing.value.id)
    viewing.value = data
    fetchAssets()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '保存失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = { shop: '商铺', parking: '车位', stall: '摊位', equipment: '设备', other: '其他' }
  return map[type] || type
}

function getStatusLabel(status: string): string {
  const map: Record<string, string> = { leased: '已出租', idle: '空闲' }
  return map[status] || status
}

onMounted(fetchAssets)

const onSearchInput = useDebounce(() => { page.value = 0; fetchAssets() }, 300)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>资产管理 — 查看与补全</h2>
    </div>

    <div class="card" style="margin-bottom: var(--space-md); padding: var(--space-md); font-size: 0.8125rem; color: var(--color-text-secondary);">
      资产在签合同时自动创建，此处仅用于查看和补全信息
    </div>

    <div class="form-group">
      <input class="input" v-model="search" @input="onSearchInput" placeholder="搜索资产名称..." />
    </div>

    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in assets" :key="a.id" @click="openDetail(a)" style="cursor: pointer;">
            <td>{{ a.name }}</td>
            <td>{{ getTypeLabel(a.assetType) }}</td>
            <td>
              <span :class="{ 'badge badge-success': a.status === 'leased', 'badge badge-info': a.status === 'idle' }">
                {{ getStatusLabel(a.status) }}
              </span>
            </td>
            <td>{{ new Date(a.createdAt).toLocaleDateString('zh-CN') }}</td>
            <td>
              <button class="btn btn-secondary btn-sm" @click.stop="openDetail(a)">查看详情</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchAssets()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchAssets()">下一页</button>
    </div>

    <!-- Detail / Edit Modal -->
    <div v-if="showDetail" class="modal-overlay" @click.self="showDetail = false">
      <div class="modal-content" style="max-width: 640px;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-md);">
          <h3 style="margin-bottom: 0;">
            {{ editing ? '补全资产信息' : '资产详情' }}
          </h3>
          <button v-if="!editing" class="btn btn-secondary btn-sm" @click="startEdit">补全信息</button>
        </div>

        <!-- View mode -->
        <template v-if="!editing">
          <div class="form-group">
            <label class="label">名称</label>
            <div style="font-size: 0.9375rem;">{{ viewing?.name }}</div>
          </div>
          <div style="display: flex; gap: var(--space-lg);">
            <div class="form-group" style="flex: 1;">
              <label class="label">类型</label>
              <div style="font-size: 0.875rem;">{{ getTypeLabel(viewing?.assetType || '') }}</div>
            </div>
            <div class="form-group" style="flex: 1;">
              <label class="label">状态</label>
              <span :class="{ 'badge badge-success': viewing?.status === 'leased', 'badge badge-info': viewing?.status === 'idle' }">
                {{ getStatusLabel(viewing?.status || '') }}
              </span>
            </div>
          </div>
          <div class="form-group">
            <label class="label">描述</label>
            <div style="font-size: 0.875rem; color: var(--color-text-secondary);">{{ viewing?.description || '暂无描述' }}</div>
          </div>
          <div class="form-group">
            <label class="label">创建时间</label>
            <div style="font-size: 0.875rem;">{{ viewing?.createdAt ? new Date(viewing.createdAt).toLocaleDateString('zh-CN') : '' }}</div>
          </div>

          <!-- Related Contracts -->
          <div style="margin-top: var(--space-lg);">
            <label class="label" style="font-size: 0.875rem; font-weight: 600;">历史租赁记录</label>
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
                    租户：{{ c.tenant?.name || '—' }}
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
            <label class="label">名称</label>
            <input class="input" v-model="form.name" />
          </div>
          <div class="form-group">
            <label class="label">类型</label>
            <select class="input" v-model="form.assetType">
              <option value="shop">商铺</option>
              <option value="parking">车位</option>
              <option value="stall">摊位</option>
              <option value="equipment">设备</option>
              <option value="other">其他</option>
            </select>
          </div>
          <div class="form-group">
            <label class="label">描述</label>
            <input class="input" v-model="form.description" />
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
