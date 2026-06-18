<template>
  <div id="app">
    <router-view />
    <sync-progress />
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useSyncStore } from '@/store/sync'
import { useNSQStore } from '@/store/nsq'
import { useOrderStore } from '@/store/order'
import SyncProgress from '@/components/SyncProgress.vue'

const syncStore = useSyncStore()
const nsqStore = useNSQStore()
const orderStore = useOrderStore()

onMounted(() => {
  syncStore.init()
  nsqStore.init()
  orderStore.init()
  
  window.electronAPI && window.ipcRenderer.on('sync:progress', (event, progress) => {
    syncStore.updateProgress(progress)
  })
})

onUnmounted(() => {
  nsqStore.destroy()
  window.ipcRenderer && window.ipcRenderer.removeAllListeners('sync:progress')
})
</script>

<style lang="scss">
#app {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
</style>
