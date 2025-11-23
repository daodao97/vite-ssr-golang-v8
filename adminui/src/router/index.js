import { Layout } from '@okiss/oms'

import Dashboard from '../views/dashboard/index.vue'

export const routes = [
  {
    path: '/',
    component: Layout,
     meta: {
      menuType: 1
    },
    children: [
      {
        path: '/',
        name: 'dashboard',
        component: Dashboard,
        meta: { title: 'Dash', icon: 'el-icon-user', menuType: 2, keepAlive: false }
      },
    ]
  }
]
