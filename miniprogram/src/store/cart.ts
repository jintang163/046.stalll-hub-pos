import { create } from 'zustand'
import Taro from '@tarojs/taro'
import type { Product, SKU, AttributeValue } from '../services/product'

export interface CartItem {
  id: string
  product_id: number
  product_name: string
  product_image: string
  sku_id: number
  sku_name: string
  attribute_ids: number[]
  attribute_names: string[]
  price: number
  quantity: number
  subtotal: number
  remark?: string
}

interface CartStore {
  items: CartItem[]
  storeId: number
  tableNo: string
  remark: string
  couponId: number | null
  couponDiscount: number

  total: () => number
  actualTotal: () => number
  itemCount: () => number
  addItem: (product: Product, sku: SKU, attributes: { attr_id: number; attr_name: string; value: AttributeValue }[], quantity: number) => void
  updateQuantity: (id: string, quantity: number) => void
  removeItem: (id: string) => void
  setStoreId: (id: number) => void
  setTableNo: (no: string) => void
  setRemark: (remark: string) => void
  setCoupon: (couponId: number | null, discount: number) => void
  clear: () => void
}

const generateItemKey = (productId: number, skuId: number, attributeIds: number[]) => {
  const sortedAttrs = [...attributeIds].sort().join(',')
  return `${productId}-${skuId}-[${sortedAttrs}]`
}

const getStoredCart = () => {
  try {
    const stored = Taro.getStorageSync('cart')
    return stored || { items: [], storeId: 1, tableNo: '', remark: '', couponId: null, couponDiscount: 0 }
  } catch {
    return { items: [], storeId: 1, tableNo: '', remark: '', couponId: null, couponDiscount: 0 }
  }
}

const saveCart = (state: Partial<CartStore>) => {
  try {
    const data = {
      items: state.items,
      storeId: state.storeId,
      tableNo: state.tableNo,
      remark: state.remark,
      couponId: state.couponId,
      couponDiscount: state.couponDiscount
    }
    Taro.setStorageSync('cart', data)
  } catch {}
}

const initial = getStoredCart()

export const useCartStore = create<CartStore>((set, get) => ({
  items: initial.items,
  storeId: initial.storeId,
  tableNo: initial.tableNo,
  remark: initial.remark,
  couponId: initial.couponId,
  couponDiscount: initial.couponDiscount,

  total: () => {
    return get().items.reduce((sum, item) => sum + item.subtotal, 0)
  },

  actualTotal: () => {
    const state = get()
    return Math.max(0, state.total() - state.couponDiscount)
  },

  itemCount: () => {
    return get().items.reduce((sum, item) => sum + item.quantity, 0)
  },

  addItem: (product, sku, attributes, quantity) => {
    const attributeIds = attributes.map(a => a.value.id)
    const attributeNames = attributes.map(a => `${a.attr_name}:${a.value.value}`)
    const extraPrice = attributes.reduce((sum, a) => sum + (a.value.extra_price || 0), 0)
    const unitPrice = (sku.price || 0) + extraPrice
    const id = generateItemKey(product.id, sku.id, attributeIds)

    set(state => {
      const existingIndex = state.items.findIndex(item => item.id === id)
      let newItems

      if (existingIndex !== -1) {
        newItems = [...state.items]
        const existing = newItems[existingIndex]
        newItems[existingIndex] = {
          ...existing,
          quantity: existing.quantity + quantity,
          subtotal: existing.price * (existing.quantity + quantity)
        }
      } else {
        newItems = [
          ...state.items,
          {
            id,
            product_id: product.id,
            product_name: product.name,
            product_image: product.image,
            sku_id: sku.id,
            sku_name: sku.spec_name,
            attribute_ids: attributeIds,
            attribute_names: attributeNames,
            price: unitPrice,
            quantity,
            subtotal: unitPrice * quantity
          }
        ]
      }

      const newState = { ...state, items: newItems }
      saveCart(newState)
      return newState
    })
  },

  updateQuantity: (id, quantity) => {
    if (quantity <= 0) {
      get().removeItem(id)
      return
    }

    set(state => {
      const newItems = state.items.map(item =>
        item.id === id
          ? { ...item, quantity, subtotal: item.price * quantity }
          : item
      )
      const newState = { ...state, items: newItems }
      saveCart(newState)
      return newState
    })
  },

  removeItem: (id) => {
    set(state => {
      const newItems = state.items.filter(item => item.id !== id)
      const newState = { ...state, items: newItems }
      saveCart(newState)
      return newState
    })
  },

  setStoreId: (id) => {
    set(state => {
      const newState = { ...state, storeId: id }
      saveCart(newState)
      return newState
    })
  },

  setTableNo: (no) => {
    set(state => {
      const newState = { ...state, tableNo: no }
      saveCart(newState)
      return newState
    })
  },

  setRemark: (remark) => {
    set(state => {
      const newState = { ...state, remark }
      saveCart(newState)
      return newState
    })
  },

  setCoupon: (couponId, discount) => {
    set(state => {
      const newState = { ...state, couponId, couponDiscount: discount || 0 }
      saveCart(newState)
      return newState
    })
  },

  clear: () => {
    set({
      items: [],
      tableNo: '',
      remark: '',
      couponId: null,
      couponDiscount: 0
    })
    Taro.removeStorageSync('cart')
  }
}))
