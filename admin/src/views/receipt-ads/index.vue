<template>
  <div class="receipt-ads-page">
    <div class="page-header">
      <h2 class="page-title">小票广告位</h2>
      <div class="header-actions">
        <el-button type="primary" @click="openAdDialog">
          <el-icon><Plus /></el-icon>新增广告
        </el-button>
      </div>
    </div>

    <el-row :gutter="16" class="stats-cards">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-label">广告总数</div>
          <div class="stat-value">{{ stats.total }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card active">
          <div class="stat-label">启用中</div>
          <div class="stat-value">{{ stats.active }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card view">
          <div class="stat-label">总展示次数</div>
          <div class="stat-value">{{ stats.totalViews }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card click">
          <div class="stat-label">总点击次数</div>
          <div class="stat-value">{{ stats.totalClicks }}</div>
        </div>
      </el-col>
    </el-row>

    <div class="card-wrapper">
      <div class="filter-bar">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="关键词">
            <el-input
              v-model="filterForm.keyword"
              placeholder="广告标题/内容"
              clearable
              style="width: 200px"
              @keyup.enter="fetchList" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filterForm.status" placeholder="全部" clearable style="width: 120px">
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="filterForm.ad_type" placeholder="全部" clearable style="width: 120px">
              <el-option label="图片" value="image" />
              <el-option label="二维码" value="qrcode" />
              <el-option label="文字" value="text" />
            </el-select>
          </el-form-item>
          <el-form-item label="位置">
            <el-select v-model="filterForm.position" placeholder="全部位置" clearable style="width: 120px">
              <el-option label="底部" value="footer" />
              <el-option label="顶部" value="header" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchList">
              <el-icon><Search /></el-icon>搜索
            </el-button>
            <el-button @click="resetFilter">
              <el-icon><Refresh /></el-icon>重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="adList" v-loading="loading" stripe style="width: 100%">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="title" label="广告标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="ad_type_text" label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getAdTypeTag(row.ad_type)" size="small">{{ row.ad_type_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="position_text" label="位置" width="80" align="center" />
        <el-table-column prop="sort_order" label="排序" width="70" align="center" />
        <el-table-column label="展示/点击" width="160" align="center">
          <template #default="{ row }">
            <div class="stats-mini">
              <span class="view-count">展示 {{ row.view_count }}</span>
              <span class="divider">|</span>
              <span class="click-count">点击 {{ row.click_count }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="投放时间" width="200" align="center">
          <template #default="{ row }">
            <div v-if="row.start_date || row.end_date" class="date-range">
              <span>{{ row.start_date || '不限' }}</span>
              <span class="arrow">→</span>
              <span>{{ row.end_date || '不限' }}</span>
            </div>
            <span v-else class="text-muted">长期有效</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" align="center" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="handleViewStats(row)">统计</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="filterForm.page"
          v-model:page-size="filterForm.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList" />
      </div>
    </div>

    <el-dialog
      v-model="adDialogVisible"
      :title="isEdit ? '编辑广告' : '新增广告'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="adFormRef"
        :model="adForm"
        :rules="adRules"
        label-width="100px">
        <el-row :gutter="20">
          <el-col :span="14">
            <el-form-item label="广告标题" prop="title">
              <el-input v-model="adForm.title" placeholder="请输入广告标题" maxlength="100" show-word-limit />
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="广告类型" prop="ad_type">
              <el-select v-model="adForm.ad_type" style="width: 100%" @change="handleAdTypeChange">
                <el-option label="二维码广告" value="qrcode" />
                <el-option label="图片广告" value="image" />
                <el-option label="文字广告" value="text" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="adForm.ad_type === 'qrcode'">
          <el-form-item label="二维码内容" prop="qr_code_content">
            <el-input
              v-model="adForm.qr_code_content"
              type="textarea"
              :rows="2"
              placeholder="请输入二维码内容（链接或文字）"
              maxlength="500"
              show-word-limit />
          </el-form-item>
        </template>

        <template v-if="adForm.ad_type === 'image'">
          <el-form-item label="图片地址" prop="image_url">
            <el-input v-model="adForm.image_url" placeholder="请输入图片URL" />
          </el-form-item>
        </template>

        <el-form-item label="副标题">
          <el-input v-model="adForm.subtitle" placeholder="请输入副标题（可选）" maxlength="100" show-word-limit />
        </el-form-item>

        <el-form-item label="广告内容">
          <el-input
            v-model="adForm.content"
            type="textarea"
            :rows="3"
            placeholder="请输入广告内容（可选）"
            maxlength="500"
            show-word-limit />
        </el-form-item>

        <el-form-item label="跳转链接">
          <el-input v-model="adForm.link_url" placeholder="扫码或点击后跳转的链接（可选）" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="显示位置" prop="position">
              <el-select v-model="adForm.position" style="width: 100%">
                <el-option label="小票底部" value="footer" />
                <el-option label="小票顶部" value="header" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="排序权重">
              <el-input-number v-model="adForm.sort_order" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-switch v-model="adForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始日期">
              <el-date-picker
                v-model="adForm.start_date"
                type="date"
                placeholder="选择开始日期"
                value-format="YYYY-MM-DD"
                style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束日期">
              <el-date-picker
                v-model="adForm.end_date"
                type="date"
                placeholder="选择结束日期"
                value-format="YYYY-MM-DD"
                style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始时间">
              <el-time-picker
                v-model="adForm.start_time"
                placeholder="选择开始时间"
                format="HH:mm"
                value-format="HH:mm"
                style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间">
              <el-time-picker
                v-model="adForm.end_time"
                placeholder="选择结束时间"
                format="HH:mm"
                value-format="HH:mm"
                style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="备注">
          <el-input v-model="adForm.remark" type="textarea" :rows="2" placeholder="备注信息" maxlength="255" show-word-limit />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="adDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="statsDialogVisible" title="广告数据统计" width="800px">
      <div class="stats-detail">
        <div class="stats-summary">
          <div class="stat-item">
            <div class="stat-label">总展示</div>
            <div class="stat-value">{{ currentAd.view_count || 0 }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">总点击</div>
            <div class="stat-value">{{ currentAd.click_count || 0 }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">点击率</div>
            <div class="stat-value">{{ clickRate }}%</div>
          </div>
        </div>
        <div class="empty-tip">
          <el-icon><InfoFilled /></el-icon>
          数据统计周期内点击数据详情可通过对接数据分析平台查看
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, InfoFilled } from '@element-plus/icons-vue'
import {
  getReceiptAdList,
  createReceiptAd,
  updateReceiptAd,
  deleteReceiptAd,
  updateReceiptAdStatus
} from '@/api/receipt-ads'

const loading = ref(false)
const submitLoading = ref(false)
const adList = ref([])
const total = ref(0)

const stats = reactive({
  total: 0,
  active: 0,
  totalViews: 0,
  totalClicks: 0
})

const filterForm = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  status: null,
  ad_type: '',
  position: ''
})

const adDialogVisible = ref(false)
const statsDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const adFormRef = ref()
const currentAd = ref({})

const adForm = reactive({
  title: '',
  ad_type: 'qrcode',
  image_url: '',
  qr_code_content: '',
  link_url: '',
  content: '',
  subtitle: '',
  position: 'footer',
  sort_order: 0,
  status: 1,
  start_date: '',
  end_date: '',
  start_time: '',
  end_time: '',
  remark: ''
})

const adRules = {
  title: [{ required: true, message: '请输入广告标题', trigger: 'blur' }],
  ad_type: [{ required: true, message: '请选择广告类型', trigger: 'change' }],
  position: [{ required: true, message: '请选择显示位置', trigger: 'change' }]
}

const clickRate = computed(() => {
  if (!currentAd.value || !currentAd.value.view_count || currentAd.value.view_count === 0) {
    return '0.00'
  }
  return ((currentAd.value.click_count / currentAd.value.view_count) * 100).toFixed(2)
})

function getAdTypeTag(type) {
  const map = {
    image: '',
    qrcode: 'success',
    text: 'info'
  }
  return map[type] || ''
}

async function fetchList() {
  loading.value = true
  try {
    const params = { ...filterForm }
    if (params.status === null || params.status === '') delete params.status
    const res = await getReceiptAdList(params)
    adList.value = res.list || []
    total.value = res.total || 0
    calculateStats()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function calculateStats() {
  let activeCount = 0
  let totalViews = 0
  let totalClicks = 0
  adList.value.forEach(item => {
    if (item.status === 1) activeCount++
    totalViews += item.view_count || 0
    totalClicks += item.click_count || 0
  })
  stats.total = total.value
  stats.active = activeCount
  stats.totalViews = totalViews
  stats.totalClicks = totalClicks
}

function resetFilter() {
  filterForm.keyword = ''
  filterForm.status = null
  filterForm.ad_type = ''
  filterForm.position = ''
  filterForm.page = 1
  fetchList()
}

function openAdDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(adForm, {
      title: row.title,
      ad_type: row.ad_type,
      image_url: row.image_url || '',
      qr_code_content: row.qr_code_content || '',
      link_url: row.link_url || '',
      content: row.content || '',
      subtitle: row.subtitle || '',
      position: row.position,
      sort_order: row.sort_order,
      status: row.status,
      start_date: row.start_date || '',
      end_date: row.end_date || '',
      start_time: row.start_time || '',
      end_time: row.end_time || '',
      remark: row.remark || ''
    })
  } else {
    adForm.title = ''
    adForm.ad_type = 'qrcode'
    adForm.image_url = ''
    adForm.qr_code_content = ''
    adForm.link_url = ''
    adForm.content = ''
    adForm.subtitle = ''
    adForm.position = 'footer'
    adForm.sort_order = 0
    adForm.status = 1
    adForm.start_date = ''
    adForm.end_date = ''
    adForm.start_time = ''
    adForm.end_time = ''
    adForm.remark = ''
  }

  adDialogVisible.value = true
}

function handleEdit(row) {
  openAdDialog(row)
}

function handleAdTypeChange() {
  // 切换类型时清空对应字段
}

async function handleStatusChange(row) {
  try {
    await updateReceiptAdStatus(row.id, row.status)
    ElMessage.success(row.status === 1 ? '广告已启用' : '广告已禁用')
    calculateStats()
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除广告"${row.title}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteReceiptAd(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

function handleViewStats(row) {
  currentAd.value = row
  statsDialogVisible.value = true
}

async function handleSubmit() {
  try {
    await adFormRef.value.validate()
    submitLoading.value = true

    const data = { ...adForm }

    if (isEdit.value) {
      await updateReceiptAd(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createReceiptAd(data)
      ElMessage.success('创建成功')
    }

    adDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped lang="scss">
.receipt-ads-page {
  .stats-cards {
    margin-bottom: 20px;
  }

  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .stat-label {
      font-size: 14px;
      color: #909399;
      margin-bottom: 8px;
    }

    .stat-value {
      font-size: 28px;
      font-weight: 600;
      color: #303133;
    }

    &.active .stat-value {
      color: #67c23a;
    }

    &.view .stat-value {
      color: #409eff;
    }

    &.click .stat-value {
      color: #e6a23c;
    }
  }

  .filter-bar {
    margin-bottom: 16px;
  }

  .stats-mini {
    font-size: 12px;

    .view-count {
      color: #409eff;
    }

    .click-count {
      color: #e6a23c;
    }

    .divider {
      color: #dcdfe6;
      margin: 0 6px;
    }
  }

  .date-range {
    font-size: 12px;
    color: #606266;

    .arrow {
      color: #c0c4cc;
      margin: 0 4px;
    }
  }

  .text-muted {
    color: #c0c4cc;
    font-size: 12px;
  }

  .pagination {
    margin-top: 20px;
    text-align: right;
  }

  .stats-detail {
    .stats-summary {
      display: flex;
      gap: 24px;
      margin-bottom: 24px;

      .stat-item {
        flex: 1;
        text-align: center;
        padding: 20px;
        background: #f5f7fa;
        border-radius: 8px;

        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 28px;
          font-weight: 600;
          color: #303133;
        }
      }
    }

    .empty-tip {
      text-align: center;
      color: #909399;
      font-size: 13px;
      padding: 20px;
      background: #fdf6ec;
      border-radius: 8px;

      .el-icon {
        margin-right: 6px;
        color: #e6a23c;
      }
    }
  }
}
</style>
