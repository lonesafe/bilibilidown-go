<template>
  <el-container class="app-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#fb7299"><VideoPlay /></el-icon>
        <span class="logo-text">BilibiliDown</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="menu"
        @select="handleSelect"
      >
        <el-menu-item index="/">
          <el-icon><Search /></el-icon>
          <span>视频解析</span>
        </el-menu-item>
        <el-menu-item index="/downloads">
          <el-icon><Download /></el-icon>
          <span>下载任务</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-menu-item>
        <el-menu-item index="/login">
          <el-icon><User /></el-icon>
          <span>登录</span>
        </el-menu-item>
      </el-menu>
      <div class="version">v1.0.0 (Go + Vue3)</div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-title">{{ pageTitle }}</div>
        <div class="header-right">
          <el-tag v-if="loginStatus" type="success" size="small">已登录</el-tag>
          <el-tag v-else type="info" size="small">未登录</el-tag>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getLoginStatus } from './api'

const route = useRoute()
const router = useRouter()
const activeMenu = ref(route.path)
const loginStatus = ref(false)

const pageTitle = computed(() => route.meta?.title || 'BilibiliDown')

const handleSelect = (index) => {
  activeMenu.value = index
  router.push(index)
}

const checkLogin = async () => {
  try {
    const res = await getLoginStatus()
    loginStatus.value = res.data.loggedIn
  } catch (e) {
    loginStatus.value = false
  }
}

onMounted(() => {
  checkLogin()
})
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body, #app { height: 100%; }

.app-container { height: 100vh; }

.sidebar {
  background: #1a1a2e;
  color: #fff;
  display: flex;
  flex-direction: column;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
.logo-text { font-size: 18px; font-weight: bold; color: #fb7299; }

.menu {
  flex: 1;
  background: transparent;
  border-right: none;
}
.menu .el-menu-item {
  color: #ccc;
}
.menu .el-menu-item:hover,
.menu .el-menu-item.is-active {
  background: rgba(251, 114, 153, 0.15) !important;
  color: #fb7299 !important;
}

.version {
  padding: 15px;
  text-align: center;
  color: #666;
  font-size: 12px;
  border-top: 1px solid rgba(255,255,255,0.1);
}

.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  padding: 0 24px;
}
.header-title { font-size: 18px; font-weight: 600; color: #333; }

.main-content {
  background: #f5f6fa;
  padding: 24px;
}
</style>
