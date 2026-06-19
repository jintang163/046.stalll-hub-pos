'use strict';

const nsq = require('nsqjs');
const config = require('../config');
const logger = require('../utils/logger');
const { NSQ_TOPICS, NSQ_CHANNELS } = require('../constants');

class NSQConsumer {
  constructor() {
    this.consumers = new Map();
    this.readers = [];
  }

  async init() {
    logger.info('[NSQConsumer] NSQ消费者初始化中...');
  }

  subscribe(topic, channel, handler) {
    const reader = new nsq.Reader(topic, channel, {
      lookupdHTTPAddresses: config.nsq.lookupdAddress,
      maxInFlight: 100,
    });

    reader.on('message', async (msg) => {
      try {
        const data = JSON.parse(msg.body.toString());
        logger.debug('[NSQConsumer] 收到消息 topic=%s, channel=%s', topic, channel);
        const result = await handler(data, msg);
        if (result !== false) {
          msg.finish();
        }
      } catch (err) {
        logger.error('[NSQConsumer] 处理消息失败 topic=%s, error=%s', topic, err.message);
        msg.requeue(3000);
      }
    });

    reader.on('error', (err) => {
      logger.error('[NSQConsumer] Reader错误 topic=%s, error=%s', topic, err.message);
    });

    reader.on('nsqd_connected', (host, port) => {
      logger.info('[NSQConsumer] 已连接到 nsqd %s:%s for topic=%s', host, port, topic);
    });

    reader.on('nsqd_closed', (host, port) => {
      logger.warn('[NSQConsumer] nsqd 连接关闭 %s:%s for topic=%s', host, port, topic);
    });

    reader.connect();
    this.readers.push(reader);
    this.consumers.set(`${topic}:${channel}`, reader);
    logger.info('[NSQConsumer] 已订阅 topic=%s, channel=%s', topic, channel);
    return reader;
  }

  subscribeOrderCreated(handler) {
    return this.subscribe(NSQ_TOPICS.ORDER_CREATED, NSQ_CHANNELS.PRINTER_SERVICE, handler);
  }

  subscribeOrderPaid(handler) {
    return this.subscribe(NSQ_TOPICS.ORDER_PAID, NSQ_CHANNELS.PRINTER_SERVICE, handler);
  }

  subscribePrintOrder(handler) {
    return this.subscribe(NSQ_TOPICS.PRINT_ORDER, NSQ_CHANNELS.PRINTER_SERVICE, handler);
  }

  subscribeOrderUpdate(handler) {
    return this.subscribe(NSQ_TOPICS.ORDER_UPDATE, NSQ_CHANNELS.PRINTER_SERVICE, handler);
  }

  close() {
    for (const reader of this.readers) {
      reader.close();
    }
    this.readers = [];
    this.consumers.clear();
    logger.info('[NSQConsumer] 所有消费者已关闭');
  }
}

const nsqConsumer = new NSQConsumer();

module.exports = nsqConsumer;
