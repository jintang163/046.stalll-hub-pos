<template>
  <div class="products-page">
    <div class="page-header">
      <h2 class="page-title">商品管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="openBatchPriceDialog" :disabled="selectedIds.length === 0">
          <el-icon><Money /></el-icon>批量改价
        </el-button>
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增商品
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索商品名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.category_id" placeholder="选择分类" clearable style="width: 180px">
          <el-option label="全部分类" :value="0" />
          <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
        </el-select>
        <el-select v-model="query.status" placeholder="商品状态" clearable style="width: 140px">
          <el-option label="上架" :value="1" />
          <el-option label="下架" :value="0" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <div class="table-toolbar">
        <el-checkbox v-model="selectAll" :indeterminate="isIndeterminate" @change="handleSelectAll">
          全选
        </el-checkbox>
        <span style="margin-left: 12px; color: #909399;">已选择 {{ selectedIds.length }} 项</span>
      </div>

      <el-table
        :data="list"
        v-loading="loading"
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="main_image" label="图片" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.main_image"
              :src="row.main_image"
              :preview-src-list="[row.main_image]"
              style="width: 60px; height: 60px; border-radius: 4px;"
              fit="cover" />
            <div v-else style="width: 60px; height: 60px; background: #f5f7fa; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #c0c4cc;">
              无图
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="商品名称" min-width="180">
          <template #default="{ row }">
            <div class="product-name">
              <span>{{ row.name }}</span>
              <el-tag v-if="row.is_hot" type="danger" size="small" style="margin-left: 8px;">热销</el-tag>
              <el-tag v-if="row.is_recommend" type="warning" size="small" style="margin-left: 4px;">推荐</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="价格区间" width="140">
          <template #default="{ row }">
            <div class="price-range">
              <span v-if="row.min_price === row.max_price" class="price">¥{{ row.min_price }}</span>
              <span v-else class="price">¥{{ row.min_price }} - ¥{{ row.max_price }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="sku_count" label="SKU数" width="80" align="center" />
        <el-table-column prop="total_stock" label="总库存" width="100" align="center">
          <template #default="{ row }">
            <span :class="{ 'low-stock': row.total_stock <= 10 }">{{ row.total_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleCopy(row)">复制</el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList" />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑商品' : '新增商品'"
      width="900px"
      :close-on-click-modal="false">
      <el-form
        ref="productFormRef"
        :model="productForm"
        :rules="productRules"
        label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="商品名称" prop="name">
              <el-input v-model="productForm.name" placeholder="请输入商品名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属分类" prop="category_id">
              <el-select v-model="productForm.category_id" placeholder="请选择分类" style="width: 100%">
                <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="商品描述">
          <el-input
            v-model="productForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入商品描述" />
        </el-form-item>

        <el-form-item label="商品图片">
          <div class="image-upload">
            <div class="main-image">
              <p class="label">主图</p>
              <div v-if="productForm.main_image" class="image-preview">
                <el-image :src="productForm.main_image" style="width: 100px; height: 100px;" fit="cover" />
                <el-button type="danger" size="small" @click="productForm.main_image = ''">移除</el-button>
              </div>
              <el-input v-else v-model="productForm.main_image" placeholder="输入图片URL" style="width: 200px;" />
            </div>
          </div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="排序">
              <el-input-number v-model="productForm.sort_order" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="库存预警阈值">
              <el-input-number v-model="productForm.stock_warning_threshold" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="商品状态">
              <el-switch v-model="productForm.status" :active-value="1" :inactive-value="0" active-text="上架" inactive-text="下架" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <div class="checkbox-group">
            <el-checkbox v-model="productForm.is_hot">热销商品</el-checkbox>
            <el-checkbox v-model="productForm.is_recommend">推荐商品</el-checkbox>
          </div>
        </el-form-item>

        <el-divider content-position="left">SKU规格管理</el-divider>
        <el-form-item label="SKU列表">
          <div class="sku-list">
            <div v-for="(sku, index) in productForm.skus" :key="index" class="sku-item">
              <el-row :gutter="12">
                <el-col :span="3">
                  <el-input v-model="sku.sku_code" placeholder="SKU编码" />
                </el-col>
                <el-col :span="4">
                  <el-input v-model="sku.spec_name" placeholder="规格名称(如:大杯)" />
                </el-col>
                <el-col :span="3">
                  <el-input-number v-model="sku.price" :min="0" :precision="2" placeholder="售价" style="width: 100%" />
                </el-col>
                <el-col :span="3">
                  <el-input-number v-model="sku.original_price" :min="0" :precision="2" placeholder="原价" style="width: 100%" />
                </el-col>
                <el-col :span="2">
                  <el-input-number v-model="sku.stock" :min="0" placeholder="库存" style="width: 100%" />
                </el-col>
                <el-col :span="3">
                  <el-input v-model="sku.image" placeholder="图片URL" />
                </el-col>
                <el-col :span="2">
                  <el-switch v-model="sku.status" :active-value="1" :inactive-value="0" />
                </el-col>
                <el-col :span="4">
                  <div class="sku-actions">
                    <el-button type="primary" link size="small" @click="openAttributeDialog(index)">
                      属性
                    </el-button>
                    <el-button type="danger" link size="small" @click="removeSKU(index)" :disabled="productForm.skus.length <= 1">
                      删除
                    </el-button>
                  </div>
                </el-col>
              </el-row>
            </div>
            <el-button type="dashed" style="width: 100%; margin-top: 12px;" @click="addSKU">
              <el-icon><Plus /></el-icon>添加SKU
            </el-button>
          </div>
        </el-form-item>

        <el-divider content-position="left">属性管理 (辣度等)</el-divider>
        <el-form-item label="属性列表">
          <div class="attribute-list">
            <div v-for="(attr, aIndex) in productForm.attributes" :key="aIndex" class="attribute-item">
              <div class="attribute-header">
                <el-input v-model="attr.name" placeholder="属性名称(如:辣度)" style="width: 200px; margin-right: 12px;" />
                <el-input-number v-model="attr.sort_order" :min="0" placeholder="排序" style="width: 120px; margin-right: 12px;" />
                <el-switch v-model="attr.status" :active-value="1" :inactive-value="0" style="margin-right: 12px;" />
                <el-button type="danger" link @click="removeAttribute(aIndex)">删除属性</el-button>
              </div>
              <div class="attribute-values">
                <div v-for="(val, vIndex) in attr.values" :key="vIndex" class="value-item">
                  <el-input v-model="val.value" placeholder="属性值(如:微辣)" style="width: 150px;" />
                  <el-input-number v-model="val.extra_price" :min="0" :precision="2" placeholder="加价" style="width: 120px;" />
                  <el-input-number v-model="val.stock" placeholder="库存(-1不限)" style="width: 120px;" />
                  <el-input-number v-model="val.sort_order" :min="0" placeholder="排序" style="width: 100px;" />
                  <el-switch v-model="val.status" :active-value="1" :inactive-value="0" />
                  <el-button type="danger" link @click="removeAttributeValue(aIndex, vIndex)" :disabled="attr.values.length <= 1">删除</el-button>
                </div>
                <el-button type="dashed" size="small" style="margin-top: 8px;" @click="addAttributeValue(aIndex)">
                  <el-icon><Plus /></el-icon>添加属性值
                </el-button>
              </div>
            </div>
            <el-button type="dashed" style="width: 100%; margin-top: 12px;" @click="addAttribute">
              <el-icon><Plus /></el-icon>添加属性
            </el-button>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchPriceDialogVisible" title="批量改价" width="500px">
      <el-form label-width="100px">
        <el-form-item label="改价方式">
          <el-radio-group v-model="batchPriceForm.price_type">
            <el-radio value="fixed">固定价格</el-radio>
            <el-radio value="percentage">按百分比</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="batchPriceForm.price_type === 'fixed' ? '新价格' : '调整比例'">
          <el-input-number
            v-model="batchPriceForm.price"
            :min="0"
            :precision="2"
            :step="batchPriceForm.price_type === 'percentage' ? 5 : 1"
            style="width: 200px" />
          <span v-if="batchPriceForm.price_type === 'percentage'" style="margin-left: 8px; color: #909399;">% （100为原价，90为9折）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchPriceDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="batchPriceLoading" @click="handleBatchPrice">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Money } from '@element-plus/icons-vue'
import {
  getProductList,
  createProduct,
  updateProduct,
  deleteProduct,
  copyProduct,
  batchUpdatePrice,
  updateStock
} from '@/api/product'

const loading = ref(false)
const submitLoading = ref(false)
const batchPriceLoading = ref(false)

const list = ref([])
const total = ref(0)
const categories = ref([
  { id: 1, name: '热菜' },
  { id: 2, name: '凉菜' },
  { id: 3, name: '主食' },
  { id: 4, name: '汤品' },
  { id: 5, name: '饮品' },
  { id: 6, name: '烧烤' }
])

const query = reactive({
  name: '',
  category_id: 0,
  status: null,
  page: 1,
  page_size: 10,
  store_id: 1
})

const selectedIds = ref([])
const selectAll = ref(false)
const isIndeterminate = ref(false)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const productFormRef = ref()

const emptySKU = () => ({
  id: null,
  sku_code: '',
  spec_name: '',
  price: 0,
  original_price: 0,
  stock: 0,
  image: '',
  status: 1,
  attribute_values: []
})

const emptyAttribute = () => ({
  id: null,
  name: '',
  sort_order: 0,
  status: 1,
  values: [{ id: null, value: '', extra_price: 0, stock: -1, sort_order: 0, status: 1 }]
})

const productForm = reactive({
  store_id: 1,
  category_id: null,
  name: '',
  description: '',
  main_image: '',
  images: '',
  sort_order: 0,
  status: 1,
  is_hot: false,
  is_recommend: false,
  stock_warning_threshold: 10,
  skus: [emptySKU()],
  attributes: []
})

const productRules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择分类', trigger: 'change' }]
}

const batchPriceDialogVisible = ref(false)
const batchPriceForm = reactive({
  store_id: 1,
  product_ids: [],
  price: 0,
  price_type: 'fixed'
})

function handleSelectionChange(selection) {
  selectedIds.value = selection.map(item => item.id)
  isIndeterminate.value = selection.length > 0 && selection.length < list.value.length
  selectAll.value = selection.length === list.value.length && list.value.length > 0
}

function handleSelectAll(val) {
  const table = document.querySelector('.el-table')
  if (val) {
    list.value.forEach(row => table?.__vueParentComponent?.toggleRowSelection?.(row, true))
  } else {
    list.value.forEach(row => table?.__vueParentComponent?.toggleRowSelection?.(row, false))
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getProductList(query)
    list.value = res.list
    total.value = res.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.name = ''
  query.category_id = 0
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(productForm, {
      store_id: 1,
      category_id: row.category_id,
      name: row.name,
      description: row.description || '',
      main_image: row.main_image || '',
      images: row.images || '',
      sort_order: row.sort_order,
      status: row.status,
      is_hot: row.is_hot,
      is_recommend: row.is_recommend,
      stock_warning_threshold: row.stock_warning_threshold || 10,
      skus: row.skus?.map(sku => ({
        id: sku.id,
        sku_code: sku.sku_code,
        spec_name: sku.spec_name,
        price: sku.price,
        original_price: sku.original_price,
        stock: sku.stock,
        image: sku.image,
        status: sku.status,
        attribute_values: sku.attribute_values?.map(av => ({
          attribute_id: av.attribute_id,
          value_id: av.value_id
        })) || []
      })) || [emptySKU()],
      attributes: row.attributes?.map(attr => ({
        id: attr.id,
        name: attr.name,
        sort_order: attr.sort_order,
        status: attr.status,
        values: attr.values?.map(val => ({
          id: val.id,
          value: val.value,
          extra_price: val.extra_price,
          stock: val.stock,
          sort_order: val.sort_order,
          status: val.status
        })) || []
      })) || []
    })
  } else {
    productForm.store_id = 1
    productForm.category_id = null
    productForm.name = ''
    productForm.description = ''
    productForm.main_image = ''
    productForm.images = ''
    productForm.sort_order = 0
    productForm.status = 1
    productForm.is_hot = false
    productForm.is_recommend = false
    productForm.stock_warning_threshold = 10
    productForm.skus = [emptySKU()]
    productForm.attributes = []
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

async function handleStatusChange(row) {
  try {
    await updateProduct(row.id, { status: row.status })
    ElMessage.success(row.status === 1 ? '商品已上架' : '商品已下架')
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除商品"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteProduct(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleCopy(row) {
  try {
    await copyProduct({
      store_id: 1,
      product_id: row.id,
      new_name: row.name + ' (副本)'
    })
    ElMessage.success('复制成功')
    fetchList()
  } catch (e) {
    console.error(e)
  }
}

async function handleSubmit() {
  try {
    await productFormRef.value.validate()
    submitLoading.value = true

    if (isEdit.value) {
      await updateProduct(editId.value, productForm)
      ElMessage.success('更新成功')
    } else {
      await createProduct(productForm)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

function openBatchPriceDialog() {
  batchPriceForm.product_ids = selectedIds.value
  batchPriceForm.price = 0
  batchPriceForm.price_type = 'fixed'
  batchPriceDialogVisible.value = true
}

async function handleBatchPrice() {
  try {
    batchPriceLoading.value = true
    await batchUpdatePrice(batchPriceForm)
    ElMessage.success('批量改价成功')
    batchPriceDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    batchPriceLoading.value = false
  }
}

function addSKU() {
  productForm.skus.push(emptySKU())
}

function removeSKU(index) {
  productForm.skus.splice(index, 1)
}

function addAttribute() {
  productForm.attributes.push(emptyAttribute())
}

function removeAttribute(index) {
  productForm.attributes.splice(index, 1)
}

function addAttributeValue(attrIndex) {
  productForm.attributes[attrIndex].values.push({
    id: null,
    value: '',
    extra_price: 0,
    stock: -1,
    sort_order: 0,
    status: 1
  })
}

function removeAttributeValue(attrIndex, valIndex) {
  productForm.attributes[attrIndex].values.splice(valIndex, 1)
}

function openAttributeDialog(skuIndex) {
  ElMessage.info('属性关联功能开发中')
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped lang="scss">
.products-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .product-name {
    display: flex;
    align-items: center;

    .el-tag {
      transform: scale(0.85);
      transform-origin: left center;
    }
  }

  .price-range .price {
    color: #f56c6c;
    font-weight: 600;
  }

  .low-stock {
    color: #f56c6c;
    font-weight: 600;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .sku-list {
    width: 100%;

    .sku-item {
      padding: 12px;
      background: #f5f7fa;
      border-radius: 8px;
      margin-bottom: 12px;

      .sku-actions {
        display: flex;
        justify-content: center;
        gap: 8px;
      }
    }
  }

  .attribute-list {
    width: 100%;

    .attribute-item {
      padding: 16px;
      background: #f5f7fa;
      border-radius: 8px;
      margin-bottom: 16px;

      .attribute-header {
        margin-bottom: 12px;
        display: flex;
        align-items: center;
      }

      .attribute-values {
        padding-left: 20px;

        .value-item {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 8px;
        }
      }
    }
  }

  .checkbox-group {
    display: flex;
    gap: 24px;
  }
}
</style>
