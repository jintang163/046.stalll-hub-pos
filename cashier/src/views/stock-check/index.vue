<template>
  <div class="stock-check-page">
    <header class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="goBack" v-if="currentView !== 'list'">返回</el-button>
        <h1 class="page-title">
          {{ currentView === 'list' ? '库存盘点' : (currentView === 'detail' ? '盘点详情' : '差异报告') }}
        </h1>
        <el-tag :type="networkStatus ? 'success' : 'warning'" size="default" class="net-tag">
          <el-icon><component :is="networkStatus ? 'Connection' : 'Warning'" /></el-icon>
          {{ networkStatus ? '在线' : '离线' }}
        </el-tag>
      </div>
      <div class="header-right">
        <template v-if="currentView === 'list'">
          <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
            新建盘点
          </el-button>
          <el-button :icon="Refresh" @click="loadCheckList">
            刷新
          </el-button>
        </template>
        <template v-else-if="currentView === 'detail' && currentCheck">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索SKU编码/商品名"
            style="width: 240px"
            clearable
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
          <el-button :icon="Upload" @click="uploadCheck" :disabled="!networkStatus || currentCheck.status === 2">
            上传云端
          </el-button>
          <el-button type="success" :icon="Check" @click="completeCheck" :disabled="currentCheck.status === 2">
            完成盘点
          </el-button>
          <el-button type="primary" :icon="DataAnalysis" @click="showDiff">
            差异报告
          </el-button>
        </template>
      </div>
    </header>

    <div class="page-body" v-if="currentView === 'list'">
      <div class="check-list">
        <div
          v-for="check in checkList"
          :key="check.id"
          class="check-card"
          @click="openCheckDetail(check)"
        >
          <div class="check-header">
            <span class="check-title">{{ check.title }}</span>
            <el-tag :type="getStatusType(check.status)" size="small">
              {{ getStatusText(check.status) }}
            </el-tag>
          </div>
          <div class="check-no">{{ check.check_no }}</div>
          <div class="check-stats">
            <div class="stat-item">
              <span class="stat-label">SKU总数</span>
              <span class="stat-value">{{ check.total_sku }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">已盘</span>
              <span class="stat-value">{{ check.checked_sku }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">差异数</span>
              <span class="stat-value diff" :class="{ profit: check.total_diff_qty > 0, loss: check.total_diff_qty < 0 }">
                {{ check.total_diff_qty > 0 ? '+' : '' }}{{ check.total_diff_qty }}
              </span>
            </div>
          </div>
          <div class="check-footer">
            <span>操作人: {{ check.operator_name || '-' }}</span>
            <span>{{ formatDate(check.created_at) }}</span>
          </div>
        </div>

        <el-empty v-if="checkList.length === 0 && !loading" description="暂无盘点单" />
      </div>
    </div>

    <div class="page-body detail-body" v-else-if="currentView === 'detail' && currentCheck">
      <div class="detail-sidebar">
        <div class="stats-card">
          <h3>盘点统计</h3>
          <div class="stats-grid">
            <div class="stat-box">
              <div class="stat-num">{{ currentCheck.total_sku }}</div>
              <div class="stat-label">SKU总数</div>
            </div>
            <div class="stat-box success">
              <div class="stat-num">{{ currentCheck.checked_sku }}</div>
              <div class="stat-label">已盘数量</div>
            </div>
            <div class="stat-box info">
              <div class="stat-num">{{ progressPercent }}%</div>
              <div class="stat-label">完成进度</div>
            </div>
          </div>
          <el-progress :percentage="progressPercent" :status="progressPercent === 100 ? 'success' : ''" />
        </div>

        <div class="stats-card">
          <h3>差异统计</h3>
          <div class="stats-grid">
            <div class="stat-box warning">
              <div class="stat-num">{{ diffStats.profit_count || 0 }}</div>
              <div class="stat-label">盘盈SKU</div>
            </div>
            <div class="stat-box danger">
              <div class="stat-num">{{ diffStats.loss_count || 0 }}</div>
              <div class="stat-label">盘亏SKU</div>
            </div>
            <div class="stat-box">
              <div class="stat-num normal">{{ diffStats.normal_count || 0 }}</div>
              <div class="stat-label">无差异</div>
            </div>
          </div>
          <div class="diff-summary">
            <span>差异数量: </span>
            <span :class="{ profit: (diffStats.total_diff_qty || 0) > 0, loss: (diffStats.total_diff_qty || 0) < 0 }">
              {{ (diffStats.total_diff_qty || 0) > 0 ? '+' : '' }}{{ diffStats.total_diff_qty || 0 }}
            </span>
          </div>
          <div class="diff-summary">
            <span>差异金额: </span>
            <span :class="{ profit: (diffStats.total_diff_amount || 0) > 0, loss: (diffStats.total_diff_amount || 0) < 0 }">
              ¥{{ (diffStats.total_diff_amount || 0).toFixed(2) }}
            </span>
          </div>
        </div>

        <div class="scan-card">
          <h3>扫码录入</h3>
          <el-input
            ref="scanInputRef"
            v-model="scanInput"
            placeholder="请扫描条码或输入SKU编码"
            size="large"
            @keyup.enter="onScanSubmit"
          >
            <template #append>
              <el-button @click="onScanSubmit">确认</el-button>
            </template>
          </el-input>
          <el-input-number
            v-model="scanQuantity"
            :min="0"
            size="large"
            style="width: 100%; margin-top: 12px"
            placeholder="盘点数量"
            controls-position="right"
          />
          <p class="scan-tip">
            💡 扫码枪扫描条码后，按回车确认，自动增减库存
          </p>
        </div>
      </div>

      <div class="detail-main">
        <div class="filter-bar">
          <el-radio-group v-model="filterStatus" size="default" @change="loadItems">
            <el-radio-button :value="-1">全部</el-radio-button>
            <el-radio-button :value="0">未盘</el-radio-button>
            <el-radio-button :value="1">已盘</el-radio-button>
          </el-radio-group>
          <span class="item-count">共 {{ filteredItems.length }} 条</span>
        </div>

        <div class="items-list" v-loading="loading">
          <div
            v-for="item in filteredItems"
            :key="item.id"
            class="item-row"
            :class="{ checked: item.status === 1, 'has-diff': item.status === 1 && item.diff_qty !== 0 }"
          >
            <div class="item-info">
              <div class="item-name">{{ item.product_name }}
                <span class="item-spec" v-if="item.spec_name"> / {{ item.spec_name }}</span>
              </div>
              <div class="item-meta">
                <span class="sku-code">{{ item.sku_code }}</span>
                <span v-if="item.category_name">{{ item.category_name }}</span>
              </div>
            </div>
            <div class="item-stocks">
              <div class="stock-col">
                <span class="stock-label">系统库存</span>
                <span class="stock-num system">{{ item.system_stock }}</span>
              </div>
              <div class="stock-divider">=</div>
              <div class="stock-col">
                <span class="stock-label">实盘数量</span>
                <el-input-number
                  v-model="item.actual_stock"
                  :min="0"
                  size="small"
                  @change="onItemQtyChange(item)"
                />
              </div>
              <div class="stock-divider">→</div>
              <div class="stock-col">
                <span class="stock-label">差异</span>
                <span
                  class="stock-num diff"
                  :class="{ profit: item.diff_qty > 0, loss: item.diff_qty < 0, zero: item.diff_qty === 0 && item.status === 1 }"
                >
                  {{ item.status === 1 ? (item.diff_qty > 0 ? '+' : '') + item.diff_qty : '-' }}
                </span>
              </div>
            </div>
          </div>

          <el-empty v-if="filteredItems.length === 0 && !loading" description="暂无商品" />
        </div>
      </div>
    </div>

    <div class="page-body" v-else-if="currentView === 'diff' && currentCheck">
      <div class="diff-report">
        <div class="diff-header-card">
          <h2>{{ currentCheck.title }}</h2>
          <p>盘点单号: {{ currentCheck.check_no }}</p>
          <div class="diff-summary-row">
            <div class="summary-item">
              <span class="label">SKU总数</span>
              <span class="value">{{ currentCheck.total_sku }}</span>
            </div>
            <div class="summary-item">
              <span class="label">已盘数量</span>
              <span class="value success">{{ currentCheck.checked_sku }}</span>
            </div>
            <div class="summary-item">
              <span class="label">盘盈SKU</span>
              <span class="value profit">{{ diffStats.profit_count || 0 }}</span>
            </div>
            <div class="summary-item">
              <span class="label">盘亏SKU</span>
              <span class="value loss">{{ diffStats.loss_count || 0 }}</span>
            </div>
            <div class="summary-item">
              <span class="label">无差异</span>
              <span class="value normal">{{ diffStats.normal_count || 0 }}</span>
            </div>
            <div class="summary-item">
              <span class="label">差异数量</span>
              <span class="value" :class="{ profit: (diffStats.total_diff_qty || 0) > 0, loss: (diffStats.total_diff_qty || 0) < 0 }">
                {{ (diffStats.total_diff_qty || 0) > 0 ? '+' : '' }}{{ diffStats.total_diff_qty || 0 }}
              </span>
            </div>
            <div class="summary-item">
              <span class="label">差异金额</span>
              <span class="value" :class="{ profit: (diffStats.total_diff_amount || 0) > 0, loss: (diffStats.total_diff_amount || 0) < 0 }">
                ¥{{ (diffStats.total_diff_amount || 0).toFixed(2) }}
              </span>
            </div>
          </div>
        </div>

        <el-tabs v-model="diffTab" type="border-card">
          <el-tab-pane label="全部差异" name="all">
            <diff-item-table :items="diffItems.all" />
          </el-tab-pane>
          <el-tab-pane label="盘盈商品" name="profit">
            <diff-item-table :items="diffItems.profit" />
          </el-tab-pane>
          <el-tab-pane label="盘亏商品" name="loss">
            <diff-item-table :items="diffItems.loss" />
          </el-tab-pane>
          <el-tab-pane label="无差异商品" name="normal">
            <diff-item-table :items="diffItems.normal" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <el-dialog v-model="showCreateDialog" title="新建盘点单" width="500px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="盘点标题">
          <el-input v-model="createForm.title" placeholder="请输入盘点标题" />
        </el-form-item>
        <el-form-item label="盘点类型">
          <el-select v-model="createForm.check_type" style="width: 100%">
            <el-option label="全部商品盘点" value="all" />
            <el-option label="分类盘点" value="category" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="createForm.operator_name" placeholder="请输入操作人姓名" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" type="textarea" :rows="3" placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createCheck" :loading="creating">
          生成盘点单
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Refresh, Upload, Check, Search, ArrowLeft, DataAnalysis
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const router = useRouter()
const route = useRoute()

const currentView = ref('list')
const loading = ref(false)
const creating = ref(false)
const networkStatus = ref(true)
const searchKeyword = ref('')
const filterStatus = ref(-1)
const diffTab = ref('all')

const checkList = ref([])
const currentCheck = ref(null)
const allItems = ref([])
const scanInput = ref('')
const scanQuantity = ref(1)
const diffStats = reactive({
  total_sku: 0,
  checked_sku: 0,
  normal_count: 0,
  profit_count: 0,
  loss_count: 0,
  total_diff_qty: 0,
  total_diff_amount: 0,
})

const showCreateDialog = ref(false)
const createForm = reactive({
  title: '',
  check_type: 'all',
  operator_name: '',
  remark: '',
})

const progressPercent = computed(() => {
  if (!currentCheck.value?.total_sku) return 0
  return Math.round((currentCheck.value.checked_sku / currentCheck.value.total_sku) * 100)
})

const filteredItems = computed(() => {
  let items = allItems.value
  if (filterStatus.value >= 0) {
    items = items.filter(i => i.status === filterStatus.value)
  }
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    items = items.filter(i =>
      (i.sku_code && i.sku_code.toLowerCase().includes(kw)) ||
      (i.product_name && i.product_name.toLowerCase().includes(kw)) ||
      (i.spec_name && i.spec_name.toLowerCase().includes(kw))
    )
  }
  return items
})

const diffItems = computed(() => {
  const checked = allItems.value.filter(i => i.status === 1)
  return {
    all: checked.filter(i => i.diff_qty !== 0),
    profit: checked.filter(i => i.diff_qty > 0),
    loss: checked.filter(i => i.diff_qty < 0),
    normal: checked.filter(i => i.diff_qty === 0),
  }
})

const goBack = () => {
  if (currentView.value === 'diff') {
    currentView.value = 'detail'
  } else {
    currentView.value = 'list'
    currentCheck.value = null
    allItems.value = []
  }
}

const loadCheckList = async () => {
  loading.value = true
  try {
    if (networkStatus.value) {
      const res = await request.get('/stock-checks', { params: { page_size: 50 } })
      checkList.value = res.data?.data?.list || res.data?.list || []
      for (const check of checkList.value) {
        saveToLocal(check)
      }
    } else {
      const list = await window.electronAPI.invoke('db:getStockCheckList')
      checkList.value = list || []
    }
  } catch (e) {
    console.error('Load check list error:', e)
    const list = await window.electronAPI.invoke('db:getStockCheckList')
    checkList.value = list || []
  } finally {
    loading.value = false
  }
}

const saveToLocal = async (check) => {
  try {
    await window.electronAPI.invoke('db:saveStockCheck', check)
  } catch (e) {
    console.error('Save to local error:', e)
  }
}

const openCheckDetail = async (check) => {
  loading.value = true
  try {
    let detail = null
    if (networkStatus.value) {
      const res = await request.get(`/stock-checks/${check.id}/items`, { params: { page_size: 500 } })
      detail = { ...check, items: res.data?.data?.list || res.data?.list || [] }
      await window.electronAPI.invoke('db:saveStockCheck', detail)
    } else {
      detail = await window.electronAPI.invoke('db:getStockCheckById', check.id)
    }

    currentCheck.value = detail
    allItems.value = detail?.items || []
    currentView.value = 'detail'
    loadStats()
  } catch (e) {
    console.error('Open check detail error:', e)
    ElMessage.error('加载盘点详情失败')
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const stats = await window.electronAPI.invoke('db:getStockCheckStats', currentCheck.value.id)
    if (stats) {
      Object.assign(diffStats, stats)
      if (currentCheck.value) {
        currentCheck.value.total_sku = stats.total_sku
        currentCheck.value.checked_sku = stats.checked_sku
        currentCheck.value.total_diff_qty = stats.total_diff_qty
        currentCheck.value.total_diff_amount = stats.total_diff_amount
      }
    }
  } catch (e) {
    console.error('Load stats error:', e)
  }
}

const loadItems = async () => {
  if (!currentCheck.value) return
  try {
    const items = await window.electronAPI.invoke('db:searchStockCheckItems',
      currentCheck.value.id, searchKeyword.value, filterStatus.value)
    allItems.value = items || []
  } catch (e) {
    console.error('Load items error:', e)
  }
}

const onSearch = () => {
  loadItems()
}

const onItemQtyChange = async (item) => {
  if (!currentCheck.value) return
  try {
    const actualStock = Number(item.actual_stock) || 0
    item.diff_qty = actualStock - item.system_stock
    item.diff_amount = item.diff_qty * (item.cost_price || 0)
    item.status = 1

    const result = await window.electronAPI.invoke('db:updateStockCheckItem', item.id, actualStock, item.remark || '')
    if (result) {
      currentCheck.value = result
      loadStats()
    }
  } catch (e) {
    console.error('Update item error:', e)
  }
}

const onScanSubmit = async () => {
  if (!scanInput.value || !currentCheck.value) return

  const skuCode = scanInput.value.trim()
  const qty = Number(scanQuantity.value) || 1

  try {
    const item = allItems.value.find(i => i.sku_code === skuCode)
    if (!item) {
      ElMessage.warning(`未找到SKU: ${skuCode}`)
      return
    }

    const newQty = (item.status === 1 ? (item.actual_stock || 0) : item.system_stock) + qty
    item.actual_stock = Math.max(0, newQty)
    item.diff_qty = item.actual_stock - item.system_stock
    item.diff_amount = item.diff_qty * (item.cost_price || 0)
    item.status = 1

    const result = await window.electronAPI.invoke('db:updateStockCheckItem', item.id, item.actual_stock, item.remark || '')
    if (result) {
      currentCheck.value = result
      loadStats()
      ElMessage.success(`${skuCode}: ${newQty}`)
    }
  } catch (e) {
    console.error('Scan error:', e)
  } finally {
    scanInput.value = ''
    scanQuantity.value = 1
    nextTick(() => {
      const input = document.querySelector('.scan-card .el-input__inner')
      if (input) input.focus()
    })
  }
}

const createCheck = async () => {
  if (!createForm.title) {
    ElMessage.warning('请输入盘点标题')
    return
  }

  creating.value = true
  try {
    let result
    if (networkStatus.value) {
      const res = await request.post('/stock-checks', {
        ...createForm,
        store_id: 1,
      })
      result = res.data?.data || res.data
    } else {
      const checkNo = 'PD' + Date.now()
      const items = await window.electronAPI.invoke('db:raw', `SELECT id, sku_code, product_name, spec_name, stock, category_id FROM product_skus WHERE status = 1`)
      const checkItems = items.map((sku, idx) => ({
        product_id: sku.product_id || 0,
        sku_id: sku.id,
        sku_code: sku.sku_code,
        product_name: sku.product_name,
        spec_name: sku.spec_name,
        system_stock: sku.stock || 0,
        actual_stock: 0,
        diff_qty: -(sku.stock || 0),
        cost_price: 0,
        diff_amount: 0,
        status: 0,
      }))
      result = {
        id: Date.now(),
        check_no: checkNo,
        title: createForm.title,
        check_type: createForm.check_type,
        status: 0,
        total_sku: checkItems.length,
        checked_sku: 0,
        operator_name: createForm.operator_name,
        remark: createForm.remark,
        items: checkItems,
        synced: 0,
        sync_status: 0,
      }
    }

    await window.electronAPI.invoke('db:saveStockCheck', result)
    ElMessage.success('盘点单创建成功')
    showCreateDialog.value = false
    createForm.title = ''
    createForm.operator_name = ''
    createForm.remark = ''
    loadCheckList()
  } catch (e) {
    console.error('Create check error:', e)
    ElMessage.error(e.message || '创建失败')
  } finally {
    creating.value = false
  }
}

const uploadCheck = async () => {
  if (!currentCheck.value || !networkStatus.value) return

  try {
    await ElMessageBox.confirm(
      `确定要将盘点单 ${currentCheck.value.check_no} 上传到云端吗？`,
      '上传确认',
      { type: 'info' }
    )

    const items = await window.electronAPI.invoke('db:searchStockCheckItems', currentCheck.value.id, '', 1)
    const uploadItems = items.map(i => ({
      sku_id: i.sku_id,
      sku_code: i.sku_code,
      actual_stock: i.actual_stock,
      remark: i.remark,
    }))

    await request.post(`/stock-checks/${currentCheck.value.id}/upload`, { items: uploadItems })

    await window.electronAPI.invoke('db:markStockCheckSynced', currentCheck.value.id)
    ElMessage.success('上传成功')
    loadCheckList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '上传失败')
    }
  }
}

const completeCheck = async () => {
  if (!currentCheck.value) return

  try {
    await ElMessageBox.confirm(
      `确定要完成盘点吗？完成后将更新系统库存。`,
      '完成确认',
      {
        type: 'warning',
        confirmButtonText: '确认完成',
        cancelButtonText: '取消',
      }
    )

    if (networkStatus.value) {
      const res = await request.post(`/stock-checks/${currentCheck.value.id}/complete`)
      currentCheck.value = res.data?.data || res.data
    } else {
      currentCheck.value.status = 2
      currentCheck.value.end_time = new Date().toISOString()
      await window.electronAPI.invoke('db:saveStockCheck', currentCheck.value)
    }

    ElMessage.success('盘点已完成')
    loadStats()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '操作失败')
    }
  }
}

const showDiff = () => {
  currentView.value = 'diff'
}

const getStatusType = (status) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'primary'
    case 2: return 'success'
    default: return 'info'
  }
}

const getStatusText = (status) => {
  switch (status) {
    case 0: return '进行中'
    case 1: return '待审核'
    case 2: return '已完成'
    default: return '未知'
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const DiffItemTable = {
  props: ['items'],
  template: `
    <el-table :data="items" stripe border v-loading="!items.length">
      <el-table-column prop="sku_code" label="SKU编码" width="140" />
      <el-table-column prop="product_name" label="商品名称" min-width="180" />
      <el-table-column prop="spec_name" label="规格" width="120" />
      <el-table-column prop="system_stock" label="系统库存" width="100" align="right" />
      <el-table-column prop="actual_stock" label="实盘数量" width="100" align="right" />
      <el-table-column label="差异数量" width="100" align="right">
        <template #default="{ row }">
          <span :class="{ profit: row.diff_qty > 0, loss: row.diff_qty < 0 }">
            {{ row.diff_qty > 0 ? '+' : '' }}{{ row.diff_qty }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="差异金额" width="120" align="right">
        <template #default="{ row }">
          <span :class="{ profit: row.diff_amount > 0, loss: row.diff_amount < 0 }">
            ¥{{ row.diff_amount?.toFixed(2) || '0.00' }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="120" />
    </el-table>
  `
}

onMounted(() => {
  loadCheckList()
  window.addEventListener('offline', () => { networkStatus.value = false })
  window.addEventListener('online', () => { networkStatus.value = true })
})

onBeforeUnmount(() => {
  window.removeEventListener('offline', () => { networkStatus.value = false })
  window.removeEventListener('online', () => { networkStatus.value = true })
})
</script>

<style lang="scss" scoped>
.stock-check-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #303133;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.page-body {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.check-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.check-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  }

  .check-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .check-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .check-no {
    font-size: 13px;
    color: #909399;
    margin-bottom: 16px;
    font-family: monospace;
  }

  .check-stats {
    display: flex;
    justify-content: space-between;
    padding: 12px 0;
    border-top: 1px solid #f0f0f0;
    border-bottom: 1px solid #f0f0f0;
    margin-bottom: 12px;
  }

  .stat-item {
    text-align: center;

    .stat-label {
      display: block;
      font-size: 12px;
      color: #909399;
      margin-bottom: 4px;
    }

    .stat-value {
      font-size: 20px;
      font-weight: 600;
      color: #303133;

      &.diff {
        &.profit { color: #67c23a; }
        &.loss { color: #f56c6c; }
      }
    }
  }

  .check-footer {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: #909399;
  }
}

.detail-body {
  display: flex;
  gap: 20px;
  padding: 20px;
}

.detail-sidebar {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stats-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  h3 {
    margin: 0 0 16px 0;
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    margin-bottom: 12px;
  }

  .stat-box {
    text-align: center;
    padding: 10px 4px;
    background: #f5f7fa;
    border-radius: 8px;

    .stat-num {
      font-size: 20px;
      font-weight: 700;
      color: #303133;
      margin-bottom: 2px;

      &.normal { color: #909399; }
      &.success { color: #67c23a; }
    }

    .stat-label {
      font-size: 12px;
      color: #909399;
    }

    &.success .stat-num { color: #67c23a; }
    &.warning .stat-num { color: #e6a23c; }
    &.danger .stat-num { color: #f56c6c; }
    &.info .stat-num { color: #409eff; }
  }

  .diff-summary {
    display: flex;
    justify-content: space-between;
    padding: 6px 0;
    font-size: 13px;
    color: #606266;

    .profit { color: #67c23a; font-weight: 600; }
    .loss { color: #f56c6c; font-weight: 600; }
  }
}

.scan-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  h3 {
    margin: 0 0 12px 0;
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .scan-tip {
    margin: 10px 0 0 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.5;
  }
}

.detail-main {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;

  .item-count {
    font-size: 13px;
    color: #909399;
  }
}

.items-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;

  .item-row {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 4px;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    &.checked {
      background: #f0f9eb;

      &.has-diff {
        background: #fdf6ec;
      }
    }

    .item-info {
      flex: 1;

      .item-name {
        font-size: 15px;
        font-weight: 500;
        color: #303133;
        margin-bottom: 4px;
      }

      .item-spec {
        font-size: 13px;
        color: #909399;
        font-weight: normal;
      }

      .item-meta {
        display: flex;
        gap: 12px;
        font-size: 12px;
        color: #909399;

        .sku-code {
          font-family: monospace;
          color: #606266;
        }
      }
    }

    .item-stocks {
      display: flex;
      align-items: center;
      gap: 16px;

      .stock-col {
        text-align: center;
        min-width: 100px;

        .stock-label {
          display: block;
          font-size: 12px;
          color: #909399;
          margin-bottom: 4px;
        }

        .stock-num {
          font-size: 18px;
          font-weight: 600;

          &.system { color: #909399; }
          &.diff {
            &.profit { color: #67c23a; }
            &.loss { color: #f56c6c; }
            &.zero { color: #909399; }
          }
        }
      }

      .stock-divider {
        color: #dcdfe6;
        font-size: 16px;
      }
    }
  }
}

.diff-report {
  max-width: 1200px;
  margin: 0 auto;

  .diff-header-card {
    background: #fff;
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    h2 {
      margin: 0 0 8px 0;
      font-size: 22px;
      color: #303133;
    }

    p {
      margin: 0 0 16px 0;
      color: #909399;
      font-family: monospace;
    }

    .diff-summary-row {
      display: flex;
      gap: 24px;
      flex-wrap: wrap;

      .summary-item {
        .label {
          font-size: 13px;
          color: #909399;
          margin-right: 8px;
        }

        .value {
          font-size: 18px;
          font-weight: 600;
          color: #303133;

          &.profit { color: #67c23a; }
          &.loss { color: #f56c6c; }
          &.success { color: #67c23a; }
          &.normal { color: #909399; }
        }
      }
    }
  }
}

.profit { color: #67c23a !important; }
.loss { color: #f56c6c !important; }
</style>
