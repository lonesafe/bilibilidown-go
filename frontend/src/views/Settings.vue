<template>
  <div class="settings">
    <el-card>
      <template #header>
        <span style="font-size:16px;font-weight:600;">下载设置</span>
      </template>
      <el-form :model="form" label-width="160px" style="max-width: 640px;">
        <el-form-item label="下载保存路径">
          <el-input v-model="form.savePath" placeholder="例如: ./download/" />
        </el-form-item>
        <el-form-item label="FFmpeg 路径">
          <el-input v-model="form.ffmpegPath" placeholder="留空则使用系统 PATH 中的 ffmpeg" />
        </el-form-item>
        <el-form-item label="并发下载数">
          <el-input-number v-model="form.downloadPoolSize" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="下载失败重试次数">
          <el-input-number v-model="form.maxFailRetry" :min="0" :max="10" />
        </el-form-item>
        <el-form-item label="下载格式">
          <el-select v-model="form.downloadFormat">
            <el-option label="MP4 (DASH 音视频合并)" :value="0" />
            <el-option label="FLV" :value="1" />
            <el-option label="MP4 (直接下载)" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="视频编码优先级">
          <el-select
            v-model="form.videoCodecPriority"
            multiple
            collapse-tags
            placeholder="选择优先级，越靠前优先级越高"
          >
            <el-option label="HEVC (H.265)" :value="12" />
            <el-option label="AVC (H.264)" :value="7" />
            <el-option label="AV1" :value="13" />
          </el-select>
        </el-form-item>
        <el-form-item label="音频质量优先级">
          <el-select
            v-model="form.audioQualityPriority"
            multiple
            collapse-tags
            placeholder="选择优先级"
          >
            <el-option label="192K (30280)" :value="30280" />
            <el-option label="132K (30232)" :value="30232" />
            <el-option label="64K (30216)" :value="30216" />
            <el-option label="Hi-Res (30251)" :value="30251" />
            <el-option label="杜比全景声 (30250)" :value="30250" />
          </el-select>
        </el-form-item>
        <el-form-item label="HTTP 服务端口">
          <el-input-number v-model="form.serverPort" :min="1024" :max="65535" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveConfig">保存设置</el-button>
          <el-button @click="loadConfig">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px;">
      <template #header>
        <span style="font-size:16px;font-weight:600;">关于</span>
      </template>
      <p>BilibiliDown Go 版 - 由 Java Swing 应用反编译后重构为 Go + Vue3 架构</p>
      <p style="margin-top:8px;color:#999;font-size:13px;">
        原版作者: nICEnnnnnnnLee | 项目地址: https://github.com/nICEnnnnnnnLee/BilibiliDown
      </p>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getConfig, updateConfig } from '../api'

const form = ref({
  savePath: './download/',
  ffmpegPath: '',
  downloadPoolSize: 3,
  maxFailRetry: 3,
  downloadFormat: 0,
  videoCodecPriority: [12, 7, 13],
  audioQualityPriority: [30280, 30232, 30216],
  serverPort: 8787
})

const loadConfig = async () => {
  try {
    const res = await getConfig()
    Object.assign(form.value, res.data)
  } catch (e) {
    ElMessage.error('加载配置失败')
  }
}

const saveConfig = async () => {
  try {
    await updateConfig(form.value)
    ElMessage.success('设置已保存')
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.settings { max-width: 720px; margin: 0 auto; }
</style>
