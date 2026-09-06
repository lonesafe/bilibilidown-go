import Home from './views/Home.vue'
import Downloads from './views/Downloads.vue'
import Settings from './views/Settings.vue'
import Login from './views/Login.vue'

const routes = [
  { path: '/', name: 'home', component: Home, meta: { title: '首页' } },
  { path: '/downloads', name: 'downloads', component: Downloads, meta: { title: '下载任务' } },
  { path: '/settings', name: 'settings', component: Settings, meta: { title: '设置' } },
  { path: '/login', name: 'login', component: Login, meta: { title: '登录' } }
]

export default routes
