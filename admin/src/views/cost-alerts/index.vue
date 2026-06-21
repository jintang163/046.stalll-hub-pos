<template>
  <div class="cost-alerts-page">
    <div class="page-header">
      <h2>成本异常告警</h2>
      <el-badge :value="unreadCount" :max="99" class="badge">
        <el-button type="warning" @click="showUnreadOnly = !showUnreadOnly">
          未处理告警
        </el-button>
      </el-badge>
    </div>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="门店">
          <el-select v-model="searchForm.store_id" placeholder="全部门店" style="width: 150px">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" style="width: 120px">
            <el-option label="全部" :value="-1" />
            <el-option label="未处理" :value="0" />
            <el-option label="已处理" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
          <el-button @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="ingredient_name" label="食材名称" min-width="120" />
        <el-table-column prop="alert_type" label="告警类型" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.alert_type === 'price_rise'" type="danger">价格上涨</el-tag>
            <el-tag v-else type="warning">{{ row.alert_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格变动" width="200">
          <template #default="{ row }">
            <div class="price-change">
              <span class="old-price">¥{{ row.previous_price }}</span>
              <el-icon class="arrow"><Right /></el-icon>
              <span class="new-price">¥{{ row.current_price }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="change_rate" label="涨跌幅" width="100">
          <template #default="{ row }">
            <span :class="{ 'rise': row.change_rate > 0, 'fall': row.change_rate < 0 }">
              {{ row.change_rate > 0 ? '+' : '' }}{{ row.change_rate }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="threshold" label="阈值" width="100">
          <template #default="{ row }">{{ row.threshold }}%</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'warning' : 'success'">
              {{ row.status === 0 ? '未处理' : '已处理' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="handler" label="处理人" width="100" />
        <el-table-column prop="created_at" label="告警时间" width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 0" link type="primary" @click="handleAlert(row)">处理</el-button>
            <el-button link type="info" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </el-card>

    <el-dialog v-model="handleDialogVisible" title="处理告警" width="500px">
      <el-form :model="handleForm" label-width="80px">
        <el-form-item label="处理人">
          <el-input v-model="handleForm.handler" placeholder="请输入处理人姓名" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="handleForm.remark" type="textarea" :rows="4" placeholder="请输入处理备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitHandle">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Right } from '@element-plus/icons-vue'
import * as alertApi from '@/api/ingredient'
import { storeApi } from '@/api/stores'

const loading = ref(false)
const handleDialogVisible = ref(false)
const showUnreadOnly = ref(false)

const searchForm = reactive({
  store_id: 0,
  status: -1
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const tableData = ref([])
const storeList = ref([])

const handleForm = reactive({
  alert_id: 0,
  handler: '',
  remark: ''
})

const unreadCount = computed(() => {
  return tableData.value.filter(item => item.status === 0).length
})

onMounted(() => {
  loadStores()
  fetchList()
})

function loadStores() {
  storeApi.list({ page: 1, page_size: 100 }).then(res => {
    storeList.value = res.data?.list || res.data || []
    if (storeList.value.length > 0 && !searchForm.store_id) {
      searchForm.store_id = storeList.value[0].id
    }
  })
}

function fetchList() {
  loading.value = true
  const status = showUnreadOnly.value ? 0 : searchForm.status
  alertApi.getCostAlerts({
    store_id: searchForm.store_id,
    status: status,
    page: pagination.page,
    page_size: pagination.pageSize
  }).then(res => {
    const data = res.data
    tableData.value = data?.list || data || []
    pagination.total = data?.total || 0
    loading.value = false
  }).catch(() => {
    ElMessage.error('获取告警列表失败')
    loading.value = false
  })
}

function handleRefresh() {
  pagination.page = 1
  fetchList()
}

function handleAlert(row) {
  handleForm.alert_id = row.id
  handleForm.handler = ''
  handleForm.remark = ''
  handleDialogVisible.value = true
}

function submitHandle() {
  if (!handleForm.handler) {
    ElMessage.warning('请输入处理人')
    return
  }
  alertApi.handleCostAlert(handleForm).then(() => {
    ElMessage.success('处理成功')
    handleDialogVisible.value = false
    fetchList()
  }).catch(() => {
    ElMessage.error('处理失败')
  })
}

function viewDetail(row) {
  ElMessage.info(`食材 ${row.ingredient_name} 详情`)
}
</script>

<style scoped>
.cost-alerts-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.badge {
  margin-top: 0;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
  display: flex;
}

.price-change {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.old-price {
  color: #909399;
  text-decoration: line-through;
}

.new-price {
  color: #f56c6c;
  font-weight: bold;
}

.arrow {
  color: #c0c4cc;
}

.rise {
  color: #f56c6c;
  font-weight: bold;
}

.fall {
  color: #67c23a;
  font-weight: bold;
}
</style>
