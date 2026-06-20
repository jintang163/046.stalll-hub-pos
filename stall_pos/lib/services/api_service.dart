import 'dart:convert';
import 'package:dio/dio.dart';
import '../models/stall.dart';
import '../models/product.dart';
import '../models/order.dart';

class ApiService {
  final Dio _dio;
  String? _token;

  ApiService({required String baseUrl, String? token}) : _dio = Dio(
    BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
      },
    ),
  ) {
    _token = token;
    if (token != null) {
      _dio.options.headers['Authorization'] = 'Bearer $token';
    }
  }

  void updateToken(String token) {
    _token = token;
    _dio.options.headers['Authorization'] = 'Bearer $token';
  }

  Future<Map<String, dynamic>> stallLogin(String username, String password) async {
    final response = await _dio.post(
      '/stall/auth/login',
      data: {
        'username': username,
        'password': password,
      },
    );
    return response.data;
  }

  Future<List<Stall>> getAllStalls(int storeId) async {
    final response = await _dio.get('/stalls/all', queryParameters: {
      'store_id': storeId,
    });
    final data = response.data['data'] as List;
    return data.map((e) => Stall.fromJson(e)).toList();
  }

  Future<List<Product>> getProductsByStall(int stallId, {int page = 1, int pageSize = 100}) async {
    final response = await _dio.get('/products', queryParameters: {
      'stall_id': stallId,
      'page': page,
      'page_size': pageSize,
    });
    final data = response.data['data']['list'] as List? ?? [];
    return data.map((e) => Product.fromJson(e)).toList();
  }

  Future<List<Category>> getCategories(int storeId) async {
    final response = await _dio.get('/categories', queryParameters: {
      'store_id': storeId,
      'page_size': 1000,
    });
    final data = response.data['data'] as List? ?? [];
    return data.map((e) => Category.fromJson(e)).toList();
  }

  Future<Map<String, dynamic>> createOrder(Order order) async {
    final response = await _dio.post('/orders', data: order.toJson());
    return response.data;
  }

  Future<Map<String, dynamic>> deviceHeartbeat(String deviceId, String appVersion) async {
    final response = await _dio.post('/stall/heartbeat', data: {
      'device_id': deviceId,
      'app_version': appVersion,
    });
    return response.data;
  }

  Future<List<Order>> getStallOrders(int stallId, {int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/orders', queryParameters: {
      'stall_id': stallId,
      'page': page,
      'page_size': pageSize,
    });
    final data = response.data['data']['list'] as List? ?? [];
    return data.map((e) => Order.fromJson(e)).toList();
  }

  Future<Order?> getOrderDetail(String orderNo) async {
    final response = await _dio.get('/orders/$orderNo');
    return Order.fromJson(response.data['data']);
  }

  Future<Map<String, dynamic>> getDailyReport(int stallId, String date) async {
    final response = await _dio.get('/stall-reports/daily', queryParameters: {
      'stall_id': stallId,
      'date': date,
    });
    return response.data['data'];
  }
}
