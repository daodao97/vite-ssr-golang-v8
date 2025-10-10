<template>
  <div class="owner-maker-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          团队母号构建器
          <div class="actions">
            <el-button size="small" @click="clearAll">清空</el-button>
            <el-button size="small" @click="parseAll">一键解析</el-button>
            <el-button type="primary" size="small" @click="parseText">解析</el-button>
          </div>
        </div>
      </template>

      <div class="mail-fetch">
        <div class="tips">从邮箱池领取账号（自动过滤已使用记录）</div>
        <el-form :inline="true" size="small" class="mail-fetch-form" @submit.prevent>
          <el-form-item label="领取数量">
            <el-input-number v-model="mailFetch.count" :min="1" :max="200" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="mailFetch.loading" @click="fetchMails">获取账号</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- <div class="parse-actions">
        <el-space>
          <el-tag>共 {{ stats.total }} 行</el-tag>
          <el-tag type="success">成功 {{ stats.ok }}</el-tag>
          <el-tag type="warning">失败 {{ stats.fail }}</el-tag>
          <el-button size="small" @click="parseText">重新解析</el-button>
        </el-space>
      </div> -->

      <div class="global-config" style="margin: 12px 0;">
        <el-form :inline="true" size="small">
          <el-form-item label="分组ID">
            <el-select
              v-model="globalConfig.group_id"
              clearable
              filterable
              allow-create
              placeholder="可选，变更后自动应用"
              style="min-width: 200px"
            >
              <el-option v-for="gid in groupIdOptions" :key="gid" :label="gid" :value="gid" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- Cards input -->
      <el-divider />
      <div class="tips">付款卡信息（每行：卡号 空格 MMYY 空格 CVV；每张卡限用两次）</div>
      <el-input
        v-model="cardText"
        type="textarea"
        :rows="5"
        placeholder="例如：\n4549240600775409 0130 193\n4549240600774881 0629 008"
      />
      <div class="parse-actions">
        <el-space>
          <el-tag>卡片数：{{ cardStats.count }}</el-tag>
          <el-tag type="info">可用次数：{{ cardStats.uses }}</el-tag>
          <el-tag type="warning" v-if="cardFailures.length">无效：{{ cardFailures.length }}</el-tag>
          <el-button size="small" @click="parseCards">解析卡信息</el-button>
          <el-button size="small" @click="assignCards">重新分配</el-button>
        </el-space>
      </div>
      <el-table v-if="cardFailures.length" :data="cardFailures" size="small" :border="true" class="mt8">
        <el-table-column prop="lineNo" label="#" width="70" />
        <el-table-column prop="raw" label="原始文本" />
        <el-table-column prop="reason" label="原因" width="180" />
      </el-table>

      <!-- Address selector & manager -->
      <el-divider />
      <div class="tips">地址信息（表单管理，支持多个，保存在浏览器，仅用于自动支付）</div>
      <div style="display:flex; gap:10px; align-items:center; flex-wrap: wrap;">
        <el-select v-model="addrSelectedId" clearable filterable placeholder="未选择时使用默认地址" style="min-width: 320px">
          <el-option
            v-for="a in addresses"
            :key="a.id"
            :label="addrLabel(a)"
            :value="a.id"
          />
        </el-select>
        <el-button size="small" @click="openAddrAdd">新增地址</el-button>
        <el-button size="small" @click="openAddrManage" :disabled="addresses.length===0">管理地址</el-button>
      </div>

      <el-alert v-if="failures.length" title="以下行解析失败（已忽略）：" type="warning" show-icon :closable="false" />
      <el-table v-if="failures.length" :data="failures" size="small" :border="true" class="mt8">
        <el-table-column prop="lineNo" label="#" width="70" />
        <el-table-column prop="raw" label="原始文本" />
        <el-table-column prop="reason" label="原因" width="180" />
      </el-table>

      <el-divider />

      <el-table :data="rows" size="small" v-loading="loading" :border="true">
        <el-table-column type="index" width="60" label="#" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="gpt_pwd" label="GPT 密码" min-width="160" />
        <el-table-column prop="mail_pwd" label="邮箱密码" min-width="160" />
        <el-table-column prop="rt" label="RT" min-width="320">
          <template #default="{ row }">
            <span class="mono">{{ shorten(row.rt) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="支付卡" min-width="220">
          <template #default="{ row }">
            <template v-if="row.__card">
              <span class="mono">**** **** **** {{ row.__card.last4 }}｜{{ row.__card.expiry }}</span>
            </template>
            <template v-else>
              <el-tag type="danger">无可用卡</el-tag>
            </template>
          </template>
        </el-table-column>
        <el-table-column label="Workspace 名称" min-width="160">
          <template #default="{ row }">
            <el-input v-model="row.workspace" placeholder="默认 ChatGPT Business Workspace" />
          </template>
        </el-table-column>
        <el-table-column label="分组ID" width="160">
          <template #default="{ row }">
            <span class="mono">{{ row.group_id || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="席位" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.seats" :min="1" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row, $index }">
            <el-button type="primary" link size="small" :loading="row.__busy && row.__step==='create'" @click="buildTeam(row)">创建团队</el-button>
            <el-divider direction="vertical" />
            <el-button link size="small" :disabled="!row.__teamId" :loading="row.__busy && row.__step==='pay_auto'" @click="checkoutAuto(row)">
              {{ row.__autoPayStatus === 'success' ? '成功' : row.__autoPayStatus === 'retry' ? '重试' : '自动支付' }}
            </el-button>
            <el-divider direction="vertical" />
            <el-button link size="small" :disabled="!row.__teamId" :loading="row.__busy && row.__step==='pay_manual'" @click="checkoutManual(row)">手动支付</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
  
  <!-- Pay URL Dialog -->
  <el-dialog v-model="payDialog.visible" title="支付链接" width="640px">
    <el-alert title="请使用下方链接完成支付" type="info" show-icon :closable="false" class="mb8" />
    <el-input v-model="payDialog.url" readonly type="textarea" :rows="3" />
    <template #footer>
      <el-button @click="payDialog.visible=false">关闭</el-button>
      <el-button @click="copyPayUrl">复制</el-button>
      <el-button @click="markPaid" :disabled="!payDialog.teamId">已完成支付</el-button>
      <el-button type="primary" @click="openPayUrl">打开</el-button>
    </template>
  </el-dialog>

  <!-- Address Manage Dialog -->
  <el-dialog v-model="addrDlg.visible" :title="addrDlg.editingId ? '编辑地址' : '新增地址'" width="720px">
    <div class="mb8" v-if="addresses.length">
      <el-table :data="addresses" size="small" :border="true">
        <el-table-column label="姓名" width="140">
          <template #default="{ row }">{{ row.full_name }}</template>
        </el-table-column>
        <el-table-column label="地址" min-width="260">
          <template #default="{ row }">{{ row.address_line1 }} {{ row.address_line2 }}</template>
        </el-table-column>
        <el-table-column label="地区" width="200">
          <template #default="{ row }">{{ row.city }}, {{ row.state }} {{ row.zip_code }} ({{ row.country }})</template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button link size="small" @click="openAddrEdit(row)">编辑</el-button>
            <el-divider direction="vertical" />
            <el-popconfirm title="确定删除该地址？" @confirm="() => removeAddr(row)">
              <template #reference>
                <el-button link size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-form :model="addrForm" label-width="120px">
      <el-form-item label="姓名">
        <el-input v-model="addrForm.full_name" />
      </el-form-item>
      <el-form-item label="地址1">
        <el-input v-model="addrForm.address_line1" />
      </el-form-item>
      <el-form-item label="地址2">
        <el-input v-model="addrForm.address_line2" />
      </el-form-item>
      <el-form-item label="城市">
        <el-input v-model="addrForm.city" />
      </el-form-item>
      <el-form-item label="州/省">
        <el-input v-model="addrForm.state" />
      </el-form-item>
      <el-form-item label="邮编">
        <el-input v-model="addrForm.zip_code" />
      </el-form-item>
      <el-form-item label="国家">
        <el-input v-model="addrForm.country" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="addrDlg.visible=false">取消</el-button>
      <el-button type="primary" @click="saveAddr">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const rawText = ref('')
const rows = ref([])
const failures = ref([])
const loading = ref(false)
const stats = reactive({ total: 0, ok: 0, fail: 0 })
const mailFetch = reactive({ count: 4, loading: false })

const globalConfig = reactive({ group_id: 'default' })
const groupIdOptions = ref([])

const applyGroupId = () => {
  const gid = String(globalConfig.group_id || '').trim()
  for (const r of rows.value) {
    r.group_id = gid
  }
}

const ensureGroupIdOption = (gid) => {
  const val = typeof gid === 'string' ? gid.trim() : String(gid ?? '').trim()
  if (!val) return
  if (!groupIdOptions.value.includes(val)) {
    groupIdOptions.value = [...groupIdOptions.value, val].sort()
  }
}

const loadGroupIds = async () => {
  try {
    const res = await fetch('/_api/team/group_ids')
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    const list = Array.isArray(data.group_ids) ? data.group_ids : []
    const normalized = []
    for (const item of list) {
      if (typeof item !== 'string') continue
      const trimmed = item.trim()
      if (!trimmed) continue
      if (!normalized.includes(trimmed)) normalized.push(trimmed)
    }
    normalized.sort()
    groupIdOptions.value = normalized
  } catch (e) {
    // ignore fetch errors; keep existing options
  } finally {
    ensureGroupIdOption(globalConfig.group_id)
    rows.value.forEach(r => ensureGroupIdOption(r.group_id))
  }
}

const fetchMails = async () => {
  const count = Number(mailFetch.count || 0)
  if (!Number.isFinite(count) || count <= 0) {
    ElMessage.error('请输入有效的领取数量')
    return
  }
  mailFetch.loading = true
  try {
    const qs = new URLSearchParams({ count: String(Math.min(Math.max(Math.floor(count), 1), 200)) })
    const res = await fetch('/_api/team/mails/unused?' + qs.toString())
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    const items = Array.isArray(data.items) ? data.items : []
    if (!items.length) {
      rawText.value = ''
      rows.value = []
      failures.value = []
      Object.assign(stats, { total: 0, ok: 0, fail: 0 })
      ElMessage.warning('未获取到可用账号')
      return
    }
    const gid = String(globalConfig.group_id || '').trim()
    rows.value = items.map(it => ({
      email: it.mail || '',
      gpt_pwd: it.gpt_pwd || '',
      mail_pwd: it.mail_pwd || '',
      rt: it.rt || '',
      workspace: defaultWorkspace(),
      seats: 7,
      group_id: gid,
      __busy: false,
      __step: '',
      __teamId: undefined,
      __payUrl: '',
      __autoPayStatus: '',
    }))
    failures.value = []
    stats.total = items.length
    stats.ok = items.length
    stats.fail = 0
    ensureGroupIdOption(gid)
    assignCards()
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    mailFetch.loading = false
  }
}

// cards
const cardText = ref('')
const cards = ref([])
const cardFailures = ref([])
const cardStats = reactive({ count: 0, uses: 0 })

// addresses store (persisted in browser)
const ADDR_STORE_KEY = 'owner_addrs_store'
const ADDR_SELECTED_KEY = 'owner_addrs_selected'
const addresses = ref([])
const addrSelectedId = ref(null)
const addrDlg = reactive({ visible: false, editingId: null })
const addrForm = reactive({ full_name: '', address_line1: '', address_line2: '', city: '', state: '', zip_code: '', country: '' })

// auto-parse on changes (debounced)
let debounceTimer1 = null
let debounceTimer2 = null
let debounceTimer3 = null
watch(rawText, () => {
  clearTimeout(debounceTimer1)
  debounceTimer1 = setTimeout(() => {
    parseText()
    assignCards()
  }, 300)
})
watch(cardText, () => {
  clearTimeout(debounceTimer2)
  debounceTimer2 = setTimeout(() => {
    parseCards()
    assignCards()
  }, 300)
})
// removed address json watcher; addresses persisted via localStorage

watch(() => globalConfig.group_id, (v) => {
  applyGroupId()
  ensureGroupIdOption(v)
})

const clearAll = () => {
  rawText.value = ''
  rows.value = []
  failures.value = []
  Object.assign(stats, { total: 0, ok: 0, fail: 0 })
  cardText.value = ''
  cards.value = []
  cardFailures.value = []
  Object.assign(cardStats, { count: 0, uses: 0 })
  // keep saved addresses (persisted); do not clear here
}

const emailRe = /^[^@\s]+@[^@\s]+\.[^@\s]+$/
const splitRe = /-{3,}/

const defaultWorkspace = () => 'ChatGPT Business Workspace'

const shorten = (s) => {
  if (!s) return ''
  const str = String(s)
  if (str.length <= 28) return str
  return str.slice(0, 16) + '…' + str.slice(-8)
}

const parseText = () => {
  const lines = String(rawText.value || '').split(/\r?\n/)
  const ok = []
  const bad = []
  let lineNo = 0
  for (const line of lines) {
    lineNo++
    const raw = String(line || '').trim()
    if (!raw) continue
    const parts = raw.split(splitRe).map(s => s.trim()).filter(Boolean)
    if (parts.length < 4) {
      bad.push({ lineNo, raw, reason: '分隔符格式错误或字段不足' })
      continue
    }
    const [email, gpt_pwd, mail_pwd, rt, ...rest] = parts
    if (!emailRe.test(email)) {
      bad.push({ lineNo, raw, reason: '邮箱格式错误' })
      continue
    }
    if (!rt || (!rt.startsWith('rt_') || !rt.startsWith('xy'))) {
      bad.push({ lineNo, raw, reason: 'RT 缺失或格式不对' })
      continue
    }
    const item = {
      email,
      gpt_pwd: gpt_pwd || '',
      mail_pwd: mail_pwd || '',
      rt: [rt, ...rest].filter(Boolean).join('---'), // 容错将剩余拼回
      workspace: defaultWorkspace(),
      seats: 7,
      group_id: String(globalConfig.group_id || '').trim(),
      __busy: false,
      __step: '',
      __teamId: undefined,
      __payUrl: '',
      __autoPayStatus: '',
    }
    ok.push(item)
  }
  rows.value = ok
  applyGroupId()
  rows.value.forEach(r => ensureGroupIdOption(r.group_id))
  failures.value = bad
  stats.total = lines.filter(l => String(l).trim()).length
  stats.ok = ok.length
  stats.fail = bad.length
  // assign cards after parsing accounts
  assignCards()
}

const normalizeExpiry = (s) => {
  if (!s) return ''
  const raw = String(s).replace(/[^0-9]/g, '')
  if (raw.length === 4) return raw.slice(0,2) + '/' + raw.slice(2)
  if (raw.length === 6) return raw.slice(0,2) + '/' + raw.slice(4)
  if (/^\d{2}\/\d{2}$/.test(s)) return s
  return s
}

const parseCards = () => {
  const lines = String(cardText.value || '').split(/\r?\n/)
  const ok = []
  const bad = []
  let ln = 0
  for (const line of lines) {
    ln++
    const raw = String(line || '').trim()
    if (!raw) continue
    const parts = raw.split(/\s+/).filter(Boolean)
    if (parts.length < 3) { bad.push({ lineNo: ln, raw, reason: '格式需：卡号 MMYY CVV' }); continue }
    const [num, exp, cvc] = parts
    const numOnly = (num||'').replace(/\D/g, '')
    if (numOnly.length < 12) { bad.push({ lineNo: ln, raw, reason: '卡号无效' }); continue }
    const expiry = normalizeExpiry(exp)
    if (!/^\d{2}\/\d{2}$/.test(expiry)) { bad.push({ lineNo: ln, raw, reason: '有效期无效' }); continue }
    if (!/^\d{3,4}$/.test(String(cvc||''))) { bad.push({ lineNo: ln, raw, reason: 'CVV无效' }); continue }
    ok.push({ number: numOnly, last4: numOnly.slice(-4), expiry, cvc: String(cvc), uses: 0 })
  }
  cards.value = ok
  cardFailures.value = bad
  cardStats.count = ok.length
  cardStats.uses = ok.length * 2
  assignCards()
}

const assignCards = () => {
  // 按顺序每张卡分配两次：card1 -> rows[0], rows[1]; card2 -> rows[2], rows[3]...
  // 多余卡不会被使用；卡不够时，剩余账号标记为无卡
  const slots = []
  for (const c of cards.value) {
    slots.push(c)
    slots.push(c)
  }
  let si = 0
  for (const r of rows.value) {
    if (si < slots.length) {
      const c = slots[si++]
      r.__card = { last4: c.last4, expiry: c.expiry, cvc: c.cvc, number: c.number }
    } else {
      r.__card = null
    }
  }
}

// address helpers
const addrLabel = (a) => `${a.full_name || ''} — ${a.city || ''} ${a.state || ''} ${a.zip_code || ''}`.trim()
const loadAddresses = () => {
  try { addresses.value = JSON.parse(localStorage.getItem(ADDR_STORE_KEY) || '[]') } catch { addresses.value = [] }
  const sel = localStorage.getItem(ADDR_SELECTED_KEY)
  addrSelectedId.value = sel || null
}
const saveAddresses = () => { try { localStorage.setItem(ADDR_STORE_KEY, JSON.stringify(addresses.value || [])) } catch {} }
watch(addrSelectedId, (v) => { try { if (v) localStorage.setItem(ADDR_SELECTED_KEY, String(v)); else localStorage.removeItem(ADDR_SELECTED_KEY) } catch {} })
loadAddresses()
const openAddrAdd = () => { addrDlg.editingId = null; Object.assign(addrForm, { full_name: '', address_line1: '', address_line2: '', city: '', state: '', zip_code: '', country: '' }); addrDlg.visible = true }
const openAddrManage = () => { addrDlg.visible = true; if (!addrDlg.editingId) addrDlg.editingId = null }
const openAddrEdit = (row) => { addrDlg.editingId = row.id; Object.assign(addrForm, { ...row }); addrDlg.visible = true }
const removeAddr = (row) => {
  const idx = addresses.value.findIndex(a => String(a.id) === String(row.id))
  if (idx >= 0) {
    addresses.value.splice(idx, 1)
    if (String(addrSelectedId.value) === String(row.id)) addrSelectedId.value = null
    saveAddresses()
  }
}
const saveAddr = () => {
  const base = { full_name: addrForm.full_name?.trim(), address_line1: addrForm.address_line1?.trim(), address_line2: addrForm.address_line2?.trim(), city: addrForm.city?.trim(), state: addrForm.state?.trim(), zip_code: addrForm.zip_code?.trim(), country: addrForm.country?.trim() }
  if (!base.full_name || !base.address_line1 || !base.city || !base.state || !base.zip_code || !base.country) {
    ElMessage.error('请完整填写必填项（姓名/地址1/城市/州/邮编/国家）')
    return
  }
  if (addrDlg.editingId) {
    const idx = addresses.value.findIndex(a => String(a.id) === String(addrDlg.editingId))
    if (idx >= 0) addresses.value[idx] = { ...addresses.value[idx], ...base }
  } else {
    const id = Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
    addresses.value.push({ id, ...base })
    addrSelectedId.value = id
  }
  saveAddresses()
  addrDlg.visible = false
}
const getSelectedAddress = () => {
  const id = addrSelectedId.value
  if (!id) return null
  return addresses.value.find(a => String(a.id) === String(id)) || null
}

const parseAll = () => {
  parseText()
  parseCards()
  assignCards()
}

const buildTeam = async (row) => {
  if (!row || row.__busy) return
  row.__busy = true
  row.__step = 'create'
  try {
    const payload = { email: row.email, password: row.gpt_pwd, seats: Number(row.seats) || 7, status: -1, rt: row.rt }
    const gidText = String(row.group_id ?? '').trim()
    if (gidText) {
      payload.group_id = gidText
    }
    if (!payload.email || !payload.password || !payload.rt) throw new Error('email/password/rt 不能为空')
    const res = await fetch('/_api/team', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    if (!res.ok) {
      const txt = await res.text()
      let msg = txt
      try { const j = JSON.parse(txt); msg = j.message || txt } catch {}
      if (res.status === 409 || /账号已存在/.test(String(msg)) || /409/.test(String(msg))) {
        await useExistingTeam(row)
        return
      }
      throw new Error(msg)
    }
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    row.__teamId = data.id
    ElMessage.success(`团队已创建(ID=${row.__teamId || '?'})`)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    row.__busy = false
    row.__step = ''
  }
}

const useExistingTeam = async (row) => {
  const qs = new URLSearchParams({ email: row.email })
  const res = await fetch('/_api/teams?' + qs.toString())
  if (!res.ok) throw new Error(await res.text())
  const resp = await res.json()
  if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
  const list = (typeof resp.code !== 'undefined' ? (resp.data?.list || []) : (resp.list || [])) || []
  const found = list.find(i => String(i.email || '').toLowerCase() === String(row.email || '').toLowerCase())
  if (!found) throw new Error('账号已存在，但无法获取团队信息')
  const st = Number(found.status)
  if (st !== -1) throw new Error('账号已存在且非待支付状态，无法生成支付链接')
  row.__teamId = found.id
  ElMessage.success(`已关联已存在团队(ID=${row.__teamId})`)
}

const checkoutManual = async (row) => {
  if (!row || row.__busy) return
  if (!row.__teamId) { ElMessage.error('请先创建团队'); return }
  row.__busy = true
  row.__step = 'pay_manual'
  try {
    const payload = { team_id: row.__teamId, workspace_name: row.workspace || defaultWorkspace(), seat_quantity: 5, autopay: false }
    const res = await fetch('/_api/team/checkout', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    row.__payUrl = data.url || ''
    if (!row.__payUrl) throw new Error('生成支付链接失败')
    showPayDialog(row.__payUrl, row.__teamId)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    row.__busy = false
    row.__step = ''
  }
}

onMounted(() => {
  loadGroupIds()
})

const checkoutAuto = async (row) => {
  if (!row || row.__busy) return
  if (!row.__teamId) { ElMessage.error('请先创建团队'); return }
  if (!row.__card) { ElMessage.error('无可用卡，无法自动支付'); return }
  row.__autoPayStatus = ''
  row.__busy = true
  row.__step = 'pay_auto'
  let nextStatus = 'retry'
  try {
    const payload = { team_id: row.__teamId, workspace_name: row.workspace || defaultWorkspace(), seat_quantity: 5, autopay: true,
      payment_info: { card_number: row.__card.number, expiry_date: row.__card.expiry, cvc: row.__card.cvc },
      address_info: getSelectedAddress() || undefined,
    }
    const res = await fetch('/_api/team/checkout', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    if (data.paid) {
      ElMessage.success('已自动支付完成')
      nextStatus = 'success'
    } else if (data.url) {
      row.__payUrl = data.url
      showPayDialog(row.__payUrl, row.__teamId)
    } else {
      ElMessage.success('已生成支付请求')
    }
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    row.__busy = false
    row.__step = ''
    row.__autoPayStatus = nextStatus
  }
}

// Pay dialog
const payDialog = reactive({ visible: false, url: '', teamId: undefined })
const showPayDialog = (url, teamId) => { payDialog.url = url || ''; payDialog.teamId = teamId; payDialog.visible = !!payDialog.url }
const copyPayUrl = async () => {
  try { await navigator.clipboard.writeText(payDialog.url); ElMessage.success('已复制') } catch { /* ignore */ }
}
const openPayUrl = () => { try { window.open(payDialog.url, '_blank') } catch {} }
const markPaid = async () => {
  if (!payDialog.teamId) return
  try {
    const res = await fetch('/_api/team', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: payDialog.teamId, status: 1 }) })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || '标记失败')
    ElMessage.success('已标记为已支付（状态=1）')
    payDialog.visible = false
  } catch (e) {
    ElMessage.error(String(e))
  }
}

</script>

<style scoped>
.owner-maker-page { padding: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; font-weight: 600; }
.tips { color: #666; font-size: 12px; margin-bottom: 8px; }
.mail-fetch-form { margin: 8px 0 12px; }
.parse-actions { margin: 8px 0; }
.mt8 { margin-top: 8px; }
.mb8 { margin-bottom: 8px; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
</style>
