'use strict';

const config = require('../config');
const logger = require('../utils/logger');

class BackendApiClient {
  constructor() {
    this.baseUrl = config.backend.baseUrl.replace(/\/$/, '');
    this.timeout = config.backend.timeout;
    this.internalToken = config.backend.internalToken;
  }
  buildHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    if (this.internalToken) {
      headers['X-Internal-Token'] = this.internalToken;
    }
    return headers;
  }

  async request(method, path, options = {}) {
    const url = `${this.baseUrl}${path}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const fetchOptions = {
        method,
        headers: {
          ...this.buildHeaders(),
          ...(options.headers || {}),
        },
        signal: controller.signal,
      };

      if (options.body) {
        fetchOptions.body = typeof options.body === 'string'
          ? options.body
          : JSON.stringify(options.body);
      }

      const res = await fetch(url, fetchOptions);
      const text = await res.text();
      let data;
      try {
        data = JSON.parse(text);
      } catch (_) {
        data = text;
      }

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${typeof data === 'object' ? (data.message || data.error || res.statusText) : res.statusText}`);
      }

      return data;
    } catch (err) {
      if (err.name === 'AbortError') {
        throw new Error(`请求超时 (${this.timeout}ms): ${url}`);
      }
      throw err;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async get(path, query = {}) {
    let fullPath = path;
    const params = new URLSearchParams();
    let hasParams = false;
    for (const [k, v] of Object.entries(query)) {
      if (v !== undefined && v !== null && v !== '') {
        params.set(k, v);
        hasParams = true;
      }
    }
    if (hasParams) {
      fullPath += `?${params.toString()}`;
    }
    return this.request('GET', fullPath);
  }

  async getOrderByNo(orderNo) {
    logger.debug('[BackendAPI] 查询订单详情: orderNo=%s', orderNo);
    try {
      const resp = await this.get(`/orders/no/${encodeURIComponent(orderNo)}`);
      if (resp && resp.code === 0 && resp.data) {
        return resp.data;
      }
      if (resp && resp.id) {
        return resp;
      }
      logger.warn('[BackendAPI] 订单查询响应格式异常: orderNo=%s, resp=%j', orderNo, resp);
      return resp && resp.data ? resp.data : resp;
    } catch (err) {
      logger.error('[BackendAPI] 查询订单详情失败: orderNo=%s, error=%s', orderNo, err.message);
      throw err;
    }
  }

  async getOrderForPrint(orderNo) {
    logger.debug('[BackendAPI] 查询打印专用订单数据: orderNo=%s', orderNo);
    try {
      const resp = await this.get(`/orders/no/${encodeURIComponent(orderNo)}/print`);
      if (resp && resp.code === 0 && resp.data) {
        return resp.data.order || resp.data;
      }
      if (resp && resp.order) {
        return resp.order;
      }
      return resp && resp.data ? (resp.data.order || resp.data) : resp;
    } catch (err) {
      logger.warn('[BackendAPI] 打印专用接口失败，回退到普通接口: orderNo=%s, error=%s', orderNo, err.message);
      return await this.getOrderByNo(orderNo);
    }
  }

  async getAllPrinters(storeId = null) {
    logger.debug('[BackendAPI] 同步打印机配置: storeId=%s', storeId);
    try {
      const query = {};
      if (storeId) query.store_id = storeId;
      const resp = await this.get('/internal/printers', query);

      let data;
      if (resp && resp.code === 0 && resp.data) {
        data = resp.data;
      } else {
        data = resp;
      }

      if (Array.isArray(data)) {
        return data;
      }
      if (data && Array.isArray(data.list)) {
        return data.list;
      }
      logger.warn('[BackendAPI] 打印机列表响应格式异常: resp=%j', resp);
      return [];
    } catch (err) {
      logger.error('[BackendAPI] 同步打印机配置失败: error=%s', err.message);
      throw err;
    }
  }

  async getPrinterById(id) {
    logger.debug('[BackendAPI] 查询打印机详情: id=%s', id);
    try {
      const resp = await this.get(`/internal/printers/${id}`);
      if (resp && resp.code === 0 && resp.data) {
        return resp.data;
      }
      return resp;
    } catch (err) {
      logger.error('[BackendAPI] 查询打印机详情失败: id=%s, error=%s', id, err.message);
      throw err;
    }
  }

  async getReceiptAds(storeId, position = 'footer') {
    logger.debug('[BackendAPI] 查询小票广告: storeId=%s, position=%s', storeId, position);
    try {
      const resp = await this.get('/internal/receipt-ads', { store_id: storeId, position });
      let data;
      if (resp && resp.code === 0 && resp.data) {
        data = resp.data;
      } else {
        data = resp;
      }
      if (Array.isArray(data)) {
        return data;
      }
      if (data && Array.isArray(data.list)) {
        return data.list;
      }
      return [];
    } catch (err) {
      logger.warn('[BackendAPI] 查询小票广告失败: storeId=%s, error=%s', storeId, err.message);
      return [];
    }
  }

  async reportAdView(adId) {
    try {
      await this.request('POST', '/internal/receipt-ads/views', {
        body: { ad_id: adId }
      });
    } catch (err) {
      logger.debug('[BackendAPI] 广告展示上报失败: adId=%s, error=%s', adId, err.message);
    }
  }

  async reportAdClick(adId, orderId, orderNo) {
    try {
      await this.request('POST', '/r/' + adId + '/view', {
        body: { order_id: orderId, order_no: orderNo }
      });
    } catch (err) {
      logger.debug('[BackendAPI] 广告点击上报失败: adId=%s, error=%s', adId, err.message);
    }
  }
}

const backendApiClient = new BackendApiClient();

module.exports = backendApiClient;
