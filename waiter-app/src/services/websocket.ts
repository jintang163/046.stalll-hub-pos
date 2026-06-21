import { useWebSocketStore } from '../store/websocket'

export const useWebSocketService = () => {
  const wsStore = useWebSocketStore()

  const connect = () => {
    wsStore.connect()
  }

  const disconnect = () => {
    wsStore.disconnect()
  }

  return {
    connect,
    disconnect,
    isConnected: () => wsStore.isConnected,
    pendingCalls: () => wsStore.pendingCalls
  }
}
