<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { assetApi, type Asset } from '../api'

const assets = ref<Asset[]>([])
const total = ref(0)
const search = ref('')
const page = ref(0)
const pageSize = 20
const showModal = ref(false)
const editing = ref<Asset | null>(null)
const form = ref({ name: '', assetType: 'shop', description: '' })
const saving = ref(false)
const submitLock = ref(false)
const errorMessage = ref('')

async function fetchAssets() {
  const { data } = await assetApi.list({ search: search.value, offset: page.value * pageSize, limit: pageSize })
  assets.value = data.data
  total.value = data.total
}
function openCreate() { editing.value = null; form.value = { name: '', assetType: 'shop', description: '' }; showModal.value = true }
function openEdit(a: Asset) { editing.value = a; form.value = { name: a.name, assetType: a.assetType, description: a.description || '' }; showModal.value = true }
async function save() {
  if (submitLock.value) return
  if (!form.value.name) return
  submitLock.value = true
  saving.value = true
  try {
    if (editing.value) {
      await assetApi.update(editing.value.id, form.value)
    } else {
      await assetApi.create(form.value)
    }
    showModal.value = false
    errorMessage.value = ''
    fetchAssets()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '保存失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}
onMounted(fetchAssets)
</script>

<template>
  <div>
    <div class="page-header"><h2>资产管理</h2><button class="btn btn-primary" @click="openCreate">+ 新建资产</button></div>
    <div class="form-group"><input class="input" v-model="search" @input="fetchAssets" placeholder="搜索资产名称..." /></div>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>ID</th><th>名称</th><th>类型</th><th>状态</th><th>创建时间</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="a in assets" :key="a.id">
            <td>{{ a.id }}</td><td>{{ a.name }}</td><td>{{ a.assetType }}</td>
            <td><span :class="{ 'badge badge-success': a.status === 'leased', 'badge badge-info': a.status === 'idle' }">{{ a.status }}</span></td>
            <td>{{ new Date(a.createdAt).toLocaleDateString('zh-CN') }}</td>
            <td><button class="btn btn-secondary btn-sm" @click="openEdit(a)">编辑</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchAssets()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchAssets()">下一页</button>
    </div>

    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal-content">
        <h3>{{ editing ? '编辑资产' : '新建资产' }}</h3>
        <div class="form-group"><label class="label">名称</label><input class="input" v-model="form.name" /></div>
        <div class="form-group"><label class="label">类型</label>
          <select class="input" v-model="form.assetType">
            <option value="shop">商铺</option><option value="parking">车位</option><option value="booth">摊位</option><option value="equipment">设备</option><option value="other">其他</option>
          </select>
        </div>
        <div class="form-group"><label class="label">描述</label><input class="input" v-model="form.description" /></div>
        <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
