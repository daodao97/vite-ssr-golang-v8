<template>
  <div class="redeem-code-tool">
    <el-card shadow="never">
      <template #header>
        <div class="header">
          兑换码生成工具
          <div class="actions">
            <el-button size="small" @click="handleGenerate" :loading="generating">
              生成兑换码
            </el-button>
            <el-button size="small" type="success" @click="handleSave" :disabled="!codes.length" :loading="saving">
              提交保存
            </el-button>
            <el-button size="small" type="primary" @click="handleDownload" :disabled="!codes.length">
              下载 .txt
            </el-button>
            <el-button size="small" @click="handleClear" :disabled="!codes.length">
              清空
            </el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" label-width="90px" class="config-form">
        <el-form-item label="分组 ID">
          <el-select
            v-model="groupId"
            clearable
            filterable
            allow-create
            placeholder="输入或选择分组"
            style="min-width: 200px"
            @change="ensureGroupIdOption"
            @visible-change="visible => visible && syncGroupOptions()"
          >
            <el-option v-for="gid in groupIdOptions" :key="gid" :label="gid" :value="gid" />
          </el-select>
        </el-form-item>
        <el-form-item label="生成数量">
          <el-input-number v-model="quantity" :min="1" :max="500" />
        </el-form-item>
        <el-form-item label="预览数量" v-if="codes.length">
          <el-tag type="info">{{ codes.length }}</el-tag>
        </el-form-item>
      </el-form>

      <el-alert
        type="info"
        show-icon
        class="mb12"
        title="规则说明"
        description="格式为 {group_id}-{UUID}；当分组为 default 或留空时，仅使用 UUID。"
      />

      <el-table v-if="codes.length" :data="codesTable" size="small" :border="true" v-loading="saving">
        <el-table-column type="index" width="60" label="#" />
        <el-table-column prop="code" label="兑换码">
          <template #default="{ row }">
            <span class="mono">{{ row.code }}</span>
            <el-button type="primary" link size="small" @click="() => copyCode(row.code)">
              复制
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-else class="placeholder">
        <el-empty description="请先选择分组与数量，点击生成兑换码" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const groupId = ref('default')
const quantity = ref(20)
const codes = ref([])
const generating = ref(false)
const saving = ref(false)
const groupIdOptions = ref(['default'])

const codesTable = computed(() => codes.value.map(code => ({ code })))

const ensureGroupIdOption = (gid) => {
  const val = typeof gid === 'string' ? gid.trim() : String(gid ?? '').trim()
  if (!val) return
  if (!groupIdOptions.value.includes(val)) {
    groupIdOptions.value = [...groupIdOptions.value, val].sort()
  }
}

const syncGroupOptions = async () => {
  try {
    const res = await fetch('/_api/team/group_ids')
    if (!res.ok) return
    const data = await res.json()
    const payload = typeof data.code !== 'undefined' ? data.data || {} : data
    const list = Array.isArray(payload.group_ids) ? payload.group_ids : []
    const merged = new Set(['default'])
    if (groupId.value && String(groupId.value).trim()) {
      merged.add(String(groupId.value).trim())
    }
    list.forEach(item => {
      if (typeof item === 'string' && item.trim()) {
        merged.add(item.trim())
      }
    })
    groupIdOptions.value = Array.from(merged).filter(Boolean).sort()
  } catch (err) {
    // ignore
  }
}

const randomUUID = () => {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  const alphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
  const bytes = new Uint8Array(24)
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(bytes)
  }
  return Array.from(bytes).map(b => alphabet[b % alphabet.length]).join('')
}

const buildCode = () => {
  const gid = String(groupId.value || '').trim() || 'default'
  const uuid = randomUUID()
  if (!gid || gid === 'default') {
    return uuid
  }
  return `${gid}-${uuid}`
}

const handleGenerate = () => {
  if (generating.value) return
  const count = Number(quantity.value)
  if (!Number.isFinite(count) || count <= 0) {
    ElMessage.error('生成数量必须大于 0')
    return
  }
  const capped = Math.min(Math.max(Math.floor(count), 1), 500)
  quantity.value = capped
  generating.value = true
  try {
    const list = []
    const seen = new Set()
    while (list.length < capped) {
      const code = buildCode()
      if (seen.has(code)) continue
      seen.add(code)
      list.push(code)
    }
    codes.value = list
    ensureGroupIdOption(groupId.value)
    ElMessage.success(`已生成 ${list.length} 个兑换码`)
  } finally {
    generating.value = false
  }
}

const copyCode = async (code) => {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(code)
      ElMessage.success('复制成功')
      return
    }
  } catch (err) {
    // ignore
  }
  const textarea = document.createElement('textarea')
  textarea.value = code
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  try {
    document.execCommand('copy')
    ElMessage.success('复制成功')
  } catch (err) {
    ElMessage.error('复制失败')
  } finally {
    document.body.removeChild(textarea)
  }
}

const handleSave = async () => {
  if (!codes.value.length) {
    ElMessage.warning('请先生成兑换码')
    return
  }
  if (saving.value) return
  saving.value = true
  try {
    const res = await fetch('/_api/team/redeem/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ codes: codes.value }),
    })
    const text = await res.text()
    let data
    try {
      data = JSON.parse(text)
    } catch (err) {
      data = null
    }
    if (!res.ok) {
      const msg = data && typeof data === 'object' ? data.message : text
      throw new Error(msg || '提交失败')
    }
    const payload = data && typeof data.code !== 'undefined' ? data.data || {} : data
    const inserted = Number(payload && payload.inserted)
    if (Number.isFinite(inserted) && inserted >= 0) {
      ElMessage.success(`成功导入 ${inserted} 个兑换码`)
    } else {
      ElMessage.success('兑换码已提交')
    }
  } catch (err) {
    ElMessage.error(err.message || '提交失败')
  } finally {
    saving.value = false
  }
}

const handleDownload = () => {
  if (!codes.value.length) {
    ElMessage.warning('没有可下载的兑换码')
    return
  }
  const blob = new Blob([codes.value.join('\n')], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const gid = String(groupId.value || 'default').trim() || 'default'
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
  const filename = `redeem_${gid}_${timestamp}.txt`
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const handleClear = () => {
  codes.value = []
}

onMounted(() => {
  ensureGroupIdOption(groupId.value)
  syncGroupOptions()
})
</script>

<style scoped>
.redeem-code-tool {
  padding: 12px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.config-form {
  margin-bottom: 12px;
}

.mb12 {
  margin-bottom: 12px;
}

.mono {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.placeholder {
  margin-top: 16px;
}
</style>
