import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'DataBoard' }
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('@/views/products/index.vue'),
        meta: { title: '商品管理', icon: 'Goods' }
      },
      // {
      //   path: 'orders',
      //   name: 'Orders',
      //   component: () => import('@/views/orders/index.vue'),
      //   meta: { title: '订单管理', icon: 'List' }
      // },
      {
        path: 'members',
        name: 'Members',
        component: () => import('@/views/members/index.vue'),
        meta: { title: '会员管理', icon: 'User' }
      },
      {
        path: 'points-config',
        name: 'PointsConfig',
        component: () => import('@/views/points-config/index.vue'),
        meta: { title: '积分配置', icon: 'Medal' }
      },
      {
        path: 'recharge-activities',
        name: 'RechargeActivities',
        component: () => import('@/views/recharge-activities/index.vue'),
        meta: { title: '充值活动', icon: 'Wallet' }
      },
      {
        path: 'coupons',
        name: 'Coupons',
        component: () => import('@/views/coupons/index.vue'),
        meta: { title: '优惠券管理', icon: 'Ticket' }
      },
      {
        path: 'promotions',
        name: 'Promotions',
        component: () => import('@/views/promotions/index.vue'),
        meta: { title: '营销活动', icon: 'Present' }
      },
      // {
      //   path: 'inventory',
      //   name: 'Inventory',
      //   component: () => import('@/views/inventory/index.vue'),
      //   meta: { title: '库存管理', icon: 'Box' }
      // },
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('@/views/reports/index.vue'),
        meta: { title: '营业报表', icon: 'TrendCharts' }
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: () => import('@/views/analytics/index.vue'),
        meta: { title: '报表分析', icon: 'DataLine' }
      },
      {
        path: 'profit',
        name: 'Profit',
        component: () => import('@/views/profit/index.vue'),
        meta: { title: '利润分析', icon: 'Coin' }
      },
      {
        path: 'stores',
        name: 'Stores',
        component: () => import('@/views/stores/index.vue'),
        meta: { title: '门店管理', icon: 'OfficeBuilding' }
      },
      {
        path: 'printers',
        name: 'Printers',
        component: () => import('@/views/printers/index.vue'),
        meta: { title: '打印机管理', icon: 'Printer' }
      },
      {
        path: 'queues',
        name: 'Queues',
        component: () => import('@/views/queues/index.vue'),
        meta: { title: '排队管理', icon: 'Tickets' }
      },
      {
        path: 'tables',
        name: 'Tables',
        component: () => import('@/views/tables/index.vue'),
        meta: { title: '桌位管理', icon: 'Table' }
      },
      {
        path: 'reservations',
        name: 'Reservations',
        component: () => import('@/views/reservations/index.vue'),
        meta: { title: '预约管理', icon: 'Calendar' }
      },
      {
        path: 'stores/map',
        name: 'StoresMap',
        component: () => import('@/views/stores/map.vue'),
        meta: { title: '门店地图', icon: 'Location' }
      },
      {
        path: 'recommendations',
        name: 'Recommendations',
        component: () => import('@/views/recommendations/index.vue'),
        meta: { title: '智能推荐', icon: 'MagicStick' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const token = userStore.token

  if (to.meta.requiresAuth) {
    if (!token) {
      next({ path: '/login', query: { redirect: to.fullPath } })
    } else {
      if (!userStore.userInfo) {
        try {
          await userStore.getCurrentUser()
          next()
        } catch (e) {
          next({ path: '/login' })
        }
      } else {
        next()
      }
    }
  } else {
    if (token && to.path === '/login') {
      next({ path: '/' })
    } else {
      next()
    }
  }
})

export default router
