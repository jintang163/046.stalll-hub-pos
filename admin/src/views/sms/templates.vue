<template>
  <div class="sms-templates-page">
    <div class="page-header">
      <h2 class="page-title">短信模板管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增模板
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索模板名称"
          clearable
          style="width: 220px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.type" placeholder="模板类型" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="营销短信" value="marketing" />
          <el-option label="通知短信" value="notification" />
          <el-option label="验证码" value="captcha" />
        </el-select>
        <el-select v-model="query.review_status" placeholder="审核状态" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="审核通过" value="approved" />
          <el-option label="审核拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="template_code" label="模板编码" width="140" />
        <el-table-column prop="template_name" label="模板名称" min-width="160" />
        <el-table-column label="模板类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTagType(row.template_type)">
              {{ getTypeName(row.template_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sign_name" label="签名名称" width="140" />
        <el-table-column prop="template_content" label="模板内容" min-width="240" show-overflow-tooltip />
        <el-table-column label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="getReviewStatusTagType(row.review_status)">
              {{ getReviewStatusName(row.review_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="是否激活" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="usage_count" label="使用次数" width="100" align="center" />
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">查看</el-button>
            <el-button type="success" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="handleReview(row)" v-if="row.review_status === 'pending'">审核</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
      v-model="dialogVisible"
      :title="isEdit ? '编辑模板' : '新增模板'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="templateFormRef"
        :model="templateForm"
        :rules="templateRules"
        label-width="110px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="模板编码" prop="template_code">
              <el-input v-model="templateForm.template_code" placeholder="请输入模板编码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="模板名称" prop="template_name">
              <el-input v-model="templateForm.template_name" placeholder="请输入模板名称" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="模板类型" prop="template_type">
          <el-radio-group v-model="templateForm.template_type">
            <el-radio value="marketing">营销短信</el-radio>
            <el-radio value="notification">通知短信</el-radio>
            <el-radio value="captcha">验证码</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="签名名称" prop="sign_name">
          <el-input v-model="templateForm.sign_name" placeholder="请输入签名名称" />
        </el-form-item>

        <el-form-item label="模板内容" prop="template_content">
          <el-input
            v-model="templateForm.template_content"
            type="textarea"
            :rows="4"
            placeholder="请输入模板内容，变量使用${变量名}格式" />
          <div class="variable-tip">
            变量数量：<span class="variable-count">{{ variableCount }}</span> 个
            <span class="tip-text">（变量格式：${xxx}）</span>
          </div>
        </el-form-item>

        <el-form-item label="变量列表">
          <div class="variable-list">
            <div
              v-for="(variable, index) in templateForm.variables"
              :key="index"
              class="variable-item">
              <el-input
                v-model="variable.name"
                placeholder="变量名称"
                size="small"
                style="width: 200px" />
              <el-input
                v-model="variable.desc"
                placeholder="变量说明"
                size="small"
                style="flex: 1" />
              <el-button
                v-if="templateForm.variables.length > 0"
                type="danger"
                link
                size="small"
                @click="removeVariable(index)">
                删除
              </el-button>
            </div>
            <el-button type="primary" link size="small" @click="addVariable">+ 添加变量</el-button>
          </div>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="templateForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入模板描述" />
        </el-form-item>

        <el-form-item label="是否激活">
          <el-switch v-model="templateForm.is_active" active-text="激活" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="reviewDialogVisible"
      title="审核模板"
      width="650px"
      :close-on-click-modal="false">
      <div v-if="currentTemplate" class="review-detail">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="模板编码">{{ currentTemplate.template_code }}</el-descriptions-item>
          <el-descriptions-item label="模板名称">{{ currentTemplate.template_name }}</el-descriptions-item>
          <el-descriptions-item label="模板类型">
            <el-tag size="small" :type="getTypeTagType(currentTemplate.template_type)">
              {{ getTypeName(currentTemplate.template_type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签名名称">{{ currentTemplate.sign_name }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h4 class="section-title">模板内容</h4>
          <p class="content-text">{{ currentTemplate.template_content }}</p>
        </div>

        <el-form
          ref="reviewFormRef"
          :model="reviewForm"
          :rules="reviewRules"
          label-width="90px"
          style="margin-top: 20px;">
          <el-form-item label="审核结果" prop="review_result">
            <el-radio-group v-model="reviewForm.review_result">
              <el-radio value="approved">通过</el-radio>
              <el-radio value="rejected">拒绝</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="审核备注" prop="review_remark">
            <el-input
              v-model="reviewForm.review_remark"
              type="textarea"
              :rows="3"
              placeholder="请输入审核备注" />
          </el-form-item>
          <el-form-item label="审核人" prop="reviewer_name">
            <el-input v-model="reviewForm.reviewer_name" placeholder="请输入审核人姓名" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="reviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmitReview">提交审核</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="detailDialogVisible"
      title="模板详情"
      width="650px"
      :close-on-click-modal="false">
      <div v-if="currentTemplate" class="template-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentTemplate.id }}</el-descriptions-item>
          <el-descriptions-item label="模板编码">{{ currentTemplate.template_code }}</el-descriptions-item>
          <el-descriptions-item label="模板名称">{{ currentTemplate.template_name }}</el-descriptions-item>
          <el-descriptions-item label="模板类型">
            <el-tag size="small" :type="getTypeTagType(currentTemplate.template_type)">
              {{ getTypeName(currentTemplate.template_type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签名名称">{{ currentTemplate.sign_name }}</el-descriptions-item>
          <el-descriptions-item label="审核状态">
            <el-tag size="small" :type="getReviewStatusTagType(currentTemplate.review_status)">
              {{ getReviewStatusName(currentTemplate.review_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="是否激活">
            <el-tag :type="currentTemplate.is_active ? 'success' : 'info'" size="small">
              {{ currentTemplate.is_active ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="使用次数">{{ currentTemplate.usage_count || 0 }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentTemplate.created_at }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ currentTemplate.updated_at || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h4 class="section-title">模板内容</h4>
          <p class="content-text">{{ currentTemplate.template_content }}</p>
        </div>

        <div v-if="currentTemplate.variables && currentTemplate.variables.length" class="detail-section">
          <h4 class="section-title">变量列表</h4>
          <el-table :data="currentTemplate.variables" size="small" border>
            <el-table-column prop="name" label="变量名" width="160" />
            <el-table-column prop="desc" label="说明" />
          </el-table>
        </div>

        <div v-if="currentTemplate.description" class="detail-section">
          <h4 class="section-title">模板描述</h4>
          <p class="content-text">{{ currentTemplate.description }}</p>
        </div>

        <div v-if="currentTemplate.review_remark" class="detail-section">
          <h4 class="section-title">审核信息</h4>
          <p class="content-text">
            <strong>审核人：</strong>{{ currentTemplate.reviewer_name || '-' }}<br />
            <strong>审核备注：</strong>{{ currentTemplate.review_remark }}
          </p>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  createSmsTemplate,
  updateSmsTemplate,
  deleteSmsTemplate,
  getSmsTemplate,
  getSmsTemplateList,
  reviewSmsTemplate,
  getActiveTemplates
} from '@/api/sms'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  name: '',
  type: '',
  review_status: '',
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const templateFormRef = ref()

const defaultForm = () => ({
  template_code: '',
  template_name: '',
  template_type: 'notification',
  sign_name: '',
  template_content: '',
  variables: [],
  description: '',
  is_active: true
})

const templateForm = reactive(defaultForm())

const templateRules = {
  template_code: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  template_name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  template_type: [{ required: true, message: '请选择模板类型', trigger: 'change' }],
  sign_name: [{ required: true, message: '请输入签名名称', trigger: 'blur' }],
  template_content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }]
}

const variableCount = computed(() => {
  if (!templateForm.template_content) return 0
  const matches = templateForm.template_content.match(/\$\{[^}]+\}/g)
  return matches ? matches.length : 0
})

const typeMap = {
  marketing: { name: '营销短信', type: 'danger' },
  notification: { name: '通知短信', type: 'primary' },
  captcha: { name: '验证码', type: 'success' }
}

const reviewStatusMap = {
  pending: { name: '待审核', type: 'warning' },
  approved: { name: '已通过', type: 'success' },
  rejected: { name: '已拒绝', type: 'danger' }
}

function getTypeName(type) {
  return typeMap[type]?.name || '未知'
}

function getTypeTagType(type) {
  return typeMap[type]?.type || 'info'
}

function getReviewStatusName(status) {
  return reviewStatusMap[status]?.name || '未知'
}

function getReviewStatusTagType(status) {
  return reviewStatusMap[status]?.type || 'info'
}

function addVariable() {
  templateForm.variables.push({ name: '', desc: '' })
}

function removeVariable(index) {
  templateForm.variables.splice(index, 1)
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getSmsTemplateList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.name = ''
  query.type = ''
  query.review_status = ''
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  Object.assign(templateForm, defaultForm())

  if (row) {
    Object.assign(templateForm, {
      template_code: row.template_code,
      template_name: row.template_name,
      template_type: row.template_type,
      sign_name: row.sign_name,
      template_content: row.template_content,
      variables: row.variables?.length ? [...row.variables] : [],
      description: row.description || '',
      is_active: row.is_active
    })
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除模板"${row.template_name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteSmsTemplate(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await templateFormRef.value.validate()
    submitLoading.value = true

    const data = {
      ...templateForm
    }

    if (isEdit.value) {
      await updateSmsTemplate(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createSmsTemplate(data)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

const reviewDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const currentTemplate = ref(null)
const reviewFormRef = ref()

const reviewForm = reactive({
  review_result: 'approved',
  review_remark: '',
  reviewer_name: ''
})

const reviewRules = {
  review_result: [{ required: true, message: '请选择审核结果', trigger: 'change' }],
  reviewer_name: [{ required: true, message: '请输入审核人姓名', trigger: 'blur' }]
}

async function handleReview(row) {
  try {
    const res = await getSmsTemplate(row.id)
    currentTemplate.value = res
    reviewForm.review_result = 'approved'
    reviewForm.review_remark = ''
    reviewForm.reviewer_name = ''
    reviewDialogVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

async function handleSubmitReview() {
  try {
    await reviewFormRef.value.validate()
    submitLoading.value = true
    await reviewSmsTemplate(currentTemplate.value.id, reviewForm)
    ElMessage.success('审核成功')
    reviewDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

async function handleView(row) {
  try {
    const res = await getSmsTemplate(row.id)
    currentTemplate.value = res
    detailDialogVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped lang="scss">
.sms-templates-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

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

  .variable-tip {
    margin-top: 8px;
    font-size: 13px;
    color: #909399;

    .variable-count {
      color: #409eff;
      font-weight: 600;
    }

    .tip-text {
      margin-left: 12px;
      font-size: 12px;
    }
  }

  .variable-list {
    width: 100%;

    .variable-item {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 10px;
    }
  }

  .review-detail,
  .template-detail {
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
      font-size: 13px;
      word-break: break-all;
    }
  }
}
</style>
