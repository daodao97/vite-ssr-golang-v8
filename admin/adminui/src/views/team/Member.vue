<template>
  <div class="team-member-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          成员管理
          <div class="actions">
            <el-button size="small" @click="openFetch">批量取Team号</el-button>
            <el-button size="small" @click="openBulk">批量添加</el-button>
            <el-button type="primary" size="small" @click="openAdd">添加</el-button>
          </div>
        </div>
      </template>

      <!-- Filters -->
      <el-form :inline="true" :model="filters" class="filters" @submit.prevent>
        <el-form-item label="Team">
          <el-select v-model="filters.team_id" clearable filterable remote :remote-method="remoteSearchTeams" :loading="teamLoading" placeholder="搜索邮箱关键字" style="width: 260px" @change="v => filters.team_id = v === '' || v == null ? undefined : Number(v)">
            <el-option v-for="t in teamOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="filters.email" placeholder="精确匹配" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width: 140px" @change="v => filters.status = v === '' || v == null ? null : Number(v)">
            <el-option :value="1" label="已邀请(1)" />
            <el-option :value="2" label="已进入(2)" />
            <el-option :value="3" label="离开(3)" />
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
        <el-table-column prop="team_id" label="Team ID" width="100" />
        <el-table-column prop="uid" label="UID" width="100" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 2 ? 'success' : (Number(row.status)===1?'info':'warning')">
              {{ Number(row.status) === 1 ? '已邀请(1)' : (Number(row.status) === 2 ? '已进入(2)' : '离开(3)') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="180">
          <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
        </el-table-column>
        <el-table-column prop="end_time" label="结束时间" width="180">
          <template #default="{ row }">{{ formatTime(row.end_time) }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link size="small" :loading="row.__forcing === true" :disabled="!row.email" @click="forceInvite(row)">强制邀请</el-button>
            <el-button link size="small" @click="openEdit(row)">编辑</el-button>
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

    <!-- Dialog: Bulk Fetch Members -->
    <el-dialog v-model="fetchDialog.visible" title="批量取Team号" width="720px">
      <el-form label-width="120px" :model="fetchDialog" style="margin-bottom: 12px">
        <el-form-item label="取号个数">
          <el-input-number v-model="fetchDialog.count" :min="1" :max="200" :disabled="fetchDialog.busy" />
        </el-form-item>
        <el-form-item label="分组ID">
          <el-select
            v-model="fetchDialog.group_id"
            clearable
            filterable
            allow-create
            placeholder="默认 default"
            style="width: 260px"
            :disabled="fetchDialog.busy"
          >
            <el-option v-for="gid in groupIdOptions" :key="gid" :label="gid" :value="gid" />
          </el-select>
        </el-form-item>
        <el-form-item label="Team">
          <el-select
            v-model="fetchDialog.team_id"
            clearable
            filterable
            remote
            :remote-method="remoteSearchTeams"
            :loading="teamLoading"
            placeholder="未选择时自动分配"
            style="width: 260px"
            :disabled="fetchDialog.busy"
            @change="v => fetchDialog.team_id = v === '' || v == null ? undefined : Number(v)"
          >
            <el-option v-for="t in teamOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
      </el-form>

      <div class="fetch-actions">
        <el-button
          type="primary"
          size="small"
          :loading="fetchDialog.busy"
          :disabled="fetchDialog.busy"
          @click="startFetch"
        >
          开始取号
        </el-button>
        <span class="auto-export-tip">导出内容实时生成</span>
        <el-tag size="small" type="info">已成功 {{ fetchSuccessCount }} 条</el-tag>
        <el-tag size="small" v-if="fetchErrorCount" type="danger">
          失败 {{ fetchErrorCount }} 条
        </el-tag>
      </div>

      <el-table
        :data="fetchDialog.results"
        size="small"
        :border="true"
        :empty-text="fetchDialog.busy ? '取号中...' : '暂无数据'"
        v-loading="fetchDialog.busy"
        style="margin-bottom: 12px"
      >
        <el-table-column type="index" width="60" label="#" />
        <el-table-column label="邮箱" min-width="220">
          <template #default="{ row }">{{ row.email || row.response?.mail || '-' }}</template>
        </el-table-column>
        <el-table-column label="GPT 密码" min-width="160">
          <template #default="{ row }">{{ row.gpt_pwd || row.response?.gpt_pwd || '-' }}</template>
        </el-table-column>
        <el-table-column label="邮箱密码" min-width="160">
          <template #default="{ row }">{{ row.mail_pwd || row.response?.mail_pwd || '-' }}</template>
        </el-table-column>
        <el-table-column label="RT" min-width="240">
          <template #default="{ row }">{{ shorten(row.rt || row.response?.rt) || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="bulkStatusTag(row.status)">
              {{ bulkStatusLabel(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="返回信息" min-width="200">
          <template #default="{ row }">
            <span v-if="row.status === 'success'">
              ID: {{ row.response?.id }}
              <span v-if="row.response?.mail">｜{{ row.response.mail }}</span>
              <span v-if="row.response?.gpt_pwd">｜GPT: {{ row.response.gpt_pwd }}</span>
              <span v-if="row.response?.mail_pwd">｜邮箱密码: {{ row.response.mail_pwd }}</span>
              <span v-if="row.response?.rt">｜RT: {{ shorten(row.response.rt) }}</span>
            </span>
            <span v-else-if="row.status === 'error'">{{ row.error }}</span>
            <span v-else-if="row.status === 'running'">执行中...</span>
            <span v-else>待处理</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="fetchExportText" class="export-box">
        <el-form label-width="120px">
          <el-form-item label="导出结果">
            <el-input :model-value="fetchExportText" type="textarea" :rows="6" readonly />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <el-button @click="fetchDialog.visible = false" :disabled="fetchDialog.busy">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Bulk Add Members -->
    <el-dialog v-model="bulkDialog.visible" title="批量添加成员" width="860px">
      <el-form label-width="120px" style="margin-bottom: 12px">
        <el-form-item label="Team">
          <el-select v-model="bulkDialog.team_id" clearable filterable remote :remote-method="remoteSearchTeams" :loading="teamLoading" placeholder="未选择时自动分配">
            <el-option v-for="t in teamOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="成员数据">
          <el-input
            v-model="bulkDialog.raw"
            type="textarea"
            :rows="6"
            placeholder="每行一条，支持“---”或“----”分隔。例如：\nuser1@example.com---gptPass1---mailPass1---rt_xxx\nuser2@example.com---gptPass2---mailPass2---rt_yyy"
          />
        </el-form-item>
      </el-form>

      <div class="parse-actions">
        <el-space>
          <el-tag>共 {{ bulkDialog.stats.total }} 行</el-tag>
          <el-tag type="success">有效 {{ bulkDialog.stats.ok }}</el-tag>
          <el-tag type="danger" v-if="bulkDialog.failures.length">失败 {{ bulkDialog.stats.fail }}</el-tag>
          <el-button size="small" @click="parseBulk">解析</el-button>
        </el-space>
        <el-button size="small" style="margin-left: auto" :disabled="!bulkDialog.rows.length || bulkDialog.busy" :loading="bulkDialog.busy" type="primary" @click="startBulkAdd">开始添加</el-button>
        <el-button size="small" :disabled="!hasPendingBulk || bulkDialog.busy" :loading="bulkDialog.busy" @click="continueBulkAdd">继续</el-button>
      </div>

      <el-alert
        v-if="bulkDialog.failures.length"
        type="warning"
        show-icon
        :closable="false"
        class="mb8"
        title="以下行解析失败，已忽略："
      />
      <el-table v-if="bulkDialog.failures.length" :data="bulkDialog.failures" size="small" :border="true" style="margin-bottom: 12px">
        <el-table-column prop="lineNo" label="#" width="70" />
        <el-table-column prop="raw" label="原始数据" min-width="260" />
        <el-table-column prop="reason" label="原因" min-width="160" />
      </el-table>

      <el-table :data="bulkDialog.rows" size="small" :border="true" :empty-text="bulkDialog.stats.total ? '无有效数据' : '请先解析'" v-loading="bulkDialog.busy">
        <el-table-column type="index" width="60" label="#" />
        <el-table-column label="邮箱" min-width="240">
          <template #default="{ row }">
            <el-input v-model="row.email" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="GPT 密码" min-width="160">
          <template #default="{ row }">
            <el-input v-model="row.gpt_pwd" size="small" placeholder="可选" />
          </template>
        </el-table-column>
        <el-table-column label="邮箱密码" min-width="160">
          <template #default="{ row }">
            <el-input v-model="row.mail_pwd" size="small" placeholder="可选" />
          </template>
        </el-table-column>
        <el-table-column label="RT" min-width="260">
          <template #default="{ row }">
            <el-input
              v-model="row.rt"
              size="small"
              type="textarea"
              :autosize="{ minRows: 1, maxRows: 3 }"
              placeholder="可选"
              style="font-family: monospace"
            />
          </template>
        </el-table-column>
        <el-table-column label="激活时长" width="140">
          <template #default="{ row }">
            <el-select v-model="row.plan" size="small" placeholder="1个月">
              <el-option label="1个月" value="1m" />
              <el-option label="2个月" value="2m" />
              <el-option label="3个月" value="3m" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="bulkStatusTag(row.status)">
              {{ bulkStatusLabel(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="返回信息" min-width="200">
          <template #default="{ row }">
            <span v-if="row.status === 'success'">
              ID: {{ row.response?.id }}
              <span v-if="row.response?.mail">｜{{ row.response.mail }}</span>
              <span v-if="row.response?.gpt_pwd">｜GPT: {{ row.response.gpt_pwd }}</span>
              <span v-if="row.response?.mail_pwd">｜邮箱密码: {{ row.response.mail_pwd }}</span>
              <span v-if="row.response?.rt">｜RT: {{ shorten(row.response.rt) }}</span>
            </span>
            <span v-else-if="row.status === 'error'">{{ row.error }}</span>
            <span v-else-if="row.status === 'running'">执行中...</span>
            <span v-else>待处理</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link size="small" :disabled="row.status === 'success' || bulkDialog.busy" :loading="row.busy" @click="retryBulkRow(row)">
              重试
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="bulkDialog.visible = false" :disabled="bulkDialog.busy">关闭</el-button>
        <el-button @click="parseBulk" :disabled="bulkDialog.busy">重新解析</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Add/Edit Member -->
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑成员' : '添加成员'" width="560px">
      <el-form :model="form" label-width="120px">
        <el-form-item v-if="dialog.isEdit" label="ID">
          <el-input v-model.number="form.id" disabled />
        </el-form-item>
        <el-form-item label="Team">
          <el-select v-model="form.team_id" :disabled="dialog.isEdit" clearable filterable remote :remote-method="remoteSearchTeams" :loading="teamLoading" placeholder="搜索邮箱关键字">
            <el-option v-for="t in teamOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" :disabled="dialog.isEdit" />
        </el-form-item>
        <el-form-item label="UID">
          <el-input v-model="form.uid" placeholder="可选，数字" />
        </el-form-item>
        <el-form-item v-if="!dialog.isEdit" label="时间计划">
          <el-select v-model="form.plan" clearable placeholder="可选">
            <el-option label="1个月" value="1m" />
            <el-option label="2个月" value="2m" />
            <el-option label="3个月" value="3m" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const submitting = ref(false)
const filters = reactive({ team_id: undefined, email: '', status: null })
const hasQueried = ref(false)
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const rows = ref([])
const teamOptions = ref([])
const teamLoading = ref(false)
const groupIdOptions = ref([])
const route = useRoute()
const router = useRouter()

const dialog = reactive({ visible: false, isEdit: false })
const form = reactive({ id: undefined, team_id: undefined, email: '', uid: '', status: 2, plan: '' })

const splitRe = /-{3,}/
const emailRe = /^[^@\s]+@[^@\s]+\.[^@\s]+$/

const trimStr = (val) => String(val ?? '').trim()

const bulkDialog = reactive({
  visible: false,
  team_id: undefined,
  raw: '',
  rows: [],
  failures: [],
  stats: { total: 0, ok: 0, fail: 0 },
  busy: false,
})

const fetchDialog = reactive({
  visible: false,
  count: 1,
  group_id: '',
  team_id: undefined,
  busy: false,
  results: [],
})

const hasPendingBulk = computed(() => bulkDialog.rows.some(r => r.status === 'pending'))
const fetchSuccessCount = computed(() => fetchDialog.results.reduce((acc, row) => acc + (row?.status === 'success' ? 1 : 0), 0))
const fetchErrorCount = computed(() => fetchDialog.results.reduce((acc, row) => acc + (row?.status === 'error' ? 1 : 0), 0))
const fetchExportText = computed(() => {
  const lines = []
  for (const row of fetchDialog.results) {
    if (!row || row.status !== 'success') continue
    const email = trimStr(row.email || row.response?.mail || row.response?.email || '')
    const gpt = trimStr(row.gpt_pwd || row.response?.gpt_pwd)
    const mailPwd = trimStr(row.mail_pwd || row.response?.mail_pwd)
    const rt = trimStr(row.rt || row.response?.rt)
    lines.push(`${email}---${gpt}---${mailPwd}---${rt}`)
  }
  return lines.length ? lines.join('\n') : ''
})

const filterQueryKeys = ['team_id', 'email', 'group_id', 'status', 'queried']
let syncingQuery = false

const toSingleValue = (val) => (Array.isArray(val) ? val[val.length - 1] : val)

const normalizeTeamId = (val) => {
  const raw = toSingleValue(val)
  if (raw == null || raw === '') return undefined
  const num = Number(raw)
  return Number.isFinite(num) ? num : undefined
}

const normalizeStatus = (val) => {
  const raw = toSingleValue(val)
  if (raw == null || raw === '') return null
  const num = Number(raw)
  return Number.isFinite(num) ? num : null
}

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
  if (filters.team_id !== undefined && filters.team_id !== null && filters.team_id !== '') {
    query.team_id = String(filters.team_id)
  }
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
  const nextTeamId = normalizeTeamId(query.team_id)
  const nextEmail = normalizeEmail(query.email)
  const nextGroupId = normalizeGroupId(query.group_id)
  const nextStatus = normalizeStatus(query.status)
  const queryHasQueried = normalizeQueried(query.queried)
  const nextHasQueried = queryHasQueried || nextTeamId !== undefined || !!nextEmail || nextGroupId !== null || nextStatus !== null

  if (filters.team_id !== nextTeamId) filters.team_id = nextTeamId
  if (filters.email !== nextEmail) filters.email = nextEmail
  if (filters.group_id !== nextGroupId) filters.group_id = nextGroupId
  if (filters.status !== nextStatus) filters.status = nextStatus

  hasQueried.value = nextHasQueried

  if (!nextHasQueried) {
    if (triggerLoad) {
      rows.value = []
      pagination.total = 0
      pagination.page = 1
    }
    return
  }

  if (triggerLoad) {
    pagination.page = 1
    await load(true)
  }
}

const resetFilters = () => { filters.team_id = undefined; filters.email = ''; filters.group_id = null; filters.status = null }
const onSearch = async () => {
  hasQueried.value = true
  pagination.page = 1
  await load(true)
}
const onReset = async () => {
  resetFilters()
  pagination.page = 1
  hasQueried.value = false
  rows.value = []
  pagination.total = 0
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

const openBulk = () => {
  bulkDialog.visible = true
  bulkDialog.team_id = filters.team_id || undefined
}

const openFetch = () => {
  fetchDialog.visible = true
  fetchDialog.count = 1
  fetchDialog.results = []
  fetchDialog.team_id = filters.team_id || bulkDialog.team_id || undefined
  const gid = String(filters.group_id || '').trim()
  if (gid) {
    fetchDialog.group_id = gid
    ensureGroupIdOption(gid)
  } else {
    fetchDialog.group_id = ''
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
    // ignore group id load errors
  } finally {
    ensureGroupIdOption(filters.group_id)
    ensureGroupIdOption(fetchDialog.group_id)
  }
}

const startFetch = async () => {
  if (fetchDialog.busy) return
  let count = Number(fetchDialog.count)
  if (!Number.isFinite(count) || count <= 0) {
    ElMessage.warning('取号个数需为正整数')
    return
  }
  count = Math.min(200, Math.floor(count))
  fetchDialog.count = count
  const gid = String(fetchDialog.group_id || '').trim()
  if (gid) ensureGroupIdOption(gid)
  fetchDialog.results = Array.from({ length: count }, () => ({
    status: 'pending',
    email: '',
    gpt_pwd: '',
    mail_pwd: '',
    rt: '',
    response: null,
    error: '',
  }))
  fetchDialog.busy = true
  let successCount = 0
  try {
    for (const row of fetchDialog.results) {
      row.status = 'running'
      row.error = ''
      row.response = null
      try {
        const payload = {}
        if (fetchDialog.team_id) payload.team_id = Number(fetchDialog.team_id)
        if (gid) payload.group_id = gid
        const res = await fetch('/_api/team/member', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error(await res.text())
        const resp = await res.json()
        if (typeof resp.code !== 'undefined' && resp.code !== 0) {
          throw new Error(resp.message || `业务错误(code=${resp.code})`)
        }
        const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
        row.status = 'success'
        row.response = data
        row.email = trimStr(data.mail || data.email || '')
        row.gpt_pwd = trimStr(data.gpt_pwd)
        row.mail_pwd = trimStr(data.mail_pwd)
        row.rt = trimStr(data.rt)
        successCount++
      } catch (e) {
        row.status = 'error'
        row.error = String(e?.message || e)
      }
    }
    const failCount = fetchDialog.results.filter(r => r.status === 'error').length
    if (successCount && !failCount) {
      ElMessage.success('批量取号完成')
    } else if (successCount && failCount) {
      ElMessage.warning('部分取号失败，请查看列表')
    } else if (!successCount) {
      ElMessage.error('取号全部失败')
    }
    if (successCount) {
      await load(true)
    }
  } finally {
    fetchDialog.busy = false
  }
}

const isPlanToken = (val) => {
  const raw = String(val || '').trim().toLowerCase()
  if (!raw) return false
  return ['1m', '1', '1month', '1个月', '1月', '30d', '2m', '2', '2month', '2个月', '2月', '60d', '3m', '3', '3month', '3个月', '3月', '90d'].includes(raw)
}

const normalizePlan = (val) => {
  const raw = String(val || '').trim().toLowerCase()
  if (!raw) return '1m'
  if (['1m', '1', '1month', '1个月', '1月', '30d'].includes(raw)) return '1m'
  if (['2m', '2', '2month', '2个月', '2月', '60d'].includes(raw)) return '2m'
  if (['3m', '3', '3month', '3个月', '3月', '90d'].includes(raw)) return '3m'
  return '1m'
}

const shorten = (s) => {
  if (!s) return ''
  const str = String(s)
  if (str.length <= 28) return str
  return str.slice(0, 12) + '…' + str.slice(-8)
}

const parseBulk = () => {
  const lines = String(bulkDialog.raw || '').split(/\r?\n/)
  const ok = []
  const bad = []
  let lineNo = 0
  for (const line of lines) {
    lineNo++
    const raw = String(line || '').trim()
    if (!raw) continue
    const parts = raw.split(splitRe).map(s => s.trim()).filter(Boolean)
    if (!parts.length) {
      bad.push({ lineNo, raw, reason: '分隔符格式错误' })
      continue
    }
    const segments = [...parts]
    let plan = ''
    if (segments.length >= 2 && isPlanToken(segments[segments.length - 1])) {
      plan = normalizePlan(segments.pop())
    }
    const [email, gptRaw, mailRaw, rtRaw, ...rest] = segments
    if (!emailRe.test(email)) {
      bad.push({ lineNo, raw, reason: '邮箱格式错误' })
      continue
    }
    const gptPwd = trimStr(gptRaw)
    const mailPwd = trimStr(mailRaw)
    const rt = trimStr([rtRaw, ...rest].filter(Boolean).join('---'))
    const finalPlan = plan || '1m'
    ok.push({
      lineNo,
      raw,
      email,
      gpt_pwd: gptPwd,
      mail_pwd: mailPwd,
      rt,
      plan: finalPlan,
      status: 'pending',
      error: '',
      busy: false,
      response: null,
    })
  }
  bulkDialog.rows = ok
  bulkDialog.failures = bad
  const nonEmpty = lines.filter(l => String(l || '').trim()).length
  Object.assign(bulkDialog.stats, { total: nonEmpty, ok: ok.length, fail: bad.length })
}

const bulkStatusTag = (status) => {
  if (status === 'success') return 'success'
  if (status === 'error') return 'danger'
  if (status === 'running') return 'info'
  return 'warning'
}

const bulkStatusLabel = (row) => {
  switch (row.status) {
    case 'success': return '成功'
    case 'error': return '失败'
    case 'running': return '执行中'
    case 'pending':
    default:
      return '待处理'
  }
}

const ensureBulkRowValid = (row) => {
  const email = String(row.email || '').trim()
  if (!emailRe.test(email)) {
    row.status = 'error'
    row.error = '邮箱格式错误'
    return false
  }
  row.email = email
  if (!['1m', '2m', '3m'].includes(row.plan)) {
    row.plan = normalizePlan(row.plan)
  }
  if (!row.plan) {
    row.plan = '1m'
  }
  row.gpt_pwd = trimStr(row.gpt_pwd)
  row.mail_pwd = trimStr(row.mail_pwd)
  row.rt = trimStr(row.rt)
  return true
}

const callBulkAdd = async (row) => {
  if (!ensureBulkRowValid(row)) {
    throw new Error(row.error || '数据校验失败')
  }
  if (!row.plan) row.plan = '1m'
  row.busy = true
  row.status = 'running'
  row.error = ''
  row.response = null
  try {
    const payload = {
      team_id: bulkDialog.team_id ? Number(bulkDialog.team_id) : undefined,
      email: row.email,
    }
    if (row.gpt_pwd) {
      payload.gpt_pwd = row.gpt_pwd
      payload.password = row.gpt_pwd
    }
    if (row.mail_pwd) payload.mail_pwd = row.mail_pwd
    if (row.rt) payload.rt = row.rt
    payload.plan = row.plan || '1m'
    const res = await fetch('/_api/team/member', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) {
      throw new Error(resp.message || `业务错误(code=${resp.code})`)
    }
    const data = typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    row.status = 'success'
    row.response = data
  } catch (e) {
    row.status = 'error'
    row.error = String(e?.message || e)
    throw e
  } finally {
    row.busy = false
  }
}

const startBulkAdd = async () => {
  if (bulkDialog.busy) return
  if (!bulkDialog.rows.length) {
    ElMessage.warning('请先解析有效数据')
    return
  }
  bulkDialog.busy = true
  try {
    for (const row of bulkDialog.rows) {
      if (row.status !== 'success') {
        row.status = 'pending'
        row.error = ''
        row.response = null
      }
    }
    for (const row of bulkDialog.rows) {
      if (row.status !== 'pending') continue
      try {
        await callBulkAdd(row)
      } catch (e) {
        ElMessage.error(`第 ${row.lineNo} 行失败：${row.error}`)
        break
      }
    }
  } finally {
    bulkDialog.busy = false
    await load(true)
  }
}

const continueBulkAdd = async () => {
  if (bulkDialog.busy) return
  const queue = bulkDialog.rows.filter(r => r.status === 'pending')
  if (!queue.length) {
    ElMessage.info('没有待处理的记录')
    return
  }
  bulkDialog.busy = true
  try {
    for (const row of queue) {
      try {
        await callBulkAdd(row)
      } catch (e) {
        ElMessage.error(`第 ${row.lineNo} 行失败：${row.error}`)
        break
      }
    }
  } finally {
    bulkDialog.busy = false
    await load(true)
  }
}

const retryBulkRow = async (row) => {
  if (bulkDialog.busy) return
  row.status = 'pending'
  row.error = ''
  row.response = null
  bulkDialog.busy = true
  try {
    await callBulkAdd(row)
    await load(true)
  } catch (e) {
    ElMessage.error(`第 ${row.lineNo} 行失败：${row.error}`)
  } finally {
    bulkDialog.busy = false
  }
}

const load = async (force = false) => {
  if (force) hasQueried.value = true
  if (!force && !hasQueried.value) return
  loading.value = true
  try {
    const qs = new URLSearchParams()
    if (filters.team_id) qs.set('team_id', String(filters.team_id))
    if (filters.email) qs.set('email', filters.email)
    const gid = String(filters.group_id || '').trim()
    if (gid) qs.set('group_id', gid)
    if (filters.status != null && filters.status !== '') qs.set('status', String(filters.status))
    qs.set('page', String(pagination.page))
    qs.set('page_size', String(pagination.pageSize))
    const res = await fetch('/_api/team/members' + (qs.toString() ? ('?' + qs.toString()) : ''))
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = resp && typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    rows.value = data.list || []
    rows.value.forEach(item => ensureGroupIdOption(item?.group_id))
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

const openAdd = () => {
  Object.assign(form, { id: undefined, team_id: filters.team_id || undefined, email: '', uid: '', status: 2, plan: '' })
  dialog.isEdit = false
  dialog.visible = true
}

const openEdit = (row) => {
  Object.assign(form, { id: row.id, team_id: row.team_id, email: row.email, uid: String(row.uid ?? ''), status: Number(row.status) })
  dialog.isEdit = true
  dialog.visible = true
}

const submit = async () => {
  submitting.value = true
  try {
    if (!dialog.isEdit) {
      const payload = { team_id: form.team_id, email: form.email }
      if (!payload.email) throw new Error('email 必填')
      const uidNum = Number(form.uid)
      if (!Number.isNaN(uidNum) && Number.isFinite(uidNum) && uidNum > 0) payload.uid = uidNum
      if (form.plan) payload.plan = form.plan
      const res = await fetch('/_api/team/member', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
      if (!res.ok) throw new Error(await res.text())
      const resp = await res.json()
      if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
      ElMessage.success('添加成功')
    } else {
      const payload = { id: form.id }
      const uidNum = Number(form.uid)
      if (!Number.isNaN(uidNum) && Number.isFinite(uidNum)) payload.uid = uidNum
      payload.status = form.status
      const res = await fetch('/_api/team/member', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
      if (!res.ok) throw new Error(await res.text())
      const resp = await res.json()
      if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
      ElMessage.success('更新成功')
    }
    dialog.visible = false
    await load(true)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    submitting.value = false
  }
}

const forceInvite = async (row) => {
  if (!row || !row.email) {
    ElMessage.error('当前记录缺少邮箱，无法邀请')
    return
  }
  const teamId = Number(row.team_id)
  if (!teamId) {
    ElMessage.error('当前记录缺少 Team ID，无法邀请')
    return
  }
  try {
    await ElMessageBox.confirm(`确认强制邀请 ${row.email} 到团队 ${teamId} 吗？`, '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch (e) {
    return
  }
  row.__forcing = true
  try {
    const payload = { team_id: teamId, email: row.email, force: true }
    const uidNum = Number(row.uid)
    if (!Number.isNaN(uidNum) && Number.isFinite(uidNum) && uidNum > 0) {
      payload.uid = uidNum
    }
    const res = await fetch('/_api/team/member', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) {
      throw new Error(resp.message || `业务错误(code=${resp.code})`)
    }
    ElMessage.success('强制邀请成功')
    await load(true)
  } catch (e) {
    ElMessage.error(String(e))
  } finally {
    row.__forcing = false
  }
}

// remote team search for selects
const remoteSearchTeams = async (query) => {
  teamLoading.value = true
  try {
    const qs = new URLSearchParams()
    if (query) qs.set('email', query)
    qs.set('page', '1')
    qs.set('page_size', '50')
    const res = await fetch('/_api/teams' + (qs.toString() ? ('?' + qs.toString()) : ''))
    if (!res.ok) throw new Error(await res.text())
    const resp = await res.json()
    if (typeof resp.code !== 'undefined' && resp.code !== 0) throw new Error(resp.message || `业务错误(code=${resp.code})`)
    const data = resp && typeof resp.code !== 'undefined' ? (resp.data || {}) : resp
    teamOptions.value = (data.list || []).map(i => ({ value: Number(i.id), label: i.email }))
    // ensure current selection exists in options so label can render
    const ensureId = filters.team_id || form.team_id || bulkDialog.team_id || fetchDialog.team_id
    if (ensureId && !teamOptions.value.some(o => Number(o.value) === Number(ensureId))) {
      // fetch by exact email not available; inject placeholder option
      // backend list already contains id, so if not present we'll add a minimal label
      teamOptions.value.unshift({ value: Number(ensureId), label: `#${ensureId}` })
    }
  } catch (e) {
    // silent fail is fine for typeahead
  } finally {
    teamLoading.value = false
  }
}

// prefill options on first render
onMounted(async () => {
  remoteSearchTeams('')
  await loadGroupIds()
  await applyQueryToFilters(route.query, { triggerLoad: true })
})

watch(() => filters.group_id, (v) => {
  ensureGroupIdOption(v)
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

watch(() => fetchDialog.group_id, (v) => {
  ensureGroupIdOption(v)
})

// format to YYYY-MM-DD HH:mm:ss (same as Team.vue)
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
    const tmp = String(val).replace('T', ' ').replace('Z', '')
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

</script>

<style scoped>
.team-member-page { padding: 16px; }
.card-header { font-weight: 600; }
.tips { color: #666; font-size: 12px; margin-top: 8px; }
.parse-actions { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.table-pagination { margin-top: 12px; text-align: right; }
.fetch-actions { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; flex-wrap: wrap; }
.auto-export-tip { font-size: 12px; color: #666; }
.export-box { margin-top: 12px; }
.export-box :deep(.el-textarea__inner) { font-family: monospace; }
</style>
