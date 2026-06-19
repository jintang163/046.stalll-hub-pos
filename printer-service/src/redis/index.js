'use strict';

const redis = require('redis');
const config = require('../config');
const logger = require('../utils/logger');

let client = null;

async function initRedis() {
  const options = {
    socket: {
      host: config.redis.host,
      port: config.redis.port,
    },
    database: config.redis.db,
  };

  if (config.redis.password) {
    options.password = config.redis.password;
  }

  client = redis.createClient(options);

  client.on('error', (err) => {
    logger.error('[Redis] 连接错误: %s', err.message);
  });

  client.on('connect', () => {
    logger.info('[Redis] 连接成功: %s:%s', config.redis.host, config.redis.port);
  });

  client.on('reconnecting', () => {
    logger.warn('[Redis] 正在重新连接...');
  });

  await client.connect();
  return client;
}

function getClient() {
  if (!client) {
    throw new Error('Redis client not initialized');
  }
  return client;
}

async function closeRedis() {
  if (client) {
    await client.quit();
    logger.info('[Redis] 连接已关闭');
  }
}

module.exports = {
  initRedis,
  getClient,
  closeRedis,
};
