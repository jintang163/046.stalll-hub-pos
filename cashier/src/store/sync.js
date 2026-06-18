import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { syncProducts, syncCategories, syncAllProducts, getSyncCount } from '@/api/sync'
import { useProductStore } from './product'

export const useSyncStore = defineStore('sync', () => {
  const progress = ref({ percent: 0, status: 'idle', message: '' })
  const lastSyncTime = ref(null)
  const lastSyncID = ref(0)
  const storeID = ref(1)

  const isSyncing = computed(() => progress.value.status === 'syncing')

  const init = async () => {
    if (!window.electronAPI) return
    try {
      lastSyncTime.value = await window.electronAPI.db.getLastSyncTime()
      lastSyncID.value = await window.electronAPI.db.getLastSyncID()
      storeID.value = await window.electronAPI.app.getStoreID()
    } catch (e) {
      console.error('初始化同步信息失败:', e)
    }
  }

  const updateProgress = (data) => {
    progress.value = { ...progress.value, ...data }
  }

  const setProgress = async (data) => {
    if (!window.electronAPI) return
    await window.electronAPI.sync.setProgress(data)
    updateProgress(data)
  }

  const fullSync = async () => {
    if (isSyncing.value) return
    
    const productStore = useProductStore()
    
    try {
      await setProgress({ percent: 0, status: 'syncing', message: '开始全量同步...' })
      
      await setProgress({ percent: 5, message: '获取商品总数...' })
      const countData = await getSyncCount(storeID.value)
      const total = countData.total || 1
      
      await setProgress({ percent: 10, message: '同步分类数据...' })
      const categories = await syncCategories(storeID.value)
      if (window.electronAPI) {
        await window.electronAPI.db.saveCategories(categories)
      }
      
      await setProgress({ percent: 20, message: '同步商品数据...' })
      const allProducts = await syncAllProducts(storeID.value)
      
      const batchSize = 50
      const batches = Math.ceil(allProducts.length / batchSize)
      
      for (let i = 0; i < batches; i++) {
        const batch = allProducts.slice(i * batchSize, (i + 1) * batchSize)
        if (window.electronAPI) {
          await window.electronAPI.db.saveProducts(batch)
        }
        
        const percent = 20 + Math.floor(((i + 1) / batches) * 70)
        await setProgress({ 
          percent, 
          message: `同步商品中... ${Math.min((i + 1) * batchSize, allProducts.length)}/${allProducts.length}` 
        })
      }
      
      const maxID = allProducts.length > 0 ? Math.max(...allProducts.map(p => p.id)) : 0
      lastSyncID.value = maxID
      lastSyncTime.value = new Date().toISOString()
      
      if (window.electronAPI) {
        await window.electronAPI.db.setLastSyncID(maxID)
        await window.electronAPI.db.setLastSyncTime(lastSyncTime.value)
      }
      
      await productStore.loadCategories()
      await productStore.loadProducts()
      
      await setProgress({ percent: 100, status: 'completed', message: '同步完成' })
      ElMessage.success(`成功同步 ${allProducts.length} 个商品`)
      
      setTimeout(() => {
        setProgress({ percent: 0, status: 'idle', message: '' })
      }, 2000)
      
    } catch (e) {
      console.error('全量同步失败:', e)
      await setProgress({ percent: 0, status: 'error', message: `同步失败: ${e.message}` })
      ElMessage.error('同步失败: ' + e.message)
    }
  }

  const incrementalSync = async () => {
    if (isSyncing.value) return
    
    const productStore = useProductStore()
    
    try {
      await setProgress({ percent: 0, status: 'syncing', message: '开始增量同步...' })
      
      let lastID = lastSyncID.value || 0
      let hasMore = true
      let totalSynced = 0
      const limit = 100
      
      const countData = await getSyncCount(storeID.value)
      const total = countData.total - lastID
      
      if (total <= 0) {
        await setProgress({ percent: 100, status: 'completed', message: '已是最新数据' })
        setTimeout(() => setProgress({ percent: 0, status: 'idle', message: '' }), 2000)
        return
      }
      
      while (hasMore) {
        await setProgress({ 
          percent: Math.floor((totalSynced / total) * 90), 
          message: `同步中... ${totalSynced}/${total}` 
        })
        
        const result = await syncProducts(storeID.value, lastID, limit)
        
        if (result.products && result.products.length > 0) {
          if (window.electronAPI) {
            await window.electronAPI.db.saveProducts(result.products)
          }
          
          for (const p of result.products) {
            productStore.updateProduct(p)
          }
          
          totalSynced += result.products.length
          lastID = result.last_id
          lastSyncID.value = lastID
          
          if (window.electronAPI) {
            await window.electronAPI.db.setLastSyncID(lastID)
          }
        }
        
        hasMore = result.has_more
      }
      
      lastSyncTime.value = new Date().toISOString()
      if (window.electronAPI) {
        await window.electronAPI.db.setLastSyncTime(lastSyncTime.value)
      }
      
      await setProgress({ percent: 100, status: 'completed', message: `同步完成，共更新 ${totalSynced} 个商品` })
      ElMessage.success(`成功同步 ${totalSynced} 个商品`)
      
      setTimeout(() => setProgress({ percent: 0, status: 'idle', message: '' }), 2000)
      
    } catch (e) {
      console.error('增量同步失败:', e)
      await setProgress({ percent: 0, status: 'error', message: `同步失败: ${e.message}` })
      ElMessage.error('同步失败: ' + e.message)
    }
  }

  return {
    progress,
    lastSyncTime,
    lastSyncID,
    storeID,
    isSyncing,
    init,
    updateProgress,
    fullSync,
    incrementalSync
  }
})
