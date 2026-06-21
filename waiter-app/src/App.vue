<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useUserStore } from './store/user'
import { usePushService } from './services/push'
import { useWebSocketService } from './services/websocket'

onLaunch(() => {
  console.log('App Launch')
  const userStore = useUserStore()
  userStore.restoreFromStorage()
  
  const pushService = usePushService()
  pushService.init()
  
  if (userStore.token) {
    const wsService = useWebSocketService()
    wsService.connect()
  }
})

onShow(() => {
  console.log('App Show')
})

onHide(() => {
  console.log('App Hide')
})
</script>

<style lang="scss">
@import './styles/global.scss';
</style>
