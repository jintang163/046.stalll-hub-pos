'use strict';

require('dotenv').config();

const nsq = require('nsqjs');
const config = require('../src/config');

const sampleOrders = [
  {
    order_no: 'ORDER_TEST_001',
    store_id: 1,
    order_data: {
      orderId: 1001,
      orderNo: 'ORDER_TEST_001',
      storeId: 1,
      storeName: '测试门店',
      tableNo: 'A01',
      orderType: 'dine_in',
      totalAmount: '126.00',
      payAmount: '126.00',
      payMethod: 'wechat',
      payStatus: 1,
      pointsEarned: 12,
      createdAt: new Date().toLocaleString(),
      items: [
        { productName: '红烧肉', skuName: '大份', categoryName: '热菜', quantity: 1, price: '58.00', subtotal: '58.00' },
        { productName: '宫保鸡丁', skuName: '标准', categoryName: '热菜', quantity: 1, price: '38.00', subtotal: '38.00' },
        { productName: '凉拌黄瓜', skuName: '', categoryName: '凉菜', quantity: 1, price: '12.00', subtotal: '12.00' },
        { productName: '可乐', skuName: '中杯', categoryName: '饮品', quantity: 2, price: '6.00', subtotal: '12.00' },
        { productName: '米饭', skuName: '', categoryName: '主食', quantity: 2, price: '3.00', subtotal: '6.00' },
      ],
    },
    timestamp: Date.now(),
  },
];

async function publishMessage() {
  console.log('========================================');
  console.log('  NSQ 订单消息发布工具');
  console.log('========================================\n');

  const writer = new nsq.Writer(
    config.nsq.nsqdAddress.split(':')[0],
    parseInt(config.nsq.nsqdAddress.split(':')[1])
  );

  return new Promise((resolve, reject) => {
    writer.on('ready', () => {
      console.log('[NSQ] Writer 已就绪');

      let sent = 0;
      const total = sampleOrders.length;

      for (const order of sampleOrders) {
        writer.publish('order_create', JSON.stringify(order), (err) => {
          sent++;
          if (err) {
            console.error(`[NSQ] 发布消息失败 (${sent}/${total}):`, err.message);
          } else {
            console.log(`[NSQ] 已发布订单消息 (${sent}/${total}): ${order.order_no}`);
          }

          if (sent >= total) {
            console.log(`\n已完成发布 ${sent} 条消息`);
            writer.close();
            resolve();
          }
        });
      }
    });

    writer.on('error', (err) => {
      console.error('[NSQ] Writer错误:', err.message);
      reject(err);
    });

    writer.connect();
  });
}

publishMessage()
  .then(() => {
    console.log('完成');
    process.exit(0);
  })
  .catch((err) => {
    console.error('发布失败:', err.message);
    process.exit(1);
  });
