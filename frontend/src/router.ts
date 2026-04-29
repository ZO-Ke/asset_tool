import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'ProjectList',
    component: () => import('./views/ProjectList.vue'),
  },
  {
    path: '/project/:id',
    name: 'ProjectDetail',
    component: () => import('./views/ProjectDetail.vue'),
    props: true,
  },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
