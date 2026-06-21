export const usePushService = () => {
  const init = () => {
    // #ifdef APP-PLUS
    initUniPush()
    // #endif
  }

  const initUniPush = () => {
    // #ifdef APP-PLUS
    try {
      const pinf = plus.push.getClientInfo()
      console.log('[Push] clientid:', pinf.clientid)

      plus.push.addEventListener('click', (msg: any) => {
        console.log('[Push] Click:', msg)
        handlePushClick(msg)
      }, false)

      plus.push.addEventListener('receive', (msg: any) => {
        console.log('[Push] Receive:', msg)
        handlePushReceive(msg)
      }, false)
    } catch (e) {
      console.error('[Push] Init failed:', e)
    }
    // #endif
  }

  const handlePushReceive = (msg: any) => {
    uni.showModal({
      title: msg.title || '新消息',
      content: msg.content || '',
      showCancel: false
    })
  }

  const handlePushClick = (msg: any) => {
    const payload = msg.payload
    if (payload) {
      try {
        const data = typeof payload === 'string' ? JSON.parse(payload) : payload
        if (data.type === 'new_order') {
          uni.switchTab({ url: '/pages/orders/index' })
        } else if (data.type === 'call_waiter') {
          uni.switchTab({ url: '/pages/calls/index' })
        }
      } catch (e) {}
    }
  }

  const setBadge = (count: number) => {
    // #ifdef APP-PLUS
    plus.runtime.setBadgeNumber(count)
    // #endif
  }

  return {
    init,
    setBadge
  }
}
