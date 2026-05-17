<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { tenantApi, type Tenant } from '../api'

const tenants = ref<Tenant[]>([])
const total = ref(0)
const search = ref('')
const page = ref(0)
const pageSize = 20
const showModal = ref(false)
const editing = ref<Tenant | null>(null)
const form = ref({ name: '', phone: '', idCard: '' })

async function fetchTenants() {
  const { data } = await tenantApi.list({ search: search.value, offset: page.value * pageSize, limit: pageSize })
  tenants.value = data.data
  total.value = data.total
}
function openCreate() { editing.value = null; form.value = { name: '', phone: '', idCard: '' }; showModal.value = true }
function openEdit(t: Tenant) { editing.value = t; form.value = { name: t.name, phone: t.phone || '', idCard: t.idCard || '' }; showModal.value = true }
async function save() {
  if (!form.value.name) return
  if (editing.value) {
    await tenantApi.update(editing.value.id, form.value)
  } else {
    await tenantApi.create(form.value)
  }
  showModal.value = false
  fetchTenants()
}
onMounted(fetchTenants)
</script>

<template>
  <div>
    <div class="page-header"><h2>租户管理</h2><button class="btn btn-primary" @click="openCreate">+ 新建租户</button></div>
    <div class="form-group"><input class="input" v-model="search" @input="fetchTenants" placeholder="搜索姓名/电话/身份证号..." /></div>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>ID</th><th>姓名</th><th>电话</th><th>身份证号</th><th>创建时间</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="t in tenants" :key="t.id">
            <td>{{ t.id }}</td><td>{{ t.name }}</td><td>{{ t.phone }}</td><td>{{ t.idCard }}</td>
            <td>{{ new Date(t.createdAt).toLocaleDateString('zh-CN') }}</td>
            <td><button class="btn btn-secondary btn-sm" @click="openEdit(t)">编辑</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchTenants()">上一页</button>
      <span>{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchTenants()">下一页</button>
    </div>

    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal-content">
        <h3>{{ editing ? '编辑租户' : '新建租户' }}</h3>
        <div class="form-group"><label class="label">姓名</label><input class="input" v-model="form.name" /></div>
        <div class="form-group"><label class="label">电话</label><input class="input" v-model="form.phone" /></div>
        <div class="form-group"><label class="label">身份证号</label><input class="input" v-model="form.idCard" /></div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" @click="save">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>
