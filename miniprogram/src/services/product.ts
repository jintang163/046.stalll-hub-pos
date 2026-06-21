import request from './request'

export interface Category {
  id: number
  name: string
  sort_order: number
  status: number
  product_count?: number
}

export interface SKU {
  id: number
  product_id: number
  sku_code: string
  spec_name: string
  price: number
  original_price: number
  stock: number
  image: string
  status: number
  is_sold_out: boolean
}

export interface AttributeValue {
  id: number
  attribute_id: number
  value: string
  extra_price: number
  stock: number
  sort_order: number
  status: number
}

export interface Attribute {
  id: number
  product_id: number
  name: string
  sort_order: number
  status: number
  values: AttributeValue[]
}

export interface Product {
  id: number
  name: string
  description: string
  image: string
  category_id: number
  sort_order: number
  status: number
  is_hot: boolean
  is_recommend: boolean
  warning_threshold: number
  category_name?: string
  skus: SKU[]
  attributes: Attribute[]
  min_price?: number
  max_price?: number
  total_stock?: number
}

export const getCategories = (storeId: number) => {
  return request<Category[]>({
    url: `/categories?store_id=${storeId}`,
    method: 'GET',
    needLogin: false
  })
}

export const getProducts = (storeId: number, categoryId?: number, page = 1, pageSize = 20) => {
  const params = new URLSearchParams()
  params.append('store_id', storeId.toString())
  params.append('page', page.toString())
  params.append('page_size', pageSize.toString())
  if (categoryId) {
    params.append('category_id', categoryId.toString())
  }
  
  return request<{ list: Product[]; total: number }>({
    url: `/products?${params.toString()}`,
    method: 'GET',
    needLogin: false
  })
}

export const getProductDetail = (id: number) => {
  return request<Product>({
    url: `/products/${id}`,
    method: 'GET',
    needLogin: false
  })
}

export const searchProducts = (storeId: number, keyword: string) => {
  return request<Product[]>({
    url: `/products/search?store_id=${storeId}&keyword=${encodeURIComponent(keyword)}`,
    method: 'GET',
    needLogin: false
  })
}
