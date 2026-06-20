import 'package:decimal/decimal.dart';
import 'package:sqflite/sqflite.dart';
import '../models/order.dart';
import 'database_helper.dart';

class OrderDatabase {
  final DatabaseHelper _dbHelper = DatabaseHelper();

  Future<int> insertOrder(Order order) async {
    final db = await _dbHelper.database;

    int? orderId;
    await db.transaction((txn) async {
      orderId = await txn.insert(
        'orders',
        _orderToMap(order),
        conflictAlgorithm: ConflictAlgorithm.replace,
      );

      if (order.items != null && order.items!.isNotEmpty) {
        for (var item in order.items!) {
          await txn.insert(
            'order_items',
            _orderItemToMap(item, order.orderNo),
          );
        }
      }
    });

    return orderId ?? 0;
  }

  Future<List<Order>> getOrders({
    int? stallId,
    String? status,
    int? limit,
    int? offset,
  }) async {
    final db = await _dbHelper.database;

    List<String> where = [];
    List<dynamic> whereArgs = [];

    if (stallId != null) {
      where.add('stall_id = ?');
      whereArgs.add(stallId);
    }
    if (status != null) {
      where.add('status = ?');
      whereArgs.add(status);
    }

    final maps = await db.query(
      'orders',
      where: where.isEmpty ? null : where.join(' AND '),
      whereArgs: whereArgs.isEmpty ? null : whereArgs,
      orderBy: 'created_at DESC',
      limit: limit,
      offset: offset,
    );

    List<Order> orders = [];
    for (var map in maps) {
      final order = _mapToOrder(map);
      order.items = await getOrderItems(order.orderNo);
      orders.add(order);
    }

    return orders;
  }

  Future<Order?> getOrderByNo(String orderNo) async {
    final db = await _dbHelper.database;

    final maps = await db.query(
      'orders',
      where: 'order_no = ?',
      whereArgs: [orderNo],
      limit: 1,
    );

    if (maps.isEmpty) return null;

    final order = _mapToOrder(maps.first);
    order.items = await getOrderItems(orderNo);
    return order;
  }

  Future<List<OrderItem>> getOrderItems(String orderNo) async {
    final db = await _dbHelper.database;

    final maps = await db.query(
      'order_items',
      where: 'order_no = ?',
      whereArgs: [orderNo],
    );

    return maps.map((map) => _mapToOrderItem(map)).toList();
  }

  Future<List<Order>> getPendingSyncOrders({int limit = 50}) async {
    final db = await _dbHelper.database;

    final maps = await db.query(
      'orders',
      where: 'sync_status = ?',
      whereArgs: [0],
      orderBy: 'created_at ASC',
      limit: limit,
    );

    List<Order> orders = [];
    for (var map in maps) {
      final order = _mapToOrder(map);
      order.items = await getOrderItems(order.orderNo);
      orders.add(order);
    }

    return orders;
  }

  Future<int> updateSyncStatus(String orderNo, int syncStatus,
      {String? error}) async {
    final db = await _dbHelper.database;

    Map<String, dynamic> values = {
      'sync_status': syncStatus,
      if (error != null) 'sync_error': error,
    };

    return await db.update(
      'orders',
      values,
      where: 'order_no = ?',
      whereArgs: [orderNo],
    );
  }

  Future<int> updateOrderStatus(String orderNo, String status) async {
    final db = await _dbHelper.database;

    return await db.update(
      'orders',
      {
        'status': status,
        if (status == 'paid') 'paid_at': DateTime.now().toIso8601String(),
      },
      where: 'order_no = ?',
      whereArgs: [orderNo],
    );
  }

  Future<Map<String, dynamic>> getDailySummary(int stallId, String date) async {
    final db = await _dbHelper.database;

    final startDate = '${date}T00:00:00';
    final endDate = '${date}T23:59:59';

    final result = await db.rawQuery('''
      SELECT 
        COUNT(*) as order_count,
        COALESCE(SUM(CAST(total_amount AS REAL)), 0) as total_amount,
        COALESCE(SUM(CASE WHEN status = 'paid' THEN CAST(total_amount AS REAL) ELSE 0 END), 0) as paid_amount
      FROM orders 
      WHERE stall_id = ? 
      AND created_at >= ? 
      AND created_at <= ?
    ''', [stallId, startDate, endDate]);

    final row = result.first;

    final itemsResult = await db.rawQuery('''
      SELECT 
        COALESCE(SUM(CAST(stall_amount AS REAL)), 0) as stall_amount,
        COALESCE(SUM(CAST(platform_amount AS REAL)), 0) as platform_amount
      FROM order_items oi
      INNER JOIN orders o ON o.order_no = oi.order_no
      WHERE o.stall_id = ? 
      AND o.status = 'paid'
      AND o.created_at >= ? 
      AND o.created_at <= ?
    ''', [stallId, startDate, endDate]);

    final itemsRow = itemsResult.first;

    return {
      'orderCount': row['order_count'] as int? ?? 0,
      'totalAmount': Decimal.parse(row['total_amount']?.toString() ?? '0'),
      'paidAmount': Decimal.parse(row['paid_amount']?.toString() ?? '0'),
      'stallAmount': Decimal.parse(itemsRow['stall_amount']?.toString() ?? '0'),
      'platformAmount': Decimal.parse(itemsRow['platform_amount']?.toString() ?? '0'),
    };
  }

  Map<String, dynamic> _orderToMap(Order order) {
    return {
      'order_no': order.orderNo,
      'store_id': order.storeId,
      'stall_id': order.stallId,
      'table_no': order.tableNo,
      'order_type': order.orderType,
      'source': order.source,
      'total_amount': order.totalAmount.toString(),
      'pay_amount': order.payAmount.toString(),
      'discount_amount': order.discountAmount.toString(),
      'pay_method': order.payMethod,
      'status': order.status,
      'remark': order.remark,
      'created_at': order.createdAt?.toIso8601String() ?? DateTime.now().toIso8601String(),
      'updated_at': order.updatedAt?.toIso8601String() ?? DateTime.now().toIso8601String(),
      'paid_at': order.paidAt?.toIso8601String(),
      'sync_status': order.syncStatus ?? 0,
      'sync_error': order.syncError,
    };
  }

  Order _mapToOrder(Map<String, dynamic> map) {
    return Order(
      id: map['id'],
      orderNo: map['order_no'] as String,
      storeId: map['store_id'] as int,
      stallId: map['stall_id'],
      tableNo: map['table_no'],
      orderType: map['order_type'] as String? ?? 'normal',
      source: map['source'] as String? ?? 'stall_pos',
      totalAmount: Decimal.parse(map['total_amount']?.toString() ?? '0'),
      payAmount: Decimal.parse(map['pay_amount']?.toString() ?? '0'),
      discountAmount: Decimal.parse(map['discount_amount']?.toString() ?? '0'),
      payMethod: map['pay_method'],
      status: map['status'] as String? ?? 'pending',
      remark: map['remark'],
      createdAt: map['created_at'] != null ? DateTime.parse(map['created_at']) : null,
      updatedAt: map['updated_at'] != null ? DateTime.parse(map['updated_at']) : null,
      paidAt: map['paid_at'] != null ? DateTime.parse(map['paid_at']) : null,
      syncStatus: map['sync_status'] as int? ?? 0,
      syncError: map['sync_error'],
    );
  }

  Map<String, dynamic> _orderItemToMap(OrderItem item, String orderNo) {
    return {
      'order_no': orderNo,
      'product_id': item.productId,
      'product_name': item.productName,
      'sku_id': item.skuId,
      'sku_spec': item.skuSpec,
      'price': item.price.toString(),
      'quantity': item.quantity,
      'amount': item.amount.toString(),
      'stall_id': item.stallId,
      'stall_amount': item.stallAmount.toString(),
      'platform_amount': item.platformAmount.toString(),
      'attribute_values': item.attributeValues?.toString(),
    };
  }

  OrderItem _mapToOrderItem(Map<String, dynamic> map) {
    return OrderItem(
      id: map['id'],
      orderId: map['order_no'],
      productId: map['product_id'],
      productName: map['product_name'],
      skuId: map['sku_id'],
      skuSpec: map['sku_spec'],
      price: Decimal.parse(map['price']?.toString() ?? '0'),
      quantity: map['quantity'] as int,
      amount: Decimal.parse(map['amount']?.toString() ?? '0'),
      stallId: map['stall_id'],
      stallAmount: Decimal.parse(map['stall_amount']?.toString() ?? '0'),
      platformAmount: Decimal.parse(map['platform_amount']?.toString() ?? '0'),
    );
  }
}
