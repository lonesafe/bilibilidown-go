import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// Parse video URL
export function parseVideo(url, qn = 80, fetchLinks = false) {
  return api.post('/parse', { url, qn, fetchLinks })
}

// Start download
export function startDownload(task) {
  return api.post('/download', task)
}

// Get all tasks
export function getTasks() {
  return api.get('/tasks')
}

// Stop a task
export function stopTask(id) {
  return api.post('/tasks/stop', { id })
}

// Get login status
export function getLoginStatus() {
  return api.get('/login/status')
}

// Set cookies
export function setCookie(cookie) {
  return api.post('/login/cookie', { cookie })
}

// Generate QR code for login
export function genQRCode() {
  return api.get('/login/qrcode')
}

// Poll QR code login status
export function pollQRCode(qrcodeKey) {
  return api.post('/login/qrcode/poll', { qrcodeKey })
}

// Get user info
export function getUserInfo() {
  return api.get('/user')
}

// Get config
export function getConfig() {
  return api.get('/config')
}

// Update config
export function updateConfig(config) {
  return api.post('/config', config)
}

export default api
