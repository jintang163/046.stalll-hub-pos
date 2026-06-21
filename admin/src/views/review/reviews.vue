<template>
  <div class="reviews-page">
    <div class="page-header">
      <h2 class="page-title">评价管理</h2>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-select
          v-model="query.store_id"
          placeholder="选择门店"
          clearable
          style="width: 160px">
          <el-option
            v-for="store in storeList"
            :key="store.id"
            :label="store.name"
            :value="store.id" />
        </el-select>
        <el-select v-model="query.platform" placeholder="平台" clearable style="width: 140px">
          <el-option label="美团" value="meituan" />
          <el-option label="饿了么" value="eleme" />
          <el-option label="大众点评" value="dianping" />
          <el-option label="抖音" value="douyin" />
          <el-option label="小程序" value="miniprogram" />
        </el-select>
        <el-select v-model="query.rating_min" placeholder="最低评分" clearable style="width: 110px">
          <el-option label="1星" :value="1" />
          <el-option label="2星" :value="2" />
          <el-option label="3星" :value="3" />
          <el-option label="4星" :value="4" />
          <el-option label="5星" :value="5" />
        </el-select>
        <el-select v-model="query.rating_max" placeholder="最高评分" clearable style="width: 110px">
          <el-option label="1星" :value="1" />
          <el-option label="2星" :value="2" />
          <el-option label="3星" :value="3" />
          <el-option label="4星" :value="4" />
          <el-option label="5星" :value="5" />
        </el-select>
        <el-select v-model="query.is_bad" placeholder="是否差评" clearable style="width: 120px">
          <el-option label="是" :value="1" />
          <el-option label="否" :value="0" />
        </el-select>
        <el-select v-model="query.is_replied" placeholder="是否已回复" clearable style="width: 120px">
          <el-option label="已回复" :value="1" />
          <el-option label="未回复" :value="0" />
        </el-select>
        <el-input
          v-model="query.keyword"
          placeholder="搜索评价内容/用户昵称"
          clearable
          style="width: 220px"
          @keyup.enter="fetchList" />
        <el-date-picker
          v-model="query.time_range"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          style="width: 320px" />
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="平台" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="getPlatformTagType(row.platform)">
              {{ getPlatformName(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user_nickname" label="用户昵称" width="140" show-overflow-tooltip />
        <el-table-column label="评分" width="120" align="center">
          <template #default="{ row }">
            <el-rate v-model="row.rating" disabled show-score text-color="#ff9900" />
          </template>
        </el-table-column>
        <el-table-column prop="content" label="评价内容" min-width="240" show-overflow-tooltip />
        <el-table-column label="是否差评" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_bad ? 'danger' : 'success'" size="small">
              {{ row.is_bad ? '差评' : '好评' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="是否已回复" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_replied ? 'success' : 'warning'" size="small">
              {{ row.is_replied ? '已回复' : '未回复' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="评价时间" width="170" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">查看详情</el-button>
            <el-button type="success" link size="small" @click="handleReply(row)">回复</el-button>
            <el-button type="warning" link size="small" @click="handleCreateWorkOrder(row)">创建工单</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList" />
      </div>
    </div>

    <el-dialog
      v-model="detailVisible"
      title="评价详情"
      width="700px"
      :close-on-click-modal="false">
      <div v-if="currentReview" class="review-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentReview.id }}</el-descriptions-item>
          <el-descriptions-item label="平台">
            <el-tag size="small" :type="getPlatformTagType(currentReview.platform)">
              {{ getPlatformName(currentReview.platform) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用户昵称">{{ currentReview.user_nickname }}</el-descriptions-item>
          <el-descriptions-item label="评分">
            <el-rate v-model="currentReview.rating" disabled show-score text-color="#ff9900" />
          </el-descriptions-item>
          <el-descriptions-item label="是否差评">
            <el-tag :type="currentReview.is_bad ? 'danger' : 'success'" size="small">
              {{ currentReview.is_bad ? '差评' : '好评' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否已回复">
            <el-tag :type="currentReview.is_replied ? 'success' : 'warning'" size="small">
              {{ currentReview.is_replied ? '已回复' : '未回复' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="评价时间">{{ currentReview.created_at }}</el-descriptions-item>
          <el-descriptions-item label="订单号">{{ currentReview.order_no || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h4 class="section-title">评价内容</h4>
          <p class="content-text">{{ currentReview.content }}</p>
        </div>

        <div v-if="currentReview.images && currentReview.images.length" class="detail-section">
          <h4 class="section-title">评价图片</h4>
          <div class="image-list">
            <el-image
              v-for="(img, idx) in currentReview.images"
              :key="idx"
              :src="img"
              :preview-src-list="currentReview.images"
              fit="cover"
              class="review-image" />
          </div>
        </div>

        <div v-if="currentReview.reply_content" class="detail-section">
          <h4 class="section-title">商家回复</h4>
          <p class="content-text reply-text">{{ currentReview.reply_content }}</p>
          <p class="reply-time">回复时间：{{ currentReview.replied_at }}</p>
        </div>

        <div v-if="!currentReview.is_replied" class="detail-section">
          <h4 class="section-title">回复评价</h4>
          <el-form ref="replyFormRef" :model="replyForm" :rules="replyRules" label-width="80px">
            <el-form-item label="回复内容" prop="reply_content">
              <el-input
                v-model="replyForm.reply_content"
                type="textarea"
                :rows="4"
                placeholder="请输入回复内容" />
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button
          v-if="currentReview && !currentReview.is_replied"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmitReply">
          提交回复
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="replyVisible"
      title="回复评价"
      width="600px"
      :close-on-click-modal="false">
      <div v-if="currentReview" class="reply-info">
        <p><strong>用户：</strong>{{ currentReview.user_nickname }}</p>
        <p><strong>评分：</strong>
          <el-rate v-model="currentReview.rating" disabled show-score text-color="#ff9900" />
        </p>
        <p><strong>评价内容：</strong>{{ currentReview.content }}</p>
      </div>
      <el-form ref="replyFormRef2" :model="replyForm" :rules="replyRules" label-width="80px" style="margin-top: 16px;">
        <el-form-item label="回复内容" prop="reply_content">
          <el-input
            v-model="replyForm.reply_content"
            type="textarea"
            :rows="5"
            placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmitReply">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="workOrderVisible"
      title="创建差评工单"
      width="600px"
      :close-on-click-modal="false">
      <div v-if="currentReview" class="work-order-info">
        <el-alert
          :title="'差评内容：' + (currentReview.content || '')"
          type="warning"
          :closable="false"
          show-icon />
      </div>
      <el-form
        ref="workOrderFormRef"
        :model="workOrderForm"
        :rules="workOrderRules"
        label-width="100px"
        style="margin-top: 16px;">
        <el-form-item label="优先级" prop="priority">
          <el-radio-group v-model="workOrderForm.priority">
            <el-radio :value="1">低</el-radio>
            <el-radio :value="2">中</el-radio>
            <el-radio :value="3">高</el-radio>
            <el-radio :value="4">紧急</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="指派店长" prop="assignee_id">
          <el-select
            v-model="workOrderForm.assignee_id"
            placeholder="请选择店长"
            filterable
            style="width: 100%">
            <el-option
              v-for="manager in managerList"
              :key="manager.id"
              :label="manager.name"
              :value="manager.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="截止时间" prop="deadline">
          <el-date-picker
            v-model="workOrderForm.deadline"
            type="datetime"
            placeholder="选择截止时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="workOrderForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="workOrderVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmitWorkOrder">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import {
  getReviewList,
  getReviewDetail,
  replyReview,
  createWorkOrder
} from '@/api/review'
import { getStoreList } from '@/api/store'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)
const storeList = ref([])
const managerList = ref([
  { id: 1, name: '张三' },
  { id: 2, name: '李四' },
  { id: 3, name: '王五' }
])

const query = reactive({
  store_id: null,
  platform: '',
  rating_min: null,
  rating_max: null,
  is_bad: null,
  is_replied: null,
  keyword: '',
  time_range: [],
  page: 1,
  page_size: 10
})

const detailVisible = ref(false)
const replyVisible = ref(false)
const workOrderVisible = ref(false)
const currentReview = ref(null)
const replyFormRef = ref()
const replyFormRef2 = ref()
const workOrderFormRef = ref()

const replyForm = reactive({
  reply_content: ''
})

const replyRules = {
  reply_content: [{ required: true, message: '请输入回复内容', trigger: 'blur' }]
}

const workOrderForm = reactive({
  review_id: null,
  priority: 2,
  assignee_id: null,
  deadline: '',
  remark: ''
})

const workOrderRules = {
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
  assignee_id: [{ required: true, message: '请选择指派店长', trigger: 'change' }]
}

const platformMap = {
  meituan: { name: '美团', type: 'primary' },
  eleme: { name: '饿了么', type: 'success' },
  dianping: { name: '大众点评', type: 'warning' },
  douyin: { name: '抖音', type: 'danger' },
  miniprogram: { name: '小程序', type: 'info' }
}

function getPlatformName(platform) {
  return platformMap[platform]?.name || '未知'
}

function getPlatformTagType(platform) {
  return platformMap[platform]?.type || 'info'
}

async function fetchStoreList() {
  try {
    const res = await getStoreList({ page: 1, page_size: 100 })
    storeList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchList() {
  loading.value = true
  try {
    const params = { ...query }
    if (params.time_range && params.time_range.length === 2) {
      params.start_time = params.time_range[0]
      params.end_time = params.time_range[1]
    }
    delete params.time_range
    const res = await getReviewList(params)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.store_id = null
  query.platform = ''
  query.rating_min = null
  query.rating_max = null
  query.is_bad = null
  query.is_replied = null
  query.keyword = ''
  query.time_range = []
  query.page = 1
  fetchList()
}

async function handleViewDetail(row) {
  try {
    const res = await getReviewDetail(row.id)
    currentReview.value = res
    replyForm.reply_content = ''
    detailVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

function handleReply(row) {
  currentReview.value = row
  replyForm.reply_content = ''
  replyVisible.value = true
}

function handleCreateWorkOrder(row) {
  currentReview.value = row
  workOrderForm.review_id = row.id
  workOrderForm.priority = row.is_bad ? 3 : 2
  workOrderForm.assignee_id = null
  workOrderForm.deadline = ''
  workOrderForm.remark = ''
  workOrderVisible.value = true
}

async function handleSubmitReply() {
  try {
    const formRef = detailVisible.value ? replyFormRef.value : replyFormRef2.value
    await formRef.validate()
    submitLoading.value = true
    await replyReview(currentReview.value.id, { reply_content: replyForm.reply_content })
    ElMessage.success('回复成功')
    detailVisible.value = false
    replyVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

async function handleSubmitWorkOrder() {
  try {
    await workOrderFormRef.value.validate()
    submitLoading.value = true
    await createWorkOrder({
      ...workOrderForm,
      store_id: currentReview.value.store_id
    })
    ElMessage.success('工单创建成功')
    workOrderVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchStoreList()
  fetchList()
})
</script>

<style scoped lang="scss">
.reviews-page {
  .search-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 20px;
    align-items: center;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .review-detail {
    .detail-section {
      margin-top: 20px;

      .section-title {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }

    .content-text {
      margin: 0;
      padding: 12px;
      background: #f5f7fa;
      border-radius: 4px;
      line-height: 1.6;
      color: #303133;
    }

    .reply-text {
      background: #ecf5ff;
    }

    .reply-time {
      margin: 8px 0 0 0;
      font-size: 12px;
      color: #909399;
    }

    .image-list {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
    }

    .review-image {
      width: 100px;
      height: 100px;
      border-radius: 4px;
      cursor: pointer;
    }
  }

  .reply-info {
    padding: 12px;
    background: #f5f7fa;
    border-radius: 4px;

    p {
      margin: 6px 0;
      font-size: 13px;
      color: #303133;
    }
  }

  .work-order-info {
    margin-bottom: 4px;
  }
}
</style>
