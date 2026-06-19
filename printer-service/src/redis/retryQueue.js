'use strict';

const { getClient } = require('./index');
const logger = require('../utils/logger');
const config = require('../config');
const { REDIS_KEYS, PRINT_STATUS } = require('../constants');

class RetryQueue {
  constructor() {
    this.retryTimer = null;
    this.isRunning = false;
  }

  async init() {
    this.startRetryLoop();
    logger.info('[RetryQueue] 重试队列已启动，重试间隔: %dms', config.print.retryInterval);
  }

  startRetryLoop() {
    if (this.retryTimer) {
      clearInterval(this.retryTimer);
    }
    this.retryTimer = setInterval(() => {
      this.processRetryQueue().catch((err) => {
        logger.error('[RetryQueue] 处理重试队列失败: %s', err.message);
      });
    }, config.print.retryInterval);
  }

  async enqueue(task) {
    const client = getClient();
    const taskWithRetry = {
      ...task,
      retryCount: task.retryCount || 0,
      maxRetries: task.maxRetries || config.print.maxRetries,
      enqueuedAt: Date.now(),
      status: PRINT_STATUS.PENDING,
    };
    const key = this.getTaskKey(task.id || `${task.orderNo}_${task.printerId}_${Date.now()}`);
    await client.hSet(REDIS_KEYS.PRINT_RETRY_QUEUE, key, JSON.stringify(taskWithRetry));
    logger.info('[RetryQueue] 任务已加入重试队列: key=%s, retryCount=%d', key, taskWithRetry.retryCount);
    return taskWithRetry;
  }

  async dequeue(taskId) {
    const client = getClient();
    const key = this.getTaskKey(taskId);
    await client.hDel(REDIS_KEYS.PRINT_RETRY_QUEUE, key);
    logger.info('[RetryQueue] 任务已从重试队列移除: key=%s', key);
  }

  async processRetryQueue() {
    if (this.isRunning) {
      return;
    }
    this.isRunning = true;

    try {
      const client = getClient();
      const allTasks = await client.hGetAll(REDIS_KEYS.PRINT_RETRY_QUEUE);
      const taskKeys = Object.keys(allTasks);

      if (taskKeys.length === 0) {
        return;
      }

      logger.info('[RetryQueue] 开始处理重试队列，待处理任务数: %d', taskKeys.length);

      const results = [];
      for (const key of taskKeys) {
        const task = JSON.parse(allTasks[key]);
        if (this.shouldRetry(task)) {
          task.retryCount += 1;
          task.status = PRINT_STATUS.RETRYING;
          task.lastRetryAt = Date.now();

          try {
            const result = await this.executeTask(task);
            if (result && result.success) {
              await this.dequeue(key);
              task.status = PRINT_STATUS.SUCCESS;
              logger.info('[RetryQueue] 任务重试成功: key=%s, 重试次数: %d', key, task.retryCount);
            } else {
              await this.updateTask(key, task);
              logger.warn('[RetryQueue] 任务重试失败，稍后重试: key=%s, 重试次数: %d', key, task.retryCount);
            }
          } catch (err) {
            logger.error('[RetryQueue] 任务重试异常: key=%s, error=%s', key, err.message);
            if (task.retryCount >= task.maxRetries) {
              task.status = PRINT_STATUS.FAILED;
              task.error = err.message;
              await this.updateTask(key, task);
              logger.error('[RetryQueue] 任务已达最大重试次数，放弃: key=%s', key);
            } else {
              await this.updateTask(key, task);
            }
          }
        } else if (task.retryCount >= task.maxRetries) {
          task.status = PRINT_STATUS.FAILED;
          await this.updateTask(key, task);
          logger.error('[RetryQueue] 任务已达最大重试次数，放弃: key=%s', key);
        }
      }

      return results;
    } finally {
      this.isRunning = false;
    }
  }

  shouldRetry(task) {
    return task.retryCount < task.maxRetries && task.status !== PRINT_STATUS.SUCCESS;
  }

  async updateTask(key, task) {
    const client = getClient();
    await client.hSet(REDIS_KEYS.PRINT_RETRY_QUEUE, key, JSON.stringify(task));
  }

  async getTask(taskId) {
    const client = getClient();
    const key = this.getTaskKey(taskId);
    const data = await client.hGet(REDIS_KEYS.PRINT_RETRY_QUEUE, key);
    return data ? JSON.parse(data) : null;
  }

  async getAllTasks() {
    const client = getClient();
    const allTasks = await client.hGetAll(REDIS_KEYS.PRINT_RETRY_QUEUE);
    return Object.values(allTasks).map((t) => JSON.parse(t));
  }

  async clearAll() {
    const client = getClient();
    await client.del(REDIS_KEYS.PRINT_RETRY_QUEUE);
    logger.info('[RetryQueue] 重试队列已清空');
  }

  setTaskExecutor(executor) {
    this.taskExecutor = executor;
  }

  async executeTask(task) {
    if (this.taskExecutor) {
      return await this.taskExecutor(task);
    }
    logger.warn('[RetryQueue] 未设置任务执行器');
    return { success: false };
  }

  getTaskKey(id) {
    return `${REDIS_KEYS.PRINT_RETRY_QUEUE}:${id}`;
  }

  stop() {
    if (this.retryTimer) {
      clearInterval(this.retryTimer);
      this.retryTimer = null;
      logger.info('[RetryQueue] 重试队列已停止');
    }
  }
}

const retryQueue = new RetryQueue();

module.exports = retryQueue;
