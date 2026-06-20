import 'dart:async';
import 'dart:convert';
import 'package:connectivity_plus/connectivity_plus.dart';
import '../models/order.dart';
import '../database/order_database.dart';
import 'api_service.dart';

typedef OrderCallback = void Function(Order order);
typedef StallChangeCallback = void Function(dynamic data);
typedef DeviceAlertCallback = void Function(dynamic alert);

class NsqService {
  final ApiService _api;
  final String stallId;

  bool _connected = false;
  Timer? _reconnectTimer;
  Timer? _orderPollTimer;
  Timer? _alertPollTimer;
  Timer? _stallChangePollTimer;
  StreamSubscription? _connectivitySubscription;

  final OrderDatabase _orderDb = OrderDatabase();

  final List<OrderCallback> _orderListeners = [];
  final List<StallChangeCallback> _stallChangeListeners = [];
  final List<DeviceAlertCallback> _deviceAlertListeners = [];

  String _lastProcessedOrderNo = '';
  bool _isPolling = false;

  NsqService({
    required ApiService api,
    required this.stallId,
  }) : _api = api;

  bool get isConnected => _connected;

  void addOrderListener(OrderCallback callback) {
    _orderListeners.add(callback);
  }

  void removeOrderListener(OrderCallback callback) {
    _orderListeners.remove(callback);
  }

  void addStallChangeListener(StallChangeCallback callback) {
    _stallChangeListeners.add(callback);
  }

  void removeStallChangeListener(StallChangeCallback callback) {
    _stallChangeListeners.remove(callback);
  }

  void addDeviceAlertListener(DeviceAlertCallback callback) {
    _deviceAlertListeners.add(callback);
  }

  void removeDeviceAlertListener(DeviceAlertCallback callback) {
    _deviceAlertListeners.remove(callback);
  }

  Future<void> connect() async {
    try {
      await _subscribeStallOrder();
      await _subscribeStallChanged();
      await _subscribeDeviceAlerts();

      _setupConnectivity();
      _connected = true;
      print('[NSQ/Poll] Connected successfully (stall=$stallId)');
    } catch (e) {
      print('[NSQ/Poll] Connection error: $e');
      _connected = false;
      _scheduleReconnect();
    }
  }

  void _setupConnectivity() {
    _connectivitySubscription?.cancel();
    _connectivitySubscription =
        Connectivity().onConnectivityChanged.listen((result) {
      if (result != ConnectivityResult.none && !_connected) {
        connect();
      }
    });
  }

  Future<void> _subscribeStallOrder() async {
    final topic = 'stall_order_stall_$stallId';
    print('[NSQ] Attempting NSQ subscribe topic: $topic (fallback to HTTP poll)');

    _orderPollTimer?.cancel();
    _orderPollTimer = Timer.periodic(const Duration(seconds: 10), (_) {
      _pollNewOrders();
    });

    await _pollNewOrders();
  }

  Future<void> _subscribeStallChanged() async {
    print('[NSQ] Subscribing stall_change events (HTTP poll fallback)');
    _stallChangePollTimer?.cancel();
    _stallChangePollTimer = Timer.periodic(const Duration(minutes: 2), (_) {
      _pollStallChange();
    });
  }

  Future<void> _subscribeDeviceAlerts() async {
    print('[NSQ] Subscribing device alert events (HTTP poll)');
    _alertPollTimer?.cancel();
    _alertPollTimer = Timer.periodic(const Duration(minutes: 1), (_) {
      _pollDeviceAlerts();
    });

    await _pollDeviceAlerts();
  }

  Future<void> _pollNewOrders() async {
    if (_isPolling) return;
    final conn = await Connectivity().checkConnectivity();
    if (conn == ConnectivityResult.none) return;

    _isPolling = true;
    try {
      final sid = int.tryParse(stallId) ?? 0;
      if (sid <= 0) return;

      final orders = await _api.getNewOrders(
        sid,
        lastOrderNo: _lastProcessedOrderNo.isEmpty ? null : _lastProcessedOrderNo,
        limit: 50,
      );

      for (final order in orders) {
        if (order.orderNo.compareTo(_lastProcessedOrderNo) > 0) {
          _lastProcessedOrderNo = order.orderNo;
        }
        final exists = await _orderDb.orderExists(order.orderNo);
        if (!exists) {
          await _saveOrderToLocal(order);
          _notifyOrderListeners(order);
        }
      }
    } catch (e) {
      print('[NSQ] Poll orders error: $e');
    } finally {
      _isPolling = false;
    }
  }

  Future<void> _pollStallChange() async {
    try {
      for (final listener in _stallChangeListeners) {
        listener({'type': 'refresh', 'timestamp': DateTime.now().toIso8601String()});
      }
    } catch (e) {
      print('[NSQ] Poll stall change error: $e');
    }
  }

  Future<void> _pollDeviceAlerts() async {
    try {
      final sid = int.tryParse(stallId);
      final alerts = await _api.getDeviceAlerts(stallId: sid);
      for (final alert in alerts) {
        _notifyDeviceAlertListeners(alert);
      }
    } catch (e) {
      print('[NSQ] Poll alerts error: $e');
    }
  }

  void _handleStallOrderMessage(String message) {
    try {
      final data = jsonDecode(message);
      final orderData = data['order'] ?? data['Items'] != null ? data : data;
      final order = Order.fromJson(orderData);

      _saveOrderToLocal(order);
      _notifyOrderListeners(order);
    } catch (e) {
      print('[NSQ] Parse order message error: $e');
    }
  }

  void _handleStallChangeMessage(String message) {
    try {
      final data = jsonDecode(message);
      _notifyStallChangeListeners(data);
    } catch (e) {
      print('[NSQ] Parse stall change message error: $e');
    }
  }

  Future<void> _saveOrderToLocal(Order order) async {
    try {
      await _orderDb.insertOrder(order);
    } catch (e) {
      print('[NSQ] Save order to local error: $e');
    }
  }

  void _notifyOrderListeners(Order order) {
    for (final listener in _orderListeners) {
      try {
        listener(order);
      } catch (e) {
        print('[NSQ] Order listener error: $e');
      }
    }
  }

  void _notifyStallChangeListeners(dynamic data) {
    for (final listener in _stallChangeListeners) {
      try {
        listener(data);
      } catch (e) {
        print('[NSQ] Stall change listener error: $e');
      }
    }
  }

  void _notifyDeviceAlertListeners(dynamic alert) {
    for (final listener in _deviceAlertListeners) {
      try {
        listener(alert);
      } catch (e) {
        print('[NSQ] Device alert listener error: $e');
      }
    }
  }

  void _scheduleReconnect() {
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 30), () {
      connect();
    });
  }

  Future<void> publish(String topic, dynamic data) async {
    try {
      final message = jsonEncode(data);
      print('[NSQ] Publish to $topic: $message');
    } catch (e) {
      print('[NSQ] Publish error: $e');
    }
  }

  Future<void> disconnect() async {
    _reconnectTimer?.cancel();
    _orderPollTimer?.cancel();
    _alertPollTimer?.cancel();
    _stallChangePollTimer?.cancel();
    _connectivitySubscription?.cancel();
    _connected = false;
    _orderListeners.clear();
    _stallChangeListeners.clear();
    _deviceAlertListeners.clear();
  }
}
