<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authApi } from '../api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

interface User {
  id: number
  username: string
  role: string
  createdAt: string
}

const users = ref<User[]>([])
const showCreate = ref(false)
const form = ref({ username: '', password: '', role: 'operator' })
const saving = ref(false)
const submitLock = ref(false)
const error = ref('')

async function fetchUsers() {
  try {
    const { data } = await authApi.listUsers()
    users.value = (data as any).data || data
  } catch {
    // handled by interceptor
  }
}

async function createUser() {
  if (submitLock.value) return
  error.value = ''
  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'
    return
  }
  submitLock.value = true
  saving.value = true
  try {
    await authApi.createUser(form.value)
    showCreate.value = false
    error.value = ''
    form.value = { username: '', password: '', role: 'operator' }
    fetchUsers()
  } catch (e: any) {
    error.value = e.response?.data?.error || '创建失败'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

async function deleteUser(id: number, username: string) {
  if (!confirm(`确定要删除用户 "${username}" 吗？`)) return
  await authApi.deleteUser(id)
  fetchUsers()
}

onMounted(fetchUsers)
</script>

<template>
  <div>
    <div class="page-header"><h2>用户管理</h2><button class="btn btn-primary" @click="showCreate = true">+ 新建用户</button></div>

    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>角色</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>#{{ u.id }}</td>
            <td>{{ u.username }}</td>
            <td>
              <span class="badge" :class="u.role === 'admin' ? 'badge-danger' : 'badge-info'">
                {{ u.role === 'admin' ? '管理员' : '操作员' }}
              </span>
            </td>
            <td>{{ new Date(u.createdAt).toLocaleDateString('zh-CN') }}</td>
            <td>
              <button
                v-if="u.id !== auth.user?.id"
                class="btn btn-secondary btn-sm"
                @click="deleteUser(u.id, u.username)"
              >
                删除
              </button>
              <span v-else style="font-size: 0.75rem; color: var(--color-text-tertiary);">当前用户</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal-content">
        <h3>新建用户</h3>
        <div v-if="error" style="color: var(--color-danger); font-size: 0.8125rem; margin-bottom: 8px;">{{ error }}</div>
        <div class="form-group"><label class="label">用户名</label><input class="input" v-model="form.username" /></div>
        <div class="form-group"><label class="label">密码</label><input class="input" type="password" v-model="form.password" /></div>
        <div class="form-group"><label class="label">角色</label>
          <select class="input" v-model="form.role">
            <option value="admin">管理员</option>
            <option value="operator">操作员</option>
          </select>
        </div>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-secondary" @click="showCreate = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="createUser">{{ saving ? '创建中...' : '创建' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
