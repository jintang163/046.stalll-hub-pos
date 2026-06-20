import 'package:decimal/decimal.dart';

class OrderItem {
  final int? id;
  final String? orderId;
  final int productId;
  final String productName;
  final int? skuId;
  final String? skuSpec;
  final Decimal price;
  final int quantity;
  final Decimal amount;
  final int? stallId;
  final Decimal stallAmount;
  final Decimal platformAmount;
  final List<dynamic>? attributeValues;

  OrderItem({
    this.id,
    this.orderId,
    required this.productId,
    required this.productName,
    this.skuId,
    this.skuSpec,
    required this.price,
    required this.quantity,
    required this.amount,
    this.stallId,
    this.stallAmount = const Decimal.fromInt(0),
    this.platformAmount = const Decimal.fromInt(0),
    this.attributeValues,
  });

  factory OrderItem.fromJson(Map<String, dynamic> json) {
    return OrderItem(
      id: json['id'],
      orderId: json['order_id'] ?? json['orderId'],
      productId: json['product_id'] ?? json['productId'] ?? 0,
      productName: json['product_name'] ?? json['productName'] ?? '',
      skuId: json['sku_id'] ?? json['skuId'],
      skuSpec: json['sku_spec'] ?? json['skuSpec'],
      price: Decimal.parse(json['price']?.toString() ?? '0'),
      quantity: json['quantity'] ?? 0,
      amount: Decimal.parse(json['amount']?.toString() ?? json['total_price']?.toString() ?? '0'),
      stallId: json['stall_id'] ?? json['stallId'],
      stallAmount: Decimal.parse(json['stall_amount']?.toString() ?? json['stallAmount']?.toString() ?? '0'),
      platformAmount: Decimal.parse(json['platform_amount']?.toString() ?? json['platformAmount']?.toString() ?? '0'),
      attributeValues: json['attribute_values'] ?? json['attributeValues'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'order_id': orderId,
      'product_id': productId,
      'product_name': productName,
      'sku_id': skuId,
      'sku_spec': skuSpec,
      'price': price.toString(),
      'quantity': quantity,
      'amount': amount.toString(),
      'stall_id': stallId,
      'stall_amount': stallAmount.toString(),
      'platform_amount': platformAmount.toString(),
      'attribute_values': attributeValues,
    };
  }
}

class Order {
  final int? id;
  final String orderNo;
  final int storeId;
  final int? stallId;
  final String? tableNo;
  final String orderType;
  final String source;
  final Decimal totalAmount;
  final Decimal payAmount;
  final Decimal discountAmount;
  final String? payMethod;
  final String status;
  final String? remark;
  final List<OrderItem>? items;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final DateTime? paidAt;
  final int? syncStatus;
  final String? syncError;

  Order({
    this.id,
    required this.orderNo,
    required this.storeId,
    this.stallId,
    this.tableNo,
    this.orderType = 'normal',
    this.source = 'stall_pos',
    required this.totalAmount,
    required this.payAmount,
    this.discountAmount = const Decimal.fromInt(0),
    this.payMethod,
    this.status = 'pending',
    this.remark,
    this.items,
    this.createdAt,
    this.updatedAt,
    this.paidAt,
    this.syncStatus = 0,
    this.syncError,
  });

  factory Order.fromJson(Map<String, dynamic> json) {
    List<OrderItem>? items;
    if (json['items'] != null && json['items'] is List) {
      items = (json['items'] as List).map((i) => OrderItem.fromJson(i)).toList();
    }

    return Order(
      id: json['id'],
      orderNo: json['order_no'] ?? json['orderNo'] ?? '',
      storeId: json['store_id'] ?? json['storeId'] ?? 0,
      stallId: json['stall_id'] ?? json['stallId'],
      tableNo: json['table_no'] ?? json['tableNo'],
      orderType: json['order_type'] ?? json['orderType'] ?? 'normal',
      source: json['source'] ?? 'stall_pos',
      totalAmount: Decimal.parse(json['total_amount']?.toString() ?? json['totalAmount']?.toString() ?? '0'),
      payAmount: Decimal.parse(json['pay_amount']?.toString() ?? json['payAmount']?.toString() ?? '0'),
      discountAmount: Decimal.parse(json['discount_amount']?.toString() ?? json['discountAmount']?.toString() ?? '0'),
      payMethod: json['pay_method'] ?? json['payMethod'],
      status: json['status'] ?? 'pending',
      remark: json['remark'],
      items: items,
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : null,
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at']) : null,
      paidAt: json['paid_at'] != null ? DateTime.parse(json['paid_at']) : null,
      syncStatus: json['sync_status'] ?? json['syncStatus'] ?? 0,
      syncError: json['sync_error'] ?? json['syncError'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'order_no': orderNo,
      'store_id': storeId,
      'stall_id': stallId,
      'table_no': tableNo,
      'order_type': orderType,
      'source': source,
      'total_amount': totalAmount.toString(),
      'pay_amount': payAmount.toString(),
      'discount_amount': discountAmount.toString(),
      'pay_method': payMethod,
      'status': status,
      'remark': remark,
      'items': items?.map((e) => e.toJson()).toList(),
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
      'paid_at': paidAt?.toIso8601String(),
      'sync_status': syncStatus,
      'sync_error': syncError,
    };
  }
}
