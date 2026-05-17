<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { contractApi, type Contract } from '../api'

const contracts = ref<Contract[]>([])
const total = ref(0)
const search = ref('')
const statusFilter = ref('')
const page = ref(0)
const pageSize = 20
const showDetail = ref(false)
const detailContract = ref<Contract | null>(null)
const showEditModal = ref(false)
const editing = ref<Contract | null>(null)
const form = ref({ startDate: '', endDate: '', monthlyRent: 0, totalReceivable: 0, deposit: 0, notes: '' })
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')

async function fetchContracts() {
  const params: any = { search: search.value, offset: page.value * pageSize, limit: pageSize }
  if (statusFilter.value) params.status = statusFilter.value
  const { data } = await contractApi.list(params)
  contracts.value = data.data
  total.value = data.total
}
function openDetail(c: Contract) { detailContract.value = c; showDetail.value = true }
function openEdit(c: Contract) {
  editing.value = c
  form.value = {
    startDate: c.startDate?.substring(0, 10),
    endDate: c.endDate?.substring(0, 10),
    monthlyRent: c.monthlyRent,
    totalReceivable: c.totalReceivable,
    deposit: c.deposit,
    notes: c.notes || '',
  }
  showEditModal.value = true
}
async function save() {
  if (submitLock.value) return
  if (!editing.value || !form.value.startDate || !form.value.endDate) return
  submitLock.value = true
  saving.value = true
  try {
    await contractApi.update(editing.value.id, form.value)
    showEditModal.value = false
    errorMessage.value = ''
    fetchContracts()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '保存失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

onMounted(fetchContracts)
</script>

<template>
  <div>
    <div class="page-header"><h2>合同管理</h2></div>

    <div style="display: flex; gap: 12px; margin-bottom: var(--space-lg);">
      <input class="input" v-model="search" @input="fetchContracts" placeholder="搜索租户/资产名称..." style="flex: 1;" />
      <select class="input" v-model="statusFilter" @change="fetchContracts" style="width: 140px;">
        <option value="">全部状态</option>
        <option value="active">执行中</option>
        <option value="paidup">已缴清</option>
        <option value="arrears">欠费中</option>
        <option value="expired">已到期</option>
      </select>
    </div>

    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>租户</th>
            <th>资产</th>
            <th>租期</th>
            <th>月租金</th>
            <th>已收/应收</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in contracts" :key="c.id">
            <td>#{{ c.id }}</td>
            <td>{{ c.tenant?.name || '-' }}</td>
            <td>{{ c.asset?.name || '-' }}</td>
            <td style="font-size: 0.8125rem;">{{ c.startDate?.substring(0, 10) }} ~ {{ c.endDate?.substring(0, 10) }}</td>
            <td>¥{{ c.monthlyRent?.toLocaleString() }}</td>
            <td>¥{{ c.totalReceived?.toLocaleString() }} / ¥{{ c.totalReceivable?.toLocaleString() }}</td>
            <td>
              <span class="badge" :class="{
                'badge-success': c.status === 'paidup',
                'badge-warning': c.status === 'active',
                'badge-danger': c.status === 'arrears',
                'badge-info': c.status === 'expired',
              }">{{ c.status }}</span>
            </td>
            <td>
              <button class="btn btn-secondary btn-sm" @click="openDetail(c)">详情</button>
              <button class="btn btn-secondary btn-sm" style="margin-left: 4px;" @click="openEdit(c)">编辑</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchContracts()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchContracts()">下一页</button>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetail && detailContract" class="modal-overlay" @click.self="showDetail = false">
      <div class="modal-content" style="max-width: 600px;">
        <h3>合同详情 #{{ detailContract.id }}</h3>
        <table>
          <tbody>
            <tr><td style="color: var(--color-text-secondary);">租户</td><td>{{ detailContract.tenant?.name || '-' }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">资产</td><td>{{ detailContract.asset?.name || '-' }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">租期</td><td>{{ detailContract.startDate?.substring(0, 10) }} ~ {{ detailContract.endDate?.substring(0, 10) }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">月租金</td><td>¥{{ detailContract.monthlyRent?.toLocaleString() }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">应收总额</td><td>¥{{ detailContract.totalReceivable?.toLocaleString() }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">已收金额</td><td>¥{{ detailContract.totalReceived?.toLocaleString() }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">押金</td><td>¥{{ detailContract.deposit?.toLocaleString() }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">状态</td>
              <td>
                <span class="badge" :class="{
                  'badge-success': detailContract.status === 'paidup',
                  'badge-warning': detailContract.status === 'active',
                  'badge-danger': detailContract.status === 'arrears',
                  'badge-info': detailContract.status === 'expired',
                }">{{ detailContract.status }}</span>
              </td>
            </tr>
            <tr v-if="detailContract.notes"><td style="color: var(--color-text-secondary);">备注</td><td>{{ detailContract.notes }}</td></tr>
            <tr><td style="color: var(--color-text-secondary);">创建时间</td><td>{{ new Date(detailContract.createdAt).toLocaleString('zh-CN') }}</td></tr>
          </tbody>
        </table>
        <button class="btn btn-secondary" style="margin-top: 16px;" @click="showDetail = false">关闭</button>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal-content">
        <h3>编辑合同</h3>
        <div class="form-group"><label class="label">开始日期</label><input class="input" type="date" v-model="form.startDate" /></div>
        <div class="form-group"><label class="label">结束日期</label><input class="input" type="date" v-model="form.endDate" /></div>
        <div class="form-group"><label class="label">月租金</label><input class="input" type="number" v-model="form.monthlyRent" /></div>
        <div class="form-group"><label class="label">应收总额</label><input class="input" type="number" v-model="form.totalReceivable" /></div>
        <div class="form-group"><label class="label">押金</label><input class="input" type="number" v-model="form.deposit" /></div>
        <div class="form-group"><label class="label">备注</label><input class="input" v-model="form.notes" /></div>
        <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showEditModal = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
