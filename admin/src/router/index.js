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
      {
        path: 'sms',
        name: 'SmsMarketing',
        component: () => import('@/views/sms/index.vue'),
        meta: { title: '短信营销', icon: 'Message' }
      },
      {
        path: 'sms/templates',
        name: 'SmsTemplates',
        component: () => import('@/views/sms/templates.vue'),
        meta: { title: '短信模板', icon: 'Document' }
      },
      {
        path: 'sms/tasks',
        name: 'SmsTasks',
        component: () => import('@/views/sms/tasks.vue'),
        meta: { title: '短信任务', icon: 'List' }
      },
      {
        path: 'receipt-ads',
        name: 'ReceiptAds',
        component: () => import('@/views/receipt-ads/index.vue'),
        meta: { title: '小票广告', icon: 'Promotion' }
      },
      {
        path: 'time-slot-pricing',
        name: 'TimeSlotPricing',
        component: () => import('@/views/time-slot-pricing/index.vue'),
        meta: { title: '时段定价', icon: 'Clock' }
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
        path: 'review/trend',
        name: 'ReviewTrend',
        component: () => import('@/views/review/index.vue'),
        meta: { title: '点评趋势', icon: 'TrendCharts' }
      },
      {
        path: 'review/reviews',
        name: 'ReviewManage',
        component: () => import('@/views/review/reviews.vue'),
        meta: { title: '评价管理', icon: 'ChatDotRound' }
      },
      {
        path: 'review/workorders',
        name: 'ReviewWorkOrders',
        component: () => import('@/views/review/workorders.vue'),
        meta: { title: '工单告警', icon: 'Warning' }
      },
      {
        path: 'ingredients',
        name: 'Ingredients',
        component: () => import('@/views/ingredients/index.vue'),
        meta: { title: '食材管理', icon: 'Apple' }
      },
      {
        path: 'bom',
        name: 'BOM',
        component: () => import('@/views/bom/index.vue'),
        meta: { title: 'BOM管理', icon: 'Grid' }
      },
      {
        path: 'cost-alerts',
        name: 'CostAlerts',
        component: () => import('@/views/cost-alerts/index.vue'),
        meta: { title: '成本告警', icon: 'Bell' }
      },
      {
        path: 'transfers',
        name: 'Transfers',
        component: () => import('@/views/transfers/index.vue'),
        meta: { title: '库存调拨', icon: 'Switch' }
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
      },
      {
        path: 'suppliers',
        name: 'Suppliers',
        component: () => import('@/views/suppliers/index.vue'),
        meta: { title: '供应商管理', icon: 'ShoppingCart' }
      },
      {
        path: 'purchase-orders',
        name: 'PurchaseOrders',
        component: () => import('@/views/purchase-orders/index.vue'),
        meta: { title: '采购订单', icon: 'Purchase' }
      },
      {
        path: 'payables',
        name: 'Payables',
        component: () => import('@/views/payables/index.vue'),
        meta: { title: '应付账款', icon: 'Money' }
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
