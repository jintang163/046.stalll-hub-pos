import React, { useState, useRef, useCallback, useEffect } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Toast, Popup, Cell } from '@nutui/nutui-react-taro'
import { parseVoiceText, VoiceMatchResult, VoiceParseResponse } from '../../services/voice'
import { useAppStore } from '../../store/app'
import { useCartStore } from '../../store/cart'
import type { Product, SKU, Attribute, AttributeValue } from '../../services/product'
import { getProductDetail } from '../../services/product'
import styles from './index.module.scss'

declare const wx: any

const WECHAT_SI_PLUGIN = 'WechatSI'

const VoiceOrder: React.FC = () => {
  const currentStore = useAppStore(state => state.currentStore)
  const addItem = useCartStore(state => state.addItem)

  const [recording, setRecording] = useState(false)
  const [recognizedText, setRecognizedText] = useState('')
  const [parseResult, setParseResult] = useState<VoiceParseResponse | null>(null)
  const [parsing, setParsing] = useState(false)
  const [addingToCart, setAddingToCart] = useState(false)
  const [itemsAdded, setItemsAdded] = useState(0)

  const [needSelectProduct, setNeedSelectProduct] = useState<{
    index: number
    result: VoiceMatchResult
  } | null>(null)
  const [loadedProduct, setLoadedProduct] = useState<Product | null>(null)
  const [selectedSku, setSelectedSku] = useState<SKU | null>(null)
  const [selectedAttrs, setSelectedAttrs] = useState<Map<number, AttributeValue>>(new Map())
  const [selectQuantity, setSelectQuantity] = useState(1)

  const managerRef = useRef<any>(null)
  const fallbackTimerRef = useRef<any>(null)
  const recordingStateRef = useRef(false)

  const getWechatSI = useCallback((): any => {
    try {
      if (typeof wx !== 'undefined' && wx.requirePlugin) {
        return wx.requirePlugin(WECHAT_SI_PLUGIN)
      }
      if (typeof Taro !== 'undefined' && (Taro as any).requirePlugin) {
        return (Taro as any).requirePlugin(WECHAT_SI_PLUGIN)
      }
      return null
    } catch {
      return null
    }
  }, [])

  const handleRecognized = useCallback((text: string) => {
    if (!text) return
    const cleanText = text.trim().replace(/[。？！\s]+$/, '')
    if (cleanText && cleanText !== recognizedText) {
      setRecognizedText(cleanText)
      handleParse(cleanText)
    }
  }, [recognizedText])

  const startWechatSI = useCallback(() => {
    const plugin = getWechatSI()
    if (!plugin || !plugin.getRecordRecognitionManager) {
      return false
    }

    try {
      const manager = plugin.getRecordRecognitionManager()
      managerRef.current = manager

      manager.onStart = () => {
        recordingStateRef.current = true
        setRecording(true)
      }

      manager.onRecognize = (res: any) => {
        const text = res?.result || res?.ret || ''
        if (text) {
          setRecognizedText(text.trim())
        }
      }

      manager.onStop = (res: any) => {
        recordingStateRef.current = false
        setRecording(false)
        const text = (res?.result || res?.ret || '').trim()
        if (text) {
          handleRecognized(text)
        } else {
          Toast.show('未能识别语音，请重试')
        }
      }

      manager.onError = (err: any) => {
        console.warn('WechatSI onError', err)
        recordingStateRef.current = false
        setRecording(false)
        const msg = err?.msg || err?.message || ''
        if (msg.includes('权限') || msg.includes('auth')) {
          Toast.show('请在设置中开启麦克风权限')
        } else {
          startFallbackRecord()
        }
      }

      manager.start({
        lang: 'zh_CN',
        duration: 60000
      })

      return true
    } catch (e) {
      console.warn('WechatSI start error', e)
      return false
    }
  }, [getWechatSI, handleRecognized])

  const stopWechatSI = useCallback(() => {
    try {
      if (managerRef.current && managerRef.current.stop) {
        managerRef.current.stop()
      }
    } catch {}
    managerRef.current = null
  }, [])

  const startFallbackRecord = useCallback(() => {
    Toast.show('语音识别加载中，使用录音模式...')
    try {
      Taro.startRecord({
        success: () => {
          recordingStateRef.current = true
          setRecording(true)
          fallbackTimerRef.current = setTimeout(() => {
            stopFallbackRecord()
          }, 5000)
        },
        fail: () => {
          setRecording(false)
          Toast.show('录音失败，请检查麦克风权限')
        }
      })
    } catch {
      setRecording(false)
      Toast.show('录音初始化失败')
    }
  }, [])

  const stopFallbackRecord = useCallback(() => {
    try {
      if (fallbackTimerRef.current) {
        clearTimeout(fallbackTimerRef.current)
        fallbackTimerRef.current = null
      }
      Taro.stopRecord({
        success: (res: any) => {
          recordingStateRef.current = false
          setRecording(false)
          const filePath = res.tempFilePath
          if (filePath) {
            translateWithPlugin(filePath)
          }
        },
        fail: () => {
          recordingStateRef.current = false
          setRecording(false)
        }
      })
    } catch {
      recordingStateRef.current = false
      setRecording(false)
    }
  }, [])

  const translateWithPlugin = useCallback((filePath: string) => {
    const plugin = getWechatSI()
    if (!plugin || !plugin.translateVoice) {
      Toast.show('语音识别插件不可用')
      return
    }
    try {
      plugin.translateVoice({
        lfrom: 'zh_CN',
        lto: 'zh_CN',
        content: filePath,
        tts: false,
        success: (res: any) => {
          const text = (res?.translateResult || res?.result || '').trim()
          if (text) {
            handleRecognized(text)
          } else {
            Toast.show('未能识别语音，请放慢语速重试')
          }
        },
        fail: (err: any) => {
          console.warn('translateVoice fail', err)
          Toast.show('语音识别失败，请重试或手动输入')
        }
      })
    } catch {
      Toast.show('语音识别失败')
    }
  }, [getWechatSI, handleRecognized])

  const handleStartRecord = () => {
    if (!currentStore) {
      Toast.show('请先选择门店')
      return
    }
    if (recordingStateRef.current) return

    setRecognizedText('')
    setParseResult(null)
    setItemsAdded(0)

    const started = startWechatSI()
    if (!started) {
      startFallbackRecord()
    }
  }

  const handleStopRecord = () => {
    if (!recordingStateRef.current) return

    if (managerRef.current) {
      stopWechatSI()
    } else {
      stopFallbackRecord()
    }
  }

  useEffect(() => {
    return () => {
      if (fallbackTimerRef.current) clearTimeout(fallbackTimerRef.current)
      if (recordingStateRef.current) {
        try { managerRef.current && managerRef.current.stop && managerRef.current.stop() } catch {}
        try { Taro.stopRecord && Taro.stopRecord({}) } catch {}
      }
    }
  }, [])

  const handleParse = async (text: string) => {
    if (!currentStore || !text.trim()) return
    setParsing(true)
    try {
      const result = await parseVoiceText(currentStore.id, text.trim())
      setParseResult(result)
    } catch (e: any) {
      Toast.show(e.message || '解析失败')
    } finally {
      setParsing(false)
    }
  }

  const resolveProductDefault = async (result: VoiceMatchResult): Promise<{
    product: Product,
    sku: SKU,
    attrs: { attr_id: number; attr_name: string; value: AttributeValue }[]
  } | null> => {
    try {
      const product = await getProductDetail(result.product_id)
      if (!product) return null

      const availSkus = (product.skus || []).filter((s: SKU) => s.status === 1)
      if (availSkus.length === 0) return null

      let sku = availSkus.find((s: SKU) => s.id === result.sku_id)
      if (!sku) {
        sku = availSkus.find((s: SKU) => s.stock > 0) || availSkus[0]
      }
      if (!sku) return null

      const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
      const requiredAttrs = (product.attributes || []).filter((a: Attribute) => a.status === 1)

      for (const attr of requiredAttrs) {
        let matched = false
        for (const av of (attr.values || [])) {
          if (av.status === 1 && (av.stock === -1 || av.stock > 0)) {
            attrs.push({ attr_id: attr.id, attr_name: attr.name, value: av })
            matched = true
            break
          }
        }
        if (!matched && (attr.values || []).length > 0) {
          const fallback = attr.values[0]
          attrs.push({ attr_id: attr.id, attr_name: attr.name, value: fallback })
        }
      }

      return { product, sku, attrs }
    } catch {
      return null
    }
  }

  const addResultToCart = async (
    result: VoiceMatchResult,
    customSku?: SKU,
    customAttrs?: { attr_id: number; attr_name: string; value: AttributeValue }[],
    customQty?: number
  ): Promise<boolean> => {
    try {
      if (customSku && customAttrs) {
        const product = await getProductDetail(result.product_id)
        addItem(product, customSku, customAttrs, customQty || result.quantity)
        return true
      }

      const resolved = await resolveProductDefault(result)
      if (!resolved) return false

      const { product, sku, attrs } = resolved
      const needUserSelect = (product.attributes || []).some(
        (a: Attribute) => a.status === 1 && (a.values || []).length > 1
      )

      if (needUserSelect && !customSku) {
        return false
      }

      addItem(product, sku, attrs, customQty || result.quantity)
      return true
    } catch {
      return false
    }
  }

  const handleAddAllToCart = async () => {
    if (!parseResult || parseResult.items.length === 0) return

    setAddingToCart(true)
    let added = 0
    let pendingSelectIdx = -1

    for (let idx = 0; idx < parseResult.items.length; idx++) {
      const item = parseResult.items[idx]
      const success = await addResultToCart(item)
      if (success) {
        added++
      } else {
        pendingSelectIdx = idx
        break
      }
    }

    setItemsAdded(added)

    if (pendingSelectIdx >= 0) {
      const pendingItem = parseResult.items[pendingSelectIdx]
      try {
        const product = await getProductDetail(pendingItem.product_id)
        setLoadedProduct(product)

        const availSkus = (product.skus || []).filter((s: SKU) => s.status === 1)
        const defaultSku = availSkus.find((s: SKU) => s.id === pendingItem.sku_id)
          || availSkus.find((s: SKU) => s.stock > 0)
          || (availSkus[0] || null)
        setSelectedSku(defaultSku)

        const defaultAttrs = new Map<number, AttributeValue>()
        for (const attr of (product.attributes || []).filter((a: Attribute) => a.status === 1)) {
          for (const av of (attr.values || [])) {
            if (av.status === 1 && (av.stock === -1 || av.stock > 0)) {
              defaultAttrs.set(attr.id, av)
              break
            }
          }
        }
        setSelectedAttrs(defaultAttrs)
        setSelectQuantity(pendingItem.quantity)
        setNeedSelectProduct({ index: pendingSelectIdx, result: pendingItem })
      } catch {
        Toast.show(`商品「${pendingItem.product_name}」加购失败`)
      }
    }

    setAddingToCart(false)

    if (pendingSelectIdx < 0 && added > 0) {
      Toast.show(`已添加 ${added} 道菜到购物车`)
      setTimeout(() => { Taro.navigateBack() }, 1200)
    } else if (pendingSelectIdx < 0 && added === 0) {
      Toast.show('添加失败，请手动点餐')
    }
  }

  const confirmSelectProduct = async () => {
    if (!needSelectProduct || !loadedProduct || !selectedSku) return

    const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
    for (const attr of (loadedProduct.attributes || []).filter((a: Attribute) => a.status === 1)) {
      const val = selectedAttrs.get(attr.id)
      if (val) {
        attrs.push({ attr_id: attr.id, attr_name: attr.name, value: val })
      } else if ((attr.values || []).length > 0) {
        attrs.push({ attr_id: attr.id, attr_name: attr.name, value: attr.values[0] })
      }
    }

    addItem(loadedProduct, selectedSku, attrs, selectQuantity)
    const currentAdded = itemsAdded + 1
    setItemsAdded(currentAdded)
    setNeedSelectProduct(null)
    setLoadedProduct(null)

    const pendingIdx = needSelectProduct.index + 1
    let pendingLater = -1
    let addedInLoop = 0

    for (let idx = pendingIdx; idx < (parseResult?.items.length || 0); idx++) {
      const item = parseResult!.items[idx]
      const success = await addResultToCart(item)
      if (success) {
        addedInLoop++
      } else {
        pendingLater = idx
        break
      }
    }

    const totalAdded = currentAdded + addedInLoop
    setItemsAdded(totalAdded)

    if (pendingLater >= 0) {
      const pendingItem = parseResult!.items[pendingLater]
      try {
        const product = await getProductDetail(pendingItem.product_id)
        setLoadedProduct(product)
        const availSkus = (product.skus || []).filter((s: SKU) => s.status === 1)
        const defaultSku = availSkus.find((s: SKU) => s.id === pendingItem.sku_id)
          || availSkus.find((s: SKU) => s.stock > 0) || (availSkus[0] || null)
        setSelectedSku(defaultSku)
        const defaultAttrs = new Map<number, AttributeValue>()
        for (const attr of (product.attributes || []).filter((a: Attribute) => a.status === 1)) {
          for (const av of (attr.values || [])) {
            if (av.status === 1 && (av.stock === -1 || av.stock > 0)) {
              defaultAttrs.set(attr.id, av)
              break
            }
          }
        }
        setSelectedAttrs(defaultAttrs)
        setSelectQuantity(pendingItem.quantity)
        setNeedSelectProduct({ index: pendingLater, result: pendingItem })
      } catch {
        Toast.show(`商品「${pendingItem.product_name}」加购失败`)
      }
      return
    }

    Toast.show(`已添加 ${totalAdded} 道菜到购物车`)
    setTimeout(() => { Taro.navigateBack() }, 1200)
  }

  const suggestedPhrases = [
    '来份老坛酸菜鱼',
    '两碗白米饭',
    '加一瓶可口可乐',
    '来个宫保鸡丁和麻婆豆腐',
    '三杯珍珠奶茶',
    '要一份红烧肉再来个清炒时蔬'
  ]

  const handleSuggestionClick = (text: string) => {
    setRecognizedText(text)
    handleParse(text)
  }

  const selectPrice = () => {
    if (!selectedSku) return 0
    let price = selectedSku.price || 0
    selectedAttrs.forEach(av => { price += av.extra_price || 0 })
    return price
  }

  return (
    <View className={styles.container}>
      <View className={styles.header}>
        <View className={styles.title}>🎤 语音点餐</View>
        <View className={styles.subtitle}>说出你想吃的，智能识别加入购物车</View>
      </View>

      <ScrollView scrollY className={styles.voicePanel}>
        <View
          className={`${styles.micButton} ${recording ? styles.micButtonRecording : ''}`}
          onTouchStart={handleStartRecord}
          onTouchEnd={handleStopRecord}
          onTouchCancel={handleStopRecord}
          onClick={() => recording ? handleStopRecord() : handleStartRecord()}
        >
          <Text className={styles.micIcon}>{recording ? '🔴' : '🎙️'}</Text>
        </View>

        <Text className={styles.statusText}>
          {recording ? '正在听您说话...' : parsing ? '智能解析中...' : '长按或点击开始说话'}
        </Text>
        <Text className={styles.statusHint}>
          {recording ? '松开结束录音（最长60秒）' : '支持：来份酸菜鱼、两碗米饭、加个可乐'}
        </Text>

        {recognizedText && (
          <View className={styles.textPreview}>
            <View className={styles.previewLabel}>识别结果</View>
            <View className={styles.previewText}>{recognizedText}</View>
          </View>
        )}

        {parseResult && (
          <View className={styles.parseSection}>
            <View className={styles.sectionTitle}>
              匹配结果 {parseResult.items.length > 0 ? `(${parseResult.items.length}道菜)` : ''}
            </View>

            {parseResult.items.length > 0 ? (
              parseResult.items.map((item, idx) => (
                <View key={idx} className={styles.matchCard}>
                  <Image
                    className={styles.matchImage}
                    src={item.image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=delicious%20food%20dish&image_size=square'}
                  />
                  <View className={styles.matchInfo}>
                    <View className={styles.matchName}>{item.product_name}</View>
                    {item.sku_name && <View className={styles.matchSku}>{item.sku_name}</View>}
                    <View className={styles.matchMeta}>
                      <View className={styles.matchPrice}>
                        <Text className={styles.priceSymbol}>¥</Text>
                        {Number(item.price).toFixed(2)}
                      </View>
                      <Text className={styles.matchQty}>× {item.quantity}</Text>
                      {item.match_score < 1 && (
                        <Text className={styles.matchScore}>
                          匹配{Math.round(item.match_score * 100)}%
                        </Text>
                      )}
                    </View>
                  </View>
                </View>
              ))
            ) : (
              <View className={styles.emptyResult}>
                <View className={styles.emptyIcon}>🤔</View>
                <View className={styles.emptyText}>未匹配到菜品，请重试或手动点餐</View>
              </View>
            )}

            {parseResult.unmatched.length > 0 && (
              <View className={styles.unmatchedCard}>
                <View className={styles.unmatchedLabel}>未匹配项</View>
                {parseResult.unmatched.map((name, idx) => (
                  <View key={idx} className={styles.unmatchedItem}>{name}</View>
                ))}
              </View>
            )}
          </View>
        )}

        <View className={styles.historySection}>
          <View className={styles.historyLabel}>试试这样说：</View>
          <View className={styles.historyTags}>
            {suggestedPhrases.map((phrase, idx) => (
              <View
                key={idx}
                className={styles.historyTag}
                onClick={() => handleSuggestionClick(phrase)}
              >
                {phrase}
              </View>
            ))}
          </View>
        </View>
      </ScrollView>

      {parseResult && parseResult.items.length > 0 && (
        <View className={styles.actionBar}>
          <View
            className={`${styles.addToCartBtn} ${addingToCart ? styles.disabled : ''}`}
            onClick={handleAddAllToCart}
          >
            {addingToCart ? '加购中...' : `加入购物车 (${parseResult.items.length}道菜)`}
          </View>
        </View>
      )}

      <Popup
        visible={!!needSelectProduct && !!loadedProduct}
        position='bottom'
        round
        onClose={() => setNeedSelectProduct(null)}
        style={{ height: '80vh' }}
      >
        {needSelectProduct && loadedProduct && (
          <View style={{ padding: '24px' }}>
            <View style={{ fontSize: '32px', fontWeight: '700', marginBottom: '24px' }}>
              请选择规格：{loadedProduct.name}
            </View>
            <ScrollView scrollY style={{ maxHeight: '60vh' }}>
              {(loadedProduct.skus || []).length > 1 && (
                <View style={{ marginBottom: '24px' }}>
                  <View style={{ fontSize: '28px', color: '#333', fontWeight: '600', marginBottom: '12px' }}>规格</View>
                  <View style={{ display: 'flex', flexWrap: 'wrap', gap: '12px' }}>
                    {(loadedProduct.skus || []).filter((s: SKU) => s.status === 1).map((s: SKU) => (
                      <View
                        key={s.id}
                        style={{
                          padding: '12px 24px',
                          borderRadius: '12px',
                          border: `2px solid ${selectedSku?.id === s.id ? '#667eea' : '#eee'}`,
                          background: selectedSku?.id === s.id ? '#eef0ff' : '#f7f8fa',
                          fontSize: '26px',
                          color: selectedSku?.id === s.id ? '#667eea' : '#333'
                        }}
                        onClick={() => setSelectedSku(s)}
                      >
                        {s.spec_name} ¥{s.price.toFixed(2)}
                      </View>
                    ))}
                  </View>
                </View>
              )}

              {(loadedProduct.attributes || []).filter((a: Attribute) => a.status === 1).map((attr: Attribute) => (
                <View key={attr.id} style={{ marginBottom: '24px' }}>
                  <View style={{ fontSize: '28px', color: '#333', fontWeight: '600', marginBottom: '12px' }}>{attr.name}</View>
                  <View style={{ display: 'flex', flexWrap: 'wrap', gap: '12px' }}>
                    {(attr.values || []).filter((v: AttributeValue) => v.status === 1).map((v: AttributeValue) => (
                      <View
                        key={v.id}
                        style={{
                          padding: '12px 24px',
                          borderRadius: '12px',
                          border: `2px solid ${selectedAttrs.get(attr.id)?.id === v.id ? '#667eea' : '#eee'}`,
                          background: selectedAttrs.get(attr.id)?.id === v.id ? '#eef0ff' : '#f7f8fa',
                          fontSize: '26px',
                          color: selectedAttrs.get(attr.id)?.id === v.id ? '#667eea' : '#333'
                        }}
                        onClick={() => {
                          const next = new Map(selectedAttrs)
                          next.set(attr.id, v)
                          setSelectedAttrs(next)
                        }}
                      >
                        {v.value}
                        {v.extra_price > 0 && ` +¥${v.extra_price.toFixed(2)}`}
                      </View>
                    ))}
                  </View>
                </View>
              ))}

              <View style={{ marginBottom: '24px' }}>
                <View style={{ fontSize: '28px', color: '#333', fontWeight: '600', marginBottom: '12px' }}>数量</View>
                <View style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
                  <View
                    style={{
                      width: '60px', height: '60px', borderRadius: '50%',
                      background: '#f2f3f5', fontSize: '32px', fontWeight: '700',
                      display: 'flex', alignItems: 'center', justifyContent: 'center'
                    }}
                    onClick={() => setSelectQuantity(Math.max(1, selectQuantity - 1))}
                  >−</View>
                  <View style={{ fontSize: '32px', fontWeight: '700', minWidth: '60px', textAlign: 'center' }}>
                    {selectQuantity}
                  </View>
                  <View
                    style={{
                      width: '60px', height: '60px', borderRadius: '50%',
                      background: '#f2f3f5', fontSize: '32px', fontWeight: '700',
                      display: 'flex', alignItems: 'center', justifyContent: 'center'
                    }}
                    onClick={() => setSelectQuantity(selectQuantity + 1)}
                  >+</View>
                </View>
              </View>
            </ScrollView>

            <View style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              marginTop: '24px', paddingTop: '24px', borderTop: '1px solid #eee'
            }}>
              <View>
                <Text style={{ fontSize: '24px', color: '#999' }}>合计：</Text>
                <Text style={{ fontSize: '36px', fontWeight: '700', color: '#ff6b6b' }}>
                  ¥{(selectPrice() * selectQuantity).toFixed(2)}
                </Text>
              </View>
              <View
                style={{
                  padding: '0 48px', height: '80px', borderRadius: '40px',
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  color: '#fff', fontSize: '28px', fontWeight: '600',
                  display: 'flex', alignItems: 'center', justifyContent: 'center'
                }}
                onClick={confirmSelectProduct}
              >确认选择并加购</View>
            </View>
          </View>
        )}
      </Popup>
    </View>
  )
}

export default VoiceOrder
