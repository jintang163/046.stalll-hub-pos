'use strict';

const fs = require('fs');
const path = require('path');
const logger = require('../utils/logger');
const { getClient } = require('../redis');
const { REDIS_KEYS, CATEGORY_TYPES, PRINTER_CONNECTION_TYPES } = require('../constants');
const { PrinterFactory } = require('../printers/Printer');

class PrinterConfigManager {
  constructor() {
    this.printers = new Map();
    this.categoryPrinterMap = new Map();
    this.configFilePath = path.join(process.cwd(), 'data', 'printers.json');
    this.ensureConfigDir();
  }

  ensureConfigDir() {
    const dir = path.dirname(this.configFilePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
  }

  async init() {
    await this.loadFromRedis();
    await this.loadFromFile();
    logger.info('[PrinterConfigManager] 已加载 %d 台打印机配置', this.printers.size);
    this.initPrinterInstances();
  }

  async loadFromRedis() {
    try {
      const client = getClient();
      const data = await client.get(REDIS_KEYS.PRINTER_CONFIG);
      if (data) {
        const configs = JSON.parse(data);
        for (const config of configs) {
          this.printers.set(config.id, config);
        }
      }

      const categoryMap = await client.get(REDIS_KEYS.CATEGORY_PRINTER_MAP);
      if (categoryMap) {
        const map = JSON.parse(categoryMap);
        for (const [category, printerIds] of Object.entries(map)) {
          this.categoryPrinterMap.set(category, printerIds);
        }
      }
    } catch (err) {
      logger.warn('[PrinterConfigManager] 从Redis加载配置失败: %s', err.message);
    }
  }

  loadFromFile() {
    try {
      if (fs.existsSync(this.configFilePath)) {
        const content = fs.readFileSync(this.configFilePath, 'utf-8');
        const data = JSON.parse(content);
        if (data.printers) {
          for (const config of data.printers) {
            if (!this.printers.has(config.id)) {
              this.printers.set(config.id, config);
            }
          }
        }
        if (data.categoryPrinterMap) {
          for (const [category, printerIds] of Object.entries(data.categoryPrinterMap)) {
            if (!this.categoryPrinterMap.has(category)) {
              this.categoryPrinterMap.set(category, printerIds);
            }
          }
        }
      }
    } catch (err) {
      logger.warn('[PrinterConfigManager] 从文件加载配置失败: %s', err.message);
    }
  }

  async saveToRedis() {
    try {
      const client = getClient();
      const printers = Array.from(this.printers.values());
      await client.set(REDIS_KEYS.PRINTER_CONFIG, JSON.stringify(printers));

      const categoryMap = {};
      for (const [category, printerIds] of this.categoryPrinterMap.entries()) {
        categoryMap[category] = printerIds;
      }
      await client.set(REDIS_KEYS.CATEGORY_PRINTER_MAP, JSON.stringify(categoryMap));
    } catch (err) {
      logger.error('[PrinterConfigManager] 保存配置到Redis失败: %s', err.message);
    }
  }

  saveToFile() {
    try {
      const data = {
        printers: Array.from(this.printers.values()),
        categoryPrinterMap: Object.fromEntries(this.categoryPrinterMap),
        updatedAt: Date.now(),
      };
      fs.writeFileSync(this.configFilePath, JSON.stringify(data, null, 2), 'utf-8');
    } catch (err) {
      logger.error('[PrinterConfigManager] 保存配置到文件失败: %s', err.message);
    }
  }

  initPrinterInstances() {
    for (const config of this.printers.values()) {
      if (!config.instance) {
        try {
          config.instance = PrinterFactory.create(config);
        } catch (err) {
          logger.error('[PrinterConfigManager] 创建打印机实例失败 id=%d: %s', config.id, err.message);
        }
      }
    }
  }

  addPrinter(config) {
    const id = config.id || Date.now();
    const printerConfig = {
      ...config,
      id,
      createdAt: config.createdAt || Date.now(),
      updatedAt: Date.now(),
    };

    try {
      printerConfig.instance = PrinterFactory.create(printerConfig);
    } catch (err) {
      logger.error('[PrinterConfigManager] 创建打印机实例失败: %s', err.message);
    }

    this.printers.set(id, printerConfig);
    this.saveToRedis();
    this.saveToFile();
    logger.info('[PrinterConfigManager] 已添加打印机: id=%d, name=%s', id, config.name);
    return printerConfig;
  }

  updatePrinter(id, updates) {
    if (!this.printers.has(id)) {
      throw new Error(`打印机不存在: id=${id}`);
    }

    const existing = this.printers.get(id);
    const updated = {
      ...existing,
      ...updates,
      id,
      updatedAt: Date.now(),
    };

    if (updates.connectionType || updates.ipAddress || updates.port ||
        updates.vendorId || updates.productId || updates.address) {
      try {
        if (existing.instance) {
          existing.instance.disconnect().catch(() => {});
        }
        updated.instance = PrinterFactory.create(updated);
      } catch (err) {
        logger.error('[PrinterConfigManager] 更新打印机实例失败 id=%d: %s', id, err.message);
        updated.instance = existing.instance;
      }
    }

    this.printers.set(id, updated);
    this.saveToRedis();
    this.saveToFile();
    logger.info('[PrinterConfigManager] 已更新打印机: id=%d', id);
    return updated;
  }

  removePrinter(id) {
    const printer = this.printers.get(id);
    if (printer && printer.instance) {
      printer.instance.disconnect().catch(() => {});
    }

    this.printers.delete(id);

    for (const [category, printerIds] of this.categoryPrinterMap.entries()) {
      const filtered = printerIds.filter((pid) => pid !== id);
      if (filtered.length === 0) {
        this.categoryPrinterMap.delete(category);
      } else {
        this.categoryPrinterMap.set(category, filtered);
      }
    }

    this.saveToRedis();
    this.saveToFile();
    logger.info('[PrinterConfigManager] 已删除打印机: id=%d', id);
  }

  getPrinter(id) {
    return this.printers.get(id);
  }

  getAllPrinters(storeId = null) {
    const printers = Array.from(this.printers.values());
    if (storeId) {
      return printers.filter((p) => p.storeId === storeId);
    }
    return printers;
  }

  getPrintersByCategory(category, storeId = null) {
    const printerIds = this.categoryPrinterMap.get(category) || [];
    const printers = printerIds
      .map((id) => this.printers.get(id))
      .filter((p) => p && p.status === 1);

    if (storeId) {
      return printers.filter((p) => p.storeId === storeId);
    }

    if (printers.length === 0) {
      return this.getDefaultPrinters(storeId);
    }

    return printers;
  }

  getDefaultPrinters(storeId = null) {
    const all = this.getAllPrinters(storeId);
    return all.filter((p) => p.isDefault && p.status === 1);
  }

  setCategoryPrinters(category, printerIds) {
    this.categoryPrinterMap.set(category, printerIds);
    this.saveToRedis();
    this.saveToFile();
    logger.info('[PrinterConfigManager] 已设置分类打印机映射: category=%s, printers=%s', category, printerIds);
  }

  getCategoryPrinterMap() {
    return Object.fromEntries(this.categoryPrinterMap);
  }

  getPrinterIdsByItemCategory(item) {
    const category = this.detectItemCategory(item);
    return this.getPrintersByCategory(category, item.storeId);
  }

  detectItemCategory(item) {
    if (item.categoryType) {
      return item.categoryType;
    }
    if (item.categoryName) {
      const name = item.categoryName.toLowerCase();
      if (name.includes('热') || name.includes('炒') || name.includes('烧') || name.includes('蒸')) {
        return CATEGORY_TYPES.HOT_DISH;
      }
      if (name.includes('凉') || name.includes('冷')) {
        return CATEGORY_TYPES.COLD_DISH;
      }
      if (name.includes('饮') || name.includes('酒') || name.includes('水') || name.includes('茶') || name.includes('汁') || name.includes('咖啡')) {
        return CATEGORY_TYPES.DRINK;
      }
      if (name.includes('主食') || name.includes('饭') || name.includes('面') || name.includes('粉')) {
        return CATEGORY_TYPES.STAPLE;
      }
      if (name.includes('汤')) {
        return CATEGORY_TYPES.SOUP;
      }
    }
    if (item.categoryId) {
      return `category_${item.categoryId}`;
    }
    return CATEGORY_TYPES.OTHER;
  }

  async getPrinterStatus(id) {
    const printer = this.printers.get(id);
    if (!printer) {
      return { online: false, error: '打印机不存在' };
    }
    if (!printer.instance) {
      return { online: false, error: '打印机实例未初始化' };
    }
    try {
      const connected = await printer.instance.isConnected();
      return {
        online: connected,
        name: printer.name,
        type: printer.connectionType,
      };
    } catch (err) {
      return { online: false, error: err.message };
    }
  }
}

const printerConfigManager = new PrinterConfigManager();

module.exports = printerConfigManager;
