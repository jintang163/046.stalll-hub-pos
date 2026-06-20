import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useStallStore = defineStore('stall', () => {
  const stalls = ref([])
  const currentStall = ref(null)
  const stallMode = ref(false)
  const loading = ref(false)

  const currentStallId = computed(() => currentStall.value?.id || null)

  async function loadStalls() {
    loading.value = true
    try {
      if (window.electronAPI) {
        const result = await window.electronAPI.invoke('db:getStalls')
        stalls.value = result || []
      }
    } catch (error) {
      console.error('加载摊位列表失败:', error)
    } finally {
      loading.value = false
    }
  }

  async function saveStalls(stallList) {
    try {
      if (window.electronAPI) {
        await window.electronAPI.invoke('db:saveStalls', stallList)
        stalls.value = stallList
      }
    } catch (error) {
      console.error('保存摊位列表失败:', error)
    }
  }

  function setCurrentStall(stall) {
    currentStall.value = stall
    stallMode.value = !!stall
  }

  function toggleStallMode() {
    stallMode.value = !stallMode.value
    if (!stallMode.value) {
      currentStall.value = null
    }
  }

  async function getProductsByStall(stallId) {
    try {
      if (window.electronAPI) {
        return await window.electronAPI.invoke('db:getProductsByStall', stallId)
      }
      return []
    } catch (error) {
      console.error('获取摊位商品失败:', error)
      return []
    }
  }

  async function getStallDailySales(stallId, date) {
    try {
      if (window.electronAPI) {
        return await window.electronAPI.invoke('db:getStallDailySales', stallId, date)
      }
      return { orderCount: 0, totalAmount: 0, stallAmount: 0, platformAmount: 0 }
    } catch (error) {
      console.error('获取摊位日销售数据失败:', error)
      return { orderCount: 0, totalAmount: 0, stallAmount: 0, platformAmount: 0 }
    }
  }

  return {
    stalls,
    currentStall,
    currentStallId,
    stallMode,
    loading,
    loadStalls,
    saveStalls,
    setCurrentStall,
    toggleStallMode,
    getProductsByStall,
    getStallDailySales
  }
})
