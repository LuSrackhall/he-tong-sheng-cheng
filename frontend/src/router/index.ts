import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { guest: true, title: '登录' },
    },
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue'),
      meta: { title: '概览' },
    },
    {
      path: '/new-contract',
      name: 'NewContract',
      component: () => import('../views/NewContract.vue'),
      meta: { title: '签新合同' },
    },
    {
      path: '/collect-rent',
      name: 'CollectRent',
      component: () => import('../views/CollectRent.vue'),
      meta: { title: '收租金' },
    },
    {
      path: '/arrears',
      name: 'Arrears',
      component: () => import('../views/ArrearsList.vue'),
      meta: { title: '催缴清单' },
    },
    {
      path: '/assets',
      name: 'Assets',
      component: () => import('../views/AssetList.vue'),
      meta: { title: '资产管理' },
    },
    {
      path: '/tenants',
      name: 'Tenants',
      component: () => import('../views/TenantList.vue'),
      meta: { title: '租户管理' },
    },
    {
      path: '/contracts',
      name: 'Contracts',
      component: () => import('../views/ContractList.vue'),
      meta: { title: '合同管理' },
    },
    {
      path: '/receipt-books',
      name: 'ReceiptBooks',
      component: () => import('../views/ReceiptBookList.vue'),
      meta: { title: '收据本' },
    },
    {
      path: '/receipts',
      name: 'Receipts',
      component: () => import('../views/ReceiptList.vue'),
      meta: { title: '收据记录' },
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/UserManagement.vue'),
      meta: { admin: true, title: '用户管理' },
    },
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('../views/Settings.vue'),
      meta: { title: '设置' },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.guest) {
    next()
    return
  }
  if (!auth.token) {
    next('/login')
    return
  }
  if (to.meta.admin && auth.user?.role !== 'admin') {
    next('/')
    return
  }
  next()
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - 租赁管家` : '租赁管家'
})

export default router
