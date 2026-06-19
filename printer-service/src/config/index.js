'use strict';

require('dotenv').config();

const config = {
  port: process.env.PORT || 3001,

  nsq: {
    nsqdAddress: process.env.NSQ_NSQD_ADDRESS || '127.0.0.1:4150',
    lookupdAddress: process.env.NSQ_LOOKUPD_ADDRESS || '127.0.0.1:4161',
  },

  redis: {
    host: process.env.REDIS_HOST || '127.0.0.1',
    port: parseInt(process.env.REDIS_PORT) || 6379,
    password: process.env.REDIS_PASSWORD || '',
    db: parseInt(process.env.REDIS_DB) || 0,
  },

  socket: {
    port: parseInt(process.env.SOCKET_IO_PORT) || 3002,
    corsOrigin: process.env.SOCKET_IO_CORS_ORIGIN || '*',
  },

  print: {
    retryInterval: parseInt(process.env.PRINT_RETRY_INTERVAL) || 30000,
    maxRetries: parseInt(process.env.PRINT_MAX_RETRIES) || 5,
  },

  log: {
    level: process.env.LOG_LEVEL || 'info',
  },
};

module.exports = config;
