<template>
  <div class="store-map-page">
    <div class="page-header">
      <h2 class="page-title">门店地图</h2>
      <div class="header-actions">
        <div class="legend">
          <span class="legend-item">
            <span class="legend-dot open"></span>
            <span>营业中</span>
          </span>
          <span class="legend-item">
            <span class="legend-dot closed"></span>
            <span>停业</span>
          </span>
        </div>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索门店名称"
          clearable
          style="width: 240px"
          :prefix-icon="Search"
          @keyup.enter="handleSearch"
          @clear="handleSearch" />
        <el-select v-model="statusFilter" placeholder="门店状态" style="width: 140px" @change="handleSearch">
          <el-option label="全部" :value="null" />
          <el-option label="营业中" :value="1" />
          <el-option label="停业" :value="0" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <div class="map-container">
        <div class="map-left">
          <div ref="chartRef" class="chart"></div>
        </div>
        <div class="map-right">
          <div class="store-list-header">
            <span>门店列表</span>
            <span class="store-count">共 {{ filteredStores.length }} 家</span>
          </div>
          <div class="store-list" v-loading="loading">
            <div
              v-for="store in filteredStores"
              :key="store.id"
              class="store-item"
              :class="{ active: selectedStore?.id === store.id }"
              @click="handleStoreClick(store)">
              <div class="store-status" :class="store.status === 1 ? 'open' : 'closed'">
                {{ store.status === 1 ? '营业中' : '停业' }}
              </div>
              <div class="store-info">
                <div class="store-name">{{ store.name }}</div>
                <div class="store-address">
                  <el-icon><Location /></el-icon>
                  {{ store.address }}
                </div>
                <div class="store-phone">
                  <el-icon><Phone /></el-icon>
                  {{ store.phone }}
                </div>
              </div>
            </div>
            <el-empty v-if="!loading && filteredStores.length === 0" description="暂无门店数据" />
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="detailVisible"
      title="门店详情"
      width="500px">
      <div v-if="selectedStore" class="store-detail">
        <div class="detail-header">
          <div class="detail-status" :class="selectedStore.status === 1 ? 'open' : 'closed'">
            {{ selectedStore.status === 1 ? '营业中' : '已停业' }}
          </div>
          <h3>{{ selectedStore.name }}</h3>
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="门店编码">
            {{ selectedStore.code }}
          </el-descriptions-item>
          <el-descriptions-item label="门店地址">
            <el-icon><Location /></el-icon>
            {{ selectedStore.address }}
          </el-descriptions-item>
          <el-descriptions-item label="联系电话">
            <el-icon><Phone /></el-icon>
            {{ selectedStore.phone }}
          </el-descriptions-item>
          <el-descriptions-item label="联系人">
            {{ selectedStore.contact }}
          </el-descriptions-item>
          <el-descriptions-item label="营业时间">
            <el-icon><Clock /></el-icon>
            {{ selectedStore.business_hours || (selectedStore.open_time + '-' + selectedStore.close_time) }}
          </el-descriptions-item>
          <el-descriptions-item label="桌位数">
            <el-icon><Grid /></el-icon>
            {{ selectedStore.table_count || 0 }} 桌
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import { Search, Refresh, Location, Phone, Clock, Grid } from '@element-plus/icons-vue'
import { storeMapApi } from '@/api/stores'

const loading = ref(false)
const chartRef = ref(null)
const chartInstance = ref(null)
const stores = ref([])
const detailVisible = ref(false)
const selectedStore = ref(null)

const searchKeyword = ref('')
const statusFilter = ref(null)

const mockStores = [
  { id: 1, name: '中心店总店', code: 'ST001', address: '北京市朝阳区建国路88号', phone: '010-88888888', contact: '张经理', status: 1, open_time: '08:00', close_time: '22:00', business_hours: '08:00-22:00', table_count: 30, lng: 116.46, lat: 39.92, x: 460, y: 320 },
  { id: 2, name: '海淀分店', code: 'ST002', address: '北京市海淀区中关村大街1号', phone: '010-66666666', contact: '李经理', status: 1, open_time: '09:00', close_time: '21:00', business_hours: '09:00-21:00', table_count: 20, lng: 116.31, lat: 39.98, x: 310, y: 380 },
  { id: 3, name: '西城分店', code: 'ST003', address: '北京市西城区金融街15号', phone: '010-77777777', contact: '王经理', status: 0, open_time: '08:30', close_time: '21:30', business_hours: '08:30-21:30', table_count: 25, lng: 116.36, lat: 39.91, x: 360, y: 310 },
  { id: 4, name: '东城分店', code: 'ST004', address: '北京市东城区王府井大街88号', phone: '010-55555555', contact: '赵经理', status: 1, open_time: '10:00', close_time: '22:00', business_hours: '10:00-22:00', table_count: 18, lng: 116.42, lat: 39.91, x: 420, y: 310 },
  { id: 5, name: '朝阳分店', code: 'ST005', address: '北京市朝阳区三里屯路19号', phone: '010-99999999', contact: '刘经理', status: 1, open_time: '11:00', close_time: '23:00', business_hours: '11:00-23:00', table_count: 35, lng: 116.45, lat: 39.94, x: 450, y: 340 },
  { id: 6, name: '丰台分店', code: 'ST006', address: '北京市丰台区丰台路5号', phone: '010-44444444', contact: '陈经理', status: 0, open_time: '08:00', close_time: '20:00', business_hours: '08:00-20:00', table_count: 15, lng: 116.28, lat: 39.85, x: 280, y: 250 },
  { id: 7, name: '通州分店', code: 'ST007', address: '北京市通州区新华大街100号', phone: '010-33333333', contact: '孙经理', status: 1, open_time: '09:00', close_time: '21:00', business_hours: '09:00-21:00', table_count: 22, lng: 116.66, lat: 39.91, x: 660, y: 310 },
  { id: 8, name: '昌平分店', code: 'ST008', address: '北京市昌平区回龙观大街10号', phone: '010-22222222', contact: '周经理', status: 1, open_time: '08:00', close_time: '22:00', business_hours: '08:00-22:00', table_count: 28, lng: 116.34, lat: 40.06, x: 340, y: 460 }
]

const filteredStores = computed(() => {
  let result = stores.value
  if (searchKeyword.value) {
    result = result.filter(s => s.name.includes(searchKeyword.value))
  }
  if (statusFilter.value !== null) {
    result = result.filter(s => s.status === statusFilter.value)
  }
  return result
})

function handleSearch() {
  updateChart()
}

function resetQuery() {
  searchKeyword.value = ''
  statusFilter.value = null
  updateChart()
}

function handleStoreClick(store) {
  selectedStore.value = store
  detailVisible.value = true
  highlightStoreOnMap(store)
}

function highlightStoreOnMap(store) {
  if (!chartInstance.value) return
  chartInstance.value.dispatchAction({
    type: 'highlight',
    seriesIndex: 0,
    dataIndex: filteredStores.value.findIndex(s => s.id === store.id)
  })
  chartInstance.value.dispatchAction({
    type: 'showTip',
    seriesIndex: 0,
    dataIndex: filteredStores.value.findIndex(s => s.id === store.id)
  })
}

function initChart() {
  if (!chartRef.value) return

  chartInstance.value = echarts.init(chartRef.value)

  const option = {
    backgroundColor: '#f5f7fa',
    tooltip: {
      trigger: 'item',
      formatter: function(params) {
        const data = params.data
        return `
          <div style="padding: 8px;">
            <div style="font-weight: bold; margin-bottom: 4px;">${data.name}</div>
            <div style="font-size: 12px; color: #666;">地址: ${data.address}</div>
            <div style="font-size: 12px; color: #666;">电话: ${data.phone}</div>
            <div style="font-size: 12px; color: ${data.status === 1 ? '#67c23a' : '#909399'}; margin-top: 4px;">
              状态: ${data.status === 1 ? '营业中' : '停业'}
            </div>
          </div>
        `
      }
    },
    grid: {
      left: 40,
      right: 40,
      top: 40,
      bottom: 40
    },
    xAxis: {
      show: false,
      min: 0,
      max: 800
    },
    yAxis: {
      show: false,
      min: 0,
      max: 600
    },
    series: [
      {
        name: '地图背景',
        type: 'custom',
        renderItem: function(params, api) {
          return {
            type: 'group',
            children: [
              {
                type: 'rect',
                shape: {
                  x: 50,
                  y: 150,
                  width: 700,
                  height: 380
                },
                style: {
                  fill: '#e8f4fd',
                  stroke: '#409eff',
                  lineWidth: 2,
                  opacity: 0.6
                }
              },
              {
                type: 'rect',
                shape: {
                  x: 180,
                  y: 180,
                  width: 180,
                  height: 120
                },
                style: {
                  fill: '#d4e8fc',
                  stroke: '#409eff',
                  lineWidth: 1
                }
              },
              {
                type: 'rect',
                shape: {
                  x: 380,
                  y: 180,
                  width: 180,
                  height: 120
                },
                style: {
                  fill: '#d4e8fc',
                  stroke: '#409eff',
                  lineWidth: 1
                }
              },
              {
                type: 'rect',
                shape: {
                  x: 280,
                  y: 310,
                  width: 200,
                  height: 130
                },
                style: {
                  fill: '#d4e8fc',
                  stroke: '#409eff',
                  lineWidth: 1
                }
              },
              {
                type: 'text',
                style: {
                  text: '海淀区',
                  x: 250,
                  y: 240,
                  fontSize: 14,
                  fill: '#666',
                  textAlign: 'center'
                }
              },
              {
                type: 'text',
                style: {
                  text: '朝阳区',
                  x: 470,
                  y: 240,
                  fontSize: 14,
                  fill: '#666',
                  textAlign: 'center'
                }
              },
              {
                type: 'text',
                style: {
                  text: '市中心',
                  x: 380,
                  y: 375,
                  fontSize: 14,
                  fill: '#666',
                  textAlign: 'center'
                }
              },
              {
                type: 'text',
                style: {
                  text: '丰台区',
                  x: 280,
                  y: 280,
                  fontSize: 12,
                  fill: '#999',
                  textAlign: 'center'
                }
              },
              {
                type: 'text',
                style: {
                  text: '通州区',
                  x: 660,
                  y: 280,
                  fontSize: 12,
                  fill: '#999',
                  textAlign: 'center'
                }
              },
              {
                type: 'text',
                style: {
                  text: '昌平区',
                  x: 340,
                  y: 500,
                  fontSize: 12,
                  fill: '#999',
                  textAlign: 'center'
                }
              }
            ]
          }
        },
        data: [0]
      },
      {
        name: '门店分布',
        type: 'scatter',
        symbolSize: function(val) {
          return 14 + (val[2] || 0) * 0.3
        },
        data: [],
        label: {
          show: true,
          formatter: function(params) {
            return params.data.name
          },
          position: 'top',
          fontSize: 11,
          color: '#333',
          fontWeight: 500
        },
        itemStyle: {
          color: function(params) {
            return params.data.status === 1 ? '#67c23a' : '#909399'
          },
          shadowBlur: 8,
          shadowColor: 'rgba(0, 0, 0, 0.3)',
          borderColor: '#fff',
          borderWidth: 2
        },
        emphasis: {
          label: {
            show: true,
            fontWeight: 'bold',
            fontSize: 12
          },
          itemStyle: {
            shadowBlur: 15,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
            borderWidth: 3
          }
        }
      },
      {
        name: '营业中门店',
        type: 'effectScatter',
        symbolSize: function(val) {
          return 10 + (val[2] || 0) * 0.2
        },
        data: [],
        rippleEffect: {
          brushType: 'stroke',
          scale: 3,
          period: 4
        },
        itemStyle: {
          color: '#67c23a',
          opacity: 0.7
        },
        label: {
          show: false
        }
      }
    ]
  }

  chartInstance.value.setOption(option)

  chartInstance.value.on('click', function(params) {
    if (params.componentType === 'series' && params.seriesName === '门店分布') {
      const store = filteredStores.value.find(s => s.id === params.data.id)
      if (store) {
        handleStoreClick(store)
      }
    }
  })

  window.addEventListener('resize', handleResize)
}

function updateChart() {
  if (!chartInstance.value) return

  const scatterData = filteredStores.value.map(store => ({
    id: store.id,
    name: store.name,
    value: [store.x, store.y, store.table_count],
    status: store.status,
    address: store.address,
    phone: store.phone
  }))

  const effectData = filteredStores.value
    .filter(s => s.status === 1)
    .map(store => ({
      id: store.id,
      name: store.name,
      value: [store.x, store.y, store.table_count],
      status: store.status
    }))

  chartInstance.value.setOption({
    series: [
      {},
      {
        data: scatterData
      },
      {
        data: effectData
      }
    ]
  })
}

function handleResize() {
  chartInstance.value?.resize()
}

async function fetchStores() {
  loading.value = true
  try {
    const res = await storeMapApi.getStores()
    stores.value = res.list || res.data || mockStores
    if (!stores.value || stores.value.length === 0) {
      stores.value = mockStores
    }
    stores.value = stores.value.map((store, index) => ({
      ...mockStores[index % mockStores.length],
      ...store
    }))
    await nextTick()
    if (!chartInstance.value) {
      initChart()
    }
    updateChart()
  } catch (e) {
    console.error(e)
    stores.value = mockStores
    await nextTick()
    if (!chartInstance.value) {
      initChart()
    }
    updateChart()
  } finally {
    loading.value = false
  }
}

watch(filteredStores, () => {
  updateChart()
})

onMounted(() => {
  fetchStores()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance.value?.dispose()
})
</script>

<style scoped lang="scss">
.store-map-page {
  .header-actions {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .legend {
    display: flex;
    gap: 20px;

    .legend-item {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 14px;
      color: #606266;
    }

    .legend-dot {
      width: 12px;
      height: 12px;
      border-radius: 50%;

      &.open {
        background: #67c23a;
        box-shadow: 0 0 6px rgba(103, 194, 58, 0.6);
      }

      &.closed {
        background: #909399;
        box-shadow: 0 0 6px rgba(144, 147, 153, 0.6);
      }
    }
  }

  .map-container {
    display: flex;
    gap: 20px;
    height: calc(100vh - 280px);
    min-height: 500px;

    .map-left {
      flex: 1;
      background: #fff;
      border: 1px solid #ebeef5;
      border-radius: 8px;
      overflow: hidden;

      .chart {
        width: 100%;
        height: 100%;
      }
    }

    .map-right {
      width: 360px;
      display: flex;
      flex-direction: column;
      background: #fff;
      border: 1px solid #ebeef5;
      border-radius: 8px;
      overflow: hidden;

      .store-list-header {
        padding: 16px 20px;
        border-bottom: 1px solid #ebeef5;
        display: flex;
        justify-content: space-between;
        align-items: center;

        span:first-child {
          font-size: 16px;
          font-weight: 600;
          color: #303133;
        }

        .store-count {
          font-size: 13px;
          color: #909399;
        }
      }

      .store-list {
        flex: 1;
        overflow-y: auto;
        padding: 8px;

        .store-item {
          padding: 12px;
          margin-bottom: 8px;
          border: 1px solid #ebeef5;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            border-color: #409eff;
            background: #f5f7fa;
          }

          &.active {
            border-color: #409eff;
            background: #ecf5ff;
          }

          .store-status {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 12px;
            margin-bottom: 8px;

            &.open {
              background: #f0f9eb;
              color: #67c23a;
            }

            &.closed {
              background: #f4f4f5;
              color: #909399;
            }
          }

          .store-info {
            .store-name {
              font-size: 14px;
              font-weight: 600;
              color: #303133;
              margin-bottom: 6px;
            }

            .store-address,
            .store-phone {
              font-size: 12px;
              color: #606266;
              display: flex;
              align-items: center;
              gap: 4px;
              margin-bottom: 4px;

              .el-icon {
                color: #c0c4cc;
              }
            }
          }
        }
      }
    }
  }

  .store-detail {
    .detail-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 20px;
      padding-bottom: 16px;
      border-bottom: 1px solid #ebeef5;

      h3 {
        margin: 0;
        font-size: 18px;
        color: #303133;
      }

      .detail-status {
        padding: 4px 12px;
        border-radius: 12px;
        font-size: 13px;

        &.open {
          background: #f0f9eb;
          color: #67c23a;
        }

        &.closed {
          background: #f4f4f5;
          color: #909399;
        }
      }
    }

    .el-descriptions {
      :deep(.el-descriptions__label) {
        width: 100px;
      }

      :deep(.el-descriptions__content) {
        display: flex;
        align-items: center;
        gap: 4px;

        .el-icon {
          color: #409eff;
        }
      }
    }
  }
}
</style>
