import request from './request'

export interface VoiceMatchResult {
  product_id: number
  product_name: string
  sku_id: number
  sku_name: string
  price: number
  quantity: number
  match_score: number
  image: string
}

export interface VoiceParseResponse {
  original_text: string
  items: VoiceMatchResult[]
  unmatched: string[]
}

export const parseVoiceText = (store_id: number, text: string) => {
  return request<VoiceParseResponse>({
    url: '/mini/voice/parse',
    method: 'POST',
    data: { store_id, text }
  })
}
