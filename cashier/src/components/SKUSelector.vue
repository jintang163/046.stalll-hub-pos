<template>
  <el-dialog 
    v-model="visible" 
    title="选择规格" 
    width="480px"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <div v-if="product" class="sku-selector">
      <div class="product-info">
        <el-image 
          :src="product.image || defaultImg" 
          class="product-image"
          fit="cover"
        />
        <div class="product-detail">
          <h3 class="product-name">{{ product.name }}</h3>
          <p class="product-desc">{{ product.description || '暂无描述' }}</p>
        </div>
      </div>

      <div class="sku-section" v-if="product.skus && product.skus.length > 0">
        <div class="section-title">规格选择</div>
        <div class="sku-list">
          <el-button
            v-for="sku in availableSKUs"
            :key="sku.id"
            :type="selectedSKU?.id === sku.id ? 'primary' : 'default'"
            :disabled="!sku.status || sku.stock === 0"
            class="sku-btn"
            @click="selectSKU(sku)"
          >
            <span>{{ sku.spec_name }}</span>
            <span class="sku-price">¥{{ sku.price }}</span>
            <span v-if="sku.stock === 0" class="out-of-stock">售罄</span>
            <span v-else-if="sku.stock > 0 && sku.stock <= 10" class="low-stock">
              仅剩{{ sku.stock }}份
            </span>
          </el-button>
        </div>
      </div>

      <div 
        class="attr-section" 
        v-for="attr in availableAttrs" 
        :key="attr.id"
      >
        <div class="section-title">{{ attr.name }}</div>
        <div class="attr-list">
          <el-button
            v-for="val in attr.values"
            :key="val.id"
            :type="isAttrSelected(attr.id, val.id) ? 'primary' : 'default'"
            :disabled="!val.status || val.stock === 0"
            class="attr-btn"
            @click="toggleAttr(attr, val)"
          >
            <span>{{ val.value }}</span>
            <span v-if="val.extra_price > 0" class="extra-price">+¥{{ val.extra_price }}</span>
          </el-button>
        </div>
      </div>

      <div class="quantity-section">
        <div class="section-title">数量</div>
        <el-input-number 
          v-model="quantity" 
          :min="1" 
          :max="maxQuantity"
          size="large"
          @change="calculatePrice"
        />
      </div>

      <div class="price-preview">
        <span class="label">单价:</span>
        <span class="price">¥{{ unitPrice.toFixed(2) }}</span>
        <span class="label" style="margin-left: 20px;">小计:</span>
        <span class="price total">¥{{ subtotal.toFixed(2) }}</span>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button 
        type="primary" 
        :disabled="!canAdd"
        @click="handleAdd"
      >
        加入购物车
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: Boolean,
  product: Object
})

const emit = defineEmits(['update:modelValue', 'add'])

const defaultImg = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23f0f0f0" width="100" height="100"/%3E%3Ctext x="50" y="55" text-anchor="middle" fill="%23999"%3E暂无图%3C/text%3E%3C/svg%3E'

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const selectedSKU = ref(null)
const selectedAttrs = ref([])
const quantity = ref(1)

const availableSKUs = computed(() => {
  return (props.product?.skus || []).filter(s => s.status)
})

const availableAttrs = computed(() => {
  return (props.product?.attributes || []).filter(a => a.status)
})

const unitPrice = computed(() => {
  let price = selectedSKU.value?.price || 0
  selectedAttrs.value.forEach(a => {
    price += a.extra_price || 0
  })
  return price
})

const subtotal = computed(() => unitPrice.value * quantity.value)

const maxQuantity = computed(() => {
  if (!selectedSKU.value) return 99
  if (selectedSKU.value.stock === -1) return 99
  return Math.min(selectedSKU.value.stock, 99)
})

const canAdd = computed(() => {
  return selectedSKU.value && quantity.value > 0 && quantity.value <= maxQuantity.value
})

const isAttrSelected = (attrId, valueId) => {
  return selectedAttrs.value.some(a => a.attr_id === attrId && a.value_id === valueId)
}

const selectSKU = (sku) => {
  selectedSKU.value = sku
}

const toggleAttr = (attr, val) => {
  const idx = selectedAttrs.value.findIndex(
    a => a.attr_id === attr.id && a.value_id === val.id
  )
  
  if (idx !== -1) {
    selectedAttrs.value.splice(idx, 1)
  } else {
    selectedAttrs.value = selectedAttrs.value.filter(a => a.attr_id !== attr.id)
    selectedAttrs.value.push({
      attr_id: attr.id,
      attr_name: attr.name,
      value_id: val.id,
      value_name: val.value,
      extra_price: val.extra_price
    })
  }
}

const calculatePrice = () => {}

const handleAdd = () => {
  if (!canAdd.value) return
  
  emit('add', {
    product: props.product,
    sku: selectedSKU.value,
    attributes: selectedAttrs.value,
    quantity: quantity.value
  })
  
  ElMessage.success('已加入购物车')
  visible.value = false
}

const handleClosed = () => {
  selectedSKU.value = null
  selectedAttrs.value = []
  quantity.value = 1
}

watch(() => props.product, () => {
  if (props.product && availableSKUs.value.length > 0) {
    selectedSKU.value = availableSKUs.value[0]
  }
}, { immediate: true })
</script>

<style lang="scss" scoped>
.sku-selector {
  .product-info {
    display: flex;
    gap: 16px;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid #eee;
  }
  
  .product-image {
    width: 80px;
    height: 80px;
    border-radius: 8px;
    flex-shrink: 0;
  }
  
  .product-detail {
    flex: 1;
  }
  
  .product-name {
    margin: 0 0 8px;
    font-size: 18px;
    font-weight: 600;
  }
  
  .product-desc {
    margin: 0;
    color: #909399;
    font-size: 14px;
  }
  
  .section-title {
    font-weight: 600;
    margin-bottom: 12px;
    color: #303133;
  }
  
  .sku-section,
  .attr-section,
  .quantity-section {
    margin-bottom: 20px;
  }
  
  .sku-list,
  .attr-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .sku-btn,
  .attr-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    min-width: 90px;
    padding: 8px 16px;
    
    .sku-price {
      color: #f56c6c;
      font-weight: 600;
    }
    
    .extra-price {
      color: #f56c6c;
      font-size: 12px;
    }
    
    .out-of-stock {
      color: #909399;
      font-size: 12px;
    }
    
    .low-stock {
      color: #e6a23c;
      font-size: 12px;
    }
  }
  
  .price-preview {
    padding: 16px;
    background: #f5f7fa;
    border-radius: 8px;
    display: flex;
    align-items: center;
    
    .label {
      color: #606266;
    }
    
    .price {
      color: #f56c6c;
      font-size: 18px;
      font-weight: 600;
      
      &.total {
        font-size: 24px;
      }
    }
  }
}
</style>
