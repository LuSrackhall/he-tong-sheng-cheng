<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import api from '../api'

interface Template {
  id: number
  name: string
  filePath: string
  fieldMap: string
  createdAt: string
}

const templates = ref<Template[]>([])
const showUpload = ref(false)
const form = ref({ name: '' })
const selectedFile = ref<File | null>(null)
const uploadProgress = ref(0)
const isUploading = ref(false)
const mapping = ref<Record<number, string>>({})
const saving = ref<Record<number, boolean>>({})
const submitLock = ref(false)
const error = ref('')
const jsonErrors = ref<Record<number, string>>({})
const mappingSavedTag = ref<Record<number, boolean>>({})
const uploadSuccessMsg = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const presetFieldGroups = [
  {
    category: '合同类',
    fields: ['contractId', 'startDate', 'endDate', 'monthlyRent', 'totalReceivable', 'deposit'],
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

function hasFile(t: Template): boolean {
  return !!t.filePath && t.filePath.trim().length > 0
}

function hasFieldMapConfigured(t: Template): boolean {
  if (!t.fieldMap || !t.fieldMap.trim()) return false
  try {
    const parsed = JSON.parse(t.fieldMap)
    return typeof parsed === 'object' && parsed !== null && Object.keys(parsed).length > 0
  } catch {
    return false
  }
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  } catch {
    return iso
  }
}

async function fetchTemplates() {
  try {
    const { data } = await api.get('/templates')
    templates.value = (data as any).data || data
    templates.value.forEach((t) => {
      if (t.fieldMap) {
        try {
          mapping.value[t.id] = t.fieldMap
        } catch {
          mapping.value[t.id] = ''
        }
      } else {
        mapping.value[t.id] = ''
      }
      mappingSavedTag.value[t.id] = hasFieldMapConfigured(t)
    })
  } catch {
    // handled by interceptor
  }
}

function handleFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFile.value = input.files[0]
    error.value = ''
  }
}

function removeFile() {
  selectedFile.value = null
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

async function createAndUploadTemplate() {
  if (submitLock.value) return
  error.value = ''

  if (!form.value.name.trim()) {
    error.value = '请输入模板名称'
    return
  }
  if (!selectedFile.value) {
    error.value = '请选择 .docx 模板文件'
    return
  }

  submitLock.value = true
  isUploading.value = true
  uploadProgress.value = 0

  try {
    // Step 1: Create template record
    const { data: createRes } = await api.post('/templates', { name: form.value.name.trim() })
    const templateId: number = (createRes as any).data?.id ?? (createRes as any).id

    if (!templateId) {
      throw new Error('创建模板记录失败，未获取到模板 ID')
    }

    // Step 2: Upload the .docx file
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    await api.post(`/templates/${templateId}/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        }
      },
    })

    // Success
    showUpload.value = false
    uploadSuccessMsg.value = '模板上传成功！请在签合同时选择模板。'
    form.value = { name: '' }
    selectedFile.value = null
    uploadProgress.value = 0
    error.value = ''
    await fetchTemplates()

    // Clear success message after 5 seconds
    setTimeout(() => {
      uploadSuccessMsg.value = ''
    }, 5000)
  } catch (e: any) {
    error.value = e.response?.data?.error || e.message || '上传失败，请重试'
  } finally {
    isUploading.value = false
    submitLock.value = false
  }
}

function openUploadModal() {
  form.value = { name: '' }
  selectedFile.value = null
  uploadProgress.value = 0
  error.value = ''
  isUploading.value = false
  showUpload.value = true
  nextTick(() => {
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  })
}

function closeUploadModal() {
  if (isUploading.value) return
  showUpload.value = false
  form.value = { name: '' }
  selectedFile.value = null
  uploadProgress.value = 0
  error.value = ''
}

function insertFieldPlaceholder(templateId: number, fieldName: string) {
  const current = mapping.value[templateId] || ''
  const placeholder = `"${fieldName}": ""`

  if (!current.trim()) {
    mapping.value[templateId] = `{\n  ${placeholder}\n}`
  } else {
    // Try to insert before the last closing brace
    const trimmed = current.trim()
    if (trimmed.endsWith('}')) {
      const inner = trimmed.slice(0, -1).trimEnd()
      mapping.value[templateId] = inner + `,\n  ${placeholder}\n}`
    } else {
      mapping.value[templateId] = current + `,\n  ${placeholder}`
    }
  }
  jsonErrors.value[templateId] = ''
}

function validateJson(templateId: number): boolean {
  const raw = mapping.value[templateId] || ''
  if (!raw.trim()) {
    jsonErrors.value[templateId] = ''
    return true
  }
  try {
    const parsed = JSON.parse(raw)
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      jsonErrors.value[templateId] = '映射必须是一个 JSON 对象，例如 {"assetName": "资产名称"}'
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
  try {
    const parsed = JSON.parse(raw)
    mapping.value[templateId] = JSON.stringify(parsed, null, 2)
    jsonErrors.value[templateId] = ''
  } catch (e: any) {
    jsonErrors.value[templateId] = 'JSON 格式错误: ' + e.message
  }
}

async function saveMapping(t: Template) {
  if (!validateJson(t.id)) return
  const raw = mapping.value[t.id] || ''
  if (!raw.trim()) return

  saving.value[t.id] = true
  try {
    await api.patch(`/templates/${t.id}`, { fieldMap: raw })
    mappingSavedTag.value[t.id] = true
    uploadSuccessMsg.value = '字段映射已保存'
    setTimeout(() => {
      uploadSuccessMsg.value = ''
    }, 3000)
    await fetchTemplates()
  } catch (e: any) {
    jsonErrors.value[t.id] = e.response?.data?.error || '保存失败'
  } finally {
    saving.value[t.id] = false
  }
}

function getMappingStatusText(t: Template): string {
  if (hasFieldMapConfigured(t)) return '已配置'
  const raw = mapping.value[t.id] || ''
  return raw.trim() ? '格式有误' : '未配置'
}

function getMappingStatusClass(t: Template): string {
  if (hasFieldMapConfigured(t)) return 'badge-success'
  const raw = mapping.value[t.id] || ''
  return raw.trim() ? 'badge-danger' : 'badge-warning'
}

onMounted(fetchTemplates)
</script>

<template>
  <div>
    <!-- Success toast -->
    <div v-if="uploadSuccessMsg" class="success-toast">
      {{ uploadSuccessMsg }}
    </div>

    <!-- Page header -->
    <div class="page-header">
      <h2>合同模板管理</h2>
      <button class="btn btn-primary" @click="openUploadModal">+ 上传模板</button>
    </div>

    <!-- Guide: empty state -->
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
        还没有上传合同模板。请先准备好 Word 格式的合同模板文件，使用
        <code>${字段名}</code> 格式插入占位符，然后上传。
      </p>
      <button class="btn btn-primary" @click="openUploadModal">创建第一个模板</button>
    </div>

    <!-- Template cards -->
    <div v-else style="display: grid; gap: 16px;">
      <div v-for="t in templates" :key="t.id" class="card template-card">
        <!-- Template header info -->
        <div class="template-header">
          <div class="template-info">
            <div class="template-name">{{ t.name }}</div>
            <div class="template-meta">
              <span class="meta-item">
                <span class="meta-label">文件</span>
                <span v-if="hasFile(t)" class="status-icon status-ok" title="文件已上传">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                  已上传
                </span>
                <span v-else class="status-icon status-missing" title="尚未上传文件">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                  </svg>
                  未上传
                </span>
              </span>
              <span class="meta-item">
                <span class="meta-label">映射</span>
                <span :class="['badge', getMappingStatusClass(t)]">
                  {{ getMappingStatusText(t) }}
                </span>
              </span>
              <span class="meta-item meta-date">
                <span class="meta-label">创建于</span>
                {{ formatDate(t.createdAt) }}
              </span>
            </div>
            <div v-if="hasFile(t)" class="template-filepath">
              {{ t.filePath }}
            </div>
          </div>
        </div>

        <!-- Field mapping editor -->
        <div class="mapping-section">
          <div class="mapping-header">
            <span class="mapping-title">字段映射配置</span>
            <span class="mapping-hint">
              模板中的 <code>${字段名}</code> 占位符将替换为对应值
            </span>
          </div>

          <!-- Preset field tags -->
          <div class="preset-fields">
            <span class="preset-label">可用字段：</span>
            <template v-for="group in presetFieldGroups" :key="group.category">
              <span class="preset-category">{{ group.category }}</span>
              <button
                v-for="field in group.fields"
                :key="field"
                class="preset-tag"
                type="button"
                @click="insertFieldPlaceholder(t.id, field)"
                :title="`插入 $\{${field}\} 映射`"
              >
                ${{ '{' + field + '}' }}
              </button>
            </template>
          </div>

          <!-- JSON editor -->
          <div class="form-group" style="margin-bottom: 8px;">
            <textarea
              class="input mapping-textarea"
              :value="mapping[t.id] || ''"
              @input="mapping[t.id] = ($event.target as HTMLTextAreaElement).value"
              @blur="validateJson(t.id)"
              rows="5"
              spellcheck="false"
              placeholder='{"assetName": "资产名称", "tenantName": "租户姓名"}'
            />
          </div>

          <!-- JSON error -->
          <div v-if="jsonErrors[t.id]" class="json-error">
            {{ jsonErrors[t.id] }}
          </div>

          <!-- Action buttons -->
          <div class="mapping-actions">
            <button
              class="btn btn-secondary btn-sm"
              type="button"
              @click="formatJson(t.id)"
              :disabled="!mapping[t.id] || !mapping[t.id].trim()"
            >
              格式化
            </button>
            <button
              class="btn btn-primary btn-sm"
              type="button"
              @click="saveMapping(t)"
              :disabled="saving[t.id]"
            >
              {{ saving[t.id] ? '保存中...' : '保存映射' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Upload modal -->
    <div v-if="showUpload" class="modal-overlay" @click.self="closeUploadModal">
      <div class="modal-content" style="max-width: 540px;">
        <h3>上传合同模板</h3>

        <!-- Error -->
        <div v-if="error" class="alert alert-danger" style="margin-bottom: 12px;">{{ error }}</div>

        <!-- Form -->
        <div class="form-group">
          <label class="label">模板名称</label>
          <input
            class="input"
            v-model="form.name"
            placeholder="例如：标准商铺租赁合同"
            :disabled="isUploading"
          />
        </div>

        <div class="form-group">
          <label class="label">模板文件（.docx）</label>
          <div class="file-upload-area" :class="{ 'has-file': selectedFile }">
            <input
              ref="fileInputRef"
              type="file"
              accept=".docx"
              class="file-input-native"
              @change="handleFileChange"
              :disabled="isUploading"
            />
            <div v-if="!selectedFile" class="file-placeholder">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
              <span>点击选择 .docx 文件</span>
            </div>
            <div v-else class="file-selected">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              <span class="file-name">{{ selectedFile.name }}</span>
              <span class="file-size">({{ (selectedFile.size / 1024).toFixed(1) }} KB)</span>
              <button
                v-if="!isUploading"
                type="button"
                class="file-remove-btn"
                @click="removeFile"
                title="移除文件"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Upload progress -->
        <div v-if="isUploading" class="upload-progress">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
          </div>
          <span class="progress-text">{{ uploadProgress }}%</span>
        </div>

        <!-- Template instructions -->
        <div class="template-instructions">
          <p>
            支持 <strong>.docx</strong> 格式的 Word 模板文件。请在模板中使用
            <code>${字段名}</code> 作为占位符，生成合同时会自动替换为实际数据。
          </p>
          <p style="margin-top: 4px;">
            可用的占位符请参考下方卡片中的"可用字段"列表。
          </p>
        </div>

        <!-- Modal actions -->
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button
            class="btn btn-secondary"
            @click="closeUploadModal"
            :disabled="isUploading"
          >
            取消
          </button>
          <button
            class="btn btn-primary"
            :disabled="isUploading || !form.name.trim() || !selectedFile"
            @click="createAndUploadTemplate"
          >
            {{ isUploading ? '上传中...' : '上传模板' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Success toast */
.success-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-success);
  color: #fff;
  padding: 10px 24px;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  z-index: 2000;
  box-shadow: var(--shadow-lg);
  animation: toast-in 0.3s ease;
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-12px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

/* Guide empty state */
.guide-empty {
  text-align: center;
  padding: var(--space-2xl);
}

.guide-icon {
  color: var(--color-text-tertiary);
  margin-bottom: var(--space-md);
}

.guide-text {
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
  line-height: 1.7;
  max-width: 460px;
  margin: 0 auto var(--space-lg);
}

.guide-text code {
  background: rgba(0, 0, 0, 0.06);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 0.875rem;
}

/* Template card */
.template-card {
  padding: 20px;
}

.template-header {
  margin-bottom: 16px;
}

.template-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.template-name {
  font-weight: 600;
  font-size: 1.0625rem;
}

.template-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  font-size: 0.8125rem;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--color-text-secondary);
}

.meta-label {
  color: var(--color-text-tertiary);
}

.meta-date {
  color: var(--color-text-tertiary);
}

.status-icon {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 0.8125rem;
}

.status-ok {
  color: var(--color-success);
}

.status-missing {
  color: var(--color-danger);
}

.template-filepath {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  font-family: var(--font-mono);
  word-break: break-all;
}

/* Mapping section */
.mapping-section {
  border-top: 1px solid var(--color-border);
  padding-top: 16px;
}

.mapping-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}

.mapping-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.mapping-hint {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.mapping-hint code {
  background: rgba(0, 0, 0, 0.05);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 0.75rem;
}

/* Preset fields */
.preset-fields {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
}

.preset-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-right: 4px;
}

.preset-category {
  font-size: 0.6875rem;
  color: var(--color-text-tertiary);
  font-weight: 500;
  margin-right: -2px;
}

.preset-tag {
  display: inline-block;
  padding: 2px 8px;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  background: var(--color-surface-solid);
  color: var(--color-primary);
  font-size: 0.75rem;
  font-family: var(--font-mono);
  cursor: pointer;
  transition: all var(--transition-fast);
  line-height: 1.5;
}

.preset-tag:hover {
  background: rgba(0, 122, 255, 0.08);
  border-color: var(--color-primary);
}

/* Mapping textarea */
.mapping-textarea {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  line-height: 1.6;
  resize: vertical;
  min-height: 80px;
}

/* JSON error */
.json-error {
  color: var(--color-danger);
  font-size: 0.75rem;
  margin-bottom: 8px;
  padding: 4px 0;
}

/* Mapping actions */
.mapping-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* File upload area */
.file-upload-area {
  position: relative;
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-md);
  padding: 24px;
  text-align: center;
  transition: border-color var(--transition-fast), background var(--transition-fast);
  cursor: pointer;
}

.file-upload-area:hover {
  border-color: var(--color-primary);
  background: rgba(0, 122, 255, 0.03);
}

.file-upload-area.has-file {
  border-style: solid;
  border-color: var(--color-border);
  background: var(--color-bg);
}

.file-input-native {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.file-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--color-text-tertiary);
  font-size: 0.875rem;
  pointer-events: none;
}

.file-selected {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text);
  font-size: 0.875rem;
  pointer-events: none;
}

.file-selected .file-name {
  font-weight: 500;
}

.file-selected .file-size {
  color: var(--color-text-tertiary);
}

.file-remove-btn {
  pointer-events: auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.file-remove-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  color: var(--color-danger);
}

/* Upload progress */
.upload-progress {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 36px;
  text-align: right;
}

/* Template instructions */
.template-instructions {
  background: var(--color-bg);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  margin-bottom: 16px;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.template-instructions code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
}
</style>
