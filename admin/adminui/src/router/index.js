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
        path: '/team/maker',
        name: 'team-owner-maker',
        component: () => import('../views/team/Maker.vue'),
        meta: { title: '团队母号', icon: 'el-icon-user', menuType: 2, keepAlive: false }
      },
      {
        path: '/team/owner',
        name: 'team-owner',
        component: () => import('../views/team/Owner.vue'),
        meta: { title: '团队管理', icon: 'el-icon-user', menuType: 2, keepAlive: false }
      },
      {
        path: '/team/member',
        name: 'team-member',
        component: () => import('../views/team/Member.vue'),
        meta: { title: '成员管理', icon: 'el-icon-user', menuType: 2, keepAlive: false }
      },
    ]
  }
]
