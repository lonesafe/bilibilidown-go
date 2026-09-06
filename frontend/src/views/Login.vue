<template>
  <div class="login">
    <el-card style="max-width: 560px; margin: 0 auto;">
      <template #header>
        <div class="card-header">
          <span>账号登录</span>
          <el-tag v-if="userInfo.name" type="success" size="small">已登录: {{ userInfo.name }}</el-tag>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 扫码登录 -->
        <el-tab-pane label="扫码登录" name="qrcode">
          <div class="qrcode-area">
            <div v-if="loading" class="loading">
              <el-icon class="is-loading" :size="40"><Loading /></el-icon>
              <p>生成二维码中...</p>
            </div>
            <div v-else-if="qrUrl" class="qr-container">
              <div class="qr-image">
                <img :src="qrDataUrl" alt="登录二维码" />
                <div v-if="qrStatus === 'expired'" class="qr-mask" @click="refreshQR">
                  <el-icon :size="32"><Refresh /></el-icon>
                  <p>二维码已过期，点击刷新</p>
                </div>
                <div v-else-if="qrStatus === 'scanned'" class="qr-mask">
                  <el-icon :size="32" color="#67c23a"><CircleCheck /></el-icon>
                  <p>扫码成功，请在手机上确认</p>
                </div>
              </div>
              <p class="qr-tip">{{ qrStatusText }}</p>
              <el-button text type="primary" @click="refreshQR">
                <el-icon><Refresh /></el-icon> 刷新二维码
              </el-button>
            </div>
            <el-empty v-else description="加载失败" />
          </div>
        </el-tab-pane>

        <!-- Cookie 登录 -->
        <el-tab-pane label="Cookie 登录" name="cookie">
          <el-alert
            type="info"
            :closable="false"
            title="使用方法"
            description="请从浏览器登录 B站后，按 F12 打开开发者工具，在 Application/存储 → Cookies 中复制全部 Cookie 粘贴到下方输入框。"
            show-icon
            style="margin-bottom: 20px;"
          />
          <el-input
            v-model="cookie"
            type="textarea"
            :rows="6"
            placeholder="粘贴 Cookie 字符串，例如: SESSDATA=xxx; bili_jct=xxx; DedeUserID=xxx"
          />
          <div style="margin-top: 16px;">
            <el-button type="primary" @click="saveCookie">保存 Cookie 并登录</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>

      <el-divider v-if="userInfo.name" />

      <div v-if="userInfo.name" class="user-info">
        <el-avatar :size="48" :src="userInfo.avatar" />
        <div class="user-meta">
          <div class="uname">{{ userInfo.name }}</div>
          <div class="mid">UID: {{ userInfo.mid }}</div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { setCookie, getLoginStatus, getUserInfo, genQRCode, pollQRCode } from '../api'

const activeTab = ref('qrcode')
const cookie = ref('')
const userInfo = ref({ name: '', mid: 0, avatar: '' })

// QR code state
const loading = ref(false)
const qrUrl = ref('')
const qrDataUrl = ref('')
const qrcodeKey = ref('')
const qrStatus = ref('waiting') // waiting, scanned, expired
let pollTimer = null

const qrStatusText = ref('请使用 B站 App 扫码登录')

const generateQR = async () => {
  loading.value = true
  qrStatus.value = 'waiting'
  qrStatusText.value = '请使用 B站 App 扫码登录'
  try {
    const res = await genQRCode()
    qrUrl.value = res.data.url
    qrcodeKey.value = res.data.qrcodeKey
    qrDataUrl.value = await QRCode.toDataURL(res.data.url, {
      width: 240,
      margin: 1,
      color: { dark: '#000000', light: '#ffffff' }
    })
    startPolling()
  } catch (e) {
    ElMessage.error('生成二维码失败')
  } finally {
    loading.value = false
  }
}

const startPolling = () => {
  stopPolling()
  pollTimer = setInterval(async () => {
    try {
      const res = await pollQRCode(qrcodeKey.value)
      const data = res.data
      if (data.code === 0) {
        // Login success
        stopPolling()
        ElMessage.success('登录成功')
        await checkStatus()
        qrStatus.value = 'success'
      } else if (data.code === 86090) {
        // Scanned, waiting for confirm
        qrStatus.value = 'scanned'
        qrStatusText.value = '扫码成功，请在手机上确认登录'
      } else if (data.code === 86038 || data.code === 86101) {
        // Not scanned yet
        qrStatus.value = 'waiting'
        qrStatusText.value = '请使用 B站 App 扫码登录'
      } else {
        // Expired or other error
        stopPolling()
        qrStatus.value = 'expired'
        qrStatusText.value = data.message || '二维码已过期'
      }
    } catch (e) {
      // ignore poll errors
    }
  }, 2000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const refreshQR = () => {
  stopPolling()
  generateQR()
}

const saveCookie = async () => {
  if (!cookie.value.trim()) {
    ElMessage.warning('请输入 Cookie')
    return
  }
  try {
    await setCookie(cookie.value.trim())
    ElMessage.success('Cookie 已保存')
    await checkStatus()
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

const checkStatus = async () => {
  try {
    const res = await getLoginStatus()
    if (res.data.loggedIn) {
      const userRes = await getUserInfo()
      userInfo.value = userRes.data
    } else {
      userInfo.value = { name: '', mid: 0, avatar: '' }
    }
  } catch (e) {
    ElMessage.error('检查失败')
  }
}

watch(activeTab, (val) => {
  if (val === 'qrcode' && !qrUrl.value) {
    generateQR()
  }
})

onMounted(() => {
  checkStatus()
  if (activeTab.value === 'qrcode') {
    generateQR()
  }
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; font-size: 16px; font-weight: 600; }
.qrcode-area { display: flex; flex-direction: column; align-items: center; padding: 20px 0; }
.loading { display: flex; flex-direction: column; align-items: center; gap: 12px; color: #999; }
.qr-container { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.qr-image { position: relative; width: 240px; height: 240px; border: 1px solid #eee; border-radius: 8px; padding: 8px; background: #fff; }
.qr-image img { width: 100%; height: 100%; }
.qr-mask { position: absolute; inset: 0; background: rgba(255,255,255,0.92); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; border-radius: 8px; cursor: pointer; color: #666; }
.qr-tip { color: #666; font-size: 14px; margin: 0; }
.user-info { display: flex; align-items: center; gap: 16px; }
.user-meta .uname { font-size: 16px; font-weight: 600; }
.user-meta .mid { color: #999; font-size: 13px; margin-top: 4px; }
</style>
