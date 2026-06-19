'use strict';

const net = require('net');

const PORT = 9100;

const server = net.createServer((socket) => {
  console.log(`[MockPrinter] 打印机客户端已连接: ${socket.remoteAddress}:${socket.remotePort}`);

  let buffer = Buffer.alloc(0);

  socket.on('data', (data) => {
    buffer = Buffer.concat([buffer, data]);
    console.log(`[MockPrinter] 收到打印数据 (${data.length} bytes)`);

    try {
      const text = data.toString('utf-8').replace(/[\x00-\x1F\x7F]/g, '');
      if (text.trim()) {
        console.log('[MockPrinter] 打印内容:');
        console.log('---------------------------');
        console.log(text);
        console.log('---------------------------');
      }
    } catch (err) {
      console.error('[MockPrinter] 解析打印数据失败:', err.message);
    }
  });

  socket.on('end', () => {
    console.log('[MockPrinter] 客户端已断开连接');
    buffer = Buffer.alloc(0);
  });

  socket.on('error', (err) => {
    console.error('[MockPrinter] Socket错误:', err.message);
  });
});

server.listen(PORT, () => {
  console.log(`[MockPrinter] 模拟打印机服务已启动，监听端口: ${PORT}`);
  console.log('[MockPrinter] 等待打印任务...');
});

server.on('error', (err) => {
  if (err.code === 'EADDRINUSE') {
    console.error(`[MockPrinter] 端口 ${PORT} 已被占用，请检查是否已有其他打印机服务运行`);
  } else {
    console.error('[MockPrinter] 服务启动失败:', err.message);
  }
  process.exit(1);
});
