const Database = require('better-sqlite3')
const path = require('path')
const fs = require('fs')

class SQLiteDatabase {
  constructor(dbPath) {
    this.dbPath = dbPath
    this.db = null
  }

  init(dbPath) {
    if (dbPath) {
      this.dbPath = dbPath
    }

    const dbDir = path.dirname(this.dbPath)
    if (!fs.existsSync(dbDir)) {
      fs.mkdirSync(dbDir, { recursive: true })
    }

    this.db = new Database(this.dbPath)
    this.db.pragma('journal_mode = WAL')
    this.db.pragma('foreign_keys = ON')
    this.db.pragma('synchronous = NORMAL')
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
      
      `CREATE TABLE IF NOT EXISTS skus (
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
      
      `CREATE TABLE IF NOT EXISTS attributes (
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
        FOREIGN KEY (attribute_id) REFERENCES attributes(id) ON DELETE CASCADE
      )`,
      
      `CREATE TABLE IF NOT EXISTS product_attributes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        product_id INTEGER NOT NULL,
        attribute_id INTEGER NOT NULL,
        created_at TEXT,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
        FOREIGN KEY (attribute_id) REFERENCES attributes(id) ON DELETE CASCADE,
        UNIQUE(product_id, attribute_id)
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
      
      `CREATE TABLE IF NOT EXISTS sync_records (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sync_type TEXT NOT NULL,
        status TEXT NOT NULL,
        total_count INTEGER DEFAULT 0,
        success_count INTEGER DEFAULT 0,
        fail_count INTEGER DEFAULT 0,
        start_time TEXT,
        end_time TEXT,
        error_message TEXT,
        created_at TEXT
      )`
    ]

    tables.forEach(sql => this.db.exec(sql))
  }

  createIndexes() {
    const indexes = [
      'CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id)',
      'CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)',
      'CREATE INDEX IF NOT EXISTS idx_skus_product ON skus(product_id)',
      'CREATE INDEX IF NOT EXISTS idx_skus_status ON skus(status)',
      'CREATE INDEX IF NOT EXISTS idx_skus_code ON skus(sku_code)',
      'CREATE INDEX IF NOT EXISTS idx_attributes_product ON attributes(product_id)',
      'CREATE INDEX IF NOT EXISTS idx_attr_values_attr ON attribute_values(attribute_id)',
      'CREATE INDEX IF NOT EXISTS idx_product_attr_product ON product_attributes(product_id)',
      'CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)',
      'CREATE INDEX IF NOT EXISTS idx_orders_synced ON orders(synced)',
      'CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at)',
      'CREATE INDEX IF NOT EXISTS idx_orders_paid ON orders(paid_at)',
      'CREATE INDEX IF NOT EXISTS idx_sync_records_type ON sync_records(sync_type)',
      'CREATE INDEX IF NOT EXISTS idx_sync_records_status ON sync_records(status)',
      'CREATE INDEX IF NOT EXISTS idx_sync_records_time ON sync_records(start_time)'
    ]

    indexes.forEach(sql => this.db.exec(sql))
  }

  insert(table, data) {
    const columns = Object.keys(data)
    const placeholders = columns.map(() => '?').join(', ')
    const values = columns.map(col => data[col])
    const sql = `INSERT INTO ${table} (${columns.join(', ')}) VALUES (${placeholders})`
    const stmt = this.db.prepare(sql)
    const result = stmt.run(...values)
    return { success: true, id: result.lastInsertRowid, changes: result.changes }
  }

  update(table, data, where, whereParams = []) {
    const columns = Object.keys(data)
    const setClause = columns.map(col => `${col} = ?`).join(', ')
    const values = [...columns.map(col => data[col]), ...whereParams]
    const sql = `UPDATE ${table} SET ${setClause} WHERE ${where}`
    const stmt = this.db.prepare(sql)
    const result = stmt.run(...values)
    return { success: true, changes: result.changes }
  }

  delete(table, where, whereParams = []) {
    const sql = `DELETE FROM ${table} WHERE ${where}`
    const stmt = this.db.prepare(sql)
    const result = stmt.run(...whereParams)
    return { success: true, changes: result.changes }
  }

  query(sql, params = []) {
    const stmt = this.db.prepare(sql)
    return stmt.all(...params)
  }

  queryOne(sql, params = []) {
    const stmt = this.db.prepare(sql)
    return stmt.get(...params)
  }

  transaction(callback) {
    const tx = this.db.transaction(callback)
    return tx()
  }

  batchInsert(table, dataList) {
    if (!dataList || dataList.length === 0) {
      return { success: true, count: 0 }
    }

    const columns = Object.keys(dataList[0])
    const placeholders = columns.map(() => '?').join(', ')
    const sql = `INSERT OR REPLACE INTO ${table} (${columns.join(', ')}) VALUES (${placeholders})`
    const stmt = this.db.prepare(sql)

    const tx = this.db.transaction((items) => {
      let count = 0
      for (const data of items) {
        const values = columns.map(col => data[col])
        stmt.run(...values)
        count++
      }
      return count
    })

    const count = tx(dataList)
    return { success: true, count }
  }

  batchUpdate(table, dataList, whereFields) {
    if (!dataList || dataList.length === 0) {
      return { success: true, count: 0 }
    }

    const tx = this.db.transaction((items) => {
      let count = 0
      for (const data of items) {
        const updateData = { ...data }
        const whereClauses = []
        const whereParams = []

        for (const field of whereFields) {
          whereClauses.push(`${field} = ?`)
          whereParams.push(data[field])
          delete updateData[field]
        }

        const columns = Object.keys(updateData)
        const setClause = columns.map(col => `${col} = ?`).join(', ')
        const values = [...columns.map(col => updateData[col]), ...whereParams]
        const sql = `UPDATE ${table} SET ${setClause} WHERE ${whereClauses.join(' AND ')}`
        const stmt = this.db.prepare(sql)
        const result = stmt.run(...values)
        count += result.changes
      }
      return count
    })

    const count = tx(dataList)
    return { success: true, count }
  }

  batchDelete(table, whereField, ids) {
    if (!ids || ids.length === 0) {
      return { success: true, count: 0 }
    }

    const placeholders = ids.map(() => '?').join(', ')
    const sql = `DELETE FROM ${table} WHERE ${whereField} IN (${placeholders})`
    const stmt = this.db.prepare(sql)

    const tx = this.db.transaction(() => {
      const result = stmt.run(...ids)
      return result.changes
    })

    const count = tx()
    return { success: true, count }
  }

  select(table) {
    return new QueryBuilder(this.db, table)
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

    const skus = this.db.prepare('SELECT * FROM skus WHERE product_id = ?').all(id)
    const attributes = this.db.prepare(`
      SELECT a.*, 
             json_group_array(json_object(
               'id', av.id,
               'value', av.value,
               'extra_price', av.extra_price,
               'stock', av.stock,
               'sort_order', av.sort_order,
               'status', av.status
             )) as values
      FROM attributes a
      LEFT JOIN attribute_values av ON av.attribute_id = a.id
      WHERE a.product_id = ?
      GROUP BY a.id
      ORDER BY a.sort_order
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
    return this.db.prepare('SELECT * FROM skus WHERE product_id = ? ORDER BY id').all(productId)
  }

  getAttributes(productId) {
    const attrs = this.db.prepare(`
      SELECT a.*, 
             json_group_array(json_object(
               'id', av.id,
               'value', av.value,
               'extra_price', av.extra_price,
               'stock', av.stock,
               'sort_order', av.sort_order,
               'status', av.status
             )) as values
      FROM attributes a
      LEFT JOIN attribute_values av ON av.attribute_id = a.id
      WHERE a.product_id = ?
      GROUP BY a.id
      ORDER BY a.sort_order
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
        INSERT OR REPLACE INTO skus 
        (id, product_id, sku_code, spec_name, price, original_price, stock, image, status, created_at, updated_at)
        VALUES (@id, @product_id, @sku_code, @spec_name, @price, @original_price, @stock, @image, @status, @created_at, @updated_at)
      `)

      const deleteSKUs = this.db.prepare('DELETE FROM skus WHERE product_id = ?')
      
      const insertAttr = this.db.prepare(`
        INSERT OR REPLACE INTO attributes 
        (id, product_id, name, sort_order, status, created_at, updated_at)
        VALUES (@id, @product_id, @name, @sort_order, @status, @created_at, @updated_at)
      `)

      const deleteAttrs = this.db.prepare('DELETE FROM attributes WHERE product_id = ?')
      
      const insertAttrValue = this.db.prepare(`
        INSERT OR REPLACE INTO attribute_values 
        (id, attribute_id, value, extra_price, stock, sort_order, status, created_at, updated_at)
        VALUES (@id, @attribute_id, @value, @extra_price, @stock, @sort_order, @status, @created_at, @updated_at)
      `)

      const deleteAttrValues = this.db.prepare(`
        DELETE FROM attribute_values WHERE attribute_id IN 
        (SELECT id FROM attributes WHERE product_id = ?)
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
    const result = this.db.prepare('UPDATE skus SET stock = ? WHERE id = ?').run(stock, skuId)
    return { success: result.changes > 0 }
  }

  updateProductStatus(productId, status) {
    const result = this.db.prepare('UPDATE products SET status = ? WHERE id = ?').run(status ? 1 : 0, productId)
    return { success: result.changes > 0 }
  }

  deleteProduct(productId) {
    const tx = this.db.transaction(() => {
      this.db.prepare('DELETE FROM attribute_values WHERE attribute_id IN (SELECT id FROM attributes WHERE product_id = ?)').run(productId)
      this.db.prepare('DELETE FROM product_attributes WHERE product_id = ?').run(productId)
      this.db.prepare('DELETE FROM attributes WHERE product_id = ?').run(productId)
      this.db.prepare('DELETE FROM skus WHERE product_id = ?').run(productId)
      this.db.prepare('DELETE FROM products WHERE id = ?').run(productId)
    })

    tx()
    return { success: true }
  }

  deleteCategory(categoryId) {
    const result = this.db.prepare('DELETE FROM categories WHERE id = ?').run(categoryId)
    return { success: result.changes > 0 }
  }

  clearAllProducts() {
    const tx = this.db.transaction(() => {
      this.db.exec('DELETE FROM attribute_values')
      this.db.exec('DELETE FROM product_attributes')
      this.db.exec('DELETE FROM attributes')
      this.db.exec('DELETE FROM skus')
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

  createSyncRecord(record) {
    const result = this.db.prepare(`
      INSERT INTO sync_records 
      (sync_type, status, total_count, success_count, fail_count, start_time, end_time, error_message, created_at)
      VALUES (@sync_type, @status, @total_count, @success_count, @fail_count, @start_time, @end_time, @error_message, @created_at)
    `).run({
      sync_type: record.sync_type,
      status: record.status,
      total_count: record.total_count || 0,
      success_count: record.success_count || 0,
      fail_count: record.fail_count || 0,
      start_time: record.start_time,
      end_time: record.end_time || null,
      error_message: record.error_message || null,
      created_at: record.created_at || new Date().toISOString()
    })
    return { success: true, id: result.lastInsertRowid }
  }

  updateSyncRecord(id, updates) {
    const columns = Object.keys(updates)
    const setClause = columns.map(col => `${col} = ?`).join(', ')
    const values = [...columns.map(col => updates[col]), id]
    const result = this.db.prepare(`UPDATE sync_records SET ${setClause} WHERE id = ?`).run(...values)
    return { success: result.changes > 0 }
  }

  getSyncRecords(limit = 20) {
    return this.db.prepare('SELECT * FROM sync_records ORDER BY start_time DESC LIMIT ?').all(limit)
  }

  raw(sql, params = []) {
    return this.db.exec(sql, params)
  }

  close() {
    if (this.db) {
      this.db.close()
      this.db = null
    }
  }
}

class QueryBuilder {
  constructor(db, table) {
    this.db = db
    this.table = table
    this._columns = ['*']
    this._where = []
    this._params = []
    this._order = null
    this._limit = null
    this._offset = null
    this._join = []
  }

  select(...columns) {
    this._columns = columns
    return this
  }

  where(condition, params = []) {
    this._where.push(condition)
    this._params.push(...params)
    return this
  }

  andWhere(condition, params = []) {
    return this.where(condition, params)
  }

  orderBy(column, direction = 'ASC') {
    this._order = `${column} ${direction}`
    return this
  }

  limit(limit) {
    this._limit = limit
    return this
  }

  offset(offset) {
    this._offset = offset
    return this
  }

  join(table, on) {
    this._join.push(`JOIN ${table} ON ${on}`)
    return this
  }

  leftJoin(table, on) {
    this._join.push(`LEFT JOIN ${table} ON ${on}`)
    return this
  }

  _buildQuery() {
    let sql = `SELECT ${this._columns.join(', ')} FROM ${this.table}`

    if (this._join.length > 0) {
      sql += ' ' + this._join.join(' ')
    }

    if (this._where.length > 0) {
      sql += ' WHERE ' + this._where.join(' AND ')
    }

    if (this._order) {
      sql += ' ORDER BY ' + this._order
    }

    if (this._limit !== null) {
      sql += ' LIMIT ' + this._limit
    }

    if (this._offset !== null) {
      sql += ' OFFSET ' + this._offset
    }

    return sql
  }

  get() {
    const sql = this._buildQuery()
    const stmt = this.db.prepare(sql)
    return stmt.all(...this._params)
  }

  first() {
    const sql = this._buildQuery()
    const stmt = this.db.prepare(sql)
    return stmt.get(...this._params)
  }

  count() {
    const sql = `SELECT COUNT(*) as count FROM ${this.table}` + 
                (this._where.length > 0 ? ' WHERE ' + this._where.join(' AND ') : '')
    const stmt = this.db.prepare(sql)
    const result = stmt.get(...this._params)
    return result ? result.count : 0
  }

  insert(data) {
    const columns = Object.keys(data)
    const placeholders = columns.map(() => '?').join(', ')
    const values = columns.map(col => data[col])
    const sql = `INSERT INTO ${this.table} (${columns.join(', ')}) VALUES (${placeholders})`
    const stmt = this.db.prepare(sql)
    return stmt.run(...values)
  }

  update(data) {
    const columns = Object.keys(data)
    const setClause = columns.map(col => `${col} = ?`).join(', ')
    const values = [...columns.map(col => data[col]), ...this._params]
    const sql = `UPDATE ${this.table} SET ${setClause}` + 
                (this._where.length > 0 ? ' WHERE ' + this._where.join(' AND ') : '')
    const stmt = this.db.prepare(sql)
    return stmt.run(...values)
  }

  delete() {
    const sql = `DELETE FROM ${this.table}` + 
                (this._where.length > 0 ? ' WHERE ' + this._where.join(' AND ') : '')
    const stmt = this.db.prepare(sql)
    return stmt.run(...this._params)
  }
}

module.exports = SQLiteDatabase
