'use strict';

const logger = require('../utils/logger');
const { PRINTER_CONNECTION_TYPES } = require('../constants');

class BasePrinter {
  constructor(config) {
    this.id = config.id;
    this.name = config.name;
    this.storeId = config.storeId;
    this.type = config.type;
    this.connectionType = config.connectionType;
    this.printType = config.printType;
    this.status = config.status || 1;
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
    const testData = this.generateTestData();
    return await this.print(testData);
  }

  generateTestData() {
    return {
      type: 'test',
      title: '测试打印',
      content: `打印机: ${this.name}\n类型: ${this.connectionType}\n时间: ${new Date().toLocaleString()}\n打印成功！`,
    };
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
    try {
      const Network = require('escpos-network');
      this.device = new Network(this.ipAddress, this.port);
      await new Promise((resolve, reject) => {
        this.device.open((err) => {
          if (err) reject(err);
          else resolve();
        });
      });
      logger.info('[NetworkPrinter] 打印机连接成功 %s:%s', this.ipAddress, this.port);
      return true;
    } catch (err) {
      logger.error('[NetworkPrinter] 打印机连接失败 %s:%s, error=%s', this.ipAddress, this.port, err.message);
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
        logger.warn('[NetworkPrinter] 断开连接出错: %s', err.message);
      }
      this.device = null;
    }
  }

  async isConnected() {
    return this.device !== null;
  }

  async print(printData) {
    const escpos = require('escpos');
    const iconv = require('iconv-lite');

    if (!this.device) {
      const connected = await this.connect();
      if (!connected) {
        throw new Error(`无法连接到打印机 ${this.ipAddress}:${this.port}`);
      }
    }

    const printer = new escpos.Printer(this.device);

    if (this.encoding === 'GBK' || this.encoding === 'GB2312') {
      printer.encoding = this.encoding;
    }

    return await new Promise((resolve, reject) => {
      try {
        this.renderPrintContent(printer, escpos, iconv, printData);
        printer.feed(3).cut().close();
        resolve({ success: true });
      } catch (err) {
        reject(err);
      }
    });
  }

  renderPrintContent(printer, escpos, iconv, printData) {
    const { type, data } = printData;

    printer.align('ct').style('bu').size(1, 1);
    printer.text(iconv.encode(data.storeName || '门店', this.encoding));
    printer.text(iconv.encode(`单号: ${data.orderNo || ''}`, this.encoding));
    printer.feed(1);

    printer.align('lt').style('normal');
    printer.text(iconv.encode('===============', this.encoding));

    if (type === 'kitchen' || type === 'order') {
      printer.text(iconv.encode('后厨单', this.encoding));
      if (data.tableNo) {
        printer.text(iconv.encode(`桌号: ${data.tableNo}`, this.encoding));
      }
      printer.text(iconv.encode(`类型: ${this.getOrderTypeName(data.orderType)}`, this.encoding));
      printer.text(iconv.encode('===============', this.encoding));
      printer.text(iconv.encode('菜品              数量', this.encoding));
      printer.text(iconv.encode('---------------', this.encoding));

      for (const item of data.items || []) {
        const name = item.skuName && item.skuName !== item.productName
          ? `${item.productName}(${item.skuName})`
          : item.productName;
        const truncatedName = this.truncate(name, 16);
        const line = `${truncatedName.padEnd(16)} ${String(item.quantity).padStart(4)}`;
        printer.text(iconv.encode(line, this.encoding));
      }
    } else if (type === 'receipt') {
      printer.text(iconv.encode('结账单', this.encoding));
      if (data.tableNo) {
        printer.text(iconv.encode(`桌号: ${data.tableNo}`, this.encoding));
      }
      printer.text(iconv.encode(`类型: ${this.getOrderTypeName(data.orderType)}`, this.encoding));
      printer.text(iconv.encode('===============', this.encoding));
      printer.text(iconv.encode('菜品        数量  单价  金额', this.encoding));
      printer.text(iconv.encode('---------------------------', this.encoding));

      for (const item of data.items || []) {
        const name = this.truncate(item.productName, 10);
        const line = `${name.padEnd(10)} ${String(item.quantity).padStart(2)} ${String(item.price).padStart(5)} ${String(item.subtotal).padStart(6)}`;
        printer.text(iconv.encode(line, this.encoding));
      }

      printer.text(iconv.encode('---------------------------', this.encoding));
      printer.text(iconv.encode(`合计: ${data.totalAmount || '0.00'}`, this.encoding));
      printer.text(iconv.encode(`应收: ${data.payAmount || '0.00'}`, this.encoding));
      if (data.payStatus === 1) {
        printer.text(iconv.encode(`支付: ${this.getPayMethodName(data.payMethod)}`, this.encoding));
      }
    }

    printer.feed(1);
    printer.align('ct');
    printer.text(iconv.encode(type === 'kitchen' ? '请及时备菜' : '欢迎下次光临', this.encoding));
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

  truncate(str, maxLen) {
    if (!str) return '';
    if (str.length <= maxLen) return str;
    return str.slice(0, maxLen - 3) + '...';
  }
}

class USBPrinter extends BasePrinter {
  constructor(config) {
    super(config);
    this.vendorId = config.vendorId;
    this.productId = config.productId;
    this.device = null;
  }

  async connect() {
    try {
      const USB = require('escpos-usb');
      this.device = this.vendorId && this.productId
        ? new USB(this.vendorId, this.productId)
        : new USB();
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
    if (this.device) {
      try {
        await new Promise((resolve) => {
          this.device.close(() => resolve());
        });
      } catch (err) {
        logger.warn('[USBPrinter] USB断开连接出错: %s', err.message);
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
        throw new Error('无法连接到USB打印机');
      }
    }

    const escpos = require('escpos');
    const iconv = require('iconv-lite');
    const printer = new escpos.Printer(this.device);

    return await new Promise((resolve, reject) => {
      try {
        const networkPrinter = new NetworkPrinter(this.config);
        networkPrinter.device = this.device;
        networkPrinter.renderPrintContent(printer, escpos, iconv, printData);
        printer.feed(3).cut().close();
        resolve({ success: true });
      } catch (err) {
        reject(err);
      }
    });
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
    const iconv = require('iconv-lite');
    const printer = new escpos.Printer(this.device);

    return await new Promise((resolve, reject) => {
      try {
        const networkPrinter = new NetworkPrinter(this.config);
        networkPrinter.device = this.device;
        networkPrinter.renderPrintContent(printer, escpos, iconv, printData);
        printer.feed(3).cut().close();
        resolve({ success: true });
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
