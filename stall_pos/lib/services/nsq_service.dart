import 'dart:async';
import 'dart:convert';
import '../models/order.dart';
import '../database/order_database.dart';

typedef OrderCallback = void Function(Order order);
typedef StallChangeCallback = void Function(dynamic data);

class NsqService {
  final String nsqdAddress;
  final String nsqLookupdAddress;
  final String stallId;
  
  bool _connected = false;
  Timer? _reconnectTimer;
  final OrderDatabase _orderDb = OrderDatabase();
  
  final List<OrderCallback> _orderListeners = [];
  final List<StallChangeCallback> _stallChangeListeners = [];

  NsqService({
    required this.nsqdAddress,
    required this.nsqLookupdAddress,
    required this.stallId,
  });

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

  Future<void> connect() async {
    try {
      await _subscribeStallOrder();
      await _subscribeStallChanged();
      _connected = true;
      print('[NSQ] Connected successfully');
    } catch (e) {
      print('[NSQ] Connection error: $e');
      _connected = false;
      _scheduleReconnect();
    }
  }

  Future<void> _subscribeStallOrder() async {
    final topic = 'stall_order_stall_$stallId';
    final channel = 'stall_pos_$stallId';
    
    print('[NSQ] Subscribing to topic: $topic, channel: $channel');
  }

  Future<void> _subscribeStallChanged() async {
    final topic = 'stall_changed';
    final channel = 'stall_pos_$stallId';
    
    print('[NSQ] Subscribing to topic: $topic, channel: $channel');
  }

  void _handleStallOrderMessage(String message) {
    try {
      final data = jsonDecode(message);
      final orderData = data['order'] ?? data;
      final order = Order.fromJson(orderData);
      
      _saveOrderToLocal(order);
      
      for (var listener in _orderListeners) {
        listener(order);
      }
    } catch (e) {
      print('[NSQ] Parse order message error: $e');
    }
  }

  void _handleStallChangeMessage(String message) {
    try {
      final data = jsonDecode(message);
      
      for (var listener in _stallChangeListeners) {
        listener(data);
      }
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
    _connected = false;
    _orderListeners.clear();
    _stallChangeListeners.clear();
  }
}
