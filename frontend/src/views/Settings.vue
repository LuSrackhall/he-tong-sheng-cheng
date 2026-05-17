<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
const form = ref({ name: '', filePath: '' })
const mapping = ref<Record<number, string>>({})
const saving = ref(false)
const submitLock = ref(false)
const error = ref('')

async function fetchTemplates() {
  try {
    const { data } = await api.get('/templates')
    templates.value = (data as any).data || data
    templates.value.forEach(t => {
      if (t.fieldMap) {
        try { mapping.value[t.id] = t.fieldMap }
        catch { mapping.value[t.id] = '' }
      }
    })
  } catch {
    // handled by interceptor
  }
}

async function createTemplate() {
  if (submitLock.value) return
  error.value = ''
  if (!form.value.name || !form.value.filePath) {
    error.value = '请填写模板名称和文件路径'
    return
  }
  submitLock.value = true
  saving.value = true
  try {
    await api.post('/templates', form.value)
    showUpload.value = false
    error.value = ''
    form.value = { name: '', filePath: '' }
    fetchTemplates()
  } catch (e: any) {
    error.value = e.response?.data?.error || '创建失败'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

async function saveMapping(t: Template) {
  const raw = mapping.value[t.id] || ''
  if (!raw.trim()) return
  await api.patch(`/templates/${t.id}`, { fieldMap: raw })
  alert('字段映射已保存')
  fetchTemplates()
}

onMounted(fetchTemplates)
</script>

<template>
  <div>
    <div class="page-header"><h2>系统设置</h2><button class="btn btn-primary" @click="showUpload = true">+ 上传模板</button></div>

    <h3 style="font-size: 1rem; margin-bottom: 12px;">合同模板</h3>

    <div v-if="templates.length === 0" class="empty-state">暂无模板，请上传合同模板文件</div>

    <div style="display: grid; gap: 12px;">
      <div v-for="t in templates" :key="t.id" class="card" style="padding: 16px;">
        <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 12px;">
          <div>
            <div style="font-weight: 600;">{{ t.name }}</div>
            <div style="font-size: 0.75rem; color: var(--color-text-tertiary);">{{ t.filePath }}</div>
            <div style="font-size: 0.75rem; color: var(--color-text-secondary);">创建于 {{ new Date(t.createdAt).toLocaleDateString('zh-CN') }}</div>
          </div>
        </div>
        <div class="form-group">
          <label class="label" style="font-size: 0.75rem;">字段映射（JSON，如 {"assetName": "资产名称"}）</label>
          <textarea
            class="input"
            v-model="mapping[t.id]"
            rows="3"
            style="font-family: monospace; font-size: 0.75rem;"
            placeholder='{"assetName": "资产名称"}'
          />
        </div>
        <button class="btn btn-secondary btn-sm" @click="saveMapping(t)">保存映射</button>
      </div>
    </div>

    <div v-if="showUpload" class="modal-overlay" @click.self="showUpload = false">
      <div class="modal-content">
        <h3>上传模板</h3>
        <div v-if="error" style="color: var(--color-danger); font-size: 0.8125rem; margin-bottom: 8px;">{{ error }}</div>
        <div class="form-group"><label class="label">模板名称</label><input class="input" v-model="form.name" placeholder="如: 标准商铺租赁合同" /></div>
        <div class="form-group"><label class="label">文件路径（上传到服务器 templates/ 目录）</label><input class="input" v-model="form.filePath" placeholder="如: templates/shop.docx" /></div>
        <div style="font-size: 0.75rem; color: var(--color-text-tertiary); margin-bottom: 12px;">请先将 .docx 模板文件放置到服务器的 templates/ 目录下，然后在此登记。</div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showUpload = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="createTemplate">{{ saving ? '登记中...' : '登记模板' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
