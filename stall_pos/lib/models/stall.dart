import 'package:decimal/decimal.dart';

class Stall {
  final int? id;
  final int storeId;
  final String stallNo;
  final String name;
  final String? logo;
  final Decimal revenueRatio;
  final Decimal platformRatio;
  final String? contactName;
  final String? contactPhone;
  final int status;
  final int? sort;
  final String? remark;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  Stall({
    this.id,
    required this.storeId,
    required this.stallNo,
    required this.name,
    this.logo,
    required this.revenueRatio,
    required this.platformRatio,
    this.contactName,
    this.contactPhone,
    this.status = 1,
    this.sort = 0,
    this.remark,
    this.createdAt,
    this.updatedAt,
  });

  factory Stall.fromJson(Map<String, dynamic> json) {
    return Stall(
      id: json['id'],
      storeId: json['store_id'] ?? json['storeId'] ?? 0,
      stallNo: json['stall_no'] ?? json['stallNo'] ?? '',
      name: json['name'] ?? '',
      logo: json['logo'],
      revenueRatio: Decimal.parse(json['revenue_ratio']?.toString() ?? json['revenueRatio']?.toString() ?? '0.7'),
      platformRatio: Decimal.parse(json['platform_ratio']?.toString() ?? json['platformRatio']?.toString() ?? '0.3'),
      contactName: json['contact_name'] ?? json['contactName'],
      contactPhone: json['contact_phone'] ?? json['contactPhone'],
      status: json['status'] ?? 1,
      sort: json['sort'] ?? 0,
      remark: json['remark'],
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : null,
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'store_id': storeId,
      'stall_no': stallNo,
      'name': name,
      'logo': logo,
      'revenue_ratio': revenueRatio.toString(),
      'platform_ratio': platformRatio.toString(),
      'contact_name': contactName,
      'contact_phone': contactPhone,
      'status': status,
      'sort': sort,
      'remark': remark,
      'created_at': createdAt?.toIso8601String(),
      'updated_at': updatedAt?.toIso8601String(),
    };
  }

  Stall copyWith({
    int? id,
    int? storeId,
    String? stallNo,
    String? name,
    String? logo,
    Decimal? revenueRatio,
    Decimal? platformRatio,
    String? contactName,
    String? contactPhone,
    int? status,
    int? sort,
    String? remark,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return Stall(
      id: id ?? this.id,
      storeId: storeId ?? this.storeId,
      stallNo: stallNo ?? this.stallNo,
      name: name ?? this.name,
      logo: logo ?? this.logo,
      revenueRatio: revenueRatio ?? this.revenueRatio,
      platformRatio: platformRatio ?? this.platformRatio,
      contactName: contactName ?? this.contactName,
      contactPhone: contactPhone ?? this.contactPhone,
      status: status ?? this.status,
      sort: sort ?? this.sort,
      remark: remark ?? this.remark,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
