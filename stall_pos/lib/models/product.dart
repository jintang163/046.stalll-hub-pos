import 'package:decimal/decimal.dart';

class Product {
  final int id;
  final int storeId;
  final int? stallId;
  final int? categoryId;
  final String name;
  final String? image;
  final String? description;
  final Decimal price;
  final int? stock;
  final int sort;
  final int status;
  final int isHot;
  final int isRecommend;
  final String? unit;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  Product({
    required this.id,
    required this.storeId,
    this.stallId,
    this.categoryId,
    required this.name,
    this.image,
    this.description,
    required this.price,
    this.stock,
    this.sort = 0,
    this.status = 1,
    this.isHot = 0,
    this.isRecommend = 0,
    this.unit,
    this.createdAt,
    this.updatedAt,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id'] ?? 0,
      storeId: json['store_id'] ?? json['storeId'] ?? 0,
      stallId: json['stall_id'] ?? json['stallId'],
      categoryId: json['category_id'] ?? json['categoryId'],
      name: json['name'] ?? '',
      image: json['image'],
      description: json['description'],
      price: Decimal.parse(json['price']?.toString() ?? '0'),
      stock: json['stock'],
      sort: json['sort'] ?? 0,
      status: json['status'] ?? 1,
      isHot: json['is_hot'] ?? json['isHot'] ?? 0,
      isRecommend: json['is_recommend'] ?? json['isRecommend'] ?? 0,
      unit: json['unit'],
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : null,
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'store_id': storeId,
      'stall_id': stallId,
      'category_id': categoryId,
      'name': name,
      'image': image,
      'description': description,
      'price': price.toString(),
      'stock': stock,
      'sort': sort,
      'status': status,
      'is_hot': isHot,
      'is_recommend': isRecommend,
      'unit': unit,
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }
}

class Category {
  final int id;
  final int storeId;
  final String name;
  final int sort;
  final int status;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  Category({
    required this.id,
    required this.storeId,
    required this.name,
    this.sort = 0,
    this.status = 1,
    this.createdAt,
    this.updatedAt,
  });

  factory Category.fromJson(Map<String, dynamic> json) {
    return Category(
      id: json['id'] ?? 0,
      storeId: json['store_id'] ?? json['storeId'] ?? 0,
      name: json['name'] ?? '',
      sort: json['sort'] ?? 0,
      status: json['status'] ?? 1,
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : null,
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'store_id': storeId,
      'name': name,
      'sort': sort,
      'status': status,
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }
}
