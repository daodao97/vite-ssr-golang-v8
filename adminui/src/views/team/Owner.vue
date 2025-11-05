<template>
  <div class="team-owner-list">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title">
            团队管理
            <el-tag class="remain-tag" size="small" type="info">剩余座位：{{ remain.remain }}（总 {{ remain.total_seats }}，用 {{ remain.used_seats }}）｜可用邮箱：{{ remain.available_mails }}</el-tag>
          </div>
          <div class="actions">
            <el-button type="danger" plain size="small" :loading="bulk.running" @click="startBulkCancel">一键退订</el-button>
            <el-button type="primary" size="small" @click="openAdd">添加</el-button>
          </div>
        </div>
      </template>

      <!-- Filters -->
      <el-form :inline="true" :model="filters" class="filters" @submit.prevent>
        <el-form-item label="邮箱">
          <el-input v-model="filters.email" placeholder="精确匹配" clearable />
        </el-form-item>
        <el-form-item label="分组ID">
          <el-select
            v-model="filters.group_id"
            clearable
            filterable
            allow-create
            placeholder="全部"
            style="width: 160px"
          >
            <el-option v-for="gid in groupIdOptions" :key="gid" :label="gid" :value="gid" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width: 140px"
            @change="v => filters.status = v === '' || v == null ? null : Number(v)">
            <el-option :value="1" label="正常" />
            <el-option :value="2" label="禁用" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- Table -->
      <el-table :data="rows" size="small" v-loading="loading" :border="true">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="group_id" label="分组ID" width="110" />
        <el-table-column prop="account_id" label="Account ID" width="140" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="seats" label="席位" width="80" />
        <el-table-column prop="used_seats" label="已使用席位" width="110" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'danger'">{{ Number(row.status) === 1 ? '正常' : '禁用'
              }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expire_at" label="过期时间" width="180">
          <template #default="{ row }">{{ formatTime(row.expire_at) }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link size="small" @click="openEdit(row)">编辑</el-button>
            <el-divider direction="vertical" />
            <el-button link size="small" @click="refresh(row)" :disabled="!row.rt">刷新</el-button>
            <el-divider direction="vertical" />
            <el-button link size="small" @click="openDetails(row)">详情</el-button>
            <el-divider direction="vertical" />
            <el-dropdown trigger="click" @command="(cmd) => onMoreCommand(cmd, row)">
              <el-button link size="small">更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="status">查询订阅状态</el-dropdown-item>
                  <el-dropdown-item command="cancel" :disabled="!row.id">取消订阅</el-dropdown-item>
                  <el-dropdown-item command="delete" divided :disabled="!row.id" style="color: var(--el-color-danger);">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[10, 20, 50, 100]"
          :current-page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>

    <!-- Dialog: Add/Edit -->
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑团队' : '添加团队'" width="560px">
      <el-form :model="form" label-width="120px">
        <el-form-item v-if="dialog.isEdit" label="ID">
          <el-input v-model.number="form.id" disabled />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" :disabled="dialog.isEdit" />
        </el-form-item>
        <el-form-item label="密码" v-if="!dialog.isEdit">
          <el-input v-model="form.password" show-password />
        </el-form-item>
        <el-form-item label="席位">
          <el-input-number v-model="form.seats" :min="1" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option :value="1" label="正常" />
            <el-option :value="2" label="禁用" />
          </el-select>
        </el-form-item>
        <el-form-item label="RT">
          <el-input v-model="form.rt" />
        </el-form-item>
        <el-form-item label="分组ID">
          <el-select
            v-model="form.group_id"
            clearable
            filterable
            allow-create
            placeholder="可选，可留空"
            style="width: 220px"
          >
            <el-option v-for="gid in groupIdOptions" :key="gid" :label="gid" :value="gid" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Details (Users / Pending) -->
    <el-dialog v-model="details.visible" :title="`团队详情 #${details.teamId || ''}`" width="820px" @closed="onDetailsClosed">
      <div class="details-container" v-loading="details.loading">
        <el-tabs v-model="details.activeTab" class="details-tabs" tab-position="top">
          <el-tab-pane label="用户列表" name="users">
            <el-table :data="details.usersView" size="small" :border="true">
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="role" label="角色" width="120" />
              <el-table-column prop="created_time" label="创建时间" width="180">
                <template #default="{ row }">{{ formatTime(row.created_time) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="120">
                <template #default="{ row }">
                  <el-popconfirm title="确定要移除该用户吗？" @confirm="() => removeUser(row)">
                    <template #reference>
                      <el-button link size="small" type="danger" :disabled="String(row.role) === 'account-owner'">移除</el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <template v-if="String(row.role) === 'account-owner'">
                    <el-tag type="info">owner</el-tag>
                  </template>
                  <template v-else>
                    <el-tag :type="row.isWhitelist ? 'success' : 'warning'">{{ row.isWhitelist ? '白名单' : '未知'
                      }}</el-tag>
                  </template>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="Pending 列表" name="pending">
            <el-table :data="details.pendingView" size="small" :border="true">
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="role" label="角色" width="120" />
              <el-table-column prop="created_time" label="创建时间" width="180">
                <template #default="{ row }">{{ formatTime(row.created_time) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="120">
                <template #default="{ row }">
                  <el-popconfirm title="确定要移除该邀请吗？" @confirm="() => removePending(row)">
                    <template #reference>
                      <el-button link size="small" type="danger" :disabled="String(row.role) === 'account-owner'">移除</el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <template v-if="String(row.role) === 'account-owner'">
                    <el-tag type="info">owner</el-tag>
                  </template>
                  <template v-else>
                    <el-tag :type="row.isWhitelist ? 'success' : 'warning'">{{ row.isWhitelist ? '白名单' : '未知'
                      }}</el-tag>
                  </template>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="details.visible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Bulk Cancel -->
    <el-dialog
      v-model="bulk.visible"
      title="一键退订结果"
      width="660px"
      :close-on-click-modal="!bulk.running"
      :close-on-press-escape="!bulk.running"
      :show-close="!bulk.running"
    >
      <div class="bulk-result-body" v-loading="bulk.running">
        <el-table :data="bulk.logs" size="small" :border="true">
          <el-table-column prop="id" label="ID" width="100" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <el-tag size="small" :type="bulkStatusTagType(row.status)">{{ bulkStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="结果" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="bulk.visible = false" :disabled="bulk.running">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Subscription Status -->
    <el-dialog v-model="sub.visible" :title="`订阅状态 #${sub.teamId || ''}`" width="720px">
      <div v-loading="sub.loading" style="max-height: 540px; overflow: auto;">
        <template v-if="!sub.loading">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="计划类型">
              <el-tag size="small">{{ planTypeText(sub.data?.plan_type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="计费周期">{{ billingPeriodText(sub.data?.billing_period) }}</el-descriptions-item>
            <el-descriptions-item label="席位">
              {{ Number(sub.data?.seats_in_use || 0) }} / {{ Number(sub.data?.seats_entitled || 0) }}
            </el-descriptions-item>
            <el-descriptions-item label="币种">{{ sub.data?.billing_currency || '-' }}</el-descriptions-item>
            <el-descriptions-item label="生效时间">{{ formatTime(sub.data?.active_start || '') }}</el-descriptions-item>
            <el-descriptions-item label="到期时间">{{ formatTime(sub.data?.active_until || '') }}</el-descriptions-item>
            <el-descriptions-item label="自动续订">
              <el-tag size="small" :type="sub.data?.will_renew ? 'success' : 'info'">{{ sub.data?.will_renew ? '是' : '否' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="是否欠费">
              <el-tag size="small" :type="sub.data?.is_delinquent ? 'danger' : 'success'">{{ sub.data?.is_delinquent ? '是' : '否' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="非营利折扣">
              <el-tag size="small" :type="sub.data?.non_profit_org_discount_applied ? 'success' : 'info'">{{ sub.data?.non_profit_org_discount_applied ? '是' : '否' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="订阅ID">{{ sub.data?.id || '-' }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="sub.data?.is_delinquent" style="margin-top: 8px;">
            <el-alert type="warning" :closable="false" show-icon>
              <template #title>
                账户存在欠费
                <span v-if="sub.data?.grace_period_end_timestamp">，宽限结束：{{ formatTime(sub.data?.grace_period_end_timestamp) }}</span>
              </template>
            </el-alert>
          </div>

          <div style="margin-top: 8px; text-align: right;">
            <el-button text type="primary" @click="sub.showRaw = !sub.showRaw">{{ sub.showRaw ? '隐藏原始JSON' : '查看原始JSON' }}</el-button>
          </div>
          <div v-show="sub.showRaw" style="margin-top: 6px;">
            <pre style="white-space: pre-wrap; word-break: break-all;">{{ prettyJSON(sub.data) }}</pre>
          </div>
        </template>
      </div>
      <template #footer>
        <el-button @click="sub.visible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const submitting = ref(false)
const filters = reactive({ email: '', group_id: null, status: null })
const hasQueried = ref(false)
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const rows = ref([])
const remain = reactive({ total_seats: 0, used_seats: 0, remain: 0, available_mails: 0 })

const dialog = reactive({ visible: false, isEdit: false })
const form = reactive({ id: undefined, email: '', password: '', seats: 1, status: 1, rt: '', group_id: null })
const groupIdOptions = ref([])
const route = useRoute()
const router = useRouter()

// details dialog state
const details = reactive({
  visible: false,
  loading: false,
  activeTab: 'users',
  teamId: undefined,
  usersView: [],
  pendingView: []
})

// subscription status dialog
const sub = reactive({ visible: false, loading: false, teamId: undefined, data: {}, showRaw: false })

// bulk cancel dialog
const bulk = reactive({ visible: false, running: false, logs: [] })

const filterQueryKeys = ['email', 'group_id', 'status', 'queried']
let syncingQuery = false

const toSingleValue = (val) => (Array.isArray(val) ? val[val.length - 1] : val)

const normalizeEmail = (val) => {
  const raw = toSingleValue(val)
  return typeof raw === 'string' ? raw : ''
}

const normalizeGroupId = (val) => {
  const raw = toSingleValue(val)
  if (typeof raw !== 'string') return null
  const trimmed = raw.trim()
  return trimmed ? trimmed : null
}

const normalizeStatus = (val) => {
  const raw = toSingleValue(val)
  if (raw == null || raw === '') return null
  const num = Number(raw)
  return Number.isFinite(num) ? num : null
}

const normalizeQueried = (val) => toSingleValue(val) === '1'

const extractFilterQuery = (query) => {
  const result = {}
  for (const key of filterQueryKeys) {
    const raw = query[key]
    if (Array.isArray(raw)) {
      if (!raw.length) continue
      const last = raw[raw.length - 1]
      if (last == null || last === '') continue
      result[key] = last
      continue
    }
    if (raw == null || raw === '') continue
    result[key] = raw
  }
  return result
}

const queriesEqual = (a, b) => {
  const entriesA = Object.entries(a).sort((x, y) => x[0].localeCompare(y[0]))
  const entriesB = Object.entries(b).sort((x, y) => x[0].localeCompare(y[0]))
  if (entriesA.length !== entriesB.length) return false
  for (let i = 0; i < entriesA.length; i++) {
    if (entriesA[i][0] !== entriesB[i][0] || entriesA[i][1] !== entriesB[i][1]) return false
  }
  return true
}

const buildFilterQuery = () => {
  const query = {}
  if (filters.email) {
    query.email = filters.email
  }
  const gid = typeof filters.group_id === 'string' ? filters.group_id.trim() : ''
  if (gid) {
    query.group_id = gid
  }
  if (filters.status !== null && filters.status !== undefined && filters.status !== '') {
    query.status = String(filters.status)
  }
  if (hasQueried.value) {
    query.queried = '1'
  }
  return query
}

const syncQueryFromFilters = async () => {
  const target = buildFilterQuery()
  const currentFilterPart = extractFilterQuery(route.query)
  if (queriesEqual(currentFilterPart, target)) return
  const preserved = {}
  for (const [key, value] of Object.entries(route.query)) {
    if (!filterQueryKeys.includes(key)) {
      preserved[key] = value
    }
  }
  syncingQuery = true
  try {
    await router.replace({ query: { ...preserved, ...target } })
  } finally {
    syncingQuery = false
  }
}

const applyQueryToFilters = async (query, { triggerLoad = false } = {}) => {
  const nextEmail = normalizeEmail(query.email)
  const nextGroupId = normalizeGroupId(query.group_id)
  const nextStatus = normalizeStatus(query.status)
  const queryHasQueried = normalizeQueried(query.queried)
  const nextHasQueried = queryHasQueried || !!nextEmail || nextGroupId !== null || nextStatus !== null

  if (filters.email !== nextEmail) filters.email = nextEmail
  if (filters.group_id !== nextGroupId) filters.group_id = nextGroupId
  if (filters.status !== nextStatus) filters.status = nextStatus

  hasQueried.value = nextHasQueried

  if (!nextHasQueried) {
    if (triggerLoad) {
      rows.value = []
      pagination.total = 0
      pagination.page = 1
      remain.total_seats = 0
      remain.used_seats = 0
      remain.remain = 0
      remain.available_mails = 0
    }
    return
  }

  if (triggerLoad) {
    pagination.page = 1
    await load(true)
    await loadRemain(true)
  }
}

const resetFilters = () => { filters.email = ''; filters.group_id = null; filters.status = null }
const onSearch = async () => {
  hasQueried.value = true
  pagination.page = 1
  await load(true)
  await loadRemain(true)
}
const onReset = async () => {
  resetFilters()
  pagination.page = 1
  hasQueried.value = false
  rows.value = []
  pagination.total = 0
  remain.total_seats = 0
  remain.used_seats = 0
  remain.remain = 0
  remain.available_mails = 0
  await syncQueryFromFilters()
}
const handlePageChange = (page) => {
  if (!hasQueried.value) return
  pagination.page = page
  load()
}
const handlePageSizeChange = (size) => {
  if (!hasQueried.value) return
  pagination.pageSize = size
  pagination.page = 1
  load()
}

const load = async (force = false) => {
  if (force) hasQueried.value = true
  if (!force && !hasQueried.value) return
  loading.value = true
  try {
    const qs = new URLSearchParams()
    if (filters.email) qs.set('email', filters.email)
    if (filters.status != null && filters.status !== '') qs.set('status', String(filters.status))
    const gid = String(filters.group_id || '').trim()
    if (gid) qs.set('group_id', gid)
    qs.set('page', String(pagination.page))
    qs.set('page_size', String(pagination.pageSize))
    const res = await fetch('/_api/teams' + (qs.toString() ? ('?' + qs.toString()) : ''))
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = resp && typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    rows.value = data.list || []
    rows.value.forEach(item => ensureGroupIdOption(item.group_id))
    const total = Number(data.total ?? 0)
    if (!Number.isNaN(total)) pagination.total = total
    const respPage = Number(data.page)
    if (!Number.isNaN(respPage) && respPage > 0) pagination.page = respPage
    const respSize = Number(data.page_size)
    if (!Number.isNaN(respSize) && respSize > 0) pagination.pageSize = respSize
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    if (hasQueried.value) {
      await syncQueryFromFilters().catch(() => {})
    }
    loading.value = false
  }
}

const loadRemain = async (force = false) => {
  if (force) hasQueried.value = true
  if (!force && !hasQueried.value) return
  try {
    const qs = new URLSearchParams()
    const gid = String(filters.group_id || '').trim()
    if (gid) qs.set('group_id', gid)
    const res = await fetch('/_api/app/remain' + (qs.toString() ? ('?' + qs.toString()) : ''))
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    remain.total_seats = Number(data.total_seats || 0)
    remain.used_seats = Number(data.used_seats || 0)
    remain.remain = Number(data.remain || 0)
    remain.available_mails = Number(data.available_mails || 0)
  } catch (e) {
    // 静默失败，避免影响主列表
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
    // ignore fetch errors for group ids; keep existing options
  } finally {
    ensureGroupIdOption(filters.group_id)
    ensureGroupIdOption(form.group_id)
    rows.value.forEach(item => ensureGroupIdOption(item.group_id))
  }
}

watch(() => filters.group_id, (v) => {
  ensureGroupIdOption(v)
  loadRemain()
})

watch(
  () => route.query,
  async (q) => {
    if (syncingQuery) return
    const incoming = extractFilterQuery(q)
    const current = buildFilterQuery()
    if (queriesEqual(incoming, current)) return
    await applyQueryToFilters(q, { triggerLoad: true })
  }
)

watch(() => form.group_id, (v) => {
  ensureGroupIdOption(v)
})

const openAdd = () => {
  Object.assign(form, { id: undefined, email: '', password: '', seats: 1, status: 1, rt: '', group_id: null })
  dialog.isEdit = false
  dialog.visible = true
}

const openEdit = (row) => {
  Object.assign(form, { id: row.id, email: row.email, password: '', seats: row.seats, status: Number(row.status), rt: row.rt, group_id: row.group_id || null })
  ensureGroupIdOption(form.group_id)
  dialog.isEdit = true
  dialog.visible = true
}

const submit = async () => {
  submitting.value = true
  try {
    let url = '/_api/team'
    let method = 'POST'
    let payload
    if (dialog.isEdit) {
      method = 'PUT'
      payload = { id: form.id }
      if (form.seats) payload.seats = form.seats
      if (form.password) payload.password = form.password
      if (form.status) payload.status = form.status
      if (form.rt) payload.rt = form.rt
      if (form.group_id !== undefined) {
        payload.group_id = String(form.group_id || '').trim()
      }
    } else {
      if (!form.password) throw new Error('创建时需要填写密码')
      const { id, group_id, ...rest } = form
      payload = { ...rest }
      const gid = String(group_id || '').trim()
      if (gid) payload.group_id = gid
    }
    const res = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    ElMessage.success('保存成功')
    dialog.visible = false
    const ensuredGid = payload && typeof payload.group_id === 'string' ? payload.group_id : ''
    await load(true)
    await loadRemain(true)
    ensureGroupIdOption(ensuredGid)
    loadGroupIds()
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    submitting.value = false
  }
}

const openDetails = async (row) => {
  details.visible = true
  details.activeTab = 'users'
  details.teamId = row.id
  details.loading = true
  try {
    const qs = new URLSearchParams({ team_id: String(row.id) })
    const res = await fetch('/_api/team/external?' + qs.toString())
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = resp && typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    const members = Array.isArray(data.members) ? data.members : []
    const whitelist = new Set(members.map(m => String(m.email || '').toLowerCase()))
    const toView = (items) => (Array.isArray(items) ? items : []).map(i => ({
      id: i.id,
      email: i.email_address || i.email || '',
      role: i.role || '',
      created_time: i.created_time || '',
      isWhitelist: whitelist.has(String((i.email_address || i.email || '')).toLowerCase())
    }))
    details.usersView = toView(data.users)
    details.pendingView = toView(data.pending)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    details.loading = false
  }
}

const removeUser = async (row) => {
  try {
    if (String(row.role) === 'account-owner') return
    const payload = { team_id: details.teamId, user_id: row.id, email: row.email }
    const res = await fetch('/_api/team/external', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || '删除失败')
    ElMessage.success('已移除')
    // reload details
    openDetails({ id: details.teamId })
  } catch (e) {
    ElMessage.error(String(e))
  }
}

const removePending = async (row) => {
  try {
    if (String(row.role) === 'account-owner') return
    const payload = { team_id: details.teamId, email: row.email }
    const res = await fetch('/_api/team/external', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || '删除失败')
    ElMessage.success('已移除')
    openDetails({ id: details.teamId })
  } catch (e) {
    ElMessage.error(String(e))
  }
}

const refresh = async (row) => {
  try {
    if (!row.rt) {
      ElMessage.error('该 Owner 未配置 RT，无法刷新')
      return
    }
    const res = await fetch('/_api/team/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: row.id })
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    ElMessage.success('刷新成功')
    await load(true)
  } catch (e) {
    ElMessage.error(String(e))
  }
}

onMounted(async () => {
  await loadGroupIds()
  await applyQueryToFilters(route.query, { triggerLoad: true })
})

// format to YYYY-MM-DD HH:mm:ss
const pad = (n) => (n < 10 ? '0' + n : '' + n)
const formatTime = (val) => {
  if (!val) return ''
  // already in desired format
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(val)) return val
  let d
  // try parse epoch numbers
  if (typeof val === 'number') {
    d = val > 1e12 ? new Date(val) : new Date(val * 1000)
  } else if (/^\d{10,13}$/.test(val)) {
    const num = Number(val)
    d = num > 1e12 ? new Date(num) : new Date(num * 1000)
  } else {
    const tmp = val.replace('T', ' ').replace('Z', '')
    d = new Date(tmp)
    if (isNaN(d.getTime())) {
      d = new Date(val)
    }
  }
  if (isNaN(d.getTime())) return val
  const Y = d.getFullYear()
  const M = pad(d.getMonth() + 1)
  const D = pad(d.getDate())
  const h = pad(d.getHours())
  const m = pad(d.getMinutes())
  const s = pad(d.getSeconds())
  return `${Y}-${M}-${D} ${h}:${m}:${s}`
}

const fetchSubscriptionData = async (teamId) => {
  if (!teamId) throw new Error('缺少团队ID')
  const qs = new URLSearchParams({ team_id: String(teamId) })
  const res = await fetch('/_api/team/subscription?' + qs.toString())
  if (!res.ok) throw new Error(await res.text())
  const resp = await res.json()
  if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
  let data = (typeof resp.code !== 'undefined' ? (resp.data || {}) : resp) || {}
  if (data && data.data && typeof data.data === 'object') data = data.data
  return data
}

const postCancelSubscription = async (teamId) => {
  const res = await fetch('/_api/team/subscription/cancel', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ team_id: teamId })
  })
  if (!res.ok) throw new Error(await res.text())
  const resp = await res.json()
  if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || '取消失败')
  return resp
}

const normalizeAutoRenew = (val) => {
  if (val === true || val === false) return val
  if (typeof val === 'string') return ['true', '1', 'yes', 'on'].includes(val.toLowerCase())
  if (typeof val === 'number') return val === 1
  return false
}

const startBulkCancel = async () => {
  const currentRows = Array.isArray(rows.value) ? [...rows.value] : []
  if (!currentRows.length) {
    ElMessage.info('当前列表为空')
    return
  }
  try {
    await ElMessageBox.confirm('确认对当前列表中的所有团队执行一键退订？', '提示', {
      type: 'warning',
      confirmButtonText: '开始执行',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  bulk.visible = true
  bulk.logs = []
  bulk.running = true
  try {
    for (const item of currentRows) {
      const logEntry = {
        id: item?.id,
        email: item?.email || '',
        status: 'pending',
        message: '等待处理'
      }
      bulk.logs.push(logEntry)
      if (!item?.id) {
        logEntry.status = 'skipped'
        logEntry.message = '缺少团队ID，已跳过'
        continue
      }
      try {
        logEntry.status = 'checking'
        logEntry.message = '查询订阅状态中...'
        const subscription = await fetchSubscriptionData(item.id)
        const willRenew = normalizeAutoRenew(subscription?.will_renew)
        if (!subscription || Object.keys(subscription).length === 0) {
          logEntry.status = 'skipped'
          logEntry.message = '未查询到订阅记录'
          continue
        }
        if (!willRenew) {
          logEntry.status = 'success'
          logEntry.message = '已退订'
          continue
        }
        logEntry.status = 'cancelling'
        logEntry.message = '检测到自动续费，正在取消...'
        await postCancelSubscription(item.id)
        logEntry.status = 'success'
        logEntry.message = '取消成功'
      } catch (err) {
        logEntry.status = 'error'
        logEntry.message = `失败：${String(err?.message || err)}`
      }
    }
    ElMessage.success('批量退订已完成')
  } finally {
    bulk.running = false
  }
}

const bulkStatusText = (status) => {
  const map = {
    pending: '等待处理',
    checking: '查询中',
    cancelling: '取消中',
    success: '取消成功',
    skipped: '已跳过',
    error: '失败'
  }
  return map[status] || status
}

const bulkStatusTagType = (status) => {
  const map = {
    success: 'success',
    skipped: 'info',
    error: 'danger',
    cancelling: 'warning',
    checking: 'warning',
    pending: 'info'
  }
  return map[status] || 'info'
}

// ---------- More actions (subscription) ----------
const onMoreCommand = (cmd, row) => {
  if (cmd === 'status') return querySubscription(row)
  if (cmd === 'cancel') return cancelSubscription(row)
  if (cmd === 'delete') return deleteTeam(row)
}

const querySubscription = async (row) => {
  sub.visible = true
  sub.loading = true
  sub.teamId = row?.id
  sub.showRaw = false
  // prevent flashing of previous data while loading
  sub.data = {}
  try {
    sub.data = await fetchSubscriptionData(row.id)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    sub.loading = false
  }
}

const cancelSubscription = async (row) => {
  try {
    if (!row?.id) return
    try {
      await ElMessageBox.confirm('确认取消该团队的订阅？', '提示', {
        type: 'warning',
        confirmButtonText: '确认',
        cancelButtonText: '取消'
      })
    } catch {
      return
    }
    await postCancelSubscription(row.id)
    ElMessage.success('已提交取消订阅')
  } catch (e) {
    ElMessage.error(String(e))
  }
}

const deleteTeam = async (row) => {
  try {
    if (!row?.id) return
    try {
      await ElMessageBox.confirm('确认删除该团队？删除后将清除所有关联成员，且不可恢复。', '提示', {
        type: 'warning',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      })
    } catch {
      return
    }
    const res = await fetch('/_api/team', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: row.id })
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || '删除失败')
    ElMessage.success('删除成功')
    await load(true)
    await loadRemain(true)
  } catch (e) {
    ElMessage.error(String(e))
  }
}

const prettyJSON = (obj) => {
  try { return JSON.stringify(obj || {}, null, 2) } catch { return String(obj) }
}

const planTypeText = (t) => {
  const m = { team: '团队', plus: 'Plus', pro: 'Pro', enterprise: '企业', free: '免费' }
  return m[t] || (t || '-')
}

const billingPeriodText = (t) => {
  const m = { monthly: '月付', yearly: '年付', annual: '年付' }
  return m[t] || (t || '-')
}

// clear details dialog data on close
const onDetailsClosed = () => {
  details.loading = false
  details.teamId = undefined
  details.activeTab = 'users'
  details.usersView = []
  details.pendingView = []
}
</script>

<style scoped>
.team-owner-list {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.card-header .title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.remain-tag {
  margin-left: 8px;
}

.filters {
  margin-bottom: 12px;
}

.table-pagination {
  margin-top: 12px;
  text-align: right;
}

.details-tabs :deep(.el-tabs__header) {
  margin-bottom: 8px;
}

.details-container {
  min-height: 240px;
}

.bulk-result-body {
  max-height: 360px;
  overflow: auto;
}
</style>
