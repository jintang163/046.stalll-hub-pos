'use strict';

const NSQ_TOPICS = {
  PRODUCT_CHANGED: 'product_change',
  STOCK_CHANGE: 'stock_change',
  ORDER_CREATED: 'order_create',
  ORDER_STATUS: 'order_status',
  ORDER_PAID: 'order_pay_success',
  ORDER_REFUND: 'order_refund',
  PRINT_ORDER: 'print_order',
  SYNC_PRODUCT: 'sync_product',
  MEMBER_POINTS: 'member_points',
  STOCK_WARNING: 'stock_warning',
  ORDER_UPDATE: 'order_update',
};

const NSQ_CHANNELS = {
  PRINTER_SERVICE: 'printer_service',
};

const CATEGORY_TYPES = {
  HOT_DISH: 'hot_dish',
  COLD_DISH: 'cold_dish',
  DRINK: 'drink',
  STAPLE: 'staple',
  SOUP: 'soup',
  SNACK: 'snack',
  OTHER: 'other',
};

const PRINTER_CONNECTION_TYPES = {
  NETWORK: 'network',
  USB: 'usb',
  BLUETOOTH: 'bluetooth',
};

const PRINTER_PRINT_TYPES = {
  KITCHEN: 'kitchen',
  BAR: 'bar',
  RECEIPT: 'receipt',
};

const REDIS_KEYS = {
  PRINTER_CONFIG: 'printer:config',
  PRINT_RETRY_QUEUE: 'print:retry:queue',
  PRINT_STATUS: 'print:status',
  CATEGORY_PRINTER_MAP: 'printer:category:map',
};

const PRINT_STATUS = {
  PENDING: 'pending',
  PRINTING: 'printing',
  SUCCESS: 'success',
  FAILED: 'failed',
  RETRYING: 'retrying',
};

const ORDER_TYPES = {
  DINE_IN: 'dine_in',
  TAKEAWAY: 'takeaway',
  DELIVERY: 'delivery',
};

module.exports = {
  NSQ_TOPICS,
  NSQ_CHANNELS,
  CATEGORY_TYPES,
  PRINTER_CONNECTION_TYPES,
  PRINTER_PRINT_TYPES,
  REDIS_KEYS,
  PRINT_STATUS,
  ORDER_TYPES,
};
