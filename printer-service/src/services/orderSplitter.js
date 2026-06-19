'use strict';

const logger = require('../utils/logger');
const printerConfigManager = require('../config/printerConfig');
const retryQueue = require('../redis/retryQueue');
const socketService = require('../socket');
const backendApiClient = require('../api/backendClient');
const { templateManager } = require('../templates');
const { PRINT_STATUS, CATEGORY_TYPES } = require('../constants');

class OrderSplitterService {
  constructor() {
    this.processingOrders = new Set();
  }

  async init() {
    retryQueue.setTaskExecutor((task) => this.executePrintTask(task));
    logger.info('[OrderSplitterService] 订单分单服务已初始化');
  }

  async handleOrderCreated(message) {
    logger.info('[OrderSplitterService] 收到订单创建消息: orderNo=%s, storeId=%s', message.order_no || message.orderNo, message.store_id || message.storeId);

    let orderNo = message.order_no || message.orderNo;
    let storeId = message.store_id || message.storeId;

    if (!orderNo) {
      logger.warn('[OrderSplitterService] 消息缺少订单编号，跳过: message=%j', message);
      return;
    }

    let orderData;
    try {
      orderData = await this.fetchFullOrderData(orderNo, message);
      if (!orderData) {
        logger.error('[OrderSplitterService] 无法获取订单完整数据，跳过: orderNo=%s', orderNo);
        return;
      }
    } catch (err) {
      logger.error('[OrderSplitterService] 获取订单完整数据失败: orderNo=%s, error=%s', orderNo, err.message);
      return;
    }

    if (!orderData.items || orderData.items.length === 0) {
      logger.warn('[OrderSplitterService] 订单无菜品明细，跳过: orderNo=%s', orderNo);
      return;
    }

    const processingKey = `${orderNo}_${Date.now()}`;
    if (this.processingOrders.has(orderNo)) {
      logger.warn('[OrderSplitterService] 订单正在处理中，跳过: orderNo=%s', orderNo);
      return;
    }
    this.processingOrders.add(processingKey);

    try {
      const printTasks = this.splitOrderByCategory(orderData);
      logger.info('[OrderSplitterService] 订单已拆分为 %d 个打印任务: orderNo=%s', printTasks.length, orderNo);

      for (const task of printTasks) {
        await this.processPrintTask(task);
      }
    } catch (err) {
      logger.error('[OrderSplitterService] 处理订单打印失败: orderNo=%s, error=%s', orderNo, err.message);
    } finally {
      setTimeout(() => {
        this.processingOrders.delete(processingKey);
      }, 120000);
    }
  }

  async handlePrintOrder(message) {
    logger.info('[OrderSplitterService] 收到打印指令消息: orderNo=%s, printerId=%s', message.order_no || message.orderNo, message.printer_id || message.printerId);

    let orderNo = message.order_no || message.orderNo;
    let storeId = message.store_id || message.storeId;
    let printerId = message.printer_id || message.printerId;
    let printType = message.print_type || message.printType || 'kitchen';

    if (!orderNo) {
      logger.warn('[OrderSplitterService] 打印指令缺少订单编号，跳过');
      return;
    }

    let orderData;
    try {
      orderData = await this.fetchFullOrderData(orderNo, message);
      if (!orderData) {
        logger.error('[OrderSplitterService] 无法获取订单完整数据，跳过: orderNo=%s', orderNo);
        return;
      }
    } catch (err) {
      logger.error('[OrderSplitterService] 获取订单完整数据失败: orderNo=%s, error=%s', orderNo, err.message);
      return;
    }

    const printer = printerId ? printerConfigManager.getPrinter(printerId) : null;

    const task = {
      id: `print_${orderNo}_${printerId || 'default'}_${Date.now()}`,
      orderNo: orderNo,
      orderId: message.order_id || message.orderId || orderData.id,
      storeId: storeId || orderData.storeId,
      printerId: printerId,
      printer,
      type: printType,
      template: templateManager.getTemplate(printType),
      data: orderData,
      copies: printer ? (printer.copies || 1) : 1,
      createdAt: Date.now(),
    };

    await this.processPrintTask(task);
  }

  async fetchFullOrderData(orderNo, rawMessage) {
    let orderData = null;

    if (rawMessage && rawMessage.order_data && typeof rawMessage.order_data === 'object'
        && rawMessage.order_data.items && Array.isArray(rawMessage.order_data.items)
        && rawMessage.order_data.items.length > 0) {
      const itemsWithCategory = rawMessage.order_data.items.some(
        (it) => it.category_id || it.categoryId || it.category_name || it.categoryName
      );
      if (itemsWithCategory) {
        logger.debug('[OrderSplitterService] 消息已包含完整数据，直接使用: orderNo=%s', orderNo);
        return this.normalizeOrderData(rawMessage.order_data);
      }
    }

    try {
      logger.debug('[OrderSplitterService] 从后端API回查订单完整数据: orderNo=%s', orderNo);
      const apiData = await backendApiClient.getOrderForPrint(orderNo);
      if (apiData) {
        orderData = this.normalizeOrderData(apiData);
        logger.info('[OrderSplitterService] 已从后端API获取订单完整数据: orderNo=%s, items=%d', orderNo, (orderData.items || []).length);
        return orderData;
      }
    } catch (err) {
      logger.warn('[OrderSplitterService] 后端API查询失败，尝试解析消息原始数据: orderNo=%s, error=%s', orderNo, err.message);
    }

    if (rawMessage && rawMessage.order_data) {
      const embedded = typeof rawMessage.order_data === 'string'
        ? this.safeParseJSON(rawMessage.order_data)
        : rawMessage.order_data;
      if (embedded) {
        return this.normalizeOrderData(embedded);
      }
    }

    return null;
  }

  safeParseJSON(str) {
    try {
      return JSON.parse(str);
    } catch (_) {
      return null;
    }
  }

  normalizeOrderData(data) {
    if (!data) return null;

    const orderNo = data.orderNo || data.order_no || '';
    const storeId = data.storeId || data.store_id || 0;
    const storeName = data.storeName || data.store_name || '';

    const rawItems = data.items || [];
    const items = rawItems.map((item, idx) => {
      const productName = item.productName || item.product_name || `菜品${idx + 1}`;
      return {
        id: item.id,
        productId: item.productId || item.product_id,
        skuId: item.skuId || item.sku_id,
        categoryId: item.categoryId || item.category_id || 0,
        categoryName: item.categoryName || item.category_name || '',
        categoryType: item.categoryType || item.category_type,
        productName,
        skuName: item.skuName || item.sku_name || '',
        attributeValues: item.attributeValues || item.attribute_values || '',
        price: this.normalizeDecimal(item.price),
        quantity: item.quantity || 1,
        subtotal: this.normalizeDecimal(item.subtotal),
      };
    });

    const createdAt = data.createdAt
      ? (typeof data.createdAt === 'string' ? data.createdAt : new Date(data.createdAt).toLocaleString())
      : new Date().toLocaleString();

    return {
      id: data.id,
      orderId: data.id,
      orderNo,
      storeId,
      storeName,
      tableNo: data.tableNo || data.table_no || '',
      orderType: data.orderType || data.order_type || 'dine_in',
      totalAmount: this.normalizeDecimal(data.totalAmount || data.total_amount),
      discountAmount: this.normalizeDecimal(data.discountAmount || data.discount_amount),
      couponAmount: this.normalizeDecimal(data.couponAmount || data.coupon_amount),
      payAmount: this.normalizeDecimal(data.payAmount || data.pay_amount),
      payMethod: data.payMethod || data.pay_method || '',
      payStatus: data.payStatus || data.pay_status || 0,
      payTime: data.payTime || data.pay_time,
      orderStatus: data.orderStatus || data.order_status,
      printStatus: data.printStatus || data.print_status,
      pointsEarned: data.pointsEarned || data.points_earned || 0,
      pointsUsed: data.pointsUsed || data.points_used || 0,
      remark: data.remark || '',
      source: data.source || '',
      items,
      createdAt,
    };
  }

  normalizeDecimal(val) {
    if (val === undefined || val === null) return '0.00';
    if (typeof val === 'number') return val.toFixed(2);
    if (typeof val === 'string') return val;
    if (typeof val === 'object' && val.String) return val.String;
    if (typeof val === 'object' && val.Float64 !== undefined) return Number(val.Float64).toFixed(2);
    return String(val);
  }

  splitOrderByCategory(orderData) {
    const tasks = [];
    const categoryItems = new Map();

    for (const item of orderData.items) {
      const category = printerConfigManager.detectItemCategory(item);
      if (!categoryItems.has(category)) {
        categoryItems.set(category, []);
      }
      categoryItems.get(category).push(item);
    }

    for (const [category, items] of categoryItems.entries()) {
      const printers = printerConfigManager.getPrintersByCategory(category, orderData.storeId);

      if (printers.length === 0) {
        logger.warn('[OrderSplitterService] 未找到分类对应的打印机: category=%s, orderNo=%s', category, orderData.orderNo);
        continue;
      }

      for (const printer of printers) {
        const templateType = this.getTemplateTypeForCategory(category, printer);
        const template = templateManager.getTemplate(templateType);

        tasks.push({
          id: `${orderData.orderNo}_${printer.id}_${category}_${Date.now()}`,
          orderNo: orderData.orderNo,
          orderId: orderData.orderId || orderData.id,
          storeId: orderData.storeId,
          printerId: printer.id,
          printer,
          category,
          type: templateType,
          data: {
            ...orderData,
            items,
          },
          template,
          copies: printer.copies || template.copies || 1,
          createdAt: Date.now(),
        });
      }
    }

    return tasks;
  }

  getTemplateTypeForCategory(category, printer) {
    if (printer && printer.printType) {
      return printer.printType;
    }
    const templateMap = {
      [CATEGORY_TYPES.HOT_DISH]: 'kitchen',
      [CATEGORY_TYPES.COLD_DISH]: 'cold',
      [CATEGORY_TYPES.DRINK]: 'drink',
      [CATEGORY_TYPES.STAPLE]: 'kitchen',
      [CATEGORY_TYPES.SOUP]: 'kitchen',
      [CATEGORY_TYPES.SNACK]: 'kitchen',
      [CATEGORY_TYPES.OTHER]: 'kitchen',
    };
    for (const key in templateMap) {
      if (category && String(category).includes(key)) {
        return templateMap[key];
      }
    }
    if (category && category.startsWith('category_')) {
      return 'kitchen';
    }
    return templateMap[category] || 'kitchen';
  }

  async processPrintTask(task) {
    socketService.emitPrintStatus(task.storeId, {
      taskId: task.id,
      orderNo: task.orderNo,
      printerId: task.printerId,
      printerName: task.printer ? task.printer.name : '',
      category: task.category,
      itemCount: task.data.items.length,
      status: PRINT_STATUS.PENDING,
    });

    try {
      const result = await this.executePrintTask(task);

      if (result.success) {
        socketService.emitPrintStatus(task.storeId, {
          taskId: task.id,
          orderNo: task.orderNo,
          printerId: task.printerId,
          printerName: task.printer ? task.printer.name : '',
          category: task.category,
          status: PRINT_STATUS.SUCCESS,
        });
        logger.info('[OrderSplitterService] 打印成功: orderNo=%s, printer=%s, category=%s, items=%d',
          task.orderNo, task.printer ? task.printer.name : task.printerId, task.category, task.data.items.length);
      } else {
        await this.handlePrintFailure(task, result.error || '未知错误');
      }
    } catch (err) {
      logger.error('[OrderSplitterService] 打印异常: orderNo=%s, error=%s', task.orderNo, err.message);
      await this.handlePrintFailure(task, err.message);
    }
  }

  async executePrintTask(task) {
    const printer = task.printer || printerConfigManager.getPrinter(task.printerId);
    if (!printer) {
      return { success: false, error: `打印机不存在: id=${task.printerId}` };
    }
    if (printer.status !== 1) {
      return { success: false, error: `打印机已禁用: id=${task.printerId}` };
    }
    if (!printer.instance) {
      return { success: false, error: `打印机实例未初始化: id=${task.printerId}` };
    }

    socketService.emitPrintStatus(task.storeId, {
      taskId: task.id,
      orderNo: task.orderNo,
      printerId: task.printerId,
      printerName: printer.name,
      category: task.category,
      status: PRINT_STATUS.PRINTING,
    });

    try {
      const copies = task.copies || 1;
      const template = task.template || templateManager.getTemplate(task.type || 'kitchen');

      for (let i = 0; i < copies; i++) {
        await printer.instance.print({
          type: task.type,
          template,
          data: task.data,
          category: task.category,
        });
        if (i < copies - 1) {
          await new Promise((resolve) => setTimeout(resolve, 300));
        }
      }
      return { success: true };
    } catch (err) {
      logger.error('[OrderSplitterService] 执行打印失败: printer=%s, error=%s', printer.name || task.printerId, err.message);
      return { success: false, error: err.message };
    }
  }

  async handlePrintFailure(task, error) {
    socketService.emitPrintStatus(task.storeId, {
      taskId: task.id,
      orderNo: task.orderNo,
      printerId: task.printerId,
      category: task.category,
      status: PRINT_STATUS.FAILED,
      error,
    });

    const retryTask = {
      ...task,
      error,
      lastError: error,
      lastFailedAt: Date.now(),
    };

    await retryQueue.enqueue(retryTask);

    logger.warn('[OrderSplitterService] 打印失败，已加入重试队列: orderNo=%s, printer=%s',
      task.orderNo, task.printer ? task.printer.name : task.printerId);
  }

  async testPrint(printerId) {
    const printer = printerConfigManager.getPrinter(printerId);
    if (!printer) {
      throw new Error(`打印机不存在: id=${printerId}`);
    }

    if (!printer.instance) {
      throw new Error(`打印机实例未初始化: id=${printerId}`);
    }

    socketService.emitPrintStatus(printer.storeId, {
      printerId: printer.id,
      printerName: printer.name,
      status: PRINT_STATUS.PRINTING,
      test: true,
    });

    try {
      const result = await printer.instance.testPrint();
      socketService.emitPrintStatus(printer.storeId, {
        printerId: printer.id,
        printerName: printer.name,
        status: PRINT_STATUS.SUCCESS,
        test: true,
      });
      return { success: true, printer: printer.name };
    } catch (err) {
      socketService.emitPrintStatus(printer.storeId, {
        printerId: printer.id,
        printerName: printer.name,
        status: PRINT_STATUS.FAILED,
        test: true,
        error: err.message,
      });
      throw err;
    }
  }
}

const orderSplitterService = new OrderSplitterService();

module.exports = orderSplitterService;
