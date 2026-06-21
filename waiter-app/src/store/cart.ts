import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { OrderItem } from '../types'

export interface CartItem {
  product_id: number
  product_name: string
  sku_id: number
  sku_name: string
  image: string
  price: string
  quantity: number
  attribute_values?: any[]
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const tableId = ref<number>(0)
  const tableNo = ref<string>('')
  const remark = ref<string>('')

  const totalCount = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  const totalAmount = computed(() => {
    return items.value.reduce((sum, item) => sum + parseFloat(item.price) * item.quantity, 0)
  })

  const addItem = (product: any, sku: any, quantity = 1, attributeValues?: any[]) => {
    const existingIndex = items.value.findIndex(item => item.sku_id === sku.id)
    if (existingIndex >= 0) {
      items.value[existingIndex].quantity += quantity
    } else {
      items.value.push({
        product_id: product.id,
        product_name: product.name,
        sku_id: sku.id,
        sku_name: sku.spec_name || sku.name,
        image: sku.image || product.main_image || '',
        price: sku.price,
        quantity,
        attribute_values: attributeValues
      })
    }
  }

  const removeItem = (skuId: number) => {
    const index = items.value.findIndex(item => item.sku_id === skuId)
    if (index >= 0) {
      items.value.splice(index, 1)
    }
  }

  const updateQuantity = (skuId: number, quantity: number) => {
    const item = items.value.find(item => item.sku_id === skuId)
    if (item) {
      if (quantity <= 0) {
        removeItem(skuId)
      } else {
        item.quantity = quantity
      }
    }
  }

  const clearCart = () => {
    items.value = []
    remark.value = ''
  }

  const setTable = (id: number, no: string) => {
    tableId.value = id
    tableNo.value = no
  }

  const clearTable = () => {
    tableId.value = 0
    tableNo.value = ''
  }

  return {
    items,
    tableId,
    tableNo,
    remark,
    totalCount,
    totalAmount,
    addItem,
    removeItem,
    updateQuantity,
    clearCart,
    setTable,
    clearTable
  }
})
