<template>
  <div class="ingredients-page">
    <div class="page-header">
      <h2>食材管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增食材
        </el-button>
        <el-button @click="handleSync">
          <el-icon><Refresh /></el-icon>
          同步进销存
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="门店">
          <el-select v-model="searchForm.store_id" placeholder="全部门店" style="width: 150px">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchForm.category" placeholder="全部分类" clearable style="width: 150px">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="食材名称/编号" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" style="width: 120px">
            <el-option label="全部" :value="-1" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="ingredient_no" label="食材编号" width="120" />
        <el-table-column prop="name" label="食材名称" min-width="120" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="current_price" label="当前单价" width="120">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: bold">¥{{ row.current_price }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="supplier" label="供应商" width="120" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewPriceHistory(row)">价格历史</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="门店">
          <el-select v-model="form.store_id" placeholder="请选择门店" style="width: 100%">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="食材编号">
          <el-input v-model="form.ingredient_no" placeholder="请输入食材编号" />
        </el-form-item>
        <el-form-item label="食材名称">
          <el-input v-model="form.name" placeholder="请输入食材名称" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="请输入分类" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="form.unit" placeholder="如：斤、个、份" />
        </el-form-item>
        <el-form-item label="单价">
          <el-input-number v-model="form.current_price" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="供应商">
          <el-input v-model="form.supplier" placeholder="请输入供应商" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="historyVisible" title="价格历史" width="700px">
      <el-table :data="priceHistory" border>
        <el-table-column prop="effective_date" label="日期" width="120" />
        <el-table-column prop="previous_price" label="原价" width="100">
          <template #default="{ row }">¥{{ row.previous_price }}</template>
        </el-table-column>
        <el-table-column prop="price" label="现价" width="100">
          <template #default="{ row }">
            <span style="color: #f56c6c">¥{{ row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="price_change" label="涨跌幅" width="100">
          <template #default="{ row }">
            <span :style="{ color: row.price_change >= 0 ? '#f56c6c' : '#67c23a' }">
              {{ row.price_change >= 0 ? '+' : '' }}{{ row.price_change }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column prop="supplier" label="供应商" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import * as ingredientApi from '@/api/ingredient'
import { storeApi } from '@/api/stores'

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增食材')
const historyVisible = ref(false)
const priceHistory = ref([])

const searchForm = reactive({
  store_id: 0,
  category: '',
  keyword: '',
  status: -1
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const tableData = ref([])
const storeList = ref([])
const categories = ref([])

const form = reactive({
  id: 0,
  store_id: 0,
  ingredient_no: '',
  name: '',
  category: '',
  unit: '',
  current_price: 0,
  supplier: '',
  status: 1,
  remark: ''
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
    loadCategories()
  }).catch(() => {
    storeList.value = [{ id: 1, name: '默认门店' }]
  })
}

function loadCategories() {
  ingredientApi.getIngredientCategories({ store_id: searchForm.store_id }).then(res => {
    categories.value = res.data || []
  })
}

function fetchList() {
  loading.value = true
  const params = {
    store_id: searchForm.store_id,
    category: searchForm.category,
    keyword: searchForm.keyword,
    status: searchForm.status,
    page: pagination.page,
    page_size: pagination.pageSize
  }
  ingredientApi.getIngredients(params).then(res => {
    const data = res.data
    tableData.value = data?.list || data || []
    pagination.total = data?.total || 0
    loading.value = false
  }).catch(err => {
    ElMessage.error('获取食材列表失败')
    loading.value = false
  })
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleReset() {
  searchForm.category = ''
  searchForm.keyword = ''
  searchForm.status = -1
  pagination.page = 1
  fetchList()
}

function handleAdd() {
  dialogTitle.value = '新增食材'
  Object.assign(form, {
    id: 0,
    store_id: searchForm.store_id,
    ingredient_no: '',
    name: '',
    category: '',
    unit: '',
    current_price: 0,
    supplier: '',
    status: 1,
    remark: ''
  })
  dialogVisible.value = true
}

function handleEdit(row) {
  dialogTitle.value = '编辑食材'
  Object.assign(form, row)
  dialogVisible.value = true
}

function handleSubmit() {
  if (!form.name) {
    ElMessage.warning('请输入食材名称')
    return
  }
  const api = form.id ? ingredientApi.updateIngredient(form.id, form) : ingredientApi.createIngredient(form)
  api.then(() => {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchList()
  }).catch(err => {
    ElMessage.error('保存失败')
  })
}

function handleDelete(row) {
  ElMessageBox.confirm('确定要删除该食材吗？', '提示', {
    type: 'warning'
  }).then(() => {
    ingredientApi.deleteIngredient(row.id).then(() => {
      ElMessage.success('删除成功')
      fetchList()
    })
  })
}

function viewPriceHistory(row) {
  ingredientApi.getPriceHistory({ ingredient_id: row.id, limit: 30 }).then(res => {
    priceHistory.value = res.data || []
    historyVisible.value = true
  })
}

function handleSync() {
  ElMessageBox.confirm('确定要从进销存系统同步食材数据吗？', '提示', {
    type: 'info'
  }).then(() => {
    ingredientApi.triggerInventorySync().then(() => {
      ElMessage.success('同步任务已启动，请稍后刷新查看')
    })
  })
}
</script>

<style scoped>
.ingredients-page {
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
</style>
