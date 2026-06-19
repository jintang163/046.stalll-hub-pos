'use strict';

require('dotenv').config();

const { initRedis, closeRedis } = require('../src/redis');
const retryQueue = require('../src/redis/retryQueue');
const printerConfigManager = require('../src/config/printerConfig');
const { PRINTER_CONNECTION_TYPES, CATEGORY_TYPES } = require('../src/constants');
const { templateManager } = require('../src/templates');

async function runTests() {
  console.log('========================================');
  console.log('  Printer Service - 单元测试');
  console.log('========================================\n');

  let passed = 0;
  let failed = 0;

  function assert(condition, message) {
    if (condition) {
      console.log(`  ✓ ${message}`);
      passed++;
    } else {
      console.log(`  ✗ ${message}`);
      failed++;
    }
  }

  console.log('[1/6] 测试 Redis 连接...');
  try {
    await initRedis();
    assert(true, 'Redis 连接成功');
  } catch (err) {
    assert(false, `Redis 连接失败: ${err.message}`);
  }

  console.log('\n[2/6] 测试打印机配置管理...');
  try {
    await printerConfigManager.init();

    const testPrinter = {
      id: 999,
      name: '测试打印机',
      storeId: 1,
      type: 'kitchen',
      connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
      printType: 'kitchen',
      ipAddress: '127.0.0.1',
      port: 9100,
      status: 1,
      isDefault: true,
      copies: 1,
      width: 80,
      encoding: 'UTF-8',
    };

    printerConfigManager.addPrinter(testPrinter);
    const printer = printerConfigManager.getPrinter(999);
    assert(printer !== undefined, '打印机配置已添加');
    assert(printer.name === '测试打印机', '打印机名称正确');

    printerConfigManager.setCategoryPrinters(CATEGORY_TYPES.HOT_DISH, [999]);
    const hotPrinters = printerConfigManager.getPrintersByCategory(CATEGORY_TYPES.HOT_DISH);
    assert(hotPrinters.length > 0, '分类打印机映射已设置');

    printerConfigManager.removePrinter(999);
    assert(!printerConfigManager.getPrinter(999), '打印机配置已删除');
  } catch (err) {
    assert(false, `打印机配置管理测试失败: ${err.message}`);
  }

  console.log('\n[3/6] 测试菜品分类识别...');
  try {
    const category1 = printerConfigManager.detectItemCategory({ categoryName: '热菜' });
    assert(category1 === CATEGORY_TYPES.HOT_DISH, `热菜分类识别正确: ${category1}`);

    const category2 = printerConfigManager.detectItemCategory({ categoryName: '凉菜' });
    assert(category2 === CATEGORY_TYPES.COLD_DISH, `凉菜分类识别正确: ${category2}`);

    const category3 = printerConfigManager.detectItemCategory({ categoryName: '饮品' });
    assert(category3 === CATEGORY_TYPES.DRINK, `饮品分类识别正确: ${category3}`);

    const category4 = printerConfigManager.detectItemCategory({ categoryId: 10 });
    assert(category4 === 'category_10', `按分类ID识别正确: ${category4}`);
  } catch (err) {
    assert(false, `菜品分类识别测试失败: ${err.message}`);
  }

  console.log('\n[4/6] 测试打印模板...');
  try {
    const templates = templateManager.getAllTemplates();
    assert(templates.length >= 3, `模板数量正确: ${templates.length}`);

    const kitchenTemplate = templateManager.getTemplate('kitchen');
    assert(kitchenTemplate !== undefined, '后厨模板存在');
    assert(kitchenTemplate.type === 'kitchen', '模板类型正确');

    const testOrderData = {
      storeName: '测试门店',
      orderNo: 'TEST20240101001',
      tableNo: 'A01',
      orderType: 'dine_in',
      createdAt: '2024-01-01 12:00:00',
      items: [
        { productName: '红烧肉', quantity: 2, price: '38.00', subtotal: '76.00' },
        { productName: '米饭', quantity: 2, price: '2.00', subtotal: '4.00' },
      ],
    };

    const rendered = templateManager.formatTemplate(kitchenTemplate, testOrderData);
    assert(rendered.length > 0, '模板渲染成功');
  } catch (err) {
    assert(false, `打印模板测试失败: ${err.message}`);
  }

  console.log('\n[5/6] 测试重试队列...');
  try {
    await retryQueue.init();

    const testTask = {
      id: 'test_task_001',
      orderNo: 'TEST001',
      storeId: 1,
      printerId: 1,
      category: CATEGORY_TYPES.HOT_DISH,
      type: 'kitchen',
      data: { items: [] },
    };

    retryQueue.setTaskExecutor(async () => ({ success: true }));

    await retryQueue.enqueue(testTask);
    const task = await retryQueue.getTask('test_task_001');
    assert(task !== null, '任务已加入重试队列');

    await retryQueue.dequeue('test_task_001');
    const removedTask = await retryQueue.getTask('test_task_001');
    assert(removedTask === null, '任务已从重试队列移除');
  } catch (err) {
    assert(false, `重试队列测试失败: ${err.message}`);
  }

  console.log('\n[6/6] 测试订单分单逻辑...');
  try {
    const testPrinterHot = {
      id: 101,
      name: '热菜打印机',
      storeId: 1,
      type: 'kitchen',
      connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
      printType: 'kitchen',
      ipAddress: '127.0.0.1',
      port: 9100,
      status: 1,
      copies: 1,
      width: 80,
      encoding: 'UTF-8',
    };
    const testPrinterDrink = {
      id: 102,
      name: '饮品打印机',
      storeId: 1,
      type: 'bar',
      connectionType: PRINTER_CONNECTION_TYPES.NETWORK,
      printType: 'drink',
      ipAddress: '127.0.0.1',
      port: 9101,
      status: 1,
      copies: 1,
      width: 80,
      encoding: 'UTF-8',
    };

    printerConfigManager.addPrinter(testPrinterHot);
    printerConfigManager.addPrinter(testPrinterDrink);
    printerConfigManager.setCategoryPrinters(CATEGORY_TYPES.HOT_DISH, [101]);
    printerConfigManager.setCategoryPrinters(CATEGORY_TYPES.DRINK, [102]);

    const orderSplitter = require('../src/services/orderSplitter');
    await orderSplitter.init();

    const testOrder = {
      orderNo: 'TEST_ORDER_001',
      storeId: 1,
      orderType: 'dine_in',
      tableNo: 'A01',
      items: [
        { productName: '红烧肉', categoryName: '热菜', quantity: 1, price: '38.00' },
        { productName: '凉拌黄瓜', categoryName: '凉菜', quantity: 1, price: '12.00' },
        { productName: '可乐', categoryName: '饮品', quantity: 2, price: '5.00' },
      ],
    };

    const tasks = orderSplitter.splitOrderByCategory(testOrder);
    assert(tasks.length >= 2, `订单分单成功，任务数: ${tasks.length}`);

    printerConfigManager.removePrinter(101);
    printerConfigManager.removePrinter(102);
  } catch (err) {
    assert(false, `订单分单逻辑测试失败: ${err.message}`);
  }

  console.log('\n========================================');
  console.log(`  测试结果: 通过 ${passed}, 失败 ${failed}`);
  console.log('========================================');

  retryQueue.stop();
  await closeRedis();

  process.exit(failed > 0 ? 1 : 0);
}

runTests().catch((err) => {
  console.error('测试执行失败:', err);
  process.exit(1);
});
