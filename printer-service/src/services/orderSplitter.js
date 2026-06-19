'use strict';

const logger = require('../utils/logger');
const printerConfigManager = require('../config/printerConfig');
const retryQueue = require('../redis/retryQueue');
const socketService = require('../socket');
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
    logger.info('[OrderSplitterService] 收到订单创建消息: orderNo=%s, storeId=%d', message.order_no, message.store_id);

    const orderData = this.parseOrderData(message);
    if (!orderData || !orderData.items || orderData.items.length === 0) {
      logger.warn('[OrderSplitterService] 订单数据无效，跳过: orderNo=%s', message.order_no);
      return;
    }

    if (this.processingOrders.has(orderData.orderNo)) {
      logger.warn('[OrderSplitterService] 订单正在处理中，跳过: orderNo=%s', orderData.orderNo);
      return;
    }

    this.processingOrders.add(orderData.orderNo);

    try {
      const printTasks = this.splitOrderByCategory(orderData);
      logger.info('[OrderSplitterService] 订单已拆分为 %d 个打印任务: orderNo=%s', printTasks.length, orderData.orderNo);

      for (const task of printTasks) {
        await this.processPrintTask(task);
      }
    } catch (err) {
      logger.error('[OrderSplitterService] 处理订单打印失败: orderNo=%s, error=%s', message.order_no, err.message);
    } finally {
      setTimeout(() => {
        this.processingOrders.delete(orderData.orderNo);
      }, 60000);
    }
  }

  async handlePrintOrder(message) {
    logger.info('[OrderSplitterService] 收到打印指令消息: orderNo=%s, printerId=%d', message.order_no, message.printer_id);

    const printData = this.parsePrintData(message);
    if (!printData) {
      logger.warn('[OrderSplitterService] 打印数据无效，跳过');
      return;
    }

    const task = {
      id: `print_${message.order_no}_${message.printer_id}_${Date.now()}`,
      orderNo: message.order_no,
      orderId: message.order_id,
      storeId: message.store_id,
      printerId: message.printer_id,
      type: message.print_type || 'kitchen',
      data: printData,
      createdAt: Date.now(),
    };

    await this.processPrintTask(task);
  }

  parseOrderData(message) {
    if (message.order_data) {
      if (typeof message.order_data === 'string') {
        try {
          return JSON.parse(message.order_data);
        } catch (err) {
          logger.error('[OrderSplitterService] 解析订单数据失败: %s', err.message);
          return null;
        }
      }
      return message.order_data;
    }
    return null;
  }

  parsePrintData(message) {
    if (message.print_data) {
      if (typeof message.print_data === 'string') {
        try {
          return JSON.parse(message.print_data);
        } catch (err) {
          logger.error('[OrderSplitterService] 解析打印数据失败: %s', err.message);
          return null;
        }
      }
      return message.print_data;
    }
    return null;
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
          orderId: orderData.orderId,
          storeId: orderData.storeId || orderData.store_id,
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
    if (printer.printType) {
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
    return templateMap[category] || 'kitchen';
  }

  async processPrintTask(task) {
    socketService.emitPrintStatus(task.storeId, {
      taskId: task.id,
      orderNo: task.orderNo,
      printerId: task.printerId,
      category: task.category,
      status: PRINT_STATUS.PENDING,
    });

    try {
      const result = await this.executePrintTask(task);

      if (result.success) {
        socketService.emitPrintStatus(task.storeId, {
          taskId: task.id,
          orderNo: task.orderNo,
          printerId: task.printerId,
          category: task.category,
          status: PRINT_STATUS.SUCCESS,
        });
        logger.info('[OrderSplitterService] 打印成功: orderNo=%s, printerId=%d, category=%s',
          task.orderNo, task.printerId, task.category);
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
      category: task.category,
      status: PRINT_STATUS.PRINTING,
    });

    try {
      const copies = task.copies || 1;
      for (let i = 0; i < copies; i++) {
        await printer.instance.print({
          type: task.type,
          data: task.data,
        });
        if (i < copies - 1) {
          await new Promise((resolve) => setTimeout(resolve, 200));
        }
      }
      return { success: true };
    } catch (err) {
      logger.error('[OrderSplitterService] 执行打印失败: printerId=%d, error=%s', printer.id, err.message);
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
    };

    await retryQueue.enqueue(retryTask);

    logger.warn('[OrderSplitterService] 打印失败，已加入重试队列: orderNo=%s, printerId=%d',
      task.orderNo, task.printerId);
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
      status: PRINT_STATUS.PRINTING,
      test: true,
    });

    try {
      const result = await printer.instance.testPrint();
      socketService.emitPrintStatus(printer.storeId, {
        printerId: printer.id,
        status: PRINT_STATUS.SUCCESS,
        test: true,
      });
      return { success: true, printer: printer.name };
    } catch (err) {
      socketService.emitPrintStatus(printer.storeId, {
        printerId: printer.id,
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
