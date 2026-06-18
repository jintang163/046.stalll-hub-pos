const Database = require('better-sqlite3')
const path = require('path')
const fs = require('fs')

class SQLiteDatabase {
  constructor(userDataPath) {
    this.dbPath = path.join(userDataPath, 'pos.db')
    this.db = null
  }

  init() {
    const dbDir = path.dirname(this.dbPath)
    if (!fs.existsSync(dbDir)) {
      fs.mkdirSync(dbDir, { recursive: true })
    }

    this.db = new Database(this.dbPath)
    this.db.pragma('journal_mode = WAL')
    this.db.pragma('foreign_keys = ON')
    this.createTables()
    this.createIndexes()
  }

  createTables() {
    const tables = [
      `CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        sort_order INTEGER DEFAULT 0,
        status INTEGER DEFAULT 1,
        created_at TEXT,
        updated_at TEXT
      )`,
      
      `CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        image TEXT,
        category_id INTEGER,
        sort_order INTEGER DEFAULT 0,
        status INTEGER DEFAULT 1,
        is_hot INTEGER DEFAULT 0,
        is_recommend INTEGER DEFAULT 0,
        warning_threshold INTEGER DEFAULT 10,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (category_id) REFERENCES categories(id)
      )`,
      
      `CREATE TABLE IF NOT EXISTS product_skus (
        id INTEGER PRIMARY KEY,
        product_id INTEGER NOT NULL,
        sku_code TEXT,
        spec_name TEXT NOT NULL,
        price REAL DEFAULT 0,
        original_price REAL DEFAULT 0,
        stock INTEGER DEFAULT 0,
        image TEXT,
        status INTEGER DEFAULT 1,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
      )`,
      
      `CREATE TABLE IF NOT EXISTS product_attributes (
        id INTEGER PRIMARY KEY,
        product_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        sort_order INTEGER DEFAULT 0,
        status INTEGER DEFAULT 1,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
      )`,
      
      `CREATE TABLE IF NOT EXISTS attribute_values (
        id INTEGER PRIMARY KEY,
        attribute_id INTEGER NOT NULL,
        value TEXT NOT NULL,
        extra_price REAL DEFAULT 0,
        stock INTEGER DEFAULT -1,
        sort_order INTEGER DEFAULT 0,
        status INTEGER DEFAULT 1,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (attribute_id) REFERENCES product_attributes(id) ON DELETE CASCADE
      )`,
      
      `CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_no TEXT UNIQUE NOT NULL,
        total_amount REAL DEFAULT 0,
        discount_amount REAL DEFAULT 0,
        actual_amount REAL DEFAULT 0,
        member_id INTEGER,
        member_name TEXT,
        table_no TEXT,
        remark TEXT,
        status INTEGER DEFAULT 0,
        pay_status INTEGER DEFAULT 0,
        pay_method TEXT,
        synced INTEGER DEFAULT 0,
        sync_attempts INTEGER DEFAULT 0,
        last_sync_error TEXT,
        created_at TEXT,
        updated_at TEXT,
        paid_at TEXT
      )`,
      
      `CREATE TABLE IF NOT EXISTS order_items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_no TEXT NOT NULL,
        product_id INTEGER,
        product_name TEXT,
        sku_id INTEGER,
        sku_name TEXT,
        attribute_ids TEXT,
        attribute_names TEXT,
        price REAL DEFAULT 0,
        quantity INTEGER DEFAULT 1,
        subtotal REAL DEFAULT 0,
        remark TEXT,
        created_at TEXT,
        FOREIGN KEY (order_no) REFERENCES orders(order_no) ON DELETE CASCADE
      )`,
      
      `CREATE TABLE IF NOT EXISTS sync_config (
        key TEXT PRIMARY KEY,
        value TEXT
      )`
    ]

    tables.forEach(sql => this.db.exec(sql))
  }

  createIndexes() {
    const indexes = [
      'CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id)',
      'CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)',
      'CREATE INDEX IF NOT EXISTS idx_skus_product ON product_skus(product_id)',
      'CREATE INDEX IF NOT EXISTS idx_skus_status ON product_skus(status)',
      'CREATE INDEX IF NOT EXISTS idx_attr_product ON product_attributes(product_id)',
      'CREATE INDEX IF NOT EXISTS idx_attr_values_attr ON attribute_values(attribute_id)',
      'CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)',
      'CREATE INDEX IF NOT EXISTS idx_orders_synced ON orders(synced)',
      'CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at)',
      'CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_no)'
    ]

    indexes.forEach(sql => this.db.exec(sql))
  }

  getProducts() {
    const products = this.db.prepare(`
      SELECT p.*, c.name as category_name 
      FROM products p 
      LEFT JOIN categories c ON p.category_id = c.id 
      ORDER BY c.sort_order, p.sort_order
    `).all()

    return products.map(p => ({
      ...p,
      is_hot: Boolean(p.is_hot),
      is_recommend: Boolean(p.is_recommend),
      status: Boolean(p.status)
    }))
  }

  getProductById(id) {
    const product = this.db.prepare('SELECT * FROM products WHERE id = ?').get(id)
    if (!product) return null

    const skus = this.db.prepare('SELECT * FROM product_skus WHERE product_id = ?').all(id)
    const attributes = this.db.prepare(`
      SELECT pa.*, 
             json_group_array(json_object(
               'id', av.id,
               'value', av.value,
               'extra_price', av.extra_price,
               'stock', av.stock,
               'sort_order', av.sort_order,
               'status', av.status
             )) as values
      FROM product_attributes pa
      LEFT JOIN attribute_values av ON av.attribute_id = pa.id
      WHERE pa.product_id = ?
      GROUP BY pa.id
      ORDER BY pa.sort_order
    `).all(id)

    return {
      ...product,
      is_hot: Boolean(product.is_hot),
      is_recommend: Boolean(product.is_recommend),
      status: Boolean(product.status),
      skus: skus.map(s => ({ ...s, status: Boolean(s.status) })),
      attributes: attributes.map(a => ({
        ...a,
        status: Boolean(a.status),
        values: JSON.parse(a.values || '[]').map(v => ({ ...v, status: Boolean(v.status) }))
      }))
    }
  }

  getCategories() {
    return this.db.prepare('SELECT * FROM categories ORDER BY sort_order, id').all()
  }

  getSKUs(productId) {
    return this.db.prepare('SELECT * FROM product_skus WHERE product_id = ? ORDER BY id').all(productId)
  }

  getAttributes(productId) {
    const attrs = this.db.prepare(`
      SELECT pa.*, 
             json_group_array(json_object(
               'id', av.id,
               'value', av.value,
               'extra_price', av.extra_price,
               'stock', av.stock,
               'sort_order', av.sort_order,
               'status', av.status
             )) as values
      FROM product_attributes pa
      LEFT JOIN attribute_values av ON av.attribute_id = pa.id
      WHERE pa.product_id = ?
      GROUP BY pa.id
      ORDER BY pa.sort_order
    `).all(productId)

    return attrs.map(a => ({
      ...a,
      values: JSON.parse(a.values || '[]')
    }))
  }

  saveProducts(products) {
    const tx = this.db.transaction(() => {
      const insertProduct = this.db.prepare(`
        INSERT OR REPLACE INTO products 
        (id, name, description, image, category_id, sort_order, status, is_hot, is_recommend, warning_threshold, created_at, updated_at)
        VALUES (@id, @name, @description, @image, @category_id, @sort_order, @status, @is_hot, @is_recommend, @warning_threshold, @created_at, @updated_at)
      `)

      const insertSKU = this.db.prepare(`
        INSERT OR REPLACE INTO product_skus 
        (id, product_id, sku_code, spec_name, price, original_price, stock, image, status, created_at, updated_at)
        VALUES (@id, @product_id, @sku_code, @spec_name, @price, @original_price, @stock, @image, @status, @created_at, @updated_at)
      `)

      const deleteSKUs = this.db.prepare('DELETE FROM product_skus WHERE product_id = ?')
      
      const insertAttr = this.db.prepare(`
        INSERT OR REPLACE INTO product_attributes 
        (id, product_id, name, sort_order, status, created_at, updated_at)
        VALUES (@id, @product_id, @name, @sort_order, @status, @created_at, @updated_at)
      `)

      const deleteAttrs = this.db.prepare('DELETE FROM product_attributes WHERE product_id = ?')
      
      const insertAttrValue = this.db.prepare(`
        INSERT OR REPLACE INTO attribute_values 
        (id, attribute_id, value, extra_price, stock, sort_order, status, created_at, updated_at)
        VALUES (@id, @attribute_id, @value, @extra_price, @stock, @sort_order, @status, @created_at, @updated_at)
      `)

      const deleteAttrValues = this.db.prepare(`
        DELETE FROM attribute_values WHERE attribute_id IN 
        (SELECT id FROM product_attributes WHERE product_id = ?)
      `)

      products.forEach(p => {
        insertProduct.run({
          id: p.id,
          name: p.name,
          description: p.description || '',
          image: p.image || '',
          category_id: p.category_id,
          sort_order: p.sort_order || 0,
          status: p.status ? 1 : 0,
          is_hot: p.is_hot ? 1 : 0,
          is_recommend: p.is_recommend ? 1 : 0,
          warning_threshold: p.warning_threshold || 10,
          created_at: p.created_at,
          updated_at: p.updated_at
        })

        deleteSKUs.run(p.id)
        if (p.skus && p.skus.length > 0) {
          p.skus.forEach(sku => {
            insertSKU.run({
              id: sku.id,
              product_id: p.id,
              sku_code: sku.sku_code || '',
              spec_name: sku.spec_name,
              price: sku.price,
              original_price: sku.original_price || sku.price,
              stock: sku.stock,
              image: sku.image || '',
              status: sku.status ? 1 : 0,
              created_at: sku.created_at || p.created_at,
              updated_at: sku.updated_at || p.updated_at
            })
          })
        }

        deleteAttrValues.run(p.id)
        deleteAttrs.run(p.id)
        if (p.attributes && p.attributes.length > 0) {
          p.attributes.forEach(attr => {
            insertAttr.run({
              id: attr.id,
              product_id: p.id,
              name: attr.name,
              sort_order: attr.sort_order || 0,
              status: attr.status ? 1 : 0,
              created_at: attr.created_at || p.created_at,
              updated_at: attr.updated_at || p.updated_at
            })

            if (attr.values && attr.values.length > 0) {
              attr.values.forEach(val => {
                insertAttrValue.run({
                  id: val.id,
                  attribute_id: attr.id,
                  value: val.value,
                  extra_price: val.extra_price || 0,
                  stock: val.stock !== undefined ? val.stock : -1,
                  sort_order: val.sort_order || 0,
                  status: val.status ? 1 : 0,
                  created_at: val.created_at || p.created_at,
                  updated_at: val.updated_at || p.updated_at
                })
              })
            }
          })
        }
      })
    })

    tx()
    return { success: true, count: products.length }
  }

  saveCategories(categories) {
    const tx = this.db.transaction(() => {
      const insert = this.db.prepare(`
        INSERT OR REPLACE INTO categories 
        (id, name, sort_order, status, created_at, updated_at)
        VALUES (@id, @name, @sort_order, @status, @created_at, @updated_at)
      `)

      categories.forEach(c => {
        insert.run({
          id: c.id,
          name: c.name,
          sort_order: c.sort_order || 0,
          status: c.status ? 1 : 0,
          created_at: c.created_at,
          updated_at: c.updated_at
        })
      })
    })

    tx()
    return { success: true, count: categories.length }
  }

  updateStock(skuId, stock) {
    const result = this.db.prepare('UPDATE product_skus SET stock = ? WHERE id = ?').run(stock, skuId)
    return { success: result.changes > 0 }
  }

  updateProductStatus(productId, status) {
    const result = this.db.prepare('UPDATE products SET status = ? WHERE id = ?').run(status ? 1 : 0, productId)
    return { success: result.changes > 0 }
  }

  deleteProduct(productId) {
    const tx = this.db.transaction(() => {
      this.db.prepare('DELETE FROM attribute_values WHERE attribute_id IN (SELECT id FROM product_attributes WHERE product_id = ?)').run(productId)
      this.db.prepare('DELETE FROM product_attributes WHERE product_id = ?').run(productId)
      this.db.prepare('DELETE FROM product_skus WHERE product_id = ?').run(productId)
      this.db.prepare('DELETE FROM products WHERE id = ?').run(productId)
    })

    tx()
    return { success: true }
  }

  clearAllProducts() {
    const tx = this.db.transaction(() => {
      this.db.exec('DELETE FROM attribute_values')
      this.db.exec('DELETE FROM product_attributes')
      this.db.exec('DELETE FROM product_skus')
      this.db.exec('DELETE FROM products')
      this.db.exec('DELETE FROM categories')
    })

    tx()
    return { success: true }
  }

  getLastSyncTime() {
    const row = this.db.prepare("SELECT value FROM sync_config WHERE key = 'last_sync_time'").get()
    return row ? row.value : null
  }

  setLastSyncTime(time) {
    this.db.prepare(`
      INSERT OR REPLACE INTO sync_config (key, value) 
      VALUES ('last_sync_time', ?)
    `).run(time)
    return { success: true }
  }

  getLastSyncID() {
    const row = this.db.prepare("SELECT value FROM sync_config WHERE key = 'last_sync_id'").get()
    return row ? parseInt(row.value) : 0
  }

  setLastSyncID(id) {
    this.db.prepare(`
      INSERT OR REPLACE INTO sync_config (key, value) 
      VALUES ('last_sync_id', ?)
    `).run(id.toString())
    return { success: true }
  }

  saveOrder(order) {
    const tx = this.db.transaction(() => {
      const insertOrder = this.db.prepare(`
        INSERT OR REPLACE INTO orders 
        (order_no, total_amount, discount_amount, actual_amount, member_id, member_name, 
         table_no, remark, status, pay_status, pay_method, synced, created_at, updated_at, paid_at)
        VALUES (@order_no, @total_amount, @discount_amount, @actual_amount, @member_id, @member_name,
                @table_no, @remark, @status, @pay_status, @pay_method, @synced, @created_at, @updated_at, @paid_at)
      `)

      const insertItem = this.db.prepare(`
        INSERT INTO order_items 
        (order_no, product_id, product_name, sku_id, sku_name, attribute_ids, 
         attribute_names, price, quantity, subtotal, remark, created_at)
        VALUES (@order_no, @product_id, @product_name, @sku_id, @sku_name, @attribute_ids,
                @attribute_names, @price, @quantity, @subtotal, @remark, @created_at)
      `)

      insertOrder.run({
        order_no: order.order_no,
        total_amount: order.total_amount,
        discount_amount: order.discount_amount || 0,
        actual_amount: order.actual_amount,
        member_id: order.member_id || null,
        member_name: order.member_name || '',
        table_no: order.table_no || '',
        remark: order.remark || '',
        status: order.status || 0,
        pay_status: order.pay_status || 0,
        pay_method: order.pay_method || '',
        synced: 0,
        created_at: order.created_at,
        updated_at: order.updated_at,
        paid_at: order.paid_at || null
      })

      if (order.items && order.items.length > 0) {
        order.items.forEach(item => {
          insertItem.run({
            order_no: order.order_no,
            product_id: item.product_id,
            product_name: item.product_name,
            sku_id: item.sku_id,
            sku_name: item.sku_name,
            attribute_ids: item.attribute_ids ? JSON.stringify(item.attribute_ids) : '',
            attribute_names: item.attribute_names ? JSON.stringify(item.attribute_names) : '',
            price: item.price,
            quantity: item.quantity,
            subtotal: item.subtotal,
            remark: item.remark || '',
            created_at: order.created_at
          })
        })
      }
    })

    tx()
    return { success: true, order_no: order.order_no }
  }

  getPendingOrders() {
    const orders = this.db.prepare('SELECT * FROM orders WHERE synced = 0 ORDER BY created_at').all()
    return orders.map(o => ({
      ...o,
      items: this.db.prepare('SELECT * FROM order_items WHERE order_no = ?').all(o.order_no)
    }))
  }

  updateOrderStatus(orderNo, status) {
    const result = this.db.prepare('UPDATE orders SET status = ?, updated_at = datetime() WHERE order_no = ?').run(status, orderNo)
    return { success: result.changes > 0 }
  }

  getOrdersByDate(date) {
    const orders = this.db.prepare(`
      SELECT * FROM orders 
      WHERE DATE(created_at) = DATE(?) 
      ORDER BY created_at DESC
    `).all(date)

    return orders.map(o => ({
      ...o,
      items: this.db.prepare('SELECT * FROM order_items WHERE order_no = ?').all(o.order_no)
    }))
  }

  getOrderByNo(orderNo) {
    const order = this.db.prepare('SELECT * FROM orders WHERE order_no = ?').get(orderNo)
    if (!order) return null

    order.items = this.db.prepare('SELECT * FROM order_items WHERE order_no = ?').all(orderNo)
    return order
  }

  deleteOrder(orderNo) {
    const tx = this.db.transaction(() => {
      this.db.prepare('DELETE FROM order_items WHERE order_no = ?').run(orderNo)
      this.db.prepare('DELETE FROM orders WHERE order_no = ?').run(orderNo)
    })

    tx()
    return { success: true }
  }

  close() {
    if (this.db) {
      this.db.close()
    }
  }
}

module.exports = SQLiteDatabase
