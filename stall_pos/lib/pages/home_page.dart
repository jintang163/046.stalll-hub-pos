import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:decimal/decimal.dart';

import '../providers/order_provider.dart';
import '../models/product.dart';
import '../models/stall.dart';
import 'orders_page.dart';
import 'settings_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _currentIndex = 0;
  final List<Widget> _pages = [];

  @override
  void initState() {
    super.initState();
    _pages.addAll([
      const CashierPage(),
      const OrdersPage(),
      const SettingsPage(),
    ]);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _pages[_currentIndex],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() => _currentIndex = index);
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.point_of_sale),
            label: '收银',
          ),
          NavigationDestination(
            icon: Icon(Icons.receipt_long),
            label: '订单',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings),
            label: '设置',
          ),
        ],
      ),
    );
  }
}

class CashierPage extends StatefulWidget {
  const CashierPage({super.key});

  @override
  State<CashierPage> createState() => _CashierPageState();
}

class _CashierPageState extends State<CashierPage> {
  List<Product> _products = [];
  String? _selectedCategory;
  bool _isLoading = true;
  final TextEditingController _tableNoController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadProducts();
  }

  Future<void> _loadProducts() async {
    final orderProvider = Provider.of<OrderProvider>(context, listen: false);
    setState(() => _isLoading = true);

    await Future.delayed(const Duration(milliseconds: 500));

    setState(() {
      _products = List.generate(20, (index) {
        return Product(
          id: index + 1,
          storeId: 1,
          stallId: orderProvider.currentStall?.id,
          name: '商品${index + 1}',
          price: Decimal.fromInt((index + 1) * Decimal.parse('5.5'),
          stock: 100 - index * 5,
          status: 1,
          sort: index,
          isHot: index % 3 == 0 ? 1 : 0,
          isRecommend: index % 5 == 0 ? 1 : 0,
          categoryId: (index ~/ 5 + 1),
        );
      });
      _isLoading = false;
    });
  }

  List<Product> get _filteredProducts {
    if (_selectedCategory == null) return _products;
    return _products.where((p) => p.categoryId.toString() == _selectedCategory).toList();
  }

  @override
  Widget build(BuildContext context) {
    final orderProvider = Provider.of<OrderProvider>(context);
    final stall = orderProvider.currentStall;

    return Scaffold(
      appBar: AppBar(
        title: Text(stall?.name ?? '摊位收银'),
        actions: [
          IconButton(
            icon: const Icon(Icons.bluetooth),
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const PrinterSettingsPage()),
              );
            },
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(40),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                  '今日订单: ${orderProvider.orders.length} 单',
                  style: TextStyle(color: Colors.white70, fontSize: 12),
                ),
                ),
                Text(
                  '营业额: ¥${orderProvider.cartTotal}',
                  style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ],
            ),
          ),
        ),
      ),
      body: Row(
        children: [
          Expanded(
            flex: 2,
            child: _buildProductList(),
          ),
          Container(
            width: 320,
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.grey.withOpacity(0.2),
                  blurRadius: 4,
                ),
              ],
            ),
            child: _buildCartPanel(),
          ),
        ],
      ),
    );
  }

  Widget _buildProductList() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    return GridView.builder(
      padding: const EdgeInsets.all(8),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        childAspectRatio: 0.85,
        mainAxisSpacing: 8,
        crossAxisSpacing: 8,
      ),
      itemCount: _filteredProducts.length,
      itemBuilder: (context, index) {
        final product = _filteredProducts[index];
        return _buildProductCard(product);
      },
    );
  }

  Widget _buildProductCard(Product product) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () {
          context.read<OrderProvider>().addToCart(product);
        },
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              child: Container(
                color: Colors.grey.shade200,
                child: const Icon(Icons.fastfood, size: 48, color: Colors.grey),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    product.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w500),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '¥${product.price}',
                    style: const TextStyle(
                      color: Colors.red,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCartPanel() {
    final orderProvider = Provider.of<OrderProvider>(context);

    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.blue.shade50,
          ),
          child: Row(
            children: [
              const Text(
                '购物车',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
              const Spacer(),
              Text('${orderProvider.cartItemCount} 件'),
              TextButton(
                onPressed: orderProvider.clearCart,
                child: const Text('清空'),
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: TextField(
            controller: _tableNoController,
            decoration: const InputDecoration(
              labelText: '桌号/取餐号',
              isDense: true,
            ),
            onChanged: (value) {
              orderProvider.setTableNo(value.isEmpty ? null : value);
            },
          ),
        ),
        Expanded(
          child: orderProvider.cartItems.isEmpty
              ? const Center(child: Text('购物车为空'))
              : ListView.builder(
                  itemCount: orderProvider.cartItems.length,
                  itemBuilder: (context, index) {
                    final item = orderProvider.cartItems[index];
                    return ListTile(
                      title: Text(item.product.name),
                      subtitle: Text('¥${item.price}'),
                      trailing: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.remove_circle_outline),
                            onPressed: () {
                              orderProvider.updateQuantity(
                                item.product.id,
                                item.quantity - 1,
                              );
                            },
                          ),
                          Text('${item.quantity}'),
                          IconButton(
                            icon: const Icon(Icons.add_circle_outline),
                            onPressed: () {
                              orderProvider.updateQuantity(
                                item.product.id,
                                item.quantity + 1,
                              );
                            },
                          ),
                        ],
                      ),
                    );
                  },
                ),
        ),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            border: Border(top: BorderSide(color: Colors.grey.shade200)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('合计:'),
                  Text(
                    '¥${orderProvider.cartTotal}',
                    style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: Colors.red,
                    ),
                  ),
                ],
              ),
              if (orderProvider.currentStall != null)
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      '摊位分成 (${(orderProvider.currentStall!.revenueRatio * Decimal.fromInt(100)})',
                      style: TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                    Text(
                      '¥${orderProvider.stallShare}',
                      style: const TextStyle(fontSize: 12, color: Colors.green),
                    ),
                  ],
                ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: orderProvider.cartItems.isEmpty
                      ? null
                      : () => _showPayDialog(),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    foregroundColor: Colors.white,
                  ),
                  child: const Text(
                    '收款结算',
                    style: TextStyle(fontSize: 18),
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _showPayDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('选择支付方式'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.payments),
              title: const Text('现金'),
              onTap: () => _handlePay('cash'),
            ),
            ListTile(
              leading: const Icon(Icons.qr_code),
              title: const Text('微信支付'),
              onTap: () => _handlePay('wechat'),
            ),
            ListTile(
              leading: const Icon(Icons.alternate_email),
              title: const Text('支付宝'),
              onTap: () => _handlePay('alipay'),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _handlePay(String payMethod) async {
    Navigator.of(context).pop();

    final orderProvider = Provider.of<OrderProvider>(context, listen: false);

    try {
      final order = await orderProvider.checkout(payMethod: payMethod);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('订单${order.orderNo}已提交成功'),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('结算失败: $e'),
        );
      }
    }
  }
}

class PrinterSettingsPage extends StatelessWidget {
  const PrinterSettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('打印机设置')),
      body: ListView(
        children: [
          ListTile(
            leading: const Icon(Icons.bluetooth_searching),
            title: const Text('扫描打印机'),
            onTap: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('正在扫描蓝牙设备...'),
              );
            },
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.print),
            title: const Text('测试打印'),
            onTap: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('打印机未连接'),
              );
            },
          ),
        ],
      ),
    );
  }
}
