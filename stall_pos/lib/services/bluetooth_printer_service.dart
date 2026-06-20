import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';
import 'package:esc_pos_utils/esc_pos_utils.dart';
import '../models/order.dart';

class BluetoothPrinterService {
  BluetoothDevice? _connectedDevice;
  BluetoothCharacteristic? _writeCharacteristic;
  bool _isConnected = false;

  bool get isConnected => _isConnected;
  BluetoothDevice? get connectedDevice => _connectedDevice;

  Future<List<BluetoothDevice>> scanPrinters({Duration timeout = const Duration(seconds: 10)}) async {
    try {
      await FlutterBluePlus.startScan(timeout: timeout);
      
      final results = await FlutterBluePlus.scanResults.first;
      final devices = results
          .map((r) => r.device)
          .where((d) => _isPrinterDevice(d))
          .toList();
      
      await FlutterBluePlus.stopScan();
      return devices;
    } catch (e) {
      print('[Bluetooth] Scan error: $e');
      return [];
    }
  }

  bool _isPrinterDevice(BluetoothDevice device) {
    final name = device.localName.toLowerCase();
    return name.contains('printer') ||
        name.contains('print') ||
        name.contains('bt-') ||
        name.contains('mht') ||
        name.contains('xp-') ||
        name.contains('zj-');
  }

  Future<bool> connect(BluetoothDevice device) async {
    try {
      await device.connect(autoConnect: false);
      
      final services = await device.discoverServices();
      
      for (var service in services) {
        for (var characteristic in service.characteristics) {
          if (characteristic.properties.write ||
              characteristic.properties.writeWithoutResponse) {
            _writeCharacteristic = characteristic;
            _connectedDevice = device;
            _isConnected = true;
            print('[Bluetooth] Connected to printer: ${device.localName}');
            return true;
          }
        }
      }
      
      await device.disconnect();
      return false;
    } catch (e) {
      print('[Bluetooth] Connect error: $e');
      return false;
    }
  }

  Future<void> disconnect() async {
    if (_connectedDevice != null) {
      await _connectedDevice!.disconnect();
      _connectedDevice = null;
      _writeCharacteristic = null;
      _isConnected = false;
    }
  }

  Future<void> printOrder(Order order, String stallName) async {
    if (!_isConnected || _writeCharacteristic == null) {
      throw Exception('打印机未连接');
    }

    final profile = await CapabilityProfile.load();
    final generator = Generator(PaperSize.mm58, profile);
    List<int> bytes = [];

    bytes += generator.text(
      stallName,
      styles: const PosStyles(
        align: PosAlign.center,
        bold: true,
        height: PosTextSize.size2,
        width: PosTextSize.size2,
      ),
    );

    bytes += generator.text(
      '摊位小票',
      styles: const PosStyles(
        align: PosAlign.center,
      ),
    );

    bytes += generator.hr();

    bytes += generator.text('订单号: ${order.orderNo}');
    bytes += generator.text(
      '时间: ${_formatDateTime(order.createdAt ?? DateTime.now())}',
    );
    if (order.tableNo != null && order.tableNo!.isNotEmpty) {
      bytes += generator.text('桌号: ${order.tableNo}');
    }

    bytes += generator.hr();

    bytes += generator.row([
      PosColumn(text: '商品', width: 6, styles: const PosStyles(bold: true)),
      PosColumn(text: '数量', width: 2, styles: const PosStyles(bold: true, align: PosAlign.center)),
      PosColumn(text: '金额', width: 4, styles: const PosStyles(bold: true, align: PosAlign.right)),
    ]);

    for (var item in (order.items ?? [])) {
      bytes += generator.row([
        PosColumn(text: item.productName, width: 6),
        PosColumn(text: '${item.quantity}', width: 2, styles: const PosStyles(align: PosAlign.center)),
        PosColumn(text: item.amount.toString(), width: 4, styles: const PosStyles(align: PosAlign.right)),
      ]);
      if (item.skuSpec != null && item.skuSpec!.isNotEmpty) {
        bytes += generator.text('  ${item.skuSpec}', styles: const PosStyles(fontSize: PosTextSize.size1));
      }
    }

    bytes += generator.hr();

    bytes += generator.row([
      PosColumn(text: '合计', width: 6, styles: const PosStyles(bold: true)),
      PosColumn(
        text: '¥${order.totalAmount}',
        width: 6,
        styles: const PosStyles(bold: true, align: PosAlign.right),
      ),
    ]);

    bytes += generator.hr(ch: '-');

    if (order.payMethod != null) {
      bytes += generator.text('支付方式: ${_getPayMethodText(order.payMethod!)}');
    }
    bytes += generator.text('订单状态: ${_getStatusText(order.status)}');

    bytes += generator.feed(2);
    bytes += generator.text(
      '欢迎下次光临',
      styles: const PosStyles(align: PosAlign.center),
    );
    bytes += generator.feed(3);
    bytes += generator.cut();

    await _writeData(bytes);
  }

  Future<void> printTest() async {
    if (!_isConnected || _writeCharacteristic == null) {
      throw Exception('打印机未连接');
    }

    final profile = await CapabilityProfile.load();
    final generator = Generator(PaperSize.mm58, profile);
    List<int> bytes = [];

    bytes += generator.text(
      '测试打印',
      styles: const PosStyles(
        align: PosAlign.center,
        bold: true,
        height: PosTextSize.size2,
        width: PosTextSize.size2,
      ),
    );
    bytes += generator.feed(1);
    bytes += generator.text('蓝牙打印测试成功！', styles: const PosStyles(align: PosAlign.center));
    bytes += generator.text('打印机连接正常', styles: const PosStyles(align: PosAlign.center));
    bytes += generator.feed(3);
    bytes += generator.cut();

    await _writeData(bytes);
  }

  Future<void> _writeData(List<int> data) async {
    if (_writeCharacteristic == null) return;

    const chunkSize = 20;
    for (var i = 0; i < data.length; i += chunkSize) {
      final end = (i + chunkSize < data.length) ? i + chunkSize : data.length;
      final chunk = data.sublist(i, end);
      
      await _writeCharacteristic!.write(
        chunk,
        withoutResponse: _writeCharacteristic!.properties.writeWithoutResponse,
      );
      
      await Future.delayed(const Duration(milliseconds: 10));
    }
  }

  String _formatDateTime(DateTime dateTime) {
    return '${dateTime.year}-${dateTime.month.toString().padLeft(2, '0')}-${dateTime.day.toString().padLeft(2, '0')} '
        '${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}:${dateTime.second.toString().padLeft(2, '0')}';
  }

  String _getPayMethodText(String method) {
    switch (method) {
      case 'cash':
        return '现金';
      case 'wechat':
        return '微信支付';
      case 'alipay':
        return '支付宝';
      case 'card':
        return '会员卡';
      default:
        return method;
    }
  }

  String _getStatusText(String status) {
    switch (status) {
      case 'pending':
        return '待支付';
      case 'paid':
        return '已支付';
      case 'completed':
        return '已完成';
      case 'cancelled':
        return '已取消';
      default:
        return status;
    }
  }

  void dispose() {
    disconnect();
  }
}
