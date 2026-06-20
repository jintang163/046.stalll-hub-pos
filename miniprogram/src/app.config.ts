export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/product/detail',
    'pages/cart/index',
    'pages/order/confirm',
    'pages/order/list',
    'pages/order/detail',
    'pages/user/index',
    'pages/coupon/index',
    'pages/store/select',
    'pages/table/scan',
    'pages/table/reserve',
    'pages/voice/index'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#667eea',
    navigationBarTitleText: '大排档',
    navigationBarTextStyle: 'white',
    backgroundColor: '#f5f7fa'
  },
  plugins: {
    'WechatSI': {
      'version': '0.3.5',
      'provider': 'wx069ba97219f66d99'
    }
  },
  tabBar: {
    color: '#999',
    selectedColor: '#667eea',
    backgroundColor: '#fff',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/index/index',
        text: '点餐'
      },
      {
        pagePath: 'pages/order/list',
        text: '订单'
      },
      {
        pagePath: 'pages/user/index',
        text: '我的'
      }
    ]
  }
})
