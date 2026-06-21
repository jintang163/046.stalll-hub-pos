<template>
  <div class="bom-page">
    <div class="page-header">
      <h2>菜品BOM管理</h2>
    </div>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="门店">
          <el-select v-model="searchForm.store_id" style="width: 150px" @change="loadProducts">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜品">
          <el-select v-model="searchForm.product_id" placeholder="请选择菜品" filterable style="width: 250px" @change="loadBOM">
            <el-option v-for="p in productList" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="10">
        <el-card class="product-card">
          <template #header>
            <div class="card-header">
              <span>菜品列表</span>
            </div>
          </template>
          <el-table :data="productList" border highlight-current-row @row-click="selectProduct" height="600">
            <el-table-column prop="name" label="菜品名称" />
            <el-table-column prop="price" label="售价" width="100">
              <template #default="{ row }">¥{{ row.price }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="14">
        <el-card class="bom-card" v-if="selectedProduct">
          <template #header>
            <div class="card-header">
              <span>{{ selectedProduct.name }} - 物料清单</span>
              <div>
                <el-button type="primary" size="small" @click="handleAddItem">
                  <el-icon><Plus /></el-icon>
                  添加食材
                </el-button>
                <el-button size="small" @click="viewCostDetail">成本详情</el-button>
              </div>
            </div>
          </template>

          <el-table :data="bomItems" border>
            <el-table-column type="index" label="序号" width="60" />
            <el-table-column prop="ingredient_name" label="食材名称" min-width="120" />
            <el-table-column prop="quantity" label="用量" width="100">
              <template #default="{ row }">
                {{ row.quantity }} {{ row.unit }}
              </template>
            </el-table-column>
            <el-table-column prop="wastage_rate" label="损耗率" width="100">
              <template #default="{ row }">{{ row.wastage_rate || 0 }}%</template>
            </el-table-column>
            <el-table-column label="食材单价" width="100">
              <template #default="{ row }">
                ¥{{ row.ingredient?.current_price || 0 }}
              </template>
            </el-table-column>
            <el-table-column label="成本" width="100">
              <template #default="{ row }">
                ¥{{ calculateItemCost(row).toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row, $index }">
                <el-button link type="primary" @click="editItem(row, $index)">编辑</el-button>
                <el-button link type="danger" @click="removeItem($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="bom-footer">
            <div class="total-cost">
              总成本：<span class="cost-amount">¥{{ totalCost.toFixed(2) }}</span>
            </div>
            <el-button type="primary" @click="saveBOM">保存BOM</el-button>
          </div>
        </el-card>

        <el-card v-else class="empty-card">
          <el-empty description="请选择左侧菜品查看BOM" />
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="itemDialogVisible" :title="itemDialogTitle" width="500px">
      <el-form :model="itemForm" label-width="100px">
        <el-form-item label="食材">
          <el-select v-model="itemForm.ingredient_id" filterable placeholder="请选择食材" style="width: 100%" @change="onIngredientChange">
            <el-option v-for="ing in ingredientList" :key="ing.id" :label="`${ing.name} (${ing.unit}) - ¥${ing.current_price}`" :value="ing.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="用量">
          <el-input-number v-model="itemForm.quantity" :precision="3" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="itemForm.unit" placeholder="如：斤、个、克" />
        </el-form-item>
        <el-form-item label="损耗率(%)">
          <el-input-number v-model="itemForm.wastage_rate" :precision="2" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmItem">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="costDetailVisible" title="成本构成详情" width="600px">
      <div v-if="costDetail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="菜品">{{ selectedProduct?.name }}</el-descriptions-item>
          <el-descriptions-item label="食材数量">{{ costDetail.ingredient_count }}</el-descriptions-item>
          <el-descriptions-item label="总成本" :span="2">
            <span style="color: #f56c6c; font-size: 20px; font-weight: bold">
              ¥{{ costDetail.total_cost }}
            </span>
          </el-descriptions-item>
        </el-descriptions>

        <el-table :data="costDetail.ingredient_cost" border style="margin-top: 20px">
          <el-table-column type="index" label="序号" width="60" />
          <el-table-column prop="ingredient_name" label="食材名称" />
          <el-table-column prop="unit_price" label="单价" width="100">
            <template #default="{ row }">¥{{ row.unit_price }}/{{ row.unit }}</template>
          </el-table-column>
          <el-table-column prop="quantity" label="用量" width="80">
            <template #default="{ row }">{{ row.quantity }}{{ row.unit }}</template>
          </el-table-column>
          <el-table-column prop="wastage_rate" label="损耗" width="80">
            <template #default="{ row }">{{ row.wastage_rate }}%</template>
          </el-table-column>
          <el-table-column prop="actual_qty" label="实际用量" width="100">
            <template #default="{ row }">{{ row.actual_qty }}{{ row.unit }}</template>
          </el-table-column>
          <el-table-column prop="total_cost" label="成本" width="100">
            <template #default="{ row }">¥{{ row.total_cost }}</template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as bomApi from '@/api/ingredient'
import { storeApi } from '@/api/stores'
import { productApi } from '@/api/product'

const storeList = ref([])
const productList = ref([])
const ingredientList = ref([])
const bomItems = ref([])
const selectedProduct = ref(null)
const costDetail = ref(null)

const searchForm = reactive({
  store_id: 0,
  product_id: 0
})

const itemDialogVisible = ref(false)
const itemDialogTitle = ref('添加食材')
const costDetailVisible = ref(false)
const editIndex = ref(-1)

const itemForm = reactive({
  ingredient_id: 0,
  ingredient_name: '',
  quantity: 0,
  unit: '',
  wastage_rate: 0
})

const totalCost = computed(() => {
  return bomItems.value.reduce((sum, item) => {
    return sum + calculateItemCost(item)
  }, 0)
})

onMounted(() => {
  loadStores()
})

function loadStores() {
  storeApi.list({ page: 1, page_size: 100 }).then(res => {
    storeList.value = res.data?.list || res.data || []
    if (storeList.value.length > 0) {
      searchForm.store_id = storeList.value[0].id
      loadProducts()
      loadIngredients()
    }
  })
}

function loadProducts() {
  productApi.list({ store_id: searchForm.store_id, page: 1, page_size: 200 }).then(res => {
    productList.value = res.data?.list || res.data || []
  })
}

function loadIngredients() {
  bomApi.getIngredients({
    store_id: searchForm.store_id,
    status: 1,
    page: 1,
    page_size: 500
  }).then(res => {
    ingredientList.value = res.data?.list || res.data || []
  })
}

function selectProduct(row) {
  selectedProduct.value = row
  searchForm.product_id = row.id
  loadBOM()
}

function loadBOM() {
  if (!searchForm.product_id) return
  bomApi.getProductBOM(searchForm.product_id).then(res => {
    bomItems.value = res.data || []
  })
}

function calculateItemCost(item) {
  const price = item.ingredient?.current_price || 0
  const qty = item.quantity || 0
  const wastage = item.wastage_rate || 0
  const actualQty = qty * (1 + wastage / 100)
  return price * actualQty
}

function handleAddItem() {
  if (!selectedProduct.value) {
    ElMessage.warning('请先选择菜品')
    return
  }
  itemDialogTitle.value = '添加食材'
  editIndex.value = -1
  Object.assign(itemForm, {
    ingredient_id: 0,
    ingredient_name: '',
    quantity: 0,
    unit: '',
    wastage_rate: 0
  })
  itemDialogVisible.value = true
}

function editItem(row, index) {
  itemDialogTitle.value = '编辑食材'
  editIndex.value = index
  Object.assign(itemForm, {
    ingredient_id: row.ingredient_id,
    ingredient_name: row.ingredient_name,
    quantity: row.quantity,
    unit: row.unit,
    wastage_rate: row.wastage_rate
  })
  itemDialogVisible.value = true
}

function removeItem(index) {
  bomItems.value.splice(index, 1)
}

function onIngredientChange(id) {
  const ing = ingredientList.value.find(i => i.id === id)
  if (ing) {
    itemForm.ingredient_name = ing.name
    itemForm.unit = ing.unit
  }
}

function confirmItem() {
  if (!itemForm.ingredient_id) {
    ElMessage.warning('请选择食材')
    return
  }
  if (itemForm.quantity <= 0) {
    ElMessage.warning('用量必须大于0')
    return
  }

  const ingredient = ingredientList.value.find(i => i.id === itemForm.ingredient_id)

  const item = {
    ingredient_id: itemForm.ingredient_id,
    ingredient_name: itemForm.ingredient_name,
    quantity: itemForm.quantity,
    unit: itemForm.unit,
    wastage_rate: itemForm.wastage_rate,
    ingredient: ingredient
  }

  if (editIndex.value >= 0) {
    bomItems.value[editIndex.value] = item
  } else {
    bomItems.value.push(item)
  }

  itemDialogVisible.value = false
}

function saveBOM() {
  if (!selectedProduct.value) {
    ElMessage.warning('请先选择菜品')
    return
  }

  const items = bomItems.value.map((item, index) => ({
    ingredient_id: item.ingredient_id,
    ingredient_name: item.ingredient_name,
    quantity: item.quantity,
    unit: item.unit,
    wastage_rate: item.wastage_rate,
    sort_order: index + 1
  }))

  bomApi.saveProductBOM({
    store_id: searchForm.store_id,
    product_id: searchForm.product_id,
    items: items
  }).then(() => {
    ElMessage.success('BOM保存成功')
    loadBOM()
  }).catch(err => {
    ElMessage.error('保存失败')
  })
}

function viewCostDetail() {
  if (!selectedProduct.value) return
  bomApi.getProductCostDetail(searchForm.product_id).then(res => {
    costDetail.value = res.data
    costDetailVisible.value = true
  })
}
</script>

<style scoped>
.bom-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.search-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.product-card,
.bom-card {
  margin-bottom: 20px;
}

.empty-card {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bom-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #ebeef5;
}

.total-cost {
  font-size: 16px;
}

.cost-amount {
  color: #f56c6c;
  font-size: 24px;
  font-weight: bold;
}
</style>
