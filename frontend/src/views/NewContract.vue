<script setup lang="ts">
import { ref, computed } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { assetApi, tenantApi, contractApi, type Asset, type Tenant, type Contract } from '../api'

const step = ref(1)
const assets = ref<Asset[]>([])
const tenants = ref<Tenant[]>([])
const assetSearch = ref('')
const tenantSearch = ref('')
const selectedAsset = ref<Asset | null>(null)
const selectedTenant = ref<Tenant | null>(null)
const showNewAsset = ref(false)
const showNewTenant = ref(false)
const newAsset = ref({ name: '', assetType: 'shop', description: '' })
const newTenant = ref({ name: '', phone: '', idCard: '' })
const contract = ref({ startDate: '', endDate: '', monthlyRent: 0, totalReceivable: 0, deposit: 0, notes: '' })
const saving = ref(false)
const submitLock = ref(false)
const createdContract = ref<Contract | null>(null)

const hasUnsavedChanges = computed(() =>
  selectedAsset.value !== null || selectedTenant.value !== null ||
  contract.value.startDate !== '' || contract.value.endDate !== '' ||
  contract.value.monthlyRent > 0
)

onBeforeRouteLeave((_to, _from, next) => {
  if (hasUnsavedChanges.value && step.value !== 4) {
    const confirmed = window.confirm('有未保存的合同数据，确定要离开吗？')
    if (!confirmed) return next(false)
  }
  next()
})

async function searchAssets() {
  if (assetSearch.value.length < 1) { assets.value = []; return }
  const { data } = await assetApi.list({ search: assetSearch.value, limit: 10 })
  assets.value = data.data
}
async function searchTenants() {
  if (tenantSearch.value.length < 1) { tenants.value = []; return }
  const { data } = await tenantApi.list({ search: tenantSearch.value, limit: 10 })
  tenants.value = data.data
}
async function createAsset() {
  if (!newAsset.value.name) return
  await assetApi.create(newAsset.value)
  showNewAsset.value = false
  newAsset.value = { name: '', assetType: 'shop', description: '' }
  searchAssets()
}
async function createTenant() {
  if (!newTenant.value.name) return
  await tenantApi.create(newTenant.value)
  showNewTenant.value = false
  newTenant.value = { name: '', phone: '', idCard: '' }
  searchTenants()
}
async function createContract() {
  if (submitLock.value) return
  if (!selectedAsset.value || !selectedTenant.value) return
  if (!contract.value.startDate || !contract.value.endDate || !contract.value.monthlyRent) return
  submitLock.value = true
  saving.value = true
  try {
    const { data } = await contractApi.create({
      assetId: selectedAsset.value.id,
      tenantId: selectedTenant.value.id,
      startDate: contract.value.startDate,
      endDate: contract.value.endDate,
      monthlyRent: contract.value.monthlyRent,
      totalReceivable: contract.value.totalReceivable,
      deposit: contract.value.deposit,
      notes: contract.value.notes,
    })
    createdContract.value = data
    step.value = 4
  } finally {
    saving.value = false
    submitLock.value = false
  }
}
function reset() {
  step.value = 1
  selectedAsset.value = null
  selectedTenant.value = null
  contract.value = { startDate: '', endDate: '', monthlyRent: 0, totalReceivable: 0, deposit: 0, notes: '' }
  createdContract.value = null
}
</script>

<template>
  <div>
    <div class="page-header"><h2>签新合同</h2></div>

    <div class="steps">
      <div class="step" :class="{ active: step === 1, completed: step > 1 }">① 选资产</div>
      <div class="step" :class="{ active: step === 2, completed: step > 2 }">② 录租户</div>
      <div class="step" :class="{ active: step === 3, completed: step > 3 }">③ 定合同</div>
      <div class="step" :class="{ active: step === 4 }">④ 预览打印</div>
    </div>

    <!-- Step 1: Select Asset -->
    <div v-if="step === 1" class="card">
      <h3 style="margin-bottom: 16px;">选择资产</h3>
      <div class="form-group">
        <input class="input" v-model="assetSearch" @input="searchAssets" placeholder="搜索资产名称..." />
      </div>
      <div v-if="assets.length" style="margin-bottom: 12px;">
        <div v-for="a in assets" :key="a.id" class="card" style="padding: 12px; margin-bottom: 8px; cursor: pointer;" @click="selectedAsset = a; step = 2" :style="{ borderColor: selectedAsset?.id === a.id ? 'var(--color-primary)' : '' }">
          <div style="font-weight: 500;">{{ a.name }}</div>
          <div style="font-size: 0.8125rem; color: var(--color-text-secondary);">{{ a.assetType }} · {{ a.status }}</div>
        </div>
      </div>
      <button class="btn btn-secondary" @click="showNewAsset = true">+ 新建资产</button>

      <div v-if="showNewAsset" class="modal-overlay" @click.self="showNewAsset = false">
        <div class="modal-content">
          <h3>新建资产</h3>
          <div class="form-group"><label class="label">名称</label><input class="input" v-model="newAsset.name" /></div>
          <div class="form-group"><label class="label">类型</label>
            <select class="input" v-model="newAsset.assetType">
              <option value="shop">商铺</option>
              <option value="parking">车位</option>
              <option value="booth">摊位</option>
              <option value="equipment">设备</option>
              <option value="other">其他</option>
            </select>
          </div>
          <div class="form-group"><label class="label">描述</label><input class="input" v-model="newAsset.description" /></div>
          <div style="display: flex; gap: 8px; justify-content: flex-end;">
            <button class="btn btn-secondary" @click="showNewAsset = false">取消</button>
            <button class="btn btn-primary" @click="createAsset">创建</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 2: Select Tenant -->
    <div v-if="step === 2" class="card">
      <h3 style="margin-bottom: 16px;">录入租户</h3>
      <div v-if="selectedAsset" style="margin-bottom: 12px; font-size: 0.875rem; color: var(--color-text-secondary);">已选资产：{{ selectedAsset.name }}</div>
      <div class="form-group">
        <input class="input" v-model="tenantSearch" @input="searchTenants" placeholder="搜索租户姓名/电话..." />
      </div>
      <div v-if="tenants.length" style="margin-bottom: 12px;">
        <div v-for="t in tenants" :key="t.id" class="card" style="padding: 12px; margin-bottom: 8px; cursor: pointer;" @click="selectedTenant = t; step = 3" :style="{ borderColor: selectedTenant?.id === t.id ? 'var(--color-primary)' : '' }">
          <div style="font-weight: 500;">{{ t.name }}</div>
          <div style="font-size: 0.8125rem; color: var(--color-text-secondary);">{{ t.phone }} · {{ t.idCard }}</div>
        </div>
      </div>
      <div style="display: flex; gap: 8px;">
        <button class="btn btn-secondary" @click="step = 1">← 返回</button>
        <button class="btn btn-secondary" @click="showNewTenant = true">+ 新建租户</button>
      </div>

      <div v-if="showNewTenant" class="modal-overlay" @click.self="showNewTenant = false">
        <div class="modal-content">
          <h3>新建租户</h3>
          <div class="form-group"><label class="label">姓名</label><input class="input" v-model="newTenant.name" /></div>
          <div class="form-group"><label class="label">电话</label><input class="input" v-model="newTenant.phone" /></div>
          <div class="form-group"><label class="label">身份证号</label><input class="input" v-model="newTenant.idCard" /></div>
          <div style="display: flex; gap: 8px; justify-content: flex-end;">
            <button class="btn btn-secondary" @click="showNewTenant = false">取消</button>
            <button class="btn btn-primary" @click="createTenant">创建</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3: Contract Details -->
    <div v-if="step === 3" class="card">
      <h3 style="margin-bottom: 16px;">合同详情</h3>
      <div style="margin-bottom: 12px; font-size: 0.875rem; color: var(--color-text-secondary);">
        资产：{{ selectedAsset?.name }} · 租户：{{ selectedTenant?.name }}
      </div>
      <div class="form-group"><label class="label">开始日期</label><input class="input" type="date" v-model="contract.startDate" /></div>
      <div class="form-group"><label class="label">结束日期</label><input class="input" type="date" v-model="contract.endDate" /></div>
      <div class="form-group"><label class="label">月租金</label><input class="input" type="number" v-model="contract.monthlyRent" /></div>
      <div class="form-group"><label class="label">应收总额（留空自动计算）</label><input class="input" type="number" v-model="contract.totalReceivable" placeholder="自动计算：整月×月租+零天×日租" /></div>
      <div class="form-group"><label class="label">押金</label><input class="input" type="number" v-model="contract.deposit" /></div>
      <div class="form-group"><label class="label">备注</label><input class="input" v-model="contract.notes" /></div>
      <div style="display: flex; gap: 8px;">
        <button class="btn btn-secondary" @click="step = 2">← 返回</button>
        <button class="btn btn-primary" :disabled="saving" @click="createContract">{{ saving ? '创建中...' : '创建合同' }}</button>
      </div>
    </div>

    <!-- Step 4: Preview -->
    <div v-if="step === 4 && createdContract" class="card">
      <h3 style="margin-bottom: 16px;">合同创建成功</h3>
      <table>
        <tbody>
          <tr><td style="color: var(--color-text-secondary);">合同编号</td><td>{{ createdContract.id }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">资产</td><td>{{ selectedAsset?.name }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">租户</td><td>{{ selectedTenant?.name }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">租期</td><td>{{ createdContract.startDate?.toString().substring(0,10) }} 至 {{ createdContract.endDate?.toString().substring(0,10) }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">月租金</td><td>¥{{ createdContract.monthlyRent }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">应收总额</td><td>¥{{ createdContract.totalReceivable }}</td></tr>
        </tbody>
      </table>
      <div style="margin-top: 20px; display: flex; gap: 8px;">
        <button class="btn btn-primary" @click="reset">签下一份合同</button>
      </div>
    </div>
  </div>
</template>
