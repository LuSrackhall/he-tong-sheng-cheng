<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api, { templateApi } from '../api'

interface Template {
  id: number
  name: string
  filePath: string
  fieldMap: string
  activeFields: string
  validated: boolean
  createdAt: string
}

const templates = ref<Template[]>([])

// Create template modal
const showCreate = ref(false)
const newTemplateName = ref('')
const creating = ref(false)

// Delete state
const deleting = ref<Record<number, boolean>>({})

async function deleteTemplate(t: Template) {
  if (!confirm(`确定要删除模板「${t.name}」吗？此操作不可撤销。`)) return
  deleting.value[t.id] = true
  try {
    await templateApi.delete(t.id)
    flash('模板已删除')
    await fetchTemplates()
  } catch (e: any) {
    const status = e.response?.status
    if (status === 409) {
      flashError('该模板已被合同引用，无法删除')
    } else {
      flashError(e.response?.data?.error || '删除失败')
    }
  } finally {
    deleting.value[t.id] = false
  }
}

// Custom field modal state (per template)
const showCustomField = ref<Record<number, boolean>>({})
const customFieldName = ref<Record<number, string>>({})
const customFieldLabel = ref<Record<number, string>>({})
const customFieldError = ref<Record<number, string>>({})

function openCustomFieldModal(templateId: number) {
  customFieldName.value[templateId] = ''
  customFieldLabel.value[templateId] = ''
  customFieldError.value[templateId] = ''
  showCustomField.value[templateId] = true
}

function addCustomField(templateId: number) {
  const name = (customFieldName.value[templateId] || '').trim()
  const label = (customFieldLabel.value[templateId] || '').trim()

  if (!name) {
    customFieldError.value[templateId] = '字段名不能为空'
    return
  }
  if (!label) {
    customFieldError.value[templateId] = '显示标签不能为空'
    return
  }

  const allPresetFields = presetFieldGroups.flatMap(g => g.fields)
  const existingMapKeys = getMapKeys(templateId)
  if (allPresetFields.includes(name) || existingMapKeys.includes(name)) {
    customFieldError.value[templateId] = `字段名 "${name}" 已存在`
    return
  }

  // Write directly to JSON textarea
  insertFieldPlaceholder(templateId, name, label)

  showCustomField.value[templateId] = false
  jsonErrors.value[templateId] = ''
}


// Upload state per template
const uploadProgress = ref<Record<number, number>>({})
const uploading = ref<Record<number, boolean>>({})
const fileInputRefs = ref<Record<number, HTMLInputElement | null>>({})

// Mapping state per template
const mapping = ref<Record<number, string>>({})
const saving = ref<Record<number, boolean>>({})
const jsonErrors = ref<Record<number, string>>({})

// Feedback
const successMsg = ref('')
const errorMsg = ref('')
const uploadErrors = ref<Record<number, { msg: string; missingFields?: string[] }>>({})

const presetFieldGroups = [
  {
    category: '合同类',
    fields: ['contractId', 'startDate', 'endDate', 'monthlyRent', 'totalReceivable', 'totalReceived', 'deposit', 'status', 'notes'],
  },
  {
    category: '资产类',
    fields: ['assetName', 'assetType', 'assetDescription'],
  },
  {
    category: '租户类',
    fields: ['tenantName', 'tenantIDCard', 'tenantPhone'],
  },
  {
    category: '其他',
    fields: ['today'],
  },
]

const presetFieldLabels: Record<string, string> = {
  contractId: '合同编号',
  startDate: '开始日期',
  endDate: '结束日期',
  monthlyRent: '月租金',
  totalReceivable: '应收总额',
  totalReceived: '已收总额',
  deposit: '押金',
  status: '状态',
  notes: '备注',
  assetName: '资产名称',
  assetType: '资产类型',
  assetDescription: '资产描述',
  tenantName: '租户姓名',
  tenantIDCard: '身份证号',
  tenantPhone: '联系电话',
  today: '今日日期',
}

function hasFile(t: Template): boolean {
  return !!t.filePath && t.filePath.trim().length > 0
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
  } catch { return iso }
}

function flash(msg: string) {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 4000)
}

function flashError(msg: string) {
  errorMsg.value = msg
  setTimeout(() => { errorMsg.value = '' }, 5000)
}

async function fetchTemplates() {
  try {
    const { data } = await api.get('/templates')
    templates.value = ((data as any).data || data) as Template[]
    templates.value.forEach((t) => {
      if (!mapping.value[t.id]) {
        mapping.value[t.id] = t.fieldMap || '{}'
      }
    })
  } catch { /* handled by interceptor */ }
}


// All keys found in JSON text (including commented lines)
function getAllKeysInJson(templateId: number): string[] {
  const raw = mapping.value[templateId] || ''
  const keyRegex = /"([^"]+)"\s*:/g
  const keys: string[] = []
  let match
  while ((match = keyRegex.exec(raw)) !== null) {
    if (!keys.includes(match[1])) keys.push(match[1])
  }
  return keys
}

// Computed: all keys currently in the fieldMap JSON (including commented lines)
function getMapKeys(templateId: number): string[] {
  return getAllKeysInJson(templateId)
}

// Parse uncommented keys from JSON text (single source of truth for "active")
function parseUncommentedKeys(raw: string): string[] {
  if (!raw || !raw.trim()) return []
  const uncommentedLines = raw.split('\n').filter(line => !line.trim().startsWith('//'))
  const cleanedJson = uncommentedLines.join('\n')
  try {
    const obj = JSON.parse(cleanedJson)
    return Object.keys(obj)
  } catch {
    return []
  }
}

// Parse JSON (skip comments) to key→label Record
function parseJsonLabels(templateId: number): Record<string, string> {
  const raw = mapping.value[templateId] || ''
  if (!raw.trim()) return {}
  const uncommentedLines = raw.split('\n').filter(line => !line.trim().startsWith('//'))
  const cleanedJson = uncommentedLines.join('\n')
  try {
    const obj = JSON.parse(cleanedJson)
    const result: Record<string, string> = {}
    for (const [key, value] of Object.entries(obj)) {
      if (typeof value === 'string') result[key] = value
    }
    return result
  } catch { return {} }
}

// All fields (preset + custom) for chip rendering
function allFieldChips(templateId: number): Array<{ name: string; label: string; isPreset: boolean }> {
  const jsonLabels = parseJsonLabels(templateId)
  const presetKeys = new Set(presetFieldGroups.flatMap(g => g.fields))
  const preset = presetFieldGroups.flatMap(g => g.fields.map(f => ({
    name: f,
    label: jsonLabels[f] || presetFieldLabels[f] || f,
    isPreset: true,
  })))
  const custom: Array<{ name: string; label: string; isPreset: boolean }> = []
  for (const [key, value] of Object.entries(jsonLabels)) {
    if (!presetKeys.has(key)) {
      custom.push({ name: key, label: value || key, isPreset: false })
    }
  }
  return [...preset, ...custom]
}

function isActive(templateId: number, key: string): boolean {
  const raw = mapping.value[templateId] || ''
  for (const line of raw.split('\n')) {
    if (line.includes(`"${key}"`)) {
      return !line.trim().startsWith('//')
    }
  }
  return false
}

function toggleActive(templateId: number, key: string) {
  rebuildJson(templateId, (_obj) => {}, (commented) => {
    if (commented.has(key)) {
      commented.delete(key)
    } else {
      commented.add(key)
    }
  })
}

// Template creation
async function createTemplate() {
  if (!newTemplateName.value.trim() || creating.value) return
  creating.value = true
  try {
    await templateApi.create(newTemplateName.value.trim())
    showCreate.value = false
    newTemplateName.value = ''
    flash('模板创建成功，请上传 Word 文件并配置字段映射')
    await fetchTemplates()
  } catch (e: any) {
    flashError(e.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

function openCreateModal() {
  newTemplateName.value = ''
  showCreate.value = true
}

// File upload / replace for a specific template
function handleFileChange(templateId: number, e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    uploadFile(templateId, input.files[0])
  }
}

async function uploadFile(templateId: number, file: File) {
  uploading.value[templateId] = true
  uploadProgress.value[templateId] = 0
  uploadErrors.value[templateId] = { msg: '' }

  try {
    await templateApi.uploadFile(templateId, file, (pct) => {
      uploadProgress.value[templateId] = pct
    })
    flash(hasFile(templates.value.find((t) => t.id === templateId)!) ? '文件替换成功' : '文件上传成功')
    await fetchTemplates()
  } catch (e: any) {
    const errData = e.response?.data
    uploadErrors.value[templateId] = {
      msg: errData?.error || '上传失败',
      missingFields: errData?.missingFields || [],
    }
  } finally {
    uploading.value[templateId] = false
    uploadProgress.value[templateId] = 0
    // Clear file input
    const input = fileInputRefs.value[templateId]
    if (input) input.value = ''
  }
}

// Rebuild JSON from current text, applying object and comment modifications
function rebuildJson(
  templateId: number,
  modifyObj: (obj: Record<string, string>) => void,
  modifyComments?: (commented: Set<string>) => void,
) {
  const raw = mapping.value[templateId] || ''

  const commentedKeys = new Set<string>()
  for (const line of raw.split('\n')) {
    const trimmed = line.trim()
    if (trimmed.startsWith('//')) {
      const match = trimmed.match(/"([^"]+)"/)
      if (match) commentedKeys.add(match[1])
    }
  }

  const uncommentedLines = raw.split('\n').filter(l => !l.trim().startsWith('//'))
  const cleanedJson = uncommentedLines.join('\n')
  let obj: Record<string, string> = {}
  try {
    if (cleanedJson.trim()) obj = JSON.parse(cleanedJson)
  } catch { /* keep empty */ }

  modifyObj(obj)
  if (modifyComments) modifyComments(commentedKeys)

  const formatted = JSON.stringify(obj, null, 2)
  if (commentedKeys.size === 0) {
    mapping.value[templateId] = formatted
  } else {
    const fmtLines = formatted.split('\n')
    const result = fmtLines.map(line => {
      for (const ck of commentedKeys) {
        if (line.includes(`"${ck}"`)) {
          return line.replace(/^(\s*)/, '$1// ')
        }
      }
      return line
    })
    mapping.value[templateId] = result.join('\n')
  }
}

// Field mapping
function insertFieldPlaceholder(templateId: number, fieldName: string, label?: string) {
  const displayLabel = label || fieldName
  rebuildJson(templateId, (obj) => {
    if (fieldName in obj) {
      delete obj[fieldName]
    } else {
      obj[fieldName] = displayLabel
    }
  })
  jsonErrors.value[templateId] = ''
}

function validateJson(templateId: number): boolean {
  const raw = mapping.value[templateId] || ''
  if (!raw.trim()) { jsonErrors.value[templateId] = ''; return true }
  // Strip comment lines for validation
  const uncommented = raw.split('\n').filter(line => !line.trim().startsWith('//')).join('\n')
  if (!uncommented.trim()) { jsonErrors.value[templateId] = ''; return true }
  try {
    const parsed = JSON.parse(uncommented)
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      jsonErrors.value[templateId] = '映射必须是 JSON 对象'
      return false
    }
    jsonErrors.value[templateId] = ''
    return true
  } catch (e: any) {
    jsonErrors.value[templateId] = 'JSON 格式错误: ' + e.message
    return false
  }
}

function formatJson(templateId: number) {
  const raw = mapping.value[templateId] || ''
  if (!raw.trim()) return
  const lines = raw.split('\n')
  const uncommented = lines.filter(line => !line.trim().startsWith('//')).join('\n')
  if (!uncommented.trim()) return
  try {
    const parsed = JSON.parse(uncommented)
    const formatted = JSON.stringify(parsed, null, 2)
    // Re-apply comments: find lines that were commented and comment their equivalents
    const commentedKeys = new Set<string>()
    for (const line of lines) {
      const trimmed = line.trim()
      if (trimmed.startsWith('//')) {
        const keyMatch = trimmed.match(/"([^"]+)"/)
        if (keyMatch) commentedKeys.add(keyMatch[1])
      }
    }
    if (commentedKeys.size === 0) {
      mapping.value[templateId] = formatted
    } else {
      const fmtLines = formatted.split('\n')
      const result = fmtLines.map(line => {
        for (const key of commentedKeys) {
          if (line.includes(`"${key}"`)) {
            return line.replace(/^(\s*)/, '$1// ')
          }
        }
        return line
      })
      mapping.value[templateId] = result.join('\n')
    }
    jsonErrors.value[templateId] = ''
  } catch (e: any) {
    jsonErrors.value[templateId] = 'JSON 格式错误: ' + e.message
  }
}

async function saveMapping(t: Template) {
  if (!validateJson(t.id)) return

  saving.value[t.id] = true
  try {
    const fieldMap = mapping.value[t.id] || '{}'
    // Derive active fields from uncommented keys in the JSON textarea
    const activeArr = parseUncommentedKeys(fieldMap)
    const activeFields = JSON.stringify(activeArr)
    const res = await templateApi.updateMapping(t.id, fieldMap, activeFields)
    if (res.data.filePath) {
      if (res.data.validated) {
        flash('映射已保存，Word 文件校验通过')
      } else {
        flashError('映射已保存，但 Word 文件校验未通过，请重新上传符合要求的 Word 文件')
      }
    } else {
      flash('字段映射已保存')
    }
    await fetchTemplates()
  } catch (e: any) {
    jsonErrors.value[t.id] = e.response?.data?.error || '保存失败'
  } finally {
    saving.value[t.id] = false
  }
}

function dismissUploadError(templateId: number) {
  uploadErrors.value[templateId] = { msg: '' }
}

onMounted(fetchTemplates)
</script>

<template>
  <div>
    <!-- Success toast -->
    <div v-if="successMsg" class="success-toast">{{ successMsg }}</div>
    <div v-if="errorMsg" class="error-toast">{{ errorMsg }}</div>

    <!-- Page header -->
    <div class="page-header">
      <h2>合同模板管理</h2>
      <button class="btn btn-primary" @click="openCreateModal">+ 创建模板</button>
    </div>

    <!-- Empty state -->
    <div v-if="templates.length === 0" class="guide-empty">
      <div class="guide-icon">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <line x1="12" y1="18" x2="12" y2="12"/>
          <line x1="9" y1="15" x2="15" y2="15"/>
        </svg>
      </div>
      <p class="guide-text">
        还没有合同模板。请先创建一个模板记录，然后上传 Word 文件并配置字段映射。
        Word 中请使用 <code>$&#123;字段名&#125;</code> 格式插入占位符。
      </p>
      <button class="btn btn-primary" @click="openCreateModal">创建第一个模板</button>
    </div>

    <!-- Template cards -->
    <div v-else class="template-list">
      <div v-for="t in templates" :key="t.id" class="card template-card">
        <!-- Card header -->
        <div class="card-header">
          <div class="card-title-row">
            <span class="template-name">{{ t.name }}</span>
            <span :class="['file-badge', hasFile(t) ? 'file-ok' : 'file-missing']">
              {{ hasFile(t) ? '文件已上传' : '未上传文件' }}
            </span>
            <span v-if="t.filePath" :class="['badge-validated', t.validated ? 'badge-ok' : 'badge-fail']">
              {{ t.validated ? '校验通过' : '校验未通过' }}
            </span>
            <button
              class="btn-delete-template"
              :disabled="deleting[t.id]"
              @click="deleteTemplate(t)"
              :title="deleting[t.id] ? '删除中...' : '删除模板'"
            >
              {{ deleting[t.id] ? '...' : '✕' }}
            </button>
          </div>
          <div class="card-meta">
            <span class="meta-date">创建于 {{ formatDate(t.createdAt) }}</span>
            <span v-if="hasFile(t)" class="meta-path">{{ t.filePath }}</span>
          </div>
        </div>

        <!-- File upload area -->
        <div class="section">
          <div class="section-header">
            <span class="section-title">Word 模板文件</span>
            <span class="section-hint">
              {{ hasFile(t) ? '点击下方按钮可替换已有文件' : '请上传配置好占位符的 .docx 文件' }}
            </span>
          </div>

          <div class="file-row">
            <label :for="'file-' + t.id" class="file-select-btn" :class="{ disabled: uploading[t.id] }">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
              {{ hasFile(t) ? '替换文件' : '上传文件' }}
            </label>
            <input
              :id="'file-' + t.id"
              :ref="(el: any) => fileInputRefs[t.id] = el"
              type="file"
              accept=".docx"
              class="file-input-hidden"
              @change="handleFileChange(t.id, $event)"
              :disabled="uploading[t.id]"
            />
            <span v-if="uploading[t.id]" class="upload-status">上传中 {{ uploadProgress[t.id] || 0 }}%</span>
            <span v-else-if="hasFile(t)" class="upload-status ok">已就绪</span>
            <span v-else class="upload-status warn">待上传</span>
          </div>

          <!-- Upload progress bar -->
          <div v-if="uploading[t.id]" class="progress-bar-wrap">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: (uploadProgress[t.id] || 0) + '%' }"></div>
            </div>
          </div>

          <!-- Upload error -->
          <div v-if="uploadErrors[t.id]?.msg" class="upload-error">
            <div class="upload-error-msg">
              <span>{{ uploadErrors[t.id].msg }}</span>
              <button class="dismiss-btn" @click="dismissUploadError(t.id)">&times;</button>
            </div>
            <ul v-if="uploadErrors[t.id].missingFields?.length" class="missing-fields">
              <li v-for="f in uploadErrors[t.id].missingFields" :key="f">
                <code>$&#123;{{ f }}&#125;</code>
              </li>
            </ul>
          </div>
        </div>

        <!-- Field mapping section -->
        <div class="section">
          <div class="section-header">
            <span class="section-title">字段映射配置</span>
            <span class="section-hint">
              点击标签添加映射并自动启用，关闭开关则禁用该字段（不参与替换和校验）
            </span>
          </div>

          <!-- Field chips with active toggles -->
          <div class="preset-fields">
            <template v-for="chip in allFieldChips(t.id)" :key="chip.name">
              <div
                :class="['field-chip', { 'field-in-map': getMapKeys(t.id).includes(chip.name), 'field-active': isActive(t.id, chip.name) }]"
              >
                <button
                  class="chip-label"
                  type="button"
                  @click="insertFieldPlaceholder(t.id, chip.name, chip.label)"
                  :title="getMapKeys(t.id).includes(chip.name) ? '点击从映射中移除' : '点击添加到映射'">
                  ${{ '{' + chip.name + '}' }}
                  <span class="chip-label-text">→ {{ chip.label }}</span>
                </button>
                <button
                  v-if="getMapKeys(t.id).includes(chip.name)"
                  :class="['chip-toggle', { on: isActive(t.id, chip.name) }]"
                  type="button"
                  @click="toggleActive(t.id, chip.name)"
                  :title="isActive(t.id, chip.name) ? '已启用，点击禁用' : '未启用，点击启用'">
                  {{ isActive(t.id, chip.name) ? '✓' : '○' }}
                </button>
              </div>
            </template>
          </div>

          <button
            class="btn-add-custom"
            type="button"
            @click="openCustomFieldModal(t.id)"
          >
            + 添加自定义字段
          </button>

          <!-- JSON editor -->
          <textarea
            class="input mapping-textarea"
            :value="mapping[t.id] || ''"
            @input="mapping[t.id] = ($event.target as HTMLTextAreaElement).value"
            @blur="validateJson(t.id)"
            rows="5"
            spellcheck="false"
            placeholder='{"assetName": "资产名称", "tenantName": "租户姓名"}'
          />

          <!-- Active fields summary -->
          <div v-if="parseUncommentedKeys(mapping[t.id] || '').length" class="active-summary">
            <span class="active-label">已启用字段（{{ parseUncommentedKeys(mapping[t.id] || '').length }}）：</span>
            <code v-for="f in parseUncommentedKeys(mapping[t.id] || '')" :key="f" class="active-tag">${{ '{' + f + '}' }}</code>
          </div>
          <div v-else class="active-summary empty">
            暂未启用任何映射字段，请先在映射中添加字段并开启开关
          </div>

          <!-- JSON error -->
          <div v-if="jsonErrors[t.id]" class="json-error">{{ jsonErrors[t.id] }}</div>

          <!-- Actions -->
          <div class="section-actions">
            <button class="btn btn-secondary btn-sm" type="button" @click="formatJson(t.id)" :disabled="!mapping[t.id]?.trim()">
              格式化
            </button>
            <button class="btn btn-primary btn-sm" type="button" @click="saveMapping(t)" :disabled="saving[t.id]">
              {{ saving[t.id] ? '保存中...' : '保存映射' }}
            </button>
          </div>
        </div>

        <!-- Guide for Word placeholders -->
        <details class="guide-details">
          <summary>Word 模板占位符使用说明</summary>
          <div class="guide-content">
            <p>在 Word 文档中需要动态填充的位置使用 <code>$&#123;字段名&#125;</code> 格式的占位符。</p>
            <p>例如：<code>$&#123;tenantName&#125;</code> 会在生成合同时替换为租户姓名。</p>
            <p>上传文件时，系统会校验<strong>所有已启用的</strong>字段是否都在 Word 中存在对应占位符。关闭开关的字段不会被校验。</p>
          </div>
        </details>

        <!-- Custom field modal -->
        <div v-if="showCustomField[t.id]" class="modal-overlay" @click.self="showCustomField[t.id] = false">
          <div class="modal-content" style="max-width: 380px;">
            <h3>添加自定义字段</h3>
            <div class="form-group">
              <label class="label">字段名（占位符）</label>
              <input
                class="input"
                v-model="customFieldName[t.id]"
                placeholder="例如：customField"
                @keyup.enter="addCustomField(t.id)"
              />
            </div>
            <div class="form-group">
              <label class="label">显示标签</label>
              <input
                class="input"
                v-model="customFieldLabel[t.id]"
                placeholder="例如：自定义字段"
                @keyup.enter="addCustomField(t.id)"
              />
            </div>
            <div v-if="customFieldError[t.id]" class="json-error" style="margin-bottom: 8px;">{{ customFieldError[t.id] }}</div>
            <div class="modal-actions">
              <button class="btn btn-secondary" @click="showCustomField[t.id] = false">取消</button>
              <button class="btn btn-primary" @click="addCustomField(t.id)">添加</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create template modal -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal-content" style="max-width: 420px;">
        <h3>创建模板</h3>
        <p class="modal-desc">先创建模板记录，稍后上传 Word 文件并配置字段映射。</p>
        <div class="form-group">
          <label class="label">模板名称</label>
          <input
            class="input"
            v-model="newTemplateName"
            placeholder="例如：标准商铺租赁合同"
            @keyup.enter="createTemplate"
            :disabled="creating"
          />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showCreate = false" :disabled="creating">取消</button>
          <button class="btn btn-primary" @click="createTemplate" :disabled="creating || !newTemplateName.trim()">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Toasts */
.success-toast {
  position: fixed; top: 20px; left: 50%; transform: translateX(-50%);
  background: var(--color-success); color: #fff; padding: 10px 24px;
  border-radius: var(--radius-md); font-size: 0.875rem; font-weight: 500;
  z-index: 2000; box-shadow: var(--shadow-lg);
  animation: toast-in 0.3s ease;
}
.error-toast {
  position: fixed; top: 60px; left: 50%; transform: translateX(-50%);
  background: var(--color-danger); color: #fff; padding: 10px 24px;
  border-radius: var(--radius-md); font-size: 0.875rem; font-weight: 500;
  z-index: 2000; box-shadow: var(--shadow-lg);
  animation: toast-in 0.3s ease;
}
@keyframes toast-in {
  from { opacity: 0; transform: translateX(-50%) translateY(-12px); }
  to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

/* Guide empty */
.guide-empty { text-align: center; padding: var(--space-2xl); }
.guide-icon { color: var(--color-text-tertiary); margin-bottom: var(--space-md); }
.guide-text { color: var(--color-text-secondary); font-size: 0.9375rem; line-height: 1.7; max-width: 480px; margin: 0 auto var(--space-lg); }
.guide-text code { background: rgba(0,0,0,0.06); padding: 2px 6px; border-radius: 4px; font-family: var(--font-mono); font-size: 0.875rem; }

/* Template list */
.template-list { display: grid; gap: 16px; }

.template-card { padding: 20px; }

.card-header { margin-bottom: 16px; }
.card-title-row { display: flex; align-items: center; gap: 12px; margin-bottom: 6px; }
.template-name { font-weight: 600; font-size: 1.0625rem; }
.file-badge { font-size: 0.75rem; padding: 2px 8px; border-radius: 10px; font-weight: 500; }
.file-ok { background: rgba(52,199,89,0.1); color: var(--color-success); }
.file-missing { background: rgba(255,149,0,0.1); color: #b87a00; }
.badge-validated { font-size: 0.75rem; padding: 2px 8px; border-radius: 10px; font-weight: 500; margin-left: 4px; }
.badge-ok { background: rgba(52,199,89,0.1); color: var(--color-success); }
.badge-fail { background: rgba(255,59,48,0.1); color: var(--color-danger); }
.card-meta { display: flex; gap: 16px; font-size: 0.75rem; color: var(--color-text-tertiary); }
.meta-path { font-family: var(--font-mono); word-break: break-all; }

/* Sections */
.section { border-top: 1px solid var(--color-border); padding-top: 14px; margin-top: 14px; }
.section-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 10px; }
.section-title { font-size: 0.8125rem; font-weight: 600; color: var(--color-text-secondary); }
.section-hint { font-size: 0.75rem; color: var(--color-text-tertiary); }

/* File area */
.file-row { display: flex; align-items: center; gap: 10px; }
.file-select-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 14px; border: 1px solid var(--color-primary); border-radius: var(--radius-sm);
  color: var(--color-primary); font-size: 0.8125rem; cursor: pointer;
  transition: all var(--transition-fast);
}
.file-select-btn:hover { background: rgba(0,122,255,0.06); }
.file-select-btn.disabled { opacity: 0.5; pointer-events: none; }
.file-input-hidden { position: absolute; width: 0; height: 0; opacity: 0; overflow: hidden; }
.upload-status { font-size: 0.8125rem; }
.upload-status.ok { color: var(--color-success); }
.upload-status.warn { color: #b87a00; }

.progress-bar-wrap { margin-top: 8px; }
.progress-bar { height: 4px; background: var(--color-border); border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: var(--color-primary); border-radius: 2px; transition: width 0.3s ease; }

/* Upload error */
.upload-error { margin-top: 8px; padding: 10px 12px; background: rgba(255,59,48,0.06); border: 1px solid rgba(255,59,48,0.2); border-radius: var(--radius-sm); }
.upload-error-msg { display: flex; justify-content: space-between; align-items: center; color: var(--color-danger); font-size: 0.8125rem; }
.dismiss-btn { background: none; border: none; color: var(--color-text-tertiary); font-size: 1.1rem; cursor: pointer; padding: 0 4px; }
.missing-fields { margin: 6px 0 0; padding-left: 16px; font-size: 0.8125rem; }
.missing-fields code { font-family: var(--font-mono); font-size: 0.75rem; background: rgba(0,0,0,0.06); padding: 1px 5px; border-radius: 3px; }

/* Preset fields */
.preset-fields { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin-bottom: 12px; }
.preset-category { font-size: 0.6875rem; color: var(--color-text-tertiary); font-weight: 500; margin-right: 2px; }

.field-chip { display: inline-flex; align-items: center; border: 1px solid var(--color-border); border-radius: 14px; overflow: hidden; }
.field-chip.field-in-map { border-color: var(--color-primary); }
.field-chip.field-active { background: rgba(0,122,255,0.06); }

.chip-label {
  padding: 3px 8px; background: none; border: none;
  font-family: var(--font-mono); font-size: 0.75rem; color: var(--color-primary);
  cursor: pointer; line-height: 1.4;
}
.chip-label:hover { background: rgba(0,122,255,0.08); }

.chip-toggle {
  width: 22px; height: 22px; display: inline-flex; align-items: center; justify-content: center;
  padding: 0; border: none; border-left: 1px solid var(--color-border);
  background: transparent; font-size: 0.7rem; cursor: pointer; color: var(--color-text-tertiary);
  transition: all var(--transition-fast);
}
.chip-toggle.on { background: var(--color-primary); color: #fff; border-left-color: var(--color-primary); }

/* Mapping textarea */
.mapping-textarea { font-family: var(--font-mono); font-size: 0.8125rem; line-height: 1.6; resize: vertical; min-height: 80px; }

/* Active summary */
.active-summary { margin-top: 8px; font-size: 0.75rem; color: var(--color-text-secondary); display: flex; flex-wrap: wrap; align-items: center; gap: 4px; }
.active-summary.empty { color: var(--color-text-tertiary); font-style: italic; }
.active-label { margin-right: 4px; }
.active-tag { font-family: var(--font-mono); font-size: 0.7rem; background: rgba(0,122,255,0.08); padding: 1px 6px; border-radius: 8px; color: var(--color-primary); }

/* JSON error */
.json-error { color: var(--color-danger); font-size: 0.75rem; margin-top: 6px; }

/* Section actions */
.section-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 10px; }

/* Guide details */
.guide-details { margin-top: 14px; font-size: 0.8125rem; color: var(--color-text-secondary); }
.guide-details summary { cursor: pointer; color: var(--color-primary); font-weight: 500; }
.guide-content { margin-top: 6px; padding: 10px 12px; background: var(--color-bg); border-radius: var(--radius-sm); line-height: 1.7; }
.guide-content code { background: rgba(0,0,0,0.06); padding: 1px 5px; border-radius: 3px; font-family: var(--font-mono); font-size: 0.75rem; }

/* Modal */
.modal-desc { color: var(--color-text-secondary); font-size: 0.875rem; margin-bottom: 14px; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 16px; }

/* Delete button */
.btn-delete-template {
  margin-left: auto;
  width: 26px; height: 26px;
  display: inline-flex; align-items: center; justify-content: center;
  padding: 0; border: 1px solid transparent; border-radius: 50%;
  background: transparent; font-size: 0.75rem; color: var(--color-text-tertiary);
  cursor: pointer; transition: all var(--transition-fast);
}
.btn-delete-template:hover {
  border-color: var(--color-danger);
  color: var(--color-danger);
  background: rgba(255,59,48,0.06);
}
.btn-delete-template:disabled { opacity: 0.5; cursor: not-allowed; }

/* Custom field */
.btn-add-custom {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; margin-bottom: 12px;
  border: 1px dashed var(--color-border); border-radius: var(--radius-sm);
  background: transparent; color: var(--color-text-secondary);
  font-size: 0.75rem; cursor: pointer;
  transition: all var(--transition-fast);
}
.btn-add-custom:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: rgba(0,122,255,0.04);
}

/* Chip label text */
.chip-label-text {
  font-family: inherit;
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  margin-left: 2px;
}
</style>
