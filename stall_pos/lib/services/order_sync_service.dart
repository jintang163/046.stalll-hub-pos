import 'dart:async';
import 'package:connectivity_plus/connectivity_plus.dart';
import '../database/order_database.dart';
import '../services/api_service.dart';
import '../models/order.dart';

class OrderSyncService {
  final OrderDatabase _orderDb = OrderDatabase();
  final ApiService _api;
  
  Timer? _syncTimer;
  StreamSubscription? _connectivitySubscription;
  bool _isSyncing = false;
  
  final List<Function> _syncListeners = [];

  OrderSyncService(this._api);

  bool get isSyncing => _isSyncing;

  void addSyncListener(Function listener) {
    _syncListeners.add(listener);
  }

  void removeSyncListener(Function listener) {
    _syncListeners.remove(listener);
  }

  void start() {
    _setupConnectivityListener();
    _startPeriodicSync();
  }

  void stop() {
    _syncTimer?.cancel();
    _connectivitySubscription?.cancel();
  }

  void _setupConnectivityListener() {
    _connectivitySubscription = Connectivity().onConnectivityChanged.listen((result) {
      if (result != ConnectivityResult.none) {
        triggerSync();
      }
    });
  }

  void _startPeriodicSync() {
    _syncTimer = Timer.periodic(const Duration(minutes: 5), (_) {
      triggerSync();
    });
  }

  Future<void> triggerSync() async {
    if (_isSyncing) return;

    final connectivityResult = await Connectivity().checkConnectivity();
    if (connectivityResult == ConnectivityResult.none) {
      return;
    }

    _isSyncing = true;
    _notifyListeners();

    try {
      await _syncPendingOrders();
    } catch (e) {
      print('[Sync] Error: $e');
    } finally {
      _isSyncing = false;
      _notifyListeners();
    }
  }

  Future<int> _syncPendingOrders() async {
    int syncedCount = 0;
    int retryCount = 0;
    const maxRetries = 3;

    while (retryCount < maxRetries) {
      try {
        final pendingOrders = await _orderDb.getPendingSyncOrders(limit: 50);
        
        if (pendingOrders.isEmpty) {
          break;
        }

        print('[Sync] Found ${pendingOrders.length} pending orders');

        for (var order in pendingOrders) {
          try {
            await _syncSingleOrder(order);
            syncedCount++;
          } catch (e) {
            print('[Sync] Failed to sync order ${order.orderNo}: $e');
          }
        }

        if (pendingOrders.length < 50) {
          break;
        }
      } catch (e) {
        print('[Sync] Batch sync error: $e');
        retryCount++;
        if (retryCount < maxRetries) {
          await Future.delayed(Duration(seconds: 2 * retryCount));
        }
      }
    }

    print('[Sync] Synced $syncedCount orders');
    return syncedCount;
  }

  Future<void> _syncSingleOrder(Order order) async {
    try {
      final result = await _api.createOrder(order);
      
      if (result['success'] == true || result['code'] == 0) {
        await _orderDb.updateSyncStatus(order.orderNo, 1);
      } else {
        final error = result['message'] ?? 'Unknown error';
        await _orderDb.updateSyncStatus(order.orderNo, -1, error: error);
      }
    } catch (e) {
      await _orderDb.updateSyncStatus(order.orderNo, -1, error: e.toString());
      rethrow;
    }
  }

  Future<int> getPendingOrderCount() async {
    final orders = await _orderDb.getPendingSyncOrders();
    return orders.length;
  }

  void _notifyListeners() {
    for (var listener in _syncListeners) {
      listener();
    }
  }
}
