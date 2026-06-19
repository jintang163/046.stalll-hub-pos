'use strict';

require('dotenv').config();

const printerConfigManager = require('../src/config/printerConfig');
const { PRINTER_CONNECTION_TYPES } = require('../src/constants');

const samplePrinters = [
  {
    id: 1,
    name: '热菜打印机',
    storeId: 1,
    type: 'kitchen',
    connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
    printType: 'kitchen',
    ipAddress: '192.168.1.100',
    port: 9100,
    status: 1,
    isDefault: true,
    copies: 1,
    width: 80,
    encoding: 'UTF-8',
  },
  {
    id: 2,
    name: '凉菜打印机',
    storeId: 1,
    type: 'kitchen',
    connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
    printType: 'cold',
    ipAddress: '192.168.1.101',
    port: 9100,
    status: 1,
    isDefault: false,
    copies: 1,
    width: 80,
    encoding: 'UTF-8',
  },
  {
    id: 3,
    name: '饮品打印机',
    storeId: 1,
    type: 'bar',
    connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
    printType: 'drink',
    ipAddress: '192.168.1.102',
    port: 9100,
    status: 1,
    isDefault: false,
    copies: 1,
    width: 80,
    encoding: 'UTF-8',
  },
  {
    id: 4,
    name: '前台小票机',
    storeId: 1,
    type: 'receipt',
    connectionType: PRINTER_CONNECTION_TYPES.USB,
    printType: 'receipt',
    vendorId: 6790,
    productId: 30012,
    status: 1,
    isDefault: false,
    copies: 2,
    width: 80,
    encoding: 'UTF-8',
  },
];

const categoryMap = {
  hot_dish: [1],
  cold_dish: [2],
  drink: [3],
  staple: [1],
  soup: [1],
  snack: [1],
  other: [1],
};

async function initSampleData() {
  console.log('初始化示例打印机配置...');

  try {
    await require('../src/redis').initRedis();

    for (const printer of samplePrinters) {
      printerConfigManager.addPrinter(printer);
      console.log(`已添加打印机: ${printer.name} (${printer.id})`);
    }

    for (const [category, printerIds] of Object.entries(categoryMap)) {
      printerConfigManager.setCategoryPrinters(category, printerIds);
      console.log(`已设置分类映射: ${category} -> [${printerIds.join(', ')}]`);
    }

    console.log('\n示例数据初始化完成!');
    console.log('打印机数量:', printerConfigManager.getAllPrinters().length);
    console.log('分类映射:', JSON.stringify(printerConfigManager.getCategoryPrinterMap(), null, 2));

    process.exit(0);
  } catch (err) {
    console.error('初始化失败:', err.message);
    process.exit(1);
  }
}

initSampleData();
