'use strict';

const logger = require('../utils/logger');
const { templateManager } = require('../templates');
const { PRINTER_CONNECTION_TYPES } = require('../constants');
const iconv = require('iconv-lite');

class BasePrinter {
  constructor(config) {
    this.id = config.id;
    this.name = config.name;
    this.storeId = config.storeId;
    this.type = config.type;
    this.connectionType = config.connectionType;
    this.printType = config.printType || 'kitchen';
    this.status = config.status !== undefined ? config.status : 1;
    this.copies = config.copies || 1;
    this.width = config.width || 80;
    this.encoding = config.encoding || 'UTF-8';
    this.config = config;
  }

  async connect() {
    throw new Error('connect() must be implemented by subclass');
  }

  async disconnect() {
    throw new Error('disconnect() must be implemented by subclass');
  }

  async isConnected() {
    throw new Error('isConnected() must be implemented by subclass');
  }

  async print(printData) {
    throw new Error('print() must be implemented by subclass');
  }

  async testPrint() {
    const testData = {
      type: 'kitchen',
      data: this.getTestOrderData(),
    };
    return await this.print(testData);
  }

  getTestOrderData() {
    return {
      storeName: this.name || '打印机测试',
      orderNo: `TEST${Date.now()}`,
      tableNo: 'T00',
      orderType: 'dine_in',
      totalAmount: '88.00',
      payAmount: '88.00',
      payMethod: 'cash',
      payStatus: 1,
      createdAt: new Date().toLocaleString(),
      items: [
        { productName: '测试菜品1(大份)', skuName: '大份', categoryName: '热菜', quantity: 1, price: '38.00', subtotal: '38.00' },
        { productName: '测试菜品2', skuName: '', categoryName: '凉菜', quantity: 2, price: '15.00', subtotal: '30.00' },
        { productName: '可乐', skuName: '中杯', categoryName: '饮品', quantity: 2, price: '10.00', subtotal: '20.00' },
      ],
    };
  }

  buildEscPosBytes(printData) {
    const EscPosPrinter = require('./escpos').EscPosPrinter;
    const printer = new EscPosPrinter();
    printer.setEncoding(this.encoding);

    const type = printData.type || this.printType || 'kitchen';
    const data = printData.data || {};
    const template = printData.template || templateManager.getTemplate(type);

    const renderedLines = templateManager.formatTemplate(template, data);

    for (const element of renderedLines) {
      this.applyTemplateElement(printer, element);
    }

    return printer.Bytes();
  }

  applyTemplateElement(printer, element) {
    switch (element.type) {
      case 'text':
        if (element.align) {
          printer.align(element.align);
        }
        if (element.bold) {
          printer.setTextBold(true);
        }
        if (element.doubleWidth) {
          printer.setDoubleWidth(true);
        }
        printer.text(element.content || '');
        if (element.bold) {
          printer.setTextBold(false);
        }
        if (element.doubleWidth) {
          printer.setDoubleWidth(false);
        }
        if (element.align) {
          printer.align('left');
        }
        break;

      case 'separator':
        printer.text(element.content || this.getSeparatorLine());
        break;

      case 'feed':
        printer.feed(element.lines || 1);
        break;

      case 'cut':
        printer.cut(element.mode || 'full');
        break;

      default:
        if (element.content) {
          printer.text(element.content);
        }
    }
  }

  getSeparatorLine() {
    const charsPerLine = this.width >= 80 ? 32 : 24;
    return '-'.repeat(charsPerLine);
  }

  getOrderTypeName(orderType) {
    const map = {
      dine_in: '堂食',
      takeaway: '外卖',
      delivery: '配送',
    };
    return map[orderType] || orderType;
  }

  getPayMethodName(payMethod) {
    const map = {
      wechat: '微信支付',
      alipay: '支付宝',
      cash: '现金',
      card: '刷卡',
    };
    return map[payMethod] || payMethod;
  }
}

class NetworkPrinter extends BasePrinter {
  constructor(config) {
    super(config);
    this.ipAddress = config.ipAddress || '127.0.0.1';
    this.port = config.port || 9100;
    this.device = null;
  }

  async connect() {
    return new Promise((resolve) => {
      try {
        const net = require('net');
        this.device = new net.Socket();

        const timeoutMs = 5000;
        const timer = setTimeout(() => {
          if (this.device) {
            this.device.destroy();
            this.device = null;
          }
          logger.error('[NetworkPrinter] 连接超时 %s:%s', this.ipAddress, this.port);
          resolve(false);
        }, timeoutMs);

        this.device.connect(this.port, this.ipAddress, () => {
          clearTimeout(timer);
          logger.info('[NetworkPrinter] 打印机连接成功 %s:%s', this.ipAddress, this.port);
          resolve(true);
        });

        this.device.on('error', (err) => {
          clearTimeout(timer);
          logger.error('[NetworkPrinter] 打印机连接错误 %s:%s, error=%s', this.ipAddress, this.port, err.message);
          this.device = null;
          resolve(false);
        });
      } catch (err) {
        logger.error('[NetworkPrinter] 创建Socket异常: %s', err.message);
        resolve(false);
      }
    });
  }

  async disconnect() {
    if (this.device) {
      try {
        this.device.end();
        this.device.destroy();
      } catch (err) {
        logger.warn('[NetworkPrinter] 断开连接出错: %s', err.message);
      }
      this.device = null;
    }
  }

  async isConnected() {
    return this.device && !this.device.destroyed && this.device.writable;
  }

  async print(printData) {
    if (!this.device || this.device.destroyed || !this.device.writable) {
      const connected = await this.connect();
      if (!connected) {
        throw new Error(`无法连接到打印机 ${this.ipAddress}:${this.port}`);
      }
    }

    return new Promise((resolve, reject) => {
      try {
        const escposBytes = this.buildEscPosBytes(printData);

        let settled = false;
        const finish = (err, result) => {
          if (settled) return;
          settled = true;
          if (err) reject(err);
          else resolve(result);
        };

        const timeout = setTimeout(() => {
          finish(new Error('打印数据写入超时'));
        }, 15000);

        const drained = this.device.write(escposBytes, (err) => {
          clearTimeout(timeout);
          if (err) {
            logger.error('[NetworkPrinter] 写入数据失败: %s', err.message);
            finish(err);
          } else {
            logger.info('[NetworkPrinter] 数据已发送到打印机 %s:%s (%d bytes)', this.ipAddress, this.port, escposBytes.length);
            finish(null, { success: true, bytes: escposBytes.length });
          }
        });

        if (!drained) {
          this.device.once('drain', () => {
            logger.debug('[NetworkPrinter] 缓冲区已排空');
          });
        }
      } catch (err) {
        reject(err);
      }
    });
  }
}

class USBPrinter extends BasePrinter {
  constructor(config) {
    super(config);
    this.vendorId = config.vendorId;
    this.productId = config.productId;
    this.device = null;
    this._adapter = null;
  }

  _ensureAdapter() {
    if (!this._adapter) {
      try {
        const escposUSB = require('escpos-usb');
        const escpos = require('escpos');
        this._adapter = { USB: escposUSB, Printer: escpos.Printer };
      } catch (err) {
        logger.warn('[USBPrinter] escpos-usb 未安装，回退到模拟模式: %s', err.message);
        this._adapter = null;
      }
    }
    return this._adapter;
  }

  async connect() {
    const adapter = this._ensureAdapter();
    if (!adapter) {
      this.device = { connected: true, mode: 'mock' };
      logger.warn('[USBPrinter] 模拟模式连接成功 (escpos-usb不可用)');
      return true;
    }

    try {
      this.device = this.vendorId && this.productId
        ? new adapter.USB(this.vendorId, this.productId)
        : new adapter.USB();

      await new Promise((resolve, reject) => {
        this.device.open((err) => {
          if (err) reject(err);
          else resolve();
        });
      });
      logger.info('[USBPrinter] USB打印机连接成功');
      return true;
    } catch (err) {
      logger.error('[USBPrinter] USB打印机连接失败: %s', err.message);
      this.device = null;
      return false;
    }
  }

  async disconnect() {
    if (this.device && !this.device.mode) {
      try {
        await new Promise((resolve) => {
          this.device.close(() => resolve());
        });
      } catch (err) {
        logger.warn('[USBPrinter] USB断开连接出错: %s', err.message);
      }
      this.device = null;
    }
    this.device = null;
  }

  async isConnected() {
    return this.device !== null;
  }

  async print(printData) {
    const adapter = this._ensureAdapter();

    if (!this.device) {
      const connected = await this.connect();
      if (!connected) {
        throw new Error('无法连接到USB打印机');
      }
    }

    if (this.device.mode === 'mock' || !adapter) {
      const escposBytes = this.buildEscPosBytes(printData);
      logger.info('[USBPrinter] [模拟模式] 已生成ESC/POS数据 (%d bytes) - 实际需要escpos-usb', escposBytes.length);
      this._debugOutputEscPos(escposBytes);
      return { success: true, bytes: escposBytes.length, mode: 'mock' };
    }

    const printer = new adapter.Printer(this.device, { encoding: this.encoding });
    const escposBytes = this.buildEscPosBytes(printData);

    return new Promise((resolve, reject) => {
      try {
        printer.raster.write(escposBytes);
        printer.feed(3).cut().close();
        resolve({ success: true, bytes: escposBytes.length });
      } catch (err) {
        reject(err);
      }
    });
  }

  _debugOutputEscPos(bytes) {
    try {
      const text = bytes.toString('utf-8').replace(/[\x00-\x1F\x7F-\x9F]/g, (c) => {
        if (c === '\n') return '\n';
        if (c === '\r') return '';
        return '';
      });
      if (text.trim()) {
        logger.debug('[USBPrinter] 模拟输出内容:\n' + '-'.repeat(32) + '\n' + text + '-'.repeat(32));
      }
    } catch (_) {}
  }
}

class BluetoothPrinter extends BasePrinter {
  constructor(config) {
    super(config);
    this.address = config.address;
    this.channel = config.channel || 1;
    this.device = null;
  }

  async connect() {
    try {
      const Bluetooth = require('escpos-bluetooth');
      this.device = new Bluetooth(this.address, this.channel);

      await new Promise((resolve, reject) => {
        this.device.open((err) => {
          if (err) reject(err);
          else resolve();
        });
      });
      logger.info('[BluetoothPrinter] 蓝牙打印机连接成功 %s', this.address);
      return true;
    } catch (err) {
      logger.error('[BluetoothPrinter] 蓝牙打印机连接失败 %s, error=%s', this.address, err.message);
      this.device = null;
      return false;
    }
  }

  async disconnect() {
    if (this.device) {
      try {
        await new Promise((resolve) => {
          this.device.close(() => resolve());
        });
      } catch (err) {
        logger.warn('[BluetoothPrinter] 蓝牙断开连接出错: %s', err.message);
      }
      this.device = null;
    }
  }

  async isConnected() {
    return this.device !== null;
  }

  async print(printData) {
    if (!this.device) {
      const connected = await this.connect();
      if (!connected) {
        throw new Error(`无法连接到蓝牙打印机 ${this.address}`);
      }
    }

    const escpos = require('escpos');
    const printer = new escpos.Printer(this.device, { encoding: this.encoding });
    const escposBytes = this.buildEscPosBytes(printData);

    return new Promise((resolve, reject) => {
      try {
        printer.raster.write(escposBytes);
        printer.feed(3).cut().close();
        resolve({ success: true, bytes: escposBytes.length });
      } catch (err) {
        reject(err);
      }
    });
  }
}

class PrinterFactory {
  static create(config) {
    switch (config.connectionType) {
      case PRINTER_CONNECTION_TYPES.NETWORK:
        return new NetworkPrinter(config);
      case PRINTER_CONNECTION_TYPES.USB:
        return new USBPrinter(config);
      case PRINTER_CONNECTION_TYPES.BLUETOOTH:
        return new BluetoothPrinter(config);
      default:
        if (config.ipAddress) {
          return new NetworkPrinter(config);
        }
        return new NetworkPrinter(config);
    }
  }
}

module.exports = {
  BasePrinter,
  NetworkPrinter,
  USBPrinter,
  BluetoothPrinter,
  PrinterFactory,
};
