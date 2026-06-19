'use strict';

const fs = require('fs');
const path = require('path');
const logger = require('../utils/logger');
const config = require('../config');
const { getClient } = require('../redis');
const { REDIS_KEYS, CATEGORY_TYPES, PRINTER_CONNECTION_TYPES } = require('../constants');
const { PrinterFactory } = require('../printers/Printer');
const backendApiClient = require('../api/backendClient');

class PrinterConfigManager {
  constructor() {
    this.printers = new Map();
    this.categoryPrinterMap = new Map();
    this.configFilePath = path.join(process.cwd(), 'data', 'printers.json');
    this._syncTimer = null;
    this._lastSyncedAt = 0;
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
    this.initPrinterInstances();
    logger.info('[PrinterConfigManager] 已加载 %d 台打印机配置 (来自Redis/本地文件)', this.printers.size);

    try {
      const synced = await this.syncFromBackend();
      if (synced) {
        logger.info('[PrinterConfigManager] 启动阶段已从后端API同步打印机配置');
      }
    } catch (err) {
      logger.warn('[PrinterConfigManager] 启动阶段后端API同步失败，继续使用本地配置: %s', err.message);
    }

    this.startPeriodicSync();
  }

  startPeriodicSync() {
    const interval = config.backend.printerSyncInterval;
    if (!interval || interval <= 0) {
      logger.info('[PrinterConfigManager] 后端打印机配置定时同步已禁用 (interval=%s)', interval);
      return;
    }

    if (this._syncTimer) {
      clearInterval(this._syncTimer);
    }

    this._syncTimer = setInterval(async () => {
      try {
        await this.syncFromBackend();
      } catch (err) {
        logger.warn('[PrinterConfigManager] 定时同步打印机配置出错: %s', err.message);
      }
    }, interval);

    logger.info('[PrinterConfigManager] 后端打印机配置定时同步已启动，间隔 %d ms', interval);
  }

  stopPeriodicSync() {
    if (this._syncTimer) {
      clearInterval(this._syncTimer);
      this._syncTimer = null;
      logger.info('[PrinterConfigManager] 后端打印机配置定时同步已停止');
    }
  }

  async syncFromBackend(storeId = null) {
    try {
      const rawList = await backendApiClient.getAllPrinters(storeId);
      if (!Array.isArray(rawList) || rawList.length === 0) {
        logger.debug('[PrinterConfigManager] 后端返回空打印机列表，跳过同步');
        return false;
      }

      const normalized = rawList
        .map((raw) => this.normalizeBackendPrinter(raw))
        .filter((p) => p && p.id);

      if (normalized.length === 0) {
        logger.warn('[PrinterConfigManager] 后端返回打印机数据全部无效，跳过同步');
        return false;
      }

      await this._disconnectAllPrinters();

      this.printers.clear();
      for (const cfg of normalized) {
        try {
          cfg.instance = PrinterFactory.create(cfg);
        } catch (err) {
          logger.error('[PrinterConfigManager] 创建打印机实例失败 id=%s: %s', cfg.id, err.message);
          cfg.instance = null;
        }
        this.printers.set(cfg.id, cfg);
      }

      this._rebuildCategoryMapFromPrinters();

      this._lastSyncedAt = Date.now();

      logger.info('[PrinterConfigManager] 后端API同步完成，共 %d 台打印机，分类映射 %d 条',
        this.printers.size, this.categoryPrinterMap.size);

      await this.saveToRedis();
      this.saveToFile();

      return true;
    } catch (err) {
      logger.error('[PrinterConfigManager] 从后端同步打印机配置失败: %s', err.message);
      throw err;
    }
  }

  _rebuildCategoryMapFromPrinters() {
    this.categoryPrinterMap.clear();
    for (const printer of this.printers.values()) {
      if (printer.status !== 1) continue;
      const categoryTypes = printer.categoryTypes || printer.categories || [];
      for (const ct of categoryTypes) {
        if (!this.categoryPrinterMap.has(ct)) {
          this.categoryPrinterMap.set(ct, []);
        }
        this.categoryPrinterMap.get(ct).push(printer.id);
      }
    }
  }

  async _disconnectAllPrinters() {
    for (const cfg of this.printers.values()) {
      if (cfg.instance && typeof cfg.instance.disconnect === 'function') {
        try {
          await cfg.instance.disconnect();
        } catch (_) {}
      }
      cfg.instance = null;
    }
  }

  normalizeBackendPrinter(raw) {
    if (!raw) return null;
    const id = raw.id || raw.ID;
    if (!id) return null;

    const connectionType =
      raw.connection_type || raw.connectionType ||
      (raw.ip_address || raw.ipAddress ? PRINTER_CONNECTION_TYPES.NETWORK :
       raw.vendor_id || raw.vendorId ? PRINTER_CONNECTION_TYPES.USB :
       raw.address ? PRINTER_CONNECTION_TYPES.BLUETOOTH : PRINTER_CONNECTION_TYPES.NETWORK);

    return {
      id: id,
      name: raw.name || raw.Name || `Printer-${id}`,
      storeId: raw.store_id || raw.storeId || raw.StoreID || 0,
      type: raw.type || raw.Type || 'kitchen',
      connectionType,
      printType: raw.print_type || raw.printType || raw.PrintType || 'kitchen',
      status: raw.status !== undefined ? (Number(raw.status) || 0) : 1,
      copies: Number(raw.copies || raw.Copies) || 1,
      width: Number(raw.width || raw.Width) || 80,
      encoding: raw.encoding || raw.Encoding || 'UTF-8',
      isDefault: (raw.is_default === true || raw.isDefault === true || raw.IsDefault === true),
      ipAddress: raw.ip_address || raw.ipAddress || raw.IpAddress || '',
      port: Number(raw.port || raw.Port) || 9100,
      vendorId: raw.vendor_id || raw.vendorId || raw.VendorID,
      productId: raw.product_id || raw.productId || raw.ProductID,
      address: raw.address || raw.Address || '',
      channel: Number(raw.channel || raw.Channel) || 1,
      categoryTypes: raw.category_types || raw.categoryTypes || raw.CategoryTypes || [],
      createdAt: raw.created_at || raw.createdAt || Date.now(),
      updatedAt: raw.updated_at || raw.updatedAt || Date.now(),
      _raw: raw,
    };
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
