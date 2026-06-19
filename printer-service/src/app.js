'use strict';

require('dotenv').config();

const logger = require('./utils/logger');
const { initRedis, closeRedis } = require('./redis');
const retryQueue = require('./redis/retryQueue');
const nsqConsumer = require('./nsq/consumer');
const socketService = require('./socket');
const printerConfigManager = require('./config/printerConfig');
const orderSplitterService = require('./services/orderSplitter');
const { templateManager } = require('./templates');

async function bootstrap() {
  logger.info('========================================');
  logger.info('  Stall Hub POS - Printer Service');
  logger.info('========================================');

  try {
    await initRedis();
    logger.info('[Bootstrap] Redis初始化完成');

    await printerConfigManager.init();
    logger.info('[Bootstrap] 打印机配置管理器初始化完成，打印机数量: %d', printerConfigManager.getAllPrinters().length);

    await templateManager;
    logger.info('[Bootstrap] 打印模板管理器初始化完成，模板数量: %d', templateManager.getAllTemplates().length);

    await retryQueue.init();
    logger.info('[Bootstrap] 重试队列初始化完成');

    await orderSplitterService.init();
    logger.info('[Bootstrap] 订单分单服务初始化完成');

    await socketService.init();
    logger.info('[Bootstrap] Socket.IO服务初始化完成');

    registerSocketHandlers();
    logger.info('[Bootstrap] Socket事件处理器已注册');

    await nsqConsumer.init();
    registerNSQHandlers();
    logger.info('[Bootstrap] NSQ消费者初始化完成');

    logger.info('========================================');
    logger.info('  Printer Service启动成功!');
    logger.info('  Socket.IO端口: %d', require('./config').socket.port);
    logger.info('========================================');

  } catch (err) {
    logger.error('[Bootstrap] 服务启动失败: %s', err.message);
    logger.error(err.stack);
    process.exit(1);
  }
}

function registerNSQHandlers() {
  nsqConsumer.subscribeOrderCreated(async (data, msg) => {
    try {
      await orderSplitterService.handleOrderCreated(data);
      return true;
    } catch (err) {
      logger.error('[NSQ] 处理订单创建消息失败: %s', err.message);
      return false;
    }
  });

  nsqConsumer.subscribeOrderPaid(async (data, msg) => {
    try {
      await orderSplitterService.handleOrderCreated(data);
      return true;
    } catch (err) {
      logger.error('[NSQ] 处理订单支付消息失败: %s', err.message);
      return false;
    }
  });

  nsqConsumer.subscribePrintOrder(async (data, msg) => {
    try {
      await orderSplitterService.handlePrintOrder(data);
      return true;
    } catch (err) {
      logger.error('[NSQ] 处理打印指令消息失败: %s', err.message);
      return false;
    }
  });

  nsqConsumer.subscribeOrderUpdate(async (data, msg) => {
    try {
      await orderSplitterService.handleOrderCreated(data);
      return true;
    } catch (err) {
      logger.error('[NSQ] 处理订单更新消息失败: %s', err.message);
      return false;
    }
  });

  logger.info('[Bootstrap] NSQ消息处理器已注册');
}

function registerSocketHandlers() {
  socketService.on('get-printers', async (storeId) => {
    const printers = printerConfigManager.getAllPrinters(storeId);
    return printers.map((p) => ({
      id: p.id,
      name: p.name,
      storeId: p.storeId,
      type: p.type,
      connectionType: p.connectionType,
      printType: p.printType,
      ipAddress: p.ipAddress,
      port: p.port,
      status: p.status,
      isDefault: p.isDefault,
      copies: p.copies,
      width: p.width,
      encoding: p.encoding,
    }));
  });

  socketService.on('test-print', async (printerId) => {
    return await orderSplitterService.testPrint(printerId);
  });

  socketService.on('update-printer-config', async (config) => {
    if (config.id) {
      return printerConfigManager.updatePrinter(config.id, config);
    }
    return printerConfigManager.addPrinter(config);
  });

  socketService.on('get-retry-queue', async () => {
    return await retryQueue.getAllTasks();
  });

  socketService.on('retry-task', async (taskId) => {
    const task = await retryQueue.getTask(taskId);
    if (!task) {
      throw new Error(`任务不存在: ${taskId}`);
    }
    const result = await retryQueue.executeTask(task);
    if (result.success) {
      await retryQueue.dequeue(taskId);
    }
    return result;
  });

  socketService.on('clear-retry-queue', async () => {
    await retryQueue.clearAll();
  });
}

async function shutdown(signal) {
  logger.info('[Shutdown] 收到 %s 信号，正在优雅关闭...', signal);

  try {
    nsqConsumer.close();
    logger.info('[Shutdown] NSQ消费者已关闭');

    await socketService.close();
    logger.info('[Shutdown] Socket.IO服务已关闭');

    retryQueue.stop();
    logger.info('[Shutdown] 重试队列已停止');

    await closeRedis();
    logger.info('[Shutdown] Redis连接已关闭');

    logger.info('[Shutdown] 服务已正常关闭');
    process.exit(0);
  } catch (err) {
    logger.error('[Shutdown] 关闭过程出错: %s', err.message);
    process.exit(1);
  }
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
process.on('uncaughtException', (err) => {
  logger.error('[UncaughtException] %s', err.message);
  logger.error(err.stack);
});
process.on('unhandledRejection', (reason, promise) => {
  logger.error('[UnhandledRejection] Promise: %s, Reason: %s', promise, reason);
});

bootstrap();
