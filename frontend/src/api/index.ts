import axios from 'axios'

const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.hash = '#/login'
    }
    return Promise.reject(err)
  },
)

export default api

export interface Asset { id: number; name: string; assetType: string; description?: string; status: string; extraFields?: string; createdAt: string }
export interface Tenant { id: number; name: string; idCard?: string; phone?: string; idCardImage?: string; extraFields?: string; createdAt: string }
export interface Contract { id: number; assetId: number; tenantId: number; asset?: Asset; tenant?: Tenant; startDate: string; endDate: string; monthlyRent: number; totalReceivable: number; totalReceived: number; deposit: number; status: string; notes?: string; createdAt: string }
export interface Payment { id: number; contractId: number; amount: number; paidAt: string; voided?: boolean; notes?: string }

export const assetApi = {
  list: (params?: any) => api.get<{ data: Asset[]; total: number }>('/assets', { params }),
  create: (data: Partial<Asset>) => api.post<Asset>('/assets', data),
  get: (id: number) => api.get<Asset>(`/assets/${id}`),
  update: (id: number, data: Partial<Asset>) => api.patch<Asset>(`/assets/${id}`, data),
}

export const tenantApi = {
  list: (params?: any) => api.get<{ data: Tenant[]; total: number }>('/tenants', { params }),
  create: (data: Partial<Tenant>) => api.post<Tenant>('/tenants', data),
  get: (id: number) => api.get<Tenant>(`/tenants/${id}`),
  update: (id: number, data: Partial<Tenant>) => api.patch<Tenant>(`/tenants/${id}`, data),
}

export const contractApi = {
  list: (params?: any) => api.get<{ data: Contract[]; total: number }>('/contracts', { params }),
  create: (data: Partial<Contract> & { templateId?: number }) => api.post<Contract>('/contracts', data),
  get: (id: number) => api.get<Contract>(`/contracts/${id}`),
  update: (id: number, data: Partial<Contract>) => api.patch<Contract>(`/contracts/${id}`, data),
  export: (id: number) => api.post<{ message: string }>(`/contracts/${id}/export`),
  download: (id: number) => api.get(`/contracts/${id}/download`, { responseType: 'blob' }),
  preview: async (id: number) => {
    const { data } = await api.get(`/contracts/${id}/preview`, { responseType: 'text' })
    const w = window.open('', '_blank')
    if (w) { w.document.write(data); w.document.close() }
  },
}

export interface Template {
  id: number
  name: string
  filePath: string
  fieldMap?: string
  activeFields?: string
  validated: boolean
  createdAt: string
}

export const templateApi = {
  list: () => api.get<{ data: Template[] } | Template[]>('/templates'),
  create: (name: string) => api.post<Template>('/templates', { name }),
  updateMapping: (id: number, fieldMap: string, activeFields: string) =>
    api.patch<Template>(`/templates/${id}`, { fieldMap, activeFields }),
  delete: (id: number) => api.delete<{ message: string }>(`/templates/${id}`),
  download: (id: number) => api.get(`/templates/${id}/download`, { responseType: 'blob' }),
  preview: async (id: number) => {
    const { data } = await api.get(`/templates/${id}/preview`, { responseType: 'text' })
    const w = window.open('', '_blank')
    if (w) { w.document.write(data); w.document.close() }
  },
  uploadFile: (id: number, file: File, onProgress?: (pct: number) => void) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post<Template>(`/templates/${id}/upload`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e: any) => {
        if (e.total && onProgress) onProgress(Math.round((e.loaded * 100) / e.total))
      },
    })
  },
}

export const paymentApi = {
  list: (contractId: number) => api.get<Payment[]>(`/contracts/${contractId}/payments`),
  create: (contractId: number, data: { amount: number; paidAt?: string; notes?: string }) =>
    api.post<{ payment: Payment; shortfall: number }>(`/contracts/${contractId}/payments`, data),
  void: (paymentId: number) => api.post<{ message: string }>(`/payments/${paymentId}/void`),
}

export const receiptApi = {
  print: async (paymentId: number) => {
    const { data } = await api.get(`/print/receipt/${paymentId}`, { responseType: 'text' })
    const w = window.open('', '_blank')
    if (w) {
      w.document.write(data)
      w.document.close()
    }
  },
}

export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string; user: { id: number; username: string; role: string } }>('/auth/login', { username, password }),
  me: () => api.get('/auth/me'),
  changePassword: (oldPassword: string, newPassword: string) =>
    api.put<{ message: string }>('/auth/password', { oldPassword, newPassword }),
  listUsers: () => api.get('/admin/users'),
  createUser: (data: { username: string; password: string; role: string }) => api.post('/admin/users', data),
  deleteUser: (id: number) => api.delete(`/admin/users/${id}`),
}

export const backupApi = {
  info: () => api.get<{ type: string; path: string; size?: number; lastModified?: string }>('/admin/backup/info'),
  backup: async () => {
    const response = await api.post('/admin/backup', {}, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([response.data as any]))
    const link = document.createElement('a')
    link.href = url
    const disposition = response.headers['content-disposition']
    const filename = disposition?.match(/filename="?(.+?)"?$/)?.[1] || `backup_${new Date().toISOString().slice(0, 10)}.db`
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },
  restore: (file: File) => {
    const fd = new FormData()
    fd.append('backup', file)
    return api.post('/admin/restore?confirmed=true', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
}
