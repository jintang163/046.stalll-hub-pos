import 'package:flutter/foundation.dart';
import 'package:decimal/decimal.dart';
import '../models/order.dart';
import '../models/stall.dart';
import '../models/product.dart';
import '../database/order_database.dart';
import '../services/api_service.dart';
import '../services/order_sync_service.dart';
import '../services/bluetooth_printer_service.dart';

class CartItem {
  final Product product;
  final int quantity;
  final Decimal price;
  final String? spec;

  CartItem({
    required this.product,
    required this.quantity,
    required this.price,
    this.spec,
  });

  Decimal get subtotal => price * Decimal.fromInt(quantity);

  CartItem copyWith({
    Product? product,
    int? quantity,
    Decimal? price,
    String? spec,
  }) {
    return CartItem(
      product: product ?? this.product,
      quantity: quantity ?? this.quantity,
      price: price ?? this.price,
      spec: spec ?? this.spec,
    );
  }
}

class OrderProvider with ChangeNotifier {
  final OrderDatabase _orderDb = OrderDatabase();
  final ApiService _api;
  final BluetoothPrinterService _printer;
  OrderSyncService? _syncService;
  
  List<CartItem> _cartItems = [];
  Stall? _currentStall;
  List<Order> _orders = [];
  bool _isLoading = false;
  String? _tableNo;

  OrderProvider(this._api, this._printer) {
    _syncService = OrderSyncService(_api);
    _syncService!.start();
  }

  List<CartItem> get cartItems => _cartItems;
  Stall? get currentStall => _currentStall;
  List<Order> get orders => _orders;
  bool get isLoading => _isLoading;
  String? get tableNo => _tableNo;
  OrderSyncService? get syncService => _syncService;

  int get cartItemCount => _cartItems.fold(0, (sum, item) => sum + item.quantity);

  Decimal get cartTotal => _cartItems.fold(
    Decimal.zero,
    (sum, item) => sum + item.subtotal,
  );

  Decimal get stallShare {
    if (_currentStall == null) return Decimal.zero;
    return cartTotal * _currentStall!.revenueRatio;
  }

  void setCurrentStall(Stall stall) {
    _currentStall = stall;
    notifyListeners();
  }

  void setTableNo(String? tableNo) {
    _tableNo = tableNo;
    notifyListeners();
  }

  void addToCart(Product product, {int quantity = 1, Decimal? price, String? spec}) {
    final itemPrice = price ?? product.price;
    final existingIndex = _cartItems.indexWhere(
      (item) => item.product.id == product.id && item.spec == spec,
    );

    if (existingIndex >= 0) {
      final existing = _cartItems[existingIndex];
      _cartItems[existingIndex] = existing.copyWith(
        quantity: existing.quantity + quantity,
      );
    } else {
      _cartItems.add(CartItem(
        product: product,
        quantity: quantity,
        price: itemPrice,
        spec: spec,
      ));
    }

    notifyListeners();
  }

  void removeFromCart(int productId, {String? spec}) {
    _cartItems.removeWhere(
      (item) => item.product.id == productId && item.spec == spec,
    );
    notifyListeners();
  }

  void updateQuantity(int productId, int quantity, {String? spec}) {
    final index = _cartItems.indexWhere(
      (item) => item.product.id == productId && item.spec == spec,
    );

    if (index >= 0) {
      if (quantity <= 0) {
        _cartItems.removeAt(index);
      } else {
        _cartItems[index] = _cartItems[index].copyWith(quantity: quantity);
      }
      notifyListeners();
    }
  }

  void clearCart() {
    _cartItems = [];
    _tableNo = null;
    notifyListeners();
  }

  Future<Order> checkout({
    required String payMethod,
    Decimal? discountAmount,
    String? remark,
  }) async {
    if (_cartItems.isEmpty) {
      throw Exception('购物车为空');
    }
    if (_currentStall == null) {
      throw Exception('未设置摊位');
    }

    final orderNo = _generateOrderNo();
    final totalAmount = cartTotal;
    final actualDiscount = discountAmount ?? Decimal.zero;
    final payAmount = totalAmount - actualDiscount;

    final orderItems = _cartItems.map((cartItem) {
      final itemAmount = cartItem.subtotal;
      final stallAmount = itemAmount * _currentStall!.revenueRatio;
      final platformAmount = itemAmount * _currentStall!.platformRatio;

      return OrderItem(
        productId: cartItem.product.id,
        productName: cartItem.product.name,
        skuSpec: cartItem.spec,
        price: cartItem.price,
        quantity: cartItem.quantity,
        amount: itemAmount,
        stallId: _currentStall!.id,
        stallAmount: stallAmount,
        platformAmount: platformAmount,
      );
    }).toList();

    final order = Order(
      orderNo: orderNo,
      storeId: _currentStall!.storeId,
      stallId: _currentStall!.id,
      tableNo: _tableNo,
      orderType: 'stall',
      source: 'stall_pos',
      totalAmount: totalAmount,
      payAmount: payAmount,
      discountAmount: actualDiscount,
      payMethod: payMethod,
      status: 'paid',
      remark: remark,
      items: orderItems,
      createdAt: DateTime.now(),
      paidAt: DateTime.now(),
      syncStatus: 0,
    );

    await _orderDb.insertOrder(order);

    try {
      await _api.createOrder(order);
      await _orderDb.updateSyncStatus(orderNo, 1);
    } catch (e) {
      print('[Checkout] Sync order failed: $e');
    }

    if (_printer.isConnected) {
      try {
        await _printer.printOrder(order, _currentStall!.name);
      } catch (e) {
        print('[Checkout] Print failed: $e');
      }
    }

    clearCart();
    await loadOrders();

    return order;
  }

  Future<void> loadOrders({int limit = 20}) async {
    if (_currentStall == null) return;

    _isLoading = true;
    notifyListeners();

    try {
      _orders = await _orderDb.getOrders(
        stallId: _currentStall!.id,
        limit: limit,
      );
    } catch (e) {
      print('[Orders] Load error: $e');
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<Map<String, dynamic>> getDailySummary() async {
    if (_currentStall == null) {
      return {
        'orderCount': 0,
        'totalAmount': Decimal.zero,
        'paidAmount': Decimal.zero,
        'stallAmount': Decimal.zero,
        'platformAmount': Decimal.zero,
      };
    }

    final today = DateTime.now().toIso8601String().split('T')[0];
    return await _orderDb.getDailySummary(_currentStall!.id, today);
  }

  String _generateOrderNo() {
    final now = DateTime.now();
    final timestamp = now.millisecondsSinceEpoch ~/ 1000;
    final random = DateTime.now().microsecondsSinceEpoch % 1000;
    return 'S${_currentStall?.id ?? 0}${timestamp}${random.toString().padLeft(3, '0')}';
  }

  @override
  void dispose() {
    _syncService?.stop();
    super.dispose();
  }
}
