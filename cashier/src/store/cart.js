import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])
  const tableNo = ref('')
  const remark = ref('')
  const memberId = ref(null)
  const memberName = ref('')
  const discountAmount = ref(0)

  const total = computed(() => {
    return items.value.reduce((sum, item) => sum + item.subtotal, 0)
  })

  const actualTotal = computed(() => {
    return Math.max(0, total.value - discountAmount.value)
  })

  const itemCount = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  const getItemKey = (productId, skuId, attributeIds = []) => {
    const sortedAttrs = [...attributeIds].sort().join(',')
    return `${productId}-${skuId}-[${sortedAttrs}]`
  }

  const findItem = (productId, skuId, attributeIds = []) => {
    const key = getItemKey(productId, skuId, attributeIds)
    return items.value.find(item => item._key === key)
  }

  const addItem = (product, sku, selectedAttributes = [], quantity = 1) => {
    const attributeIds = selectedAttributes.map(a => a.value_id)
    const attributeNames = selectedAttributes.map(a => `${a.attr_name}:${a.value_name}`)
    const key = getItemKey(product.id, sku.id, attributeIds)
    
    const extraPrice = selectedAttributes.reduce((sum, a) => sum + (a.extra_price || 0), 0)
    const unitPrice = (sku.price || 0) + extraPrice

    const existing = items.value.find(item => item._key === key)
    
    if (existing) {
      existing.quantity += quantity
      existing.subtotal = existing.price * existing.quantity
    } else {
      items.value.push({
        _key: key,
        product_id: product.id,
        product_name: product.name,
        product_image: product.image,
        sku_id: sku.id,
        sku_name: sku.spec_name,
        attribute_ids: attributeIds,
        attribute_names: attributeNames,
        price: unitPrice,
        quantity,
        subtotal: unitPrice * quantity,
        remark: ''
      })
    }
  }

  const updateQuantity = (index, quantity) => {
    if (quantity <= 0) {
      removeItem(index)
      return
    }
    const item = items.value[index]
    if (item) {
      item.quantity = quantity
      item.subtotal = item.price * quantity
    }
  }

  const removeItem = (index) => {
    items.value.splice(index, 1)
  }

  const updateItemRemark = (index, remark) => {
    const item = items.value[index]
    if (item) {
      item.remark = remark
    }
  }

  const clear = () => {
    items.value = []
    tableNo.value = ''
    remark.value = ''
    memberId.value = null
    memberName.value = ''
    discountAmount.value = 0
  }

  const applyDiscount = (amount) => {
    discountAmount.value = Math.max(0, Math.min(amount, total.value))
  }

  return {
    items,
    tableNo,
    remark,
    memberId,
    memberName,
    discountAmount,
    total,
    actualTotal,
    itemCount,
    addItem,
    updateQuantity,
    removeItem,
    updateItemRemark,
    clear,
    applyDiscount
  }
})
