<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { contractApi, paymentApi, type Contract, type Payment } from '../api'

function useDebounce<F extends (...args: any[]) => void>(fn: F, delay: number): F {
  let timer: ReturnType<typeof setTimeout>
  return ((...args: any[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }) as F
}

const route = useRoute()

const contracts = ref<Contract[]>([])
const total = ref(0)
const search = ref('')
const page = ref(0)
const pageSize = 20
const selectedContract = ref<Contract | null>(null)
const showPayModal = ref(false)
const paymentAmount = ref(0)
const paymentNotes = ref('')
const shortfall = ref(0)
const payments = ref<Payment[]>([])
const saving = ref(false)
const errorMessage = ref('')
const submitLock = ref(false)

const onlyArrears = ref(true)

async function fetchContracts() {
  const params: any = { search: search.value, offset: page.value * pageSize, limit: pageSize }
  if (onlyArrears.value) params.status = 'arrears'
  const { data } = await contractApi.list(params)
  contracts.value = data.data
  total.value = data.total
}
const onSearchInput = useDebounce(() => { page.value = 0; fetchContracts() }, 300)
async function openPayModal(c: Contract) {
  selectedContract.value = c
  showPayModal.value = true
  paymentAmount.value = 0
  paymentNotes.value = ''
  shortfall.value = 0
  const { data: pmts } = await paymentApi.list(c.id)
  payments.value = pmts
}
async function recordPayment() {
  if (submitLock.value) return
  if (!selectedContract.value || paymentAmount.value <= 0) return
  submitLock.value = true
  saving.value = true
  try {
    const { data } = await paymentApi.create(selectedContract.value.id, { amount: paymentAmount.value, notes: paymentNotes.value })
    shortfall.value = data.shortfall
    paymentAmount.value = 0
    paymentNotes.value = ''
    const { data: pmts } = await paymentApi.list(selectedContract.value.id)
    payments.value = pmts
    fetchContracts()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || '收款失败，请重试'
  } finally {
    saving.value = false
    submitLock.value = false
  }
}

onMounted(() => {
  // 从催缴清单跳转时，自动搜索指定合同
  const contractId = route.query.contractId as string
  if (contractId) search.value = contractId
  fetchContracts()
})
</script>

<template>
  <div>
    <div class="page-header"><h2>收租金</h2></div>

    <div class="form-group" style="display: flex; gap: 12px; align-items: center;">
      <input class="input" v-model="search" @input="onSearchInput" placeholder="搜索租户名或资产名..." style="flex: 1;" />
      <label style="display: flex; align-items: center; gap: 6px; font-size: 0.875rem; color: var(--color-text-secondary); white-space: nowrap; cursor: pointer; user-select: none;">
        <input type="checkbox" v-model="onlyArrears" @change="page = 0; fetchContracts()" />
        仅欠费
      </label>
    </div>

    <div v-if="contracts.length === 0" class="empty-state">暂无匹配的合同</div>

    <div style="display: grid; gap: 12px;">
      <div v-for="c in contracts" :key="c.id" class="card" style="padding: 16px; cursor: pointer;" @click="openPayModal(c)">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div>
            <div style="font-weight: 600; font-size: 1rem;">{{ c.tenant?.name || '租户#' + c.tenantId }}</div>
            <div style="font-size: 0.8125rem; color: var(--color-text-secondary);">{{ c.asset?.name || '资产#' + c.assetId }} · ¥{{ c.monthlyRent }}/月</div>
          </div>
          <div style="text-align: right;">
            <div style="font-weight: 600; color: var(--color-primary);">¥{{ c.totalReceived?.toLocaleString() }} / ¥{{ c.totalReceivable?.toLocaleString() }}</div>
            <div style="font-size: 0.75rem;">
              <span v-if="c.totalReceivable - c.totalReceived > 0" style="color: var(--color-danger);">还差 ¥{{ (c.totalReceivable - c.totalReceived).toLocaleString() }}</span>
              <span v-else class="badge badge-success">已缴全</span>
            </div>
          </div>
        </div>
        <div style="margin-top: 4px;">
          <span class="badge" :class="{
            'badge-success': c.status === 'paidup',
            'badge-warning': c.status === 'active',
            'badge-danger': c.status === 'arrears',
            'badge-info': c.status === 'expired'
          }">{{ c.status }}</span>
        </div>
      </div>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="btn btn-secondary btn-sm" :disabled="page === 0" @click="page--; fetchContracts()">上一页</button>
      <span style="padding: 6px 12px; font-size: 0.875rem;">{{ page + 1 }} / {{ Math.ceil(total / pageSize) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="(page + 1) * pageSize >= total" @click="page++; fetchContracts()">下一页</button>
    </div>

    <!-- Payment Modal -->
    <div v-if="showPayModal && selectedContract" class="modal-overlay" @click.self="showPayModal = false">
      <div class="modal-content" style="max-width: 560px;">
        <h3>记录收款</h3>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 500;">{{ selectedContract.tenant?.name || '租户#' + selectedContract.tenantId }}</div>
          <div style="font-size: 0.8125rem; color: var(--color-text-secondary);">
            {{ selectedContract.asset?.name || '资产#' + selectedContract.assetId }} · ¥{{ selectedContract.monthlyRent }}/月
          </div>
          <div style="margin-top: 8px; display: flex; gap: 16px;">
            <div><span style="color: var(--color-text-secondary);">已收：</span>¥{{ selectedContract.totalReceived?.toLocaleString() }}</div>
            <div><span style="color: var(--color-text-secondary);">应收：</span>¥{{ selectedContract.totalReceivable?.toLocaleString() }}</div>
            <div v-if="selectedContract.totalReceivable - selectedContract.totalReceived > 0" style="color: var(--color-danger); font-weight: 500;">
              还差：¥{{ (selectedContract.totalReceivable - selectedContract.totalReceived).toLocaleString() }}
            </div>
          </div>
          <div v-if="shortfall > 0" style="margin-top: 8px; color: var(--color-success); font-weight: 500;">
            本次收款后还差：¥{{ shortfall.toLocaleString() }}
          </div>
        </div>

        <div class="form-group"><label class="label">收款金额</label><input class="input" type="number" v-model="paymentAmount" placeholder="输入收款金额" /></div>
        <div class="form-group"><label class="label">备注</label><input class="input" v-model="paymentNotes" placeholder="备注信息" /></div>
        <div v-if="errorMessage" class="alert alert-danger" style="margin-bottom: 12px;">{{ errorMessage }}</div>
        <div style="display: flex; gap: 8px; margin-bottom: 16px;">
          <button class="btn btn-primary" style="flex: 1;" :disabled="saving || paymentAmount <= 0" @click="recordPayment">
            {{ saving ? '记录中...' : '确认收款' }}
          </button>
          <button v-if="selectedContract.totalReceivable - selectedContract.totalReceived > 0" class="btn btn-secondary" :disabled="saving"
            @click="paymentAmount = selectedContract.totalReceivable - selectedContract.totalReceived">
            缴差额 ¥{{ (selectedContract.totalReceivable - selectedContract.totalReceived).toLocaleString() }}
          </button>
        </div>

        <h4 style="font-size: 0.875rem; margin-bottom: 8px;">收款记录</h4>
        <div v-if="payments.length === 0" style="font-size: 0.8125rem; color: var(--color-text-tertiary);">暂无收款记录</div>
        <div v-for="p in payments" :key="p.id" style="display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid var(--color-border); font-size: 0.875rem;">
          <span>¥{{ p.amount.toLocaleString() }}</span>
          <span style="color: var(--color-text-secondary);">{{ new Date(p.paidAt).toLocaleDateString('zh-CN') }}</span>
          <span style="color: var(--color-text-tertiary);">{{ p.notes }}</span>
        </div>
        <button class="btn btn-secondary" style="margin-top: 12px;" @click="showPayModal = false">关闭</button>
      </div>
    </div>
  </div>
</template>
