import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import 'package:path_provider/path_provider.dart';
import 'dart:async';

class DatabaseHelper {
  static final DatabaseHelper _instance = DatabaseHelper._internal();
  factory DatabaseHelper() => _instance;
  DatabaseHelper._internal();

  static Database? _database;

  Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    final documentsDirectory = await getApplicationDocumentsDirectory();
    final path = join(documentsDirectory.path, 'stall_pos.db');

    return await openDatabase(
      path,
      version: 1,
      onCreate: _onCreate,
      onUpgrade: _onUpgrade,
    );
  }

  Future _onCreate(Database db, int version) async {
    await db.execute('''
      CREATE TABLE IF NOT EXISTS stalls (
        id INTEGER PRIMARY KEY,
        store_id INTEGER NOT NULL,
        stall_no TEXT NOT NULL,
        name TEXT NOT NULL,
        logo TEXT,
        revenue_ratio TEXT NOT NULL DEFAULT '0.7',
        platform_ratio TEXT NOT NULL DEFAULT '0.3',
        contact_name TEXT,
        contact_phone TEXT,
        status INTEGER NOT NULL DEFAULT 1,
        sort INTEGER DEFAULT 0,
        remark TEXT,
        created_at TEXT,
        updated_at TEXT
      )
    ''');

    await db.execute('''
      CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY,
        store_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        sort INTEGER DEFAULT 0,
        status INTEGER NOT NULL DEFAULT 1,
        created_at TEXT,
        updated_at TEXT
      )
    ''');

    await db.execute('''
      CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY,
        store_id INTEGER NOT NULL,
        stall_id INTEGER,
        category_id INTEGER,
        name TEXT NOT NULL,
        image TEXT,
        description TEXT,
        price TEXT NOT NULL DEFAULT '0',
        stock INTEGER,
        sort INTEGER DEFAULT 0,
        status INTEGER NOT NULL DEFAULT 1,
        is_hot INTEGER DEFAULT 0,
        is_recommend INTEGER DEFAULT 0,
        unit TEXT,
        created_at TEXT,
        updated_at TEXT
      )
    ''');

    await db.execute('''
      CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_no TEXT NOT NULL UNIQUE,
        store_id INTEGER NOT NULL,
        stall_id INTEGER,
        table_no TEXT,
        order_type TEXT NOT NULL DEFAULT 'normal',
        source TEXT NOT NULL DEFAULT 'stall_pos',
        total_amount TEXT NOT NULL DEFAULT '0',
        pay_amount TEXT NOT NULL DEFAULT '0',
        discount_amount TEXT DEFAULT '0',
        pay_method TEXT,
        status TEXT NOT NULL DEFAULT 'pending',
        remark TEXT,
        created_at TEXT,
        updated_at TEXT,
        paid_at TEXT,
        sync_status INTEGER NOT NULL DEFAULT 0,
        sync_error TEXT
      )
    ''');

    await db.execute('''
      CREATE TABLE IF NOT EXISTS order_items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_no TEXT NOT NULL,
        product_id INTEGER NOT NULL,
        product_name TEXT NOT NULL,
        sku_id INTEGER,
        sku_spec TEXT,
        price TEXT NOT NULL DEFAULT '0',
        quantity INTEGER NOT NULL DEFAULT 1,
        amount TEXT NOT NULL DEFAULT '0',
        stall_id INTEGER,
        stall_amount TEXT DEFAULT '0',
        platform_amount TEXT DEFAULT '0',
        attribute_values TEXT
      )
    ''');

    await db.execute('''
      CREATE TABLE IF NOT EXISTS sync_info (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sync_type TEXT NOT NULL,
        last_sync_time TEXT,
        last_sync_id INTEGER,
        created_at TEXT,
        updated_at TEXT
      )
    ''');

    await db.execute('CREATE INDEX idx_products_stall ON products(stall_id)');
    await db.execute('CREATE INDEX idx_products_category ON products(category_id)');
    await db.execute('CREATE INDEX idx_orders_stall ON orders(stall_id)');
    await db.execute('CREATE INDEX idx_orders_status ON orders(status)');
    await db.execute('CREATE INDEX idx_orders_sync ON orders(sync_status)');
    await db.execute('CREATE INDEX idx_order_items_order ON order_items(order_no)');
    await db.execute('CREATE INDEX idx_order_items_stall ON order_items(stall_id)');
  }

  Future _onUpgrade(Database db, int oldVersion, int newVersion) async {
    if (oldVersion < newVersion) {}
  }

  Future<void> close() async {
    final db = await database;
    await db.close();
    _database = null;
  }
}
