'use strict';

const http = require('http');
const express = require('express');
const cors = require('cors');
const { Server } = require('socket.io');
const config = require('../config');
const logger = require('../utils/logger');
const { PRINT_STATUS } = require('../constants');

class SocketIOService {
  constructor() {
    this.app = express();
    this.app.use(cors({ origin: config.socket.corsOrigin }));
    this.app.use(express.json());
    this.server = http.createServer(this.app);
    this.io = new Server(this.server, {
      cors: {
        origin: config.socket.corsOrigin,
        methods: ['GET', 'POST'],
      },
    });
    this.connectedClients = new Map();
    this.eventHandlers = new Map();
  }

  async init() {
    this.setupMiddleware();
    this.setupRoutes();
    this.setupSocketEvents();

    this.server.listen(config.socket.port, () => {
      logger.info('[SocketIO] 服务已启动，端口: %d', config.socket.port);
    });

    this.setupHealthCheck();
  }

  setupMiddleware() {
    this.app.use((req, res, next) => {
      req.io = this.io;
      next();
    });
  }

  setupRoutes() {
    this.app.get('/health', (req, res) => {
      res.json({
        status: 'ok',
        connectedClients: this.connectedClients.size,
        uptime: process.uptime(),
      });
    });

    this.app.post('/api/print/test', async (req, res) => {
      try {
        const { printerId } = req.body;
        if (this.eventHandlers.has('test-print')) {
          const result = await this.eventHandlers.get('test-print')(printerId);
          res.json({ success: true, data: result });
        } else {
          res.status(400).json({ success: false, message: '测试打印处理器未注册' });
        }
      } catch (err) {
        logger.error('[SocketIO] 测试打印API错误: %s', err.message);
        res.status(500).json({ success: false, message: err.message });
      }
    });

    this.app.get('/api/printers', async (req, res) => {
      try {
        if (this.eventHandlers.has('get-printers')) {
          const printers = await this.eventHandlers.get('get-printers')();
          res.json({ success: true, data: printers });
        } else {
          res.json({ success: true, data: [] });
        }
      } catch (err) {
        res.status(500).json({ success: false, message: err.message });
      }
    });

    this.app.get('/api/retry-queue', async (req, res) => {
      try {
        if (this.eventHandlers.has('get-retry-queue')) {
          const tasks = await this.eventHandlers.get('get-retry-queue')();
          res.json({ success: true, data: tasks });
        } else {
          res.json({ success: true, data: [] });
        }
      } catch (err) {
        res.status(500).json({ success: false, message: err.message });
      }
    });

    this.app.post('/api/retry-queue/:taskId/retry', async (req, res) => {
      try {
        const { taskId } = req.params;
        if (this.eventHandlers.has('retry-task')) {
          const result = await this.eventHandlers.get('retry-task')(taskId);
          res.json({ success: true, data: result });
        } else {
          res.status(400).json({ success: false, message: '重试处理器未注册' });
        }
      } catch (err) {
        res.status(500).json({ success: false, message: err.message });
      }
    });
  }

  setupSocketEvents() {
    this.io.on('connection', (socket) => {
      const clientId = socket.handshake.query.clientId || socket.id;
      const storeId = socket.handshake.query.storeId;

      this.connectedClients.set(socket.id, {
        id: socket.id,
        clientId,
        storeId,
        connectedAt: Date.now(),
      });

      logger.info('[SocketIO] 客户端已连接: socketId=%s, clientId=%s, storeId=%s, 当前连接数: %d',
        socket.id, clientId, storeId, this.connectedClients.size);

      socket.join(`store:${storeId}`);

      socket.emit('connected', {
        socketId: socket.id,
        serverTime: Date.now(),
      });

      socket.on('printer:get-all', async (callback) => {
        try {
          if (this.eventHandlers.has('get-printers')) {
            const printers = await this.eventHandlers.get('get-printers')(storeId);
            callback({ success: true, data: printers });
          } else {
            callback({ success: true, data: [] });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('printer:test', async (printerId, callback) => {
        try {
          if (this.eventHandlers.has('test-print')) {
            const result = await this.eventHandlers.get('test-print')(printerId);
            callback({ success: true, data: result });
          } else {
            callback({ success: false, message: '测试打印处理器未注册' });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('printer:update-config', async (config, callback) => {
        try {
          if (this.eventHandlers.has('update-printer-config')) {
            const result = await this.eventHandlers.get('update-printer-config')(config);
            callback({ success: true, data: result });
            this.broadcastPrinterUpdate(storeId, result);
          } else {
            callback({ success: false, message: '配置更新处理器未注册' });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('retry-queue:get-all', async (callback) => {
        try {
          if (this.eventHandlers.has('get-retry-queue')) {
            const tasks = await this.eventHandlers.get('get-retry-queue')();
            callback({ success: true, data: tasks });
          } else {
            callback({ success: true, data: [] });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('retry-queue:retry', async (taskId, callback) => {
        try {
          if (this.eventHandlers.has('retry-task')) {
            const result = await this.eventHandlers.get('retry-task')(taskId);
            callback({ success: true, data: result });
          } else {
            callback({ success: false, message: '重试处理器未注册' });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('retry-queue:clear', async (callback) => {
        try {
          if (this.eventHandlers.has('clear-retry-queue')) {
            await this.eventHandlers.get('clear-retry-queue')();
            callback({ success: true });
          } else {
            callback({ success: false, message: '清空队列处理器未注册' });
          }
        } catch (err) {
          callback({ success: false, message: err.message });
        }
      });

      socket.on('disconnect', () => {
        this.connectedClients.delete(socket.id);
        logger.info('[SocketIO] 客户端已断开: socketId=%s, 当前连接数: %d',
          socket.id, this.connectedClients.size);
      });
    });
  }

  setupHealthCheck() {
    setInterval(() => {
      this.io.emit('server:ping', {
        timestamp: Date.now(),
        connectedClients: this.connectedClients.size,
      });
    }, 30000);
  }

  on(event, handler) {
    this.eventHandlers.set(event, handler);
  }

  emitPrintStatus(storeId, statusData) {
    this.io.to(`store:${storeId}`).emit('print:status', {
      ...statusData,
      status: statusData.status || PRINT_STATUS.PENDING,
      timestamp: Date.now(),
    });
    logger.debug('[SocketIO] 发送打印状态: storeId=%s, orderNo=%s, status=%s',
      storeId, statusData.orderNo, statusData.status);
  }

  broadcastPrinterUpdate(storeId, printer) {
    this.io.to(`store:${storeId}`).emit('printer:updated', {
      printer,
      timestamp: Date.now(),
    });
  }

  broadcastPrintTask(storeId, task) {
    this.io.to(`store:${storeId}`).emit('print:task', {
      task,
      timestamp: Date.now(),
    });
  }

  getConnectedClients(storeId = null) {
    if (storeId) {
      return Array.from(this.connectedClients.values()).filter((c) => c.storeId === storeId);
    }
    return Array.from(this.connectedClients.values());
  }

  async close() {
    return new Promise((resolve) => {
      this.server.close(() => {
        logger.info('[SocketIO] 服务已关闭');
        resolve();
      });
    });
  }
}

const socketService = new SocketIOService();

module.exports = socketService;
