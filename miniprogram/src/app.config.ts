export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/product/detail',
    'pages/cart/index',
    'pages/order/confirm',
    'pages/order/list',
    'pages/order/detail',
    'pages/user/index',
    'pages/store/select'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#667eea',
    navigationBarTitleText: '大排档',
    navigationBarTextStyle: 'white',
    backgroundColor: '#f5f7fa'
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
