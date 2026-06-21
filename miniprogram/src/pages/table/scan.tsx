import React, { useState, useCallback } from 'react'
import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Button, Loading, Input, Popup } from '@nutui/nutui-react-taro'
import { tableApi } from '../../services/table'
import { waiterApi } from '../../services/waiter'
import type { TableInfo, TableItem } from '../../services/table'
import { useAppStore } from '../../store/app'
import { useCartStore } from '../../store/cart'
import styles from './scan.module.scss'

type PageStatus = 'idle' | 'scanning' | 'loading' | 'success' | 'error' | 'manual'

const TableScan: React.FC = () => {
  const [pageStatus, setPageStatus] = useState<PageStatus>('idle')
  const [tableInfo, setTableInfo] = useState<TableInfo | null>(null)
  const [availableTables, setAvailableTables] = useState<TableItem[]>([])
  const [manualTableNo, setManualTableNo] = useState('')
  const [selectedTableId, setSelectedTableId] = useState<number | null>(null)
  const [errorMessage, setErrorMessage] = useState('')
  const [showCallPopup, setShowCallPopup] = useState(false)
  const [callType, setCallType] = useState('service')
  const [callContent, setCallContent] = useState('')
  const [calling, setCalling] = useState(false)

  const currentStore = useAppStore(state => state.currentStore)
  const setTableNo = useCartStore(state => state.setTableNo)

  const getStatusText = (status: number) => {
    switch (status) {
      case 0:
        return { text: '空闲', className: styles.available }
      case 1:
        return { text: '已有人', className: styles.occupied }
      case 2:
        return { text: '已停用', className: styles.disabled }
      default:
        return { text: '未知', className: styles.disabled }
    }
  }

  const parseScene = (result: string): string => {
    try {
      if (result.includes('scene=')) {
        const urlParams = new URLSearchParams(result.split('?')[1] || result)
        const scene = urlParams.get('scene')
        if (scene) {
          return decodeURIComponent(scene)
        }
      }
      return result
    } catch {
      return result
    }
  }

  const handleScanCode = useCallback(async () => {
    try {
      setPageStatus('scanning')
      const res = await Taro.scanCode({
        onlyFromCamera: false,
        scanType: ['qrCode']
      })
      const scene = parseScene(res.result)
      await handleScanResult(scene)
    } catch (err: any) {
      if (err.errMsg?.includes('cancel')) {
        setPageStatus('idle')
      } else {
        setErrorMessage('扫码失败，请重试')
        setPageStatus('error')
      }
    }
  }, [])

  const handleScanResult = async (scene: string) => {
    setPageStatus('loading')
    setErrorMessage('')
    try {
      const info = await tableApi.scanQRCode(scene)
      setTableInfo(info)
      setPageStatus('success')
      if (info.message) {
        Taro.showToast({ title: info.message, icon: 'none' })
      }
    } catch (err: any) {
      setErrorMessage(err.message || '扫码解析失败，请重试')
      setPageStatus('error')
    }
  }

  const loadAvailableTables = async () => {
    if (!currentStore) return
    setPageStatus('loading')
    try {
      const tables = await tableApi.getAvailableTables(currentStore.id)
      setAvailableTables(tables)
      setIsManualMode(true)
      setIsManualMode(true)
      setPageStatus('manual')
    } catch {
      setErrorMessage('加载桌位列表失败，请重试')
      setPageStatus('error')
    }
  }

  const handleManualInput = async () => {
    if (!manualTableNo.trim()) {
      Taro.showToast({ title: '请输入桌号', icon: 'none' })
      return
    }
    await handleScanResult(manualTableNo.trim())
  }

  const handleTableSelect = (table: TableItem) => {
    setSelectedTableId(table.id)
    setManualTableNo(table.tableNo)
  }

  const handleConfirmManualSelect = async () => {
    if (!selectedTableId) {
      Taro.showToast({ title: '请选择桌号', icon: 'none' })
      return
    }
    const table = availableTables.find(t => t.id === selectedTableId)
    if (!table) return

    const info: TableInfo = {
      id: table.id,
      storeId: table.storeId,
      storeName: currentStore?.name || '',
      tableNo: table.tableNo,
      tableType: table.type,
      capacity: table.capacity,
      area: table.area,
      floor: table.floor,
      status: table.status,
      message: ''
    }
    setTableInfo(info)
    setPageStatus('success')
  }

  const handleStartOrder = () => {
    if (tableInfo) {
      setTableNo(tableInfo.tableNo)
      Taro.showToast({ title: '已入座', icon: 'success' })
      setTimeout(() => {
        Taro.switchTab({ url: '/pages/index/index' })
      }, 1000)
    }
  }

  const handleJoinOrder = () => {
    if (tableInfo) {
      setTableNo(tableInfo.tableNo)
      Taro.showToast({ title: '已加入订单', icon: 'success' })
      setTimeout(() => {
        Taro.switchTab({ url: '/pages/index/index' })
      }, 1000)
    }
  }

  const handleRetry = () => {
    setErrorMessage('')
    setPageStatus('idle')
    setTableInfo(null)
    setAvailableTables([])
    setSelectedTableId(null)
    setManualTableNo('')
  }

  const handleOpenCallPopup = () => {
    setCallType('service')
    setCallContent('')
    setShowCallPopup(true)
  }

  const handleCallWaiter = async () => {
    if (!tableInfo || !currentStore) return

    setCalling(true)
    try {
      await waiterApi.callWaiter({
        store_id: tableInfo.storeId || currentStore.id,
        table_id: tableInfo.id,
        table_no: tableInfo.tableNo,
        call_type: callType,
        content: callContent
      })
      Taro.showToast({ title: '呼叫已发送，请稍候', icon: 'success' })
      setShowCallPopup(false)
    } catch (err) {
      console.error('Call waiter failed:', err)
    } finally {
      setCalling(false)
    }
  }

  const callTypeOptions = [
    { value: 'service', label: '呼叫服务', icon: '🔔' },
    { value: 'water', label: '需要加水', icon: '💧' },
    { value: 'pay', label: '需要结账', icon: '💰' },
    { value: 'other', label: '其他需求', icon: '💬' }
  ]

  useDidShow(() => {
    if (pageStatus === 'idle') {
      handleScanCode()
    }
  })

  if (pageStatus === 'scanning') {
    return (
      <View className={styles.container}>
      </View>
    )
  }

  if (pageStatus === 'loading') {
    return (
      <View className={styles.loading}>
        <Loading type='spinner' />
        <Text className={styles.loadingText}>加载中...</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <View className={styles.header}>
        <Text className={styles.title}>扫码入座</Text>
        <Text className={styles.subtitle}>扫描桌位二维码快速入座点餐</Text>
      </View>

      <View className={styles.content}>
        {pageStatus === 'idle' && (
          <View className={styles.loading}>
            <Image
              className={styles.scanIcon}
              src='https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=QR%20code%20scan%20icon%20purple&image_size=square'
            />
            <Text className={styles.scanTip}>请扫描桌位二维码</Text>
            <Button
              className={styles.actionBtn}
              type='primary'
              onClick={handleScanCode}
            >
              点击扫码
            </Button>
            <Button
              className={styles.secondaryBtn}
              onClick={loadAvailableTables}
            >
              手动输入桌号
            </Button>
          </View>
        )}

        {pageStatus === 'error' && (
          <View>
          <View className={styles.errorBox}>
            <Text className={styles.errorText}>{errorMessage}</Text>
            <Button
              className={styles.retryBtn}
              type='primary'
              onClick={handleRetry}
            >
              重新扫码
            </Button>
          </View>
          <Button
            className={styles.secondaryBtn}
            onClick={loadAvailableTables}
            block
          >
            手动输入桌号
          </Button>
        </View>
        )}

        {pageStatus === 'success' && tableInfo && (
          <View>
            <View className={styles.tableCard}>
              <View className={styles.tableHeader}>
                <Text className={styles.storeName}>{tableInfo.storeName}</Text>
                <View className={`${styles.statusTag} ${getStatusText(tableInfo.status).className}`}>
                  {getStatusText(tableInfo.status).text}
                </View>
              </View>

              <View className={styles.tableInfo}>
                <View className={styles.infoRow}>
                  <Text className={styles.infoLabel}>桌号</Text>
                  <Text className={styles.tableNo}>{tableInfo.tableNo}</Text>
                </View>
                <View className={styles.infoRow}>
                  <Text className={styles.infoLabel}>容纳人数</Text>
                  <View className={styles.capacityBadge}>
                    <Text>👥 {tableInfo.capacity}人</Text>
                  </View>
                </View>
                <View className={styles.infoRow}>
                  <Text className={styles.infoLabel}>楼层</Text>
                  <Text className={styles.infoValue}>{tableInfo.floor}楼</Text>
                </View>
                <View className={styles.infoRow}>
                  <Text className={styles.infoLabel}>区域</Text>
                  <Text className={styles.infoValue}>{tableInfo.area}</Text>
                </View>
              </View>
            </View>

            <View className={styles.actionArea}>
              {tableInfo.status === 0 && (
                <Button
                  className={styles.primaryBtn}
                  type='primary'
                  onClick={handleStartOrder}
                  block
                >
                  开始点餐
                </Button>
              )}
              {tableInfo.status === 1 && (
                <Button
                  className={styles.primaryBtn}
                  type='primary'
                  onClick={handleJoinOrder}
                  block
                >
                  加入订单
                </Button>
              )}
              {tableInfo.status === 2 && (
                  <Text className={styles.loadingText}>该桌位已停用，请选择其他桌位</Text>
                )}

              <View className={styles.callWaiterBtn} onClick={handleOpenCallPopup}>
                <Text className={styles.callIcon}>🔔</Text>
                <Text className={styles.callText}>呼叫服务员</Text>
              </View>

              <Button
                className={styles.secondaryBtn}
                onClick={handleRetry}
                block
              >
                重新扫码
              </Button>
            </View>
          </View>
        )}

        {pageStatus === 'manual' && (
          <View className={styles.manualInputSection}>
            <Text className={styles.sectionTitle}>手动选择桌号</Text>

            <View className={styles.inputWrap}>
              <Input
                placeholder='请输入桌号'
                value={manualTableNo}
                onChange={setManualTableNo}
                clearable
              />
            </View>

            <Button
              className={styles.primaryBtn}
              type='primary'
              onClick={handleManualInput}
              block
            >
              确认桌号
            </Button>

            <View className={styles.divider}>
              <Text className={styles.dividerText}>或选择空闲桌位</Text>
            </View>

            {availableTables.length > 0 && (
              <View className={styles.tableList}>
                {availableTables.map(table => (
                  <View
                    key={table.id}
                    className={`${styles.tableItem} ${selectedTableId === table.id ? styles.active : ''} ${table.status !== 0 ? styles.disabled : ''}`}
                    onClick={() => handleTableSelect(table)}
                  >
                    <Text className={styles.tableItemNo}>{table.tableNo}</Text>
                    <Text className={styles.tableItemCapacity}>{table.capacity}人</Text>
                  </View>
                ))}
              </View>
            )}

            {availableTables.length > 0 && (
              <Button
                className={styles.primaryBtn}
                type='primary'
                onClick={handleConfirmManualSelect}
                block
                disabled={!selectedTableId}
              >
                确认选择
              </Button>
            )}

            {availableTables.length === 0 && (
              <View className={styles.empty}>
                <Text className={styles.emptyText}>暂无空闲桌位</Text>
              </View>
            )}

            <Button
              className={styles.secondaryBtn}
              onClick={handleRetry}
              block
              style={{ marginTop: '24rpx' }}
            >
              返回扫码
            </Button>
          </View>
        )}
      </View>

      <View className={styles.footer}>
        <Text className={styles.footerText}>扫码遇到问题？请联系服务员</Text>
      </View>

      <Popup
        visible={showCallPopup}
        position='bottom'
        onClose={() => setShowCallPopup(false)}
        round
      >
        <View className={styles.callPopup}>
          <View className={styles.callPopupHeader}>
            <Text className={styles.callPopupTitle}>呼叫服务员</Text>
            <View className={styles.callPopupClose} onClick={() => setShowCallPopup(false)}>
              <Text>✕</Text>
            </View>
          </View>

          <View className={styles.callTypeGrid}>
            {callTypeOptions.map(option => (
              <View
                key={option.value}
                className={`${styles.callTypeItem} ${callType === option.value ? styles.callTypeActive : ''}`}
                onClick={() => setCallType(option.value)}
              >
                <Text className={styles.callTypeIcon}>{option.icon}</Text>
                <Text className={styles.callTypeLabel}>{option.label}</Text>
              </View>
            ))}
          </View>

          <View className={styles.callContentInput}>
            <Input
              placeholder='补充说明（选填）'
              value={callContent}
              onChange={setCallContent}
              maxLength={200}
            />
          </View>

          <Button
            className={styles.callConfirmBtn}
            type='primary'
            block
            loading={calling}
            disabled={calling}
            onClick={handleCallWaiter}
          >
            {calling ? '呼叫中...' : '确认呼叫'}
          </Button>
        </View>
      </Popup>
    </View>
  )
}

export default TableScan
