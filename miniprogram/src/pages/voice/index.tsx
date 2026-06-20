import React, { useState, useRef, useCallback } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Toast } from '@nutui/nutui-react-taro'
import { parseVoiceText, VoiceMatchResult, VoiceParseResponse } from '../../services/voice'
import { useAppStore } from '../../store/app'
import { useCartStore } from '../../store/cart'
import type { Product, SKU, AttributeValue } from '../../services/product'
import { getProductDetail } from '../../services/product'
import styles from './index.module.scss'

declare const wx: any

const VoiceOrder: React.FC = () => {
  const currentStore = useAppStore(state => state.currentStore)
  const addItem = useCartStore(state => state.addItem)
  const [recording, setRecording] = useState(false)
  const [recognizedText, setRecognizedText] = useState('')
  const [parseResult, setParseResult] = useState<VoiceParseResponse | null>(null)
  const [parsing, setParsing] = useState(false)
  const [addingToCart, setAddingToCart] = useState(false)
  const pluginRef = useRef<any>(null)

  const initPlugin = useCallback(() => {
    try {
      const plugin = requirePlugin('WechatSI')
      pluginRef.current = plugin
      return plugin
    } catch {
      return null
    }
  }, [])

  const requirePlugin = (name: string) => {
    try {
      return Taro.requirePlugin(name)
    } catch {
      return null
    }
  }

  const handleStartRecord = async () => {
    if (!currentStore) {
      Toast.show('请先选择门店')
      return
    }

    setRecording(true)
    setRecognizedText('')
    setParseResult(null)

    try {
      const plugin = initPlugin()
      if (plugin && plugin.startRecord) {
        const manager = plugin.getRecordRecognitionManager()
        manager.onStart(() => {})
        manager.onRecognize((res: any) => {
          const text = res?.result || ''
          if (text) {
            setRecognizedText(text)
          }
        })
        manager.onStop((res: any) => {
          const text = res?.result || ''
          if (text) {
            setRecognizedText(text)
            handleParse(text)
          }
          setRecording(false)
        })
        manager.onError((err: any) => {
          console.error('语音识别错误', err)
          setRecording(false)
          fallbackToWxRecord()
        })
        manager.start({ lang: 'zh_CN' })
      } else {
        fallbackToWxRecord()
      }
    } catch {
      fallbackToWxRecord()
    }
  }

  const fallbackToWxRecord = () => {
    Taro.startRecord({
      success: () => {
        Toast.show('正在录音...')
      },
      fail: () => {
        setRecording(false)
        Toast.show('录音失败，请检查权限')
      }
    })

    setTimeout(() => {
      Taro.stopRecord({
        success: (res: any) => {
          setRecording(false)
          const tempFilePath = res.tempFilePath
          translateVoice(tempFilePath)
        },
        fail: () => {
          setRecording(false)
        }
      })
    }, 5000)
  }

  const translateVoice = (filePath: string) => {
    try {
      const plugin = initPlugin()
      if (plugin && plugin.translateVoice) {
        plugin.translateVoice({
          lfrom: 'zh_CN',
          lto: 'zh_CN',
          content: filePath,
          success: (res: any) => {
            const text = res?.result || res?.translateResult || ''
            if (text) {
              setRecognizedText(text)
              handleParse(text)
            } else {
              Toast.show('未能识别语音内容')
            }
          },
          fail: () => {
            Toast.show('语音识别失败，请手动输入')
          }
        })
      } else {
        Toast.show('语音识别插件未加载，请手动输入')
      }
    } catch {
      Toast.show('语音识别失败')
    }
  }

  const handleStopRecord = () => {
    setRecording(false)
    try {
      const plugin = pluginRef.current
      if (plugin) {
        const manager = plugin.getRecordRecognitionManager()
        if (manager && manager.stop) {
          manager.stop()
          return
        }
      }
    } catch {}
    Taro.stopRecord({})
  }

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

  const handleAddAllToCart = async () => {
    if (!parseResult || parseResult.items.length === 0) return

    setAddingToCart(true)
    let addedCount = 0

    for (const item of parseResult.items) {
      try {
        const product = await getProductDetail(item.product_id)
        const sku = product.skus?.find((s: SKU) => s.id === item.sku_id)
        if (!sku) {
          const defaultSku = product.skus?.find((s: SKU) => s.status === 1 && s.stock > 0)
          if (!defaultSku) continue
        }
        const matchSku = sku || product.skus?.find((s: SKU) => s.status === 1 && s.stock > 0)
        if (!matchSku) continue

        const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
        addItem(product, matchSku, attrs, item.quantity)
        addedCount++
      } catch {}
    }

    setAddingToCart(false)
    if (addedCount > 0) {
      Toast.show(`已添加 ${addedCount} 道菜到购物车`)
      setTimeout(() => {
        Taro.navigateBack()
      }, 1200)
    } else {
      Toast.show('添加失败，请手动点餐')
    }
  }

  const suggestedPhrases = [
    '来份酸菜鱼',
    '两碗米饭',
    '加一个可乐',
    '来个宫保鸡丁和麻婆豆腐',
    '三杯奶茶',
    '要一份红烧肉'
  ]

  const handleSuggestionClick = (text: string) => {
    setRecognizedText(text)
    handleParse(text)
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
          onClick={!recording ? handleStartRecord : handleStopRecord}
        >
          <Text className={styles.micIcon}>{recording ? '🔴' : '🎙️'}</Text>
        </View>

        <Text className={styles.statusText}>
          {recording ? '正在听您说话...' : parsing ? '正在解析...' : '点击开始说话'}
        </Text>
        <Text className={styles.statusHint}>
          {recording ? '松开结束录音' : '支持：来份酸菜鱼、两碗米饭、加个可乐'}
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
            {addingToCart ? '添加中...' : `加入购物车 (${parseResult.items.length}道菜)`}
          </View>
        </View>
      )}
    </View>
  )
}

export default VoiceOrder
