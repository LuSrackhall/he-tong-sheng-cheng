<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { onBeforeRouteLeave, useRouter } from 'vue-router'
import { assetApi, tenantApi, contractApi, templateApi, type Asset, type Tenant, type Contract, type Template } from '../api'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()

const router = useRouter()

// Field source classification for template-driven rendering
const SYSTEM_AUTO_FIELDS = new Set(['contractId', 'totalReceivable', 'totalReceived', 'deposit', 'status', 'notes'])
const ASSET_TENANT_FIELDS = new Set(['assetName', 'assetType', 'assetDescription', 'tenantName', 'tenantIDCard', 'tenantPhone'])
const USER_INPUT_FIELDS = new Set(['startDate', 'endDate', 'monthlyRent', 'yearlyRent'])

function classifyField(key: string): 'system-auto' | 'asset-tenant' | 'user-input' {
  if (SYSTEM_AUTO_FIELDS.has(key)) return 'system-auto'
  if (ASSET_TENANT_FIELDS.has(key)) return 'asset-tenant'
  if (USER_INPUT_FIELDS.has(key)) return 'user-input'
  return 'user-input'
}

// ---- state ----
const step = ref(0)

// Step 0: templates
const templates = ref<Template[]>([])
const selectedTemplate = ref<Template | null>(null)
const loadingTemplates = ref(false)

const requiredFieldKeys = ['startDate', 'endDate', 'monthlyRent', 'tenantName', 'assetName']

function parseActiveFieldsArray(raw: string): Record<string, boolean> {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      const obj: Record<string, boolean> = {}
      for (const k of parsed) {
        if (typeof k === 'string') obj[k] = true
      }
      return obj
    }
    if (typeof parsed === 'object' && parsed !== null) {
      const obj: Record<string, boolean> = {}
      for (const [k, v] of Object.entries(parsed)) {
        obj[k] = !!v
      }
      return obj
    }
    return {}
  } catch { return {} }
}

function isTemplateUsable(t: Template): boolean {
  if (!t.validated || !t.filePath) return false
  const activeMap = parseActiveFieldsArray(t.activeFields || '')
  return requiredFieldKeys.every(k => activeMap[k] === true)
}

// Step 1: assets
const assetSearch = ref('')
const assets = ref<Asset[]>([])
const selectedAsset = ref<Asset | null>(null)
const newAssetName = ref('')
const newAssetType = ref('shop')
const newAssetMonthlyRent = ref<number | null>(null)
const showNewAssetForm = ref(false)

// Step 2: tenants
const tenantSearch = ref('')
const tenants = ref<Tenant[]>([])
const selectedTenant = ref<Tenant | null>(null)
const newTenantName = ref('')
const newTenantIdCard = ref('')
const newTenantPhone = ref('')
const showNewTenantForm = ref(false)

// Step 3: contract details
const contractStartDate = ref('')
const contractEndDate = ref('')
const contractMonthlyRent = ref<number | null>(null)
const contractYearlyRent = ref<number | null>(null)
const contractTotalReceivable = ref<number>(0)
const manualTotal = ref(false)
const contractDeposit = ref<number | null>(null)
const contractNotes = ref('')
const rentLinked = ref(true)  // monthlyRent/yearlyRent linked conversion

// Template-driven active fields list
const activeFieldsList = computed(() => {
  if (!selectedTemplate.value) return []
  const afMap = parseActiveFieldsArray(selectedTemplate.value.activeFields || '')
  return Object.keys(afMap).filter(k => afMap[k] === true)
})

// Yearly rent linked conversion
watch(contractMonthlyRent, (val) => {
  if (rentLinked.value && val) {
    contractYearlyRent.value = Math.round(val * 12 * 100) / 100
  }
})
watch(contractYearlyRent, (val) => {
  if (rentLinked.value && val) {
    contractMonthlyRent.value = Math.round(val / 12 * 100) / 100
  }
})

function onYearlyRentInput() {
  if (!rentLinked.value && contractYearlyRent.value) {
    manualTotal.value = false
  }
}

// shared
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')
const createdContract = ref<Contract | null>(null)
const downloading = ref(false)

// ---- computed ----
const hasUnsavedChanges = computed(() =>
  selectedAsset.value !== null || selectedTenant.value !== null ||
  contractStartDate.value !== '' || contractEndDate.value !== '' ||
  (contractMonthlyRent.value && contractMonthlyRent.value > 0)
)

// Auto-calc total receivable
watch([contractStartDate, contractEndDate, contractMonthlyRent], () => {
  if (manualTotal.value) return
  if (!contractStartDate.value || !contractEndDate.value || !contractMonthlyRent.value) {
    contractTotalReceivable.value = 0
    return
  }
  const start = new Date(contractStartDate.value)
  const end = new Date(contractEndDate.value)
  if (end <= start) {
    contractTotalReceivable.value = 0
    return
  }
  const monthlyRent = contractMonthlyRent.value
  const dailyRent = monthlyRent / 30
  let months = 0
  let remainingStart = new Date(start)
  // Count full months
  while (true) {
    const candidate = new Date(remainingStart)
    candidate.setMonth(candidate.getMonth() + 1)
    if (candidate <= end) {
      months++
      remainingStart = candidate
    } else {
      break
    }
  }
  const remainingDays = Math.max(0, Math.ceil((end.getTime() - remainingStart.getTime()) / (1000 * 60 * 60 * 24)))
  const total = months * monthlyRent + remainingDays * dailyRent
  contractTotalReceivable.value = Math.round(total * 100) / 100
})

// Auto-select single template (only if usable)
watch(templates, (list) => {
  if (list.length === 1 && step.value === 0) {
    const t = list[0]
    if (isTemplateUsable(t)) {
      selectedTemplate.value = t
    }
  }
  // Also auto-select if only one usable template
  const usable = list.filter(t => isTemplateUsable(t))
  if (usable.length === 1 && step.value === 0 && !selectedTemplate.value) {
    selectedTemplate.value = usable[0]
  }
})

// Step guard
onBeforeRouteLeave((_to, _from, next) => {
  if (hasUnsavedChanges.value && step.value < 4 && !createdContract.value) {
    const confirmed = window.confirm('有未保存的合同数据，确定要离开吗？')
    if (!confirmed) return next(false)
  }
  next()
})

// ---- template-driven field helpers ----
const fieldLabels: Record<string, string> = {
  contractId: '合同编号', startDate: '开始日期', endDate: '结束日期',
  monthlyRent: '月租金', yearlyRent: '年租金', totalReceivable: '应收总额',
  totalReceived: '已收总额', deposit: '押金', status: '状态', notes: '备注',
  assetName: '资产名称', assetType: '资产类型', assetDescription: '资产描述',
  tenantName: '租户姓名', tenantIDCard: '身份证号', tenantPhone: '联系电话',
  today: '今日日期',
}

function getFieldLabel(key: string): string {
  return fieldLabels[key] || key
}

function isRequiredField(key: string): boolean {
  return requiredFieldKeys.includes(key)
}

function getSystemAutoValue(key: string): string {
  switch (key) {
    case 'contractId': return selectedTemplate.value ? '将在创建后自动生成' : ''
    case 'totalReceivable': return contractTotalReceivable.value > 0 ? contractTotalReceivable.value.toFixed(2) : '计算中...'
    case 'totalReceived': return '0'
    case 'deposit': return contractDeposit.value ? contractDeposit.value.toFixed(2) : ''
    case 'status': return 'active'
    case 'notes': return contractNotes.value
    default: return ''
  }
}

function getAssetTenantValue(key: string): string {
  if (key.startsWith('tenant')) {
    if (selectedTenant.value) {
      const t = selectedTenant.value as any
      switch (key) {
        case 'tenantName': return t.name || ''
        case 'tenantIDCard': return t.idCard || ''
        case 'tenantPhone': return t.phone || ''
        default: return ''
      }
    } else {
      switch (key) {
        case 'tenantName': return newTenantName.value
        case 'tenantIDCard': return newTenantIdCard.value || '后续补全'
        case 'tenantPhone': return newTenantPhone.value || '后续补全'
        default: return ''
      }
    }
  }
  if (selectedAsset.value) {
    const a = selectedAsset.value as any
    switch (key) {
      case 'assetName': return a.name || ''
      case 'assetType': return a.assetType || ''
      case 'assetDescription': return a.description || ''
      default: return ''
    }
  } else {
    switch (key) {
      case 'assetName': return newAssetName.value
      case 'assetType': return newAssetType.value
      case 'assetDescription': return '后续补全'
      default: return ''
    }
  }
}

// ---- data fetching ----
async function fetchTemplates() {
  loadingTemplates.value = true
  try {
    const { data } = await templateApi.list()
    templates.value = (data as any).data || (Array.isArray(data) ? data : [])
  } catch {
    templates.value = []
  } finally {
    loadingTemplates.value = false
  }
}

async function searchAssets() {
  if (assetSearch.value.length < 1) { assets.value = []; return }
  try {
    const { data } = await assetApi.list({ search: assetSearch.value, limit: 10 })
    assets.value = data.data || []
  } catch {
    assets.value = []
  }
}

async function searchTenants() {
  if (tenantSearch.value.length < 1) { tenants.value = []; return }
  try {
    const { data } = await tenantApi.list({ search: tenantSearch.value, limit: 10 })
    tenants.value = data.data || []
  } catch {
    tenants.value = []
  }
}

// ---- step navigation ----
function selectTemplate(t: Template) {
  if (!isTemplateUsable(t)) return
  selectedTemplate.value = t
}

function gotoStep1() {
  if (!selectedTemplate.value) return
  step.value = 1
}

function selectAsset(a: Asset) {
  selectedAsset.value = a
  newAssetName.value = ''
  showNewAssetForm.value = false
}


function gotoStep2() {
  if (!selectedAsset.value && !newAssetName.value.trim()) {
    errorMessage.value = '请选择已有资产或输入新资产名称'
    return
  }
  step.value = 2
  errorMessage.value = ''
}

function selectTenant(t: Tenant) {
  selectedTenant.value = t
  newTenantName.value = ''
  showNewTenantForm.value = false
}

function gotoStep3() {
  if (!selectedTenant.value && !newTenantName.value.trim()) {
    errorMessage.value = '请选择已有租户或输入新租户姓名'
    return
  }
  step.value = 3
  errorMessage.value = ''
}

// ---- Step 3 validation ----
function validateStep3(): boolean {
  if (!contractStartDate.value || !contractEndDate.value) {
    errorMessage.value = '请填写合同开始和结束日期'
    return false
  }
  if (new Date(contractEndDate.value) <= new Date(contractStartDate.value)) {
    errorMessage.value = '结束日期必须大于开始日期'
    return false
  }
  if (!contractMonthlyRent.value || contractMonthlyRent.value <= 0) {
    errorMessage.value = '月租金必须大于 0'
    return false
  }
  return true
}

// ---- create contract ----
async function createContract() {
  if (submitLock.value) return
  if (!validateStep3()) return

  // If new asset, create first
  let assetId = selectedAsset.value?.id
  if (!assetId && newAssetName.value.trim()) {
    try {
      const { data } = await assetApi.create({
        name: newAssetName.value.trim(),
        assetType: newAssetType.value,
        description: '后续可补全',
      })
      selectedAsset.value = data
      assetId = data.id
    } catch (err: any) {
      errorMessage.value = err.response?.data?.error || '资产创建失败'
      return
    }
  }

  // If new tenant, create first
  let tenantId = selectedTenant.value?.id
  if (!tenantId && newTenantName.value.trim()) {
    try {
      const { data } = await tenantApi.create({
        name: newTenantName.value.trim(),
        idCard: newTenantIdCard.value || undefined,
        phone: newTenantPhone.value || undefined,
      })
      selectedTenant.value = data
      tenantId = data.id
    } catch (err: any) {
      errorMessage.value = err.response?.data?.error || '租户创建失败'
      return
    }
  }

  if (!assetId || !tenantId) {
    errorMessage.value = '请选择资产和租户'
    return
  }

  submitLock.value = true
  saving.value = true
  errorMessage.value = ''

  try {
    const payload: any = {
      assetId,
      tenantId,
      startDate: contractStartDate.value,
      endDate: contractEndDate.value,
      monthlyRent: contractMonthlyRent.value,
      totalReceivable: contractTotalReceivable.value,
      deposit: contractDeposit.value ?? 0,
      notes: contractNotes.value || undefined,
    }
    if (selectedTemplate.value) {
      payload.templateId = selectedTemplate.value.id
    }

    const { data } = await contractApi.create(payload)
    createdContract.value = data
    step.value = 4
    toast.success('合同创建成功！')
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '合同创建失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

// ---- Step 4: download ----
async function handleDownload() {
  if (!createdContract.value || downloading.value) return
  downloading.value = true
  errorMessage.value = ''
  try {
    const response = await contractApi.download(createdContract.value.id)
    const url = window.URL.createObjectURL(new Blob([response.data as any]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `contract-${createdContract.value.id}.docx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    toast.success('下载完成，请用 Word 打开并打印')
  } catch (err: any) {
    if (err.response?.status === 400) {
      errorMessage.value = err.response?.data?.error || '模板校验未通过，请先在设置中重新上传符合要求的 Word 文件'
    } else {
      errorMessage.value = err.response?.data?.error || '下载失败'
    }
  } finally {
    downloading.value = false
  }
}

// ---- reset ----
function resetAll() {
  step.value = 0
  selectedTemplate.value = null
  selectedAsset.value = null
  selectedTenant.value = null
  assetSearch.value = ''
  tenantSearch.value = ''
  assets.value = []
  tenants.value = []
  newAssetName.value = ''
  newAssetType.value = 'shop'
  newAssetMonthlyRent.value = null
  showNewAssetForm.value = false
  newTenantName.value = ''
  newTenantIdCard.value = ''
  newTenantPhone.value = ''
  showNewTenantForm.value = false
  contractStartDate.value = ''
  contractEndDate.value = ''
  contractMonthlyRent.value = null
  contractYearlyRent.value = null
  contractTotalReceivable.value = 0
  manualTotal.value = false
  contractDeposit.value = null
  contractNotes.value = ''
  errorMessage.value = ''
  createdContract.value = null
  // Re-fetch templates in case new ones were added
  fetchTemplates()
}

function gotoStep0() {
  step.value = 0
}

// ---- init ----
onMounted(fetchTemplates)
</script>

<template>
  <div>
    <div class="page-header"><h2>签新合同</h2></div>

    <!-- Steps indicator -->
    <div class="steps">
      <div class="step" :class="{ active: step === 0, completed: step > 0 }">0. 选模板</div>
      <div class="step" :class="{ active: step === 1, completed: step > 1 }">1. 选资产</div>
      <div class="step" :class="{ active: step === 2, completed: step > 2 }">2. 录租户</div>
      <div class="step" :class="{ active: step === 3, completed: step > 3 }">3. 定合同</div>
      <div class="step" :class="{ active: step === 4 }">4. 预览导出</div>
    </div>

    <!-- ======================== Step 0: Select Template ======================== -->
    <div v-if="step === 0" class="card">
      <h3 style="margin-bottom: 16px;">选择合同模板</h3>

      <div v-if="loadingTemplates" style="text-align: center; padding: 24px; color: var(--color-text-secondary);">
        加载模板中...
      </div>

      <div v-else-if="templates.length === 0" class="empty-state">
        <p style="margin-bottom: 12px;">暂无可用的合同模板</p>
        <p style="font-size: 0.8125rem; margin-bottom: 16px;">请先去「设置」页面上传合同模板文件，然后再来签合同。</p>
        <button class="btn btn-primary" @click="router.push('/settings')">前往设置上传模板</button>
      </div>

      <div v-else style="display: grid; gap: 12px;">
        <div
          v-for="t in templates"
          :key="t.id"
          class="card"
          :style="{
            padding: '14px 16px',
            cursor: isTemplateUsable(t) ? 'pointer' : 'not-allowed',
            opacity: isTemplateUsable(t) ? 1 : 0.55,
            transition: 'all var(--transition-fast)',
            borderColor: selectedTemplate?.id === t.id ? 'var(--color-primary)' : '',
            boxShadow: selectedTemplate?.id === t.id ? '0 0 0 2px rgba(0,122,255,0.2)' : '',
          }"
          @click="selectTemplate(t)"
        >
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <div style="font-weight: 600;">{{ t.name }}</div>
              <div style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-top: 2px;">{{ t.filePath || '未上传文件' }}</div>
              <div v-if="!isTemplateUsable(t)" style="font-size: 0.75rem; color: var(--color-danger); margin-top: 2px;">
                {{ !t.filePath ? '未上传 Word 文件' : !t.validated ? 'Word 校验未通过' : '缺少必填字段映射' }}
              </div>
            </div>
            <span
              :class="['badge', isTemplateUsable(t) ? 'badge-success' : 'badge-danger']"
            >
              {{ isTemplateUsable(t) ? '可用' : '暂不可用' }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="templates.length > 0" style="margin-top: 20px; display: flex; gap: 8px; justify-content: flex-end;">
        <button
          class="btn btn-primary"
          :disabled="!selectedTemplate"
          @click="gotoStep1"
        >
          {{ selectedTemplate ? `使用「${selectedTemplate.name}」 →` : '请先选择模板' }}
        </button>
      </div>
    </div>

    <!-- ======================== Step 1: Select/Create Asset ======================== -->
    <div v-if="step === 1" class="card">
      <h3 style="margin-bottom: 4px;">选择或新建资产</h3>
      <p v-if="selectedTemplate" style="font-size: 0.8125rem; color: var(--color-text-secondary); margin-bottom: 16px;">
        模板：{{ selectedTemplate.name }}
      </p>

      <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>

      <!-- Search existing assets -->
      <div class="form-group">
        <label class="label">搜索已有资产</label>
        <input class="input" v-model="assetSearch" @input="searchAssets" placeholder="输入资产名称搜索..." />
      </div>

      <div v-if="assets.length" style="margin-bottom: 12px;">
        <div
          v-for="a in assets"
          :key="a.id"
          class="card"
          style="padding: 12px; margin-bottom: 8px; cursor: pointer;"
          :style="{ borderColor: selectedAsset?.id === a.id ? 'var(--color-primary)' : '' }"
          @click="selectAsset(a)"
        >
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <div style="font-weight: 500;">{{ a.name }}</div>
              <div style="font-size: 0.75rem; color: var(--color-text-secondary);">{{ a.assetType }} · {{ a.status }}</div>
            </div>
            <span v-if="selectedAsset?.id === a.id" style="color: var(--color-primary); font-size: 0.875rem;">已选</span>
          </div>
        </div>
      </div>

      <!-- New asset inline form -->
      <div style="border-top: 1px solid var(--color-border); padding-top: 16px; margin-top: 8px;">
        <p style="font-weight: 500; margin-bottom: 12px; font-size: 0.875rem; color: var(--color-text-secondary);">
          或者，直接创建新资产
        </p>
        <div class="form-group">
          <label class="label">资产名称 <span style="color: var(--color-danger);">*</span></label>
          <input
            class="input"
            v-model="newAssetName"
            placeholder="如：人民路128号商铺"
            @focus="showNewAssetForm = true"
          />
        </div>

        <div v-if="showNewAssetForm">
          <div class="form-group">
            <label class="label">资产类型</label>
            <select class="input" v-model="newAssetType">
              <option value="shop">商铺</option>
              <option value="parking">车位</option>
              <option value="stall">摊位</option>
              <option value="equipment">设备</option>
              <option value="other">其他</option>
            </select>
          </div>
          <div class="form-group">
            <label class="label">月租金参考（可选）</label>
            <input class="input" type="number" v-model="newAssetMonthlyRent" placeholder="仅供参考，后续可改" />
          </div>
          <p style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-bottom: 12px;">
            资产描述、图片等详细信息可在合同创建后在「资产管理」中补全。
          </p>
        </div>
      </div>

      <div style="margin-top: 20px; display: flex; gap: 8px;">
        <button class="btn btn-secondary" @click="gotoStep0">← 返回选模板</button>
        <button
          class="btn btn-primary"
          :disabled="!selectedAsset && !newAssetName.trim()"
          @click="gotoStep2"
        >
          下一步：录租户 →
        </button>
      </div>
    </div>

    <!-- ======================== Step 2: Select/Create Tenant ======================== -->
    <div v-if="step === 2" class="card">
      <h3 style="margin-bottom: 4px;">选择或新建租户</h3>
      <div v-if="selectedAsset" style="margin-bottom: 8px; font-size: 0.8125rem; color: var(--color-text-secondary);">
        已选资产：{{ selectedAsset.name }}
      </div>
      <div v-else-if="newAssetName" style="margin-bottom: 8px; font-size: 0.8125rem; color: var(--color-text-secondary);">
        新资产：{{ newAssetName }}（{{ newAssetType }}）
      </div>

      <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>

      <!-- Search existing tenants -->
      <div class="form-group">
        <label class="label">搜索已有租户</label>
        <input class="input" v-model="tenantSearch" @input="searchTenants" placeholder="输入姓名或电话搜索..." />
      </div>

      <div v-if="tenants.length" style="margin-bottom: 12px;">
        <div
          v-for="t in tenants"
          :key="t.id"
          class="card"
          style="padding: 12px; margin-bottom: 8px; cursor: pointer;"
          :style="{ borderColor: selectedTenant?.id === t.id ? 'var(--color-primary)' : '' }"
          @click="selectTenant(t)"
        >
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <div style="font-weight: 500;">{{ t.name }}</div>
              <div style="font-size: 0.75rem; color: var(--color-text-secondary);">{{ t.phone || '无电话' }} · {{ t.idCard || '无身份证号' }}</div>
            </div>
            <span v-if="selectedTenant?.id === t.id" style="color: var(--color-primary); font-size: 0.875rem;">已选</span>
          </div>
        </div>
      </div>

      <!-- New tenant inline form -->
      <div style="border-top: 1px solid var(--color-border); padding-top: 16px; margin-top: 8px;">
        <p style="font-weight: 500; margin-bottom: 12px; font-size: 0.875rem; color: var(--color-text-secondary);">
          或者，直接创建新租户
        </p>
        <div class="form-group">
          <label class="label">租户姓名 <span style="color: var(--color-danger);">*</span></label>
          <input
            class="input"
            v-model="newTenantName"
            placeholder="如：张三"
            @focus="showNewTenantForm = true"
          />
        </div>

        <div v-if="showNewTenantForm">
          <div class="form-group">
            <label class="label">身份证号（可选）</label>
            <input class="input" v-model="newTenantIdCard" placeholder="后续可在租户管理中补全" />
          </div>
          <div class="form-group">
            <label class="label">电话（可选）</label>
            <input class="input" v-model="newTenantPhone" placeholder="后续可在租户管理中补全" />
          </div>
          <p style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-bottom: 12px;">
            身份证图片等信息可在合同创建后在「租户管理」中补全。
          </p>
        </div>
      </div>

      <div style="margin-top: 20px; display: flex; gap: 8px;">
        <button class="btn btn-secondary" @click="step = 1">← 返回选资产</button>
        <button
          class="btn btn-primary"
          :disabled="!selectedTenant && !newTenantName.trim()"
          @click="gotoStep3"
        >
          下一步：定合同 →
        </button>
      </div>
    </div>

    <!-- ======================== Step 3: Contract Details ======================== -->
    <div v-if="step === 3" class="card">
      <h3 style="margin-bottom: 8px;">合同详情</h3>
      <div style="margin-bottom: 16px; font-size: 0.8125rem; color: var(--color-text-secondary);">
        <span v-if="selectedAsset">资产：{{ selectedAsset.name }}</span>
        <span v-else>新资产：{{ newAssetName }}</span>
        <span style="margin: 0 6px;">·</span>
        <span v-if="selectedTenant">租户：{{ selectedTenant.name }}</span>
        <span v-else>新租户：{{ newTenantName }}</span>
        <span v-if="selectedTemplate" style="margin-left: 6px;">· 模板：{{ selectedTemplate.name }}</span>
      </div>

      <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>

      <!-- Template-driven fields: iterate activeFieldsList -->
      <template v-if="selectedTemplate && activeFieldsList.length > 0">
        <div v-for="key in activeFieldsList" :key="key" class="form-group">
          <label class="label">
            {{ getFieldLabel(key) }}
            <span v-if="isRequiredField(key)" style="color: var(--color-danger);">*</span>
            <span v-if="classifyField(key) === 'system-auto'" style="font-size: 0.65rem; color: var(--color-text-tertiary);">（自动生成）</span>
            <span v-else-if="classifyField(key) === 'asset-tenant'" style="font-size: 0.65rem; color: var(--color-text-tertiary);">（来自资产/租户）</span>
          </label>

          <!-- System-auto read-only -->
          <template v-if="classifyField(key) === 'system-auto'">
            <input class="input" :value="getSystemAutoValue(key)" disabled
              style="background: var(--color-bg); color: var(--color-text-secondary); cursor: not-allowed;" />
          </template>

          <!-- Asset/tenant read-only -->
          <template v-else-if="classifyField(key) === 'asset-tenant'">
            <input class="input" :value="getAssetTenantValue(key)" disabled
              style="background: var(--color-bg); color: var(--color-text-secondary); cursor: not-allowed;" />
          </template>

          <!-- User-input fields: dates -->
          <template v-else-if="key === 'startDate' || key === 'endDate'">
            <input class="input" type="date" v-model="contractStartDate" v-if="key === 'startDate'" />
            <input class="input" type="date" v-model="contractEndDate" v-else />
          </template>

          <!-- User-input: monthlyRent -->
          <template v-else-if="key === 'monthlyRent'">
            <div style="display: flex; gap: 8px; align-items: center;">
              <input class="input" type="number" v-model="contractMonthlyRent" placeholder="如：5000"
                @input="manualTotal = false" style="flex: 1;" />
              <button v-if="activeFieldsList.includes('yearlyRent')"
                :class="['link-toggle', { active: rentLinked }]" type="button"
                @click="rentLinked = !rentLinked; if (rentLinked && contractMonthlyRent) contractYearlyRent = Math.round(contractMonthlyRent * 12 * 100) / 100"
                :title="rentLinked ? '关联已开启，点击关闭' : '关联已关闭，点击开启'">
                {{ rentLinked ? '🔗' : '🔓' }}
              </button>
            </div>
          </template>

          <!-- User-input: yearlyRent -->
          <template v-else-if="key === 'yearlyRent'">
            <div style="display: flex; gap: 8px; align-items: center;">
              <input class="input" type="number" v-model="contractYearlyRent"
                :placeholder="rentLinked && contractMonthlyRent ? '月租金 × 12' : '请输入年租金'"
                style="flex: 1;"
                @input="onYearlyRentInput" />
              <button v-if="activeFieldsList.includes('monthlyRent')"
                :class="['link-toggle', { active: rentLinked }]" type="button"
                @click="rentLinked = !rentLinked; if (rentLinked && contractMonthlyRent) contractYearlyRent = Math.round(contractMonthlyRent * 12 * 100) / 100"
                :title="rentLinked ? '关联已开启，点击关闭' : '关联已关闭，点击开启'">
                {{ rentLinked ? '🔗' : '🔓' }}
              </button>
            </div>
          </template>

          <!-- User-input: other (totalReceivable, deposit, notes, custom fields) -->
          <template v-else-if="key === 'totalReceivable'">
            <div style="display: flex; gap: 8px; align-items: center;">
              <input class="input" type="number" v-model="contractTotalReceivable"
                :placeholder="contractMonthlyRent && contractStartDate && contractEndDate ? '已自动计算' : '填写日期和租金后自动计算'"
                @focus="manualTotal = true" style="flex: 1;" />
              <span style="font-size: 0.75rem; color: var(--color-text-tertiary); white-space: nowrap;">
                {{ manualTotal ? '手动' : '自动' }}
              </span>
            </div>
          </template>
          <template v-else-if="key === 'deposit'">
            <input class="input" type="number" v-model="contractDeposit" placeholder="如：5000" />
          </template>
          <template v-else-if="key === 'notes'">
            <input class="input" v-model="contractNotes" placeholder="线下约定、特殊条款等" />
          </template>
          <template v-else>
            <input class="input" :placeholder="'请输入 ' + getFieldLabel(key)" />
          </template>

          <p v-if="key === 'totalReceivable'" style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-top: 4px;">
            自动计算：整月数 × 月租金 + 零天 × (月租金/30)。点击可手动微调。
          </p>
          <p v-if="key === 'contractId'" style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-top: 4px;">
            合同编号为系统自动生成的唯一标识，创建后即可在预览中查看
          </p>
        </div>
      </template>

      <!-- Fallback: no template selected, show minimum required fields -->
      <template v-else>
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px;">
          <div class="form-group">
            <label class="label">开始日期 <span style="color: var(--color-danger);">*</span></label>
            <input class="input" type="date" v-model="contractStartDate" />
          </div>
          <div class="form-group">
            <label class="label">结束日期 <span style="color: var(--color-danger);">*</span></label>
            <input class="input" type="date" v-model="contractEndDate" />
          </div>
        </div>
        <div class="form-group">
          <label class="label">月租金 <span style="color: var(--color-danger);">*</span></label>
          <input class="input" type="number" v-model="contractMonthlyRent" placeholder="如：5000" @input="manualTotal = false" />
        </div>
        <div class="form-group">
          <label class="label">应收总额</label>
          <div style="display: flex; gap: 8px; align-items: center;">
            <input class="input" type="number" v-model="contractTotalReceivable"
              :placeholder="contractMonthlyRent && contractStartDate && contractEndDate ? '已自动计算' : '填写日期和租金后自动计算'"
              @focus="manualTotal = true" />
            <span style="font-size: 0.75rem; color: var(--color-text-tertiary); white-space: nowrap;">{{ manualTotal ? '手动' : '自动' }}</span>
          </div>
        </div>
        <div class="form-group">
          <label class="label">押金（可选）</label>
          <input class="input" type="number" v-model="contractDeposit" placeholder="如：5000" />
        </div>
        <div class="form-group">
          <label class="label">备注（可选）</label>
          <input class="input" v-model="contractNotes" placeholder="线下约定、特殊条款等" />
        </div>
      </template>

      <div style="margin-top: 20px; display: flex; gap: 8px;">
        <button class="btn btn-secondary" @click="step = 2">← 返回录租户</button>
        <button class="btn btn-primary" :disabled="saving" @click="createContract">
          {{ saving ? '创建中...' : '创建合同' }}
        </button>
      </div>
    </div>

    <!-- ======================== Step 4: Preview & Export ======================== -->
    <div v-if="step === 4 && createdContract" class="card">
      <h3 style="margin-bottom: 4px;">合同创建成功</h3>
      <p style="font-size: 0.8125rem; color: var(--color-text-secondary); margin-bottom: 16px;">
        合同编号 #{{ createdContract.id }}
      </p>

      <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>

      <table>
        <tbody>
          <tr><td style="color: var(--color-text-secondary); width: 100px;">模板</td><td>{{ selectedTemplate?.name || '无' }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">资产</td><td>{{ selectedAsset?.name || createdContract.asset?.name || '-' }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">租户</td><td>{{ selectedTenant?.name || createdContract.tenant?.name || '-' }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">租期</td><td>{{ createdContract.startDate?.toString().substring(0, 10) }} 至 {{ createdContract.endDate?.toString().substring(0, 10) }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">月租金</td><td>&yen;{{ createdContract.monthlyRent?.toLocaleString() }}</td></tr>
          <tr><td style="color: var(--color-text-secondary);">应收总额</td><td>&yen;{{ createdContract.totalReceivable?.toLocaleString() }}</td></tr>
          <tr v-if="createdContract.deposit"><td style="color: var(--color-text-secondary);">押金</td><td>&yen;{{ createdContract.deposit?.toLocaleString() }}</td></tr>
          <tr v-if="createdContract.notes"><td style="color: var(--color-text-secondary);">备注</td><td>{{ createdContract.notes }}</td></tr>
          <tr>
            <td style="color: var(--color-text-secondary);">状态</td>
            <td>
              <span class="badge" :class="{
                'badge-success': createdContract.status === 'paidup',
                'badge-warning': createdContract.status === 'active',
                'badge-danger': createdContract.status === 'arrears',
                'badge-info': createdContract.status === 'expired',
              }">{{ createdContract.status }}</span>
            </td>
          </tr>
        </tbody>
      </table>

      <div style="margin-top: 24px; display: flex; gap: 8px; flex-wrap: wrap; align-items: center;">
        <button class="btn btn-primary" :disabled="downloading" @click="handleDownload">
          {{ downloading ? '下载中...' : '下载合同' }}
        </button>
        <button class="btn btn-primary" @click="resetAll">签下一份合同</button>
      </div>
      <p style="margin-top: 12px; font-size: 0.8125rem; color: var(--color-text-tertiary);">
        提示：下载的 Word 文件可直接用 Microsoft Office 或 WPS 打开并打印。
      </p>
    </div>
  </div>
</template>
