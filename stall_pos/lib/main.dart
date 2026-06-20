import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'providers/order_provider.dart';
import 'services/api_service.dart';
import 'services/bluetooth_printer_service.dart';
import 'pages/login_page.dart';
import 'pages/home_page.dart';
import 'pages/orders_page.dart';
import 'pages/settings_page.dart';

void main() {
  runApp(const StallPosApp());
}

class StallPosApp extends StatelessWidget {
  const StallPosApp({super.key});

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<Map<String, dynamic>>(
      future: _initServices(),
      builder: (context, snapshot) {
        if (!snapshot.hasData) {
          return const MaterialApp(
            home: Scaffold(
              body: Center(child: CircularProgressIndicator()),
            ),
          );
        }

        final services = snapshot.data!;
        return MultiProvider(
          providers: [
            Provider<ApiService>.value(value: services['api']),
            Provider<BluetoothPrinterService>.value(value: services['printer']),
            ChangeNotifierProvider<OrderProvider>(
              create: (_) => OrderProvider(services['api'], services['printer']),
            ),
          ],
          child: MaterialApp(
            title: '摊位POS收银',
            theme: ThemeData(
              colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
              useMaterial3: true,
            ),
            home: services['isLoggedIn'] == true ? const HomePage() : const LoginPage(),
            routes: {
              '/login': (_) => const LoginPage(),
              '/home': (_) => const HomePage(),
              '/orders': (_) => const OrdersPage(),
              '/settings': (_) => const SettingsPage(),
            },
          ),
        );
      },
    );
  }

  Future<Map<String, dynamic>> _initServices() async {
    final prefs = await SharedPreferences.getInstance();
    final apiBaseUrl = prefs.getString('api_base_url') ?? 'http://localhost:8080/api';
    final token = prefs.getString('token');
    final isLoggedIn = token != null && token.isNotEmpty;

    final api = ApiService(baseUrl: apiBaseUrl, token: token);
    final printer = BluetoothPrinterService();

    return {
      'api': api,
      'printer': printer,
      'isLoggedIn': isLoggedIn,
    };
  }
}
