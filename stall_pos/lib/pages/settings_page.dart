import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../services/bluetooth_printer_service.dart';
import 'login_page.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final printer = Provider.of<BluetoothPrinterService>(context);

    return Scaffold(
      appBar: AppBar(title: const Text('设置')),
      body: ListView(
        children: [
          _buildSectionHeader('摊位信息'),
          ListTile(
            leading: const Icon(Icons.store),
            title: const Text('摊位名称'),
            subtitle: Text('烧烤档'),
          ),
          ListTile(
            leading: const Icon(Icons.percent),
            title: const Text('分账比例'),
            subtitle: const Text('摊位 70% / 平台 30%'),
          ),
          
          _buildSectionHeader('设备'),
          ListTile(
            leading: const Icon(Icons.bluetooth),
            title: const Text('蓝牙打印机'),
            subtitle: Text(printer.isConnected
                ? '已连接: ${printer.connectedDevice?.localName}'
                : '未连接'),
            trailing: printer.isConnected
                ? const Icon(Icons.check_circle, color: Colors.green)
                : const Icon(Icons.chevron_right),
            onTap: () {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const PrinterSettingsPage()),
              );
            },
          ),
          ListTile(
            leading: const Icon(Icons.sync),
            title: const Text('数据同步'),
            subtitle: const Text('待同步订单: 0 单'),
            onTap: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('立即同步...')),
              );
            },
          ),

          _buildSectionHeader('系统'),
          ListTile(
            leading: const Icon(Icons.cloud),
            title: const Text('服务器地址'),
            subtitle: const Text('http://192.168.1.100:8080/api'),
            onTap: () {
              _showServerSettingDialog(context);
            },
          ),
          ListTile(
            leading: const Icon(Icons.info_outline),
            title: const Text('关于'),
            subtitle: const Text('版本 1.0.0'),
          ),
          
          const SizedBox(height: 24),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red,
                  foregroundColor: Colors.white,
                ),
                onPressed: () => _logout(context),
                child: const Text('退出登录'),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(
        title,
        style: TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: Colors.grey.shade600,
        ),
      ),
    );
  }

  void _showServerSettingDialog(BuildContext context) {
    final controller = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('服务器地址'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'http://example.com/api',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () async {
              final prefs = await SharedPreferences.getInstance();
              prefs.setString('api_base_url', controller.text);
              Navigator.of(context).pop();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('服务器地址已更新')),
              );
            },
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  Future<void> _logout(BuildContext context) async {
    final prefs = await SharedPreferences.getInstance();
    prefs.remove('token');
    prefs.remove('stall_info');

    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: (_) => const LoginPage()),
      (route) => false,
    );
  }
}

class PrinterSettingsPage extends StatefulWidget {
  const PrinterSettingsPage({super.key});

  @override
  State<PrinterSettingsPage> createState() => _PrinterSettingsPageState();
}

class _PrinterSettingsPageState extends State<PrinterSettingsPage> {
  bool _scanning = false;
  List _devices = [];

  @override
  Widget build(BuildContext context) {
    final printer = Provider.of<BluetoothPrinterService>(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('打印机设置'),
        actions: [
          IconButton(
            icon: Icon(_scanning ? Icons.stop : Icons.search),
            onPressed: () => _scanPrinters(printer),
          ),
        ],
      ),
      body: Column(
        children: [
          if (printer.isConnected)
            Container(
              color: Colors.green.shade50,
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const Icon(Icons.check_circle, color: Colors.green),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          '已连接',
                          style: TextStyle(
                            color: Colors.green,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          printer.connectedDevice?.localName ?? '',
                          style: TextStyle(color: Colors.grey.shade600),
                        ),
                      ],
                    ),
                  ),
                  TextButton(
                    onPressed: () => printer.disconnect(),
                    child: const Text('断开'),
                  ),
                ],
              ),
            ),
          Expanded(
            child: _scanning
                ? const Center(child: CircularProgressIndicator())
                : _devices.isEmpty
                    ? const Center(child: Text('未发现蓝牙打印机'))
                    : ListView.builder(
                        itemCount: _devices.length,
                        itemBuilder: (context, index) {
                          final device = _devices[index];
                          return ListTile(
                            leading: const Icon(Icons.print),
                            title: Text(device.localName),
                            subtitle: Text(device.id.toString()),
                            trailing: ElevatedButton(
                              onPressed: () => _connectDevice(printer, device),
                              child: const Text('连接'),
                            ),
                          );
                        },
                      ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: printer.isConnected
                    ? () => _printTest(printer)
                    : null,
                child: const Text('测试打印'),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _scanPrinters(BluetoothPrinterService printer) async {
    if (_scanning) {
      setState(() {
        _scanning = false;
        _devices = [];
      });
      return;
    }

    setState(() {
      _scanning = true;
      _devices = [];
    });

    try {
      final devices = await printer.scanPrinters();
      if (mounted) {
        setState(() {
          _devices = devices;
          _scanning = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _scanning = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('扫描失败: $e')),
        );
      }
    }
  }

  Future<void> _connectDevice(BluetoothPrinterService printer, device) async {
    try {
      final success = await printer.connect(device);
      if (success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('连接成功')),
        );
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('连接失败')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('连接失败: $e')),
        );
      }
    }
  }

  Future<void> _printTest(BluetoothPrinterService printer) async {
    try {
      await printer.printTest();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('打印失败: $e')),
        );
      }
    }
  }
}
