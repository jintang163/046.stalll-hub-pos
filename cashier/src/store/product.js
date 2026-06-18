import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useProductStore = defineStore('product', () => {
  const products = ref([])
  const categories = ref([])
  const currentCategory = ref('all')
  const loading = ref(false)

  const filteredProducts = computed(() => {
    let result = products.value.filter(p => p.status)
    if (currentCategory.value !== 'all') {
      result = result.filter(p => p.category_id === currentCategory.value)
    }
    return result
  })

  const productsByCategory = computed(() => {
    const groups = {}
    categories.value.forEach(cat => {
      groups[cat.id] = products.value.filter(p => 
        p.category_id === cat.id && p.status
      )
    })
    return groups
  })

  const getProductById = (id) => {
    return products.value.find(p => p.id === id)
  }

  const getSKUById = (productId, skuId) => {
    const product = getProductById(productId)
    if (product && product.skus) {
      return product.skus.find(s => s.id === skuId)
    }
    return null
  }

  const loadProducts = async () => {
    if (!window.electronAPI) return
    loading.value = true
    try {
      const data = await window.electronAPI.db.getProducts()
      products.value = data.map(p => ({
        ...p,
        skus: [],
        attributes: []
      }))
      
      for (const p of products.value) {
        const detail = await window.electronAPI.db.getProductById(p.id)
        if (detail) {
          p.skus = detail.skus || []
          p.attributes = detail.attributes || []
        }
      }
    } finally {
      loading.value = false
    }
  }

  const loadCategories = async () => {
    if (!window.electronAPI) return
    try {
      categories.value = await window.electronAPI.db.getCategories()
    } catch (e) {
      console.error('加载分类失败:', e)
    }
  }

  const updateProduct = async (product) => {
    const idx = products.value.findIndex(p => p.id === product.id)
    if (idx !== -1) {
      products.value[idx] = { ...products.value[idx], ...product }
    } else {
      products.value.push(product)
    }
  }

  const updateStock = async (skuId, stock) => {
    if (!window.electronAPI) return
    await window.electronAPI.db.updateStock(skuId, stock)
    for (const p of products.value) {
      const sku = p.skus?.find(s => s.id === skuId)
      if (sku) {
        sku.stock = stock
        break
      }
    }
  }

  const removeProduct = async (productId) => {
    products.value = products.value.filter(p => p.id !== productId)
    if (window.electronAPI) {
      await window.electronAPI.db.deleteProduct(productId)
    }
  }

  const setCategory = (catId) => {
    currentCategory.value = catId
  }

  return {
    products,
    categories,
    currentCategory,
    loading,
    filteredProducts,
    productsByCategory,
    getProductById,
    getSKUById,
    loadProducts,
    loadCategories,
    updateProduct,
    updateStock,
    removeProduct,
    setCategory
  }
})
