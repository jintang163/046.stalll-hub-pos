import request from './request'
import type { Product, ProductDetail, Category } from '../types'

export const productApi = {
  listProducts: (params: {
    store_id: number
    category_id?: number
    name?: string
    status?: number
    page?: number
    page_size?: number
  }): Promise<{ list: Product[]; total: number }> =>
    request({
      url: `/products?${new URLSearchParams(params as any).toString()}`,
      method: 'GET'
    }),

  getProduct: (id: number): Promise<ProductDetail> =>
    request({
      url: `/products/${id}`,
      method: 'GET'
    }),

  listCategories: (storeId: number): Promise<Category[]> =>
    request({
      url: `/categories?store_id=${storeId}`,
      method: 'GET'
    })
}
