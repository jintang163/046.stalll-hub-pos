import dayjs from 'dayjs'
import { v4 as uuidv4 } from 'uuid'

export const generateOrderNo = () => {
  const dateStr = dayjs().format('YYYYMMDDHHmmss')
  const random = Math.random().toString(36).substring(2, 8).toUpperCase()
  return `DD${dateStr}${random}`
}

export const generateOrder = (cart, options = {}) => {
  const now = dayjs().format('YYYY-MM-DD HH:mm:ss')
  const orderNo = generateOrderNo()
  
  const items = cart.items.map(item => ({
    product_id: item.product_id,
    product_name: item.product_name,
    sku_id: item.sku_id,
    sku_name: item.sku_name,
    attribute_ids: item.attribute_ids,
    attribute_names: item.attribute_names,
    price: item.price,
    quantity: item.quantity,
    subtotal: item.subtotal,
    remark: item.remark || ''
  }))

  const totalAmount = cart.total
  const discountAmount = options.discountAmount || 0
  const actualAmount = Math.max(0, totalAmount - discountAmount)

  return {
    order_no: orderNo,
    total_amount: totalAmount,
    discount_amount: discountAmount,
    actual_amount: actualAmount,
    member_id: options.member_id || null,
    member_name: options.member_name || '',
    table_no: options.table_no || '',
    remark: options.remark || '',
    status: 0,
    pay_status: options.paid ? 1 : 0,
    pay_method: options.pay_method || '',
    created_at: now,
    updated_at: now,
    paid_at: options.paid ? now : null,
    items
  }
}

export const formatCurrency = (amount) => {
  return `¥${Number(amount).toFixed(2)}`
}
