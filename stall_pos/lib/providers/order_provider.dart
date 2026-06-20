import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:decimal/decimal.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:device_info_plus/device_info_plus.dart';
import '../models/order.dart';
import '../models/stall.dart';
import '../models/product.dart';
import '../database/order_database.dart';
import '../services/api_service.dart';
import '../services/order_sync_service.dart';
import '../services/bluetooth_printer_service.dart';
import '../services/nsq_service.dart';

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

class DeviceAlert {
  final String deviceId;
  final String deviceName;
  final String? stallName;
  final String message;
  final int offlineMinutes;
  final DateTime time;

  DeviceAlert({
    required this.deviceId,
    required this.deviceName,
    this.stallName,
    required this.message,
    this.offlineMinutes = 0,
    DateTime? time,
  }) : time = time ?? DateTime.now();
}

class OrderProvider with ChangeNotifier {
  final OrderDatabase _orderDb = OrderDatabase();
  final ApiService _api;
  final BluetoothPrinterService _printer;
  OrderSyncService? _syncService;
  NsqService? _nsqService;

  List<CartItem> _cartItems = [];
  Stall? _currentStall;
  List<Order> _orders = [];
  bool _isLoading = false;
  String? _tableNo;

  String? _deviceId;
  String _appVersion = '1.0.0';
  Timer? _heartbeatTimer;

  final List<DeviceAlert> _deviceAlerts = [];
  List<DeviceAlert> get deviceAlerts => _deviceAlerts;

  OrderProvider(this._api, this._printer) {
    _syncService = OrderSyncService(_api);
    _syncService!.start();
    _initDeviceAndHeartbeat();
  }

  List<CartItem> get cartItems => _cartItems;
  Stall? get currentStall => _currentStall;
  List<Order> get orders => _orders;
  bool get isLoading => _isLoading;
  String? get tableNo => _tableNo;
  OrderSyncService? get syncService => _syncService;
  NsqService? get nsqService => _nsqService;
  String? get deviceId => _deviceId;

  int get cartItemCount =>
      _cartItems.fold(0, (sum, item) => sum + item.quantity);

  Decimal get cartTotal => _cartItems.fold(
        Decimal.zero,
        (sum, item) => sum + item.subtotal,
      );

  Decimal get stallShare {
    if (_currentStall == null) return Decimal.zero;
    return cartTotal * _currentStall!.revenueRatio;
  }

  Future<void> _initDeviceAndHeartbeat() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      String? savedId = prefs.getString('device_id');
      if (savedId == null || savedId.isEmpty) {
        savedId = await _generateDeviceId();
        await prefs.setString('device_id', savedId);
      }
      _deviceId = savedId;
      print('[Heartbeat] Using deviceId: $_deviceId');
    } catch (e) {
      _deviceId = 'FALLBACK-${DateTime.now().millisecondsSinceEpoch}';
      print('[Heartbeat] Fallback deviceId: $_deviceId');
    }

    _startHeartbeat();
  }

  Future<String> _generateDeviceId() async {
    try {
      final deviceInfo = DeviceInfoPlugin();
      if (defaultTargetPlatform == TargetPlatform.android) {
        final info = await deviceInfo.androidInfo;
        return 'ANDROID-${info.id}-${info.androidId ?? info.device}';
      } else if (defaultTargetPlatform == TargetPlatform.iOS) {
        final info = await deviceInfo.iosInfo;
        return 'IOS-${info.identifierForVendor ?? info.utsname.machine}';
      } else if (defaultTargetPlatform == TargetPlatform.macOS) {
        final info = await deviceInfo.macOsInfo;
        return 'MACOS-${info.systemGUID ?? info.computerName}';
      } else if (defaultTargetPlatform == TargetPlatform.windows) {
        final info = await deviceInfo.windowsInfo;
        return 'WIN-${info.deviceId}-${info.computerName}';
      }
    } catch (e) {
      print('[DeviceInfo] Error: $e');
    }
    return 'DEV-${DateTime.now().millisecondsSinceEpoch}-${DateTime.now().microsecond}';
  }

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 60), (_) {
      _sendHeartbeat();
    });
    Future.delayed(const Duration(seconds: 5), _sendHeartbeat);
  }

  Future<void> _sendHeartbeat() async {
    if (_deviceId == null) return;
    try {
      await _api.deviceHeartbeat(_deviceId!, _appVersion);
      print('[Heartbeat] Sent successfully: $_deviceId');
    } catch (e) {
      print('[Heartbeat] Failed: $e');
    }
  }

  void setCurrentStall(Stall stall) {
    _currentStall = stall;
    _setupNsqAndAlerts();
    _loadInitialData();
    notifyListeners();
  }

  Future<void> _setupNsqAndAlerts() async {
    try {
      await _nsqService?.disconnect();
    } catch (_) {}

    if (_currentStall?.id == null) return;

    _nsqService = NsqService(
      api: _api,
      stallId: _currentStall!.id.toString(),
    );

    _nsqService!.addDeviceAlertListener((alertData) {
      _handleDeviceAlert(alertData);
    });

    _nsqService!.addOrderListener((order) {
      loadOrders();
    });

    await _nsqService!.connect();
  }

  void _handleDeviceAlert(dynamic alertData) {
    try {
      final deviceName =
          alertData['device_name'] ?? alertData['deviceName'] ?? '未知设备';
      final stallName =
          alertData['stall_name'] ?? alertData['stallName'] ?? _currentStall?.name;
      final deviceId =
          alertData['device_id'] ?? alertData['deviceId'] ?? '';
      final offlineStr = alertData['offline_minutes'] ??
          alertData['offlineMinutes'] ??
          30;
      final offlineMinutes = int.tryParse(offlineStr.toString()) ?? 30;

      final msg = '设备【$deviceName】已离线超过 $offlineMinutes 分钟，请检查连接';

      final alert = DeviceAlert(
        deviceId: deviceId.toString(),
        deviceName: deviceName,
        stallName: stallName,
        message: msg,
        offlineMinutes: offlineMinutes,
      );

      final exists = _deviceAlerts.any((a) =>
          a.deviceId == alert.deviceId &&
          a.time.difference(DateTime.now()).abs().inMinutes < 30);

      if (!exists) {
        _deviceAlerts.insert(0, alert);
        if (_deviceAlerts.length > 20) _deviceAlerts.removeLast();
        notifyListeners();
      }
    } catch (e) {
      print('[Alert] Handle error: $e');
    }
  }

  void dismissAlert(int index) {
    if (index >= 0 && index < _deviceAlerts.length) {
      _deviceAlerts.removeAt(index);
      notifyListeners();
    }
  }

  Future<void> _loadInitialData() async {
    await loadProductsCache();
    await loadOrders();
  }

  Future<void> loadProductsCache() async {}

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
    _nsqService?.disconnect();
    _heartbeatTimer?.cancel();
    super.dispose();
  }
}
