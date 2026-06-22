CREATE DATABASE IF NOT EXISTS stalll_pos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE stalll_pos;

CREATE USER IF NOT EXISTS 'stalll'@'%' IDENTIFIED BY 'stalll123';
GRANT ALL PRIVILEGES ON stalll_pos.* TO 'stalll'@'%';
FLUSH PRIVILEGES;

-- 供应商表
CREATE TABLE IF NOT EXISTS suppliers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id BIGINT UNSIGNED NOT NULL,
    supplier_no VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    short_name VARCHAR(50),
    category VARCHAR(50),
    contact_person VARCHAR(50),
    phone VARCHAR(20),
    mobile VARCHAR(20),
    email VARCHAR(100),
    fax VARCHAR(20),
    address VARCHAR(255),
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    bank_name VARCHAR(100),
    bank_account VARCHAR(50),
    bank_account_name VARCHAR(100),
    tax_no VARCHAR(50),
    payment_term INT DEFAULT 0,
    payment_term_desc VARCHAR(50),
    settlement_method VARCHAR(30) DEFAULT 'bank_transfer',
    credit_limit DECIMAL(12,2) DEFAULT 0,
    current_payable DECIMAL(12,2) DEFAULT 0,
    total_purchase DECIMAL(12,2) DEFAULT 0,
    total_paid DECIMAL(12,2) DEFAULT 0,
    status INT DEFAULT 1,
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_store_id (store_id),
    INDEX idx_category (category),
    INDEX idx_status (status),
    UNIQUE INDEX uk_supplier_no (supplier_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 采购收货单表
CREATE TABLE IF NOT EXISTS purchase_receives (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id BIGINT UNSIGNED NOT NULL,
    purchase_id BIGINT UNSIGNED NOT NULL,
    purchase_no VARCHAR(50),
    supplier_id BIGINT UNSIGNED,
    supplier_name VARCHAR(100),
    receive_no VARCHAR(50) NOT NULL,
    receive_type VARCHAR(20) DEFAULT 'full',
    total_qty DECIMAL(12,2) DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    remark VARCHAR(255),
    operator_id BIGINT UNSIGNED,
    operator_name VARCHAR(50),
    received_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_store_id (store_id),
    INDEX idx_purchase_id (purchase_id),
    INDEX idx_supplier_id (supplier_id),
    INDEX idx_purchase_no (purchase_no),
    UNIQUE INDEX uk_receive_no (receive_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 采购收货明细表
CREATE TABLE IF NOT EXISTS purchase_receive_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    receive_id BIGINT UNSIGNED NOT NULL,
    purchase_item_id BIGINT UNSIGNED,
    ingredient_id BIGINT UNSIGNED NOT NULL,
    ingredient_name VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    unit VARCHAR(20),
    purchase_qty DECIMAL(10,2),
    received_qty DECIMAL(10,2),
    qualified_qty DECIMAL(10,2),
    rejected_qty DECIMAL(10,2),
    unit_price DECIMAL(10,2),
    subtotal DECIMAL(12,2),
    batch_no VARCHAR(50),
    expiry_date VARCHAR(10),
    reject_reason VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_receive_id (receive_id),
    INDEX idx_ingredient_id (ingredient_id),
    INDEX idx_batch_no (batch_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 应付账款表
CREATE TABLE IF NOT EXISTS accounts_payables (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NOT NULL,
    supplier_name VARCHAR(100),
    payable_no VARCHAR(50) NOT NULL,
    business_type VARCHAR(30) DEFAULT 'purchase',
    business_id BIGINT UNSIGNED,
    business_no VARCHAR(50),
    amount DECIMAL(12,2) NOT NULL,
    paid_amount DECIMAL(12,2) DEFAULT 0,
    balance DECIMAL(12,2) DEFAULT 0,
    due_date VARCHAR(10),
    status VARCHAR(20) DEFAULT 'unpaid',
    is_overdue INT DEFAULT 0,
    remark VARCHAR(255),
    paid_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_store_id (store_id),
    INDEX idx_supplier_id (supplier_id),
    INDEX idx_business_id (business_id),
    INDEX idx_business_no (business_no),
    INDEX idx_status (status),
    INDEX idx_due_date (due_date),
    INDEX idx_is_overdue (is_overdue),
    UNIQUE INDEX uk_payable_no (payable_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 付款记录表
CREATE TABLE IF NOT EXISTS payable_payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NOT NULL,
    supplier_name VARCHAR(100),
    payable_id BIGINT UNSIGNED,
    payment_no VARCHAR(50) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    payment_method VARCHAR(30) DEFAULT 'bank_transfer',
    payment_date VARCHAR(10),
    operator_id BIGINT UNSIGNED,
    operator_name VARCHAR(50),
    voucher_no VARCHAR(50),
    voucher_url VARCHAR(255),
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_store_id (store_id),
    INDEX idx_supplier_id (supplier_id),
    INDEX idx_payable_id (payable_id),
    INDEX idx_payment_date (payment_date),
    UNIQUE INDEX uk_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 对账单表
CREATE TABLE IF NOT EXISTS reconciliations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NOT NULL,
    supplier_name VARCHAR(100),
    reconcile_no VARCHAR(50) NOT NULL,
    period_start VARCHAR(10),
    period_end VARCHAR(10),
    system_amount DECIMAL(12,2) DEFAULT 0,
    supplier_amount DECIMAL(12,2) DEFAULT 0,
    diff_amount DECIMAL(12,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft',
    confirmed_at DATETIME,
    confirmed_by VARCHAR(50),
    remark VARCHAR(255),
    difference_remark VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_store_id (store_id),
    INDEX idx_supplier_id (supplier_id),
    INDEX idx_period_start (period_start),
    INDEX idx_period_end (period_end),
    INDEX idx_status (status),
    UNIQUE INDEX uk_reconcile_no (reconcile_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 对账明细表
CREATE TABLE IF NOT EXISTS reconciliation_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reconcile_id BIGINT UNSIGNED NOT NULL,
    business_type VARCHAR(30),
    business_id BIGINT UNSIGNED,
    business_no VARCHAR(50),
    business_date VARCHAR(10),
    system_amount DECIMAL(12,2),
    supplier_amount DECIMAL(12,2),
    diff_amount DECIMAL(12,2),
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_reconcile_id (reconcile_id),
    INDEX idx_business_id (business_id),
    INDEX idx_business_no (business_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 给采购订单表增加新字段（如果已存在则忽略错误）
-- ALTER TABLE purchase_orders ADD COLUMN supplier_id BIGINT UNSIGNED AFTER store_id;
-- ALTER TABLE purchase_orders ADD COLUMN received_amount DECIMAL(12,2) DEFAULT 0 AFTER total_amount;
-- ALTER TABLE purchase_orders ADD COLUMN received_quantity INT DEFAULT 0 AFTER total_quantity;
-- ALTER TABLE purchase_orders ADD COLUMN payment_term INT DEFAULT 0 AFTER forecast_days;
-- ALTER TABLE purchase_orders ADD COLUMN expected_date VARCHAR(10) AFTER payment_term;
-- ALTER TABLE purchase_order_items ADD COLUMN received_qty DECIMAL(10,2) DEFAULT 0 AFTER purchase_qty;
-- CREATE INDEX idx_po_supplier_id ON purchase_orders(supplier_id);

-- 示例供应商数据
INSERT INTO suppliers (store_id, supplier_no, name, short_name, category, contact_person, phone, mobile, email, address, province, city, district, bank_name, bank_account, bank_account_name, tax_no, payment_term, payment_term_desc, settlement_method, credit_limit, status, remark) VALUES
(1, 'SUP202401010001', '北京鲜优蔬果配送有限公司', '鲜优蔬果', '蔬菜', '张经理', '010-12345678', '13800138001', 'zhang@xianyou.com', '北京市朝阳区十八里店乡大洋路农副产品批发市场', '北京市', '北京市', '朝阳区', '中国工商银行北京朝阳支行', '6222021234567890123', '北京鲜优蔬果配送有限公司', '91110105MA01234567', 30, '月结30天', 'bank_transfer', 50000.00, 1, '主要叶菜供应商'),
(1, 'SUP202401010002', '上海福润肉类食品有限公司', '福润肉类', '肉类', '李主管', '021-87654321', '13900139002', 'li@furun.com', '上海市浦东新区沪南路2000号农产品中心批发市场', '上海市', '上海市', '浦东新区', '中国建设银行上海浦东分行', '6227001234567890456', '上海福润肉类食品有限公司', '91310115MA12345678', 15, '月结15天', 'bank_transfer', 80000.00, 1, '冷鲜肉专业供应商'),
(1, 'SUP202401010003', '广州海鲜水产批发中心', '海鲜水产', '水产', '王老板', '020-55667788', '13700137003', 'wang@haixian.com', '广州市荔湾区黄沙大道水产批发市场', '广东省', '广州市', '荔湾区', '中国农业银行广州荔湾支行', '6228481234567890789', '广州海鲜水产批发中心', '91440103MA12345678', 7, '7天账期', 'bank_transfer', 30000.00, 1, '海鲜水产每日配送'),
(1, 'SUP202401010004', '五常大米产地直供合作社', '五常大米', '粮油', '赵社长', '0451-22334455', '13600136004', 'zhao@wuchang.com', '黑龙江省哈尔滨市五常市民乐朝鲜族乡', '黑龙江省', '哈尔滨市', '五常市', '中国邮政储蓄银行五常支行', '6217991234567890012', '五常大米产地直供合作社', '92230184MA12345678', 0, '货到付款', 'bank_transfer', 20000.00, 1, '东北五常稻花香大米'),
(1, 'SUP202401010005', '益达调味品批发商行', '益达调味', '调味品', '孙经理', '0755-33445566', '13500135005', 'sun@yida.com', '深圳市龙岗区平湖街道白泥坑社区海吉星国际农产品物流园', '广东省', '深圳市', '龙岗区', '招商银行深圳龙岗支行', '6214831234567890345', '益达调味品批发商行', '91440307MA12345678', 30, '月结30天', 'bank_transfer', 15000.00, 1, '各类调味品、酱料供应商');
