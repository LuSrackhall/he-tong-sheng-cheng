import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      redirect: '/new-contract',
    },
    {
      path: '/new-contract',
      name: 'NewContract',
      component: () => import('../views/NewContract.vue'),
    },
    {
      path: '/collect-rent',
      name: 'CollectRent',
      component: () => import('../views/CollectRent.vue'),
    },
    {
      path: '/arrears',
      name: 'Arrears',
      component: () => import('../views/ArrearsList.vue'),
    },
    {
      path: '/assets',
      name: 'Assets',
      component: () => import('../views/AssetList.vue'),
    },
    {
      path: '/tenants',
      name: 'Tenants',
      component: () => import('../views/TenantList.vue'),
    },
    {
      path: '/contracts',
      name: 'Contracts',
      component: () => import('../views/ContractList.vue'),
    },
    {
      path: '/receipt-books',
      name: 'ReceiptBooks',
      component: () => import('../views/ReceiptBookList.vue'),
    },
    {
      path: '/receipts',
      name: 'Receipts',
      component: () => import('../views/ReceiptList.vue'),
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/UserManagement.vue'),
      meta: { admin: true },
    },
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('../views/Settings.vue'),
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

export default router
