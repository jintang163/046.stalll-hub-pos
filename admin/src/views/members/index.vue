<template>
  <div class="members-page">
    <div class="page-header">
      <h2 class="page-title">会员管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增会员
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.keyword"
          placeholder="搜索姓名/手机号"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.level" placeholder="会员等级" clearable style="width: 140px">
          <el-option label="普通会员" :value="1" />
          <el-option label="银卡会员" :value="2" />
          <el-option label="金卡会员" :value="3" />
          <el-option label="钻石会员" :value="4" />
        </el-select>
        <el-select v-model="query.status" placeholder="会员状态" clearable style="width: 140px">
          <el-option label="正常" :value="1" />
          <el-option label="已冻结" :value="0" />
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
        <el-table-column prop="avatar" label="头像" width="80">
          <template #default="{ row }">
            <el-image
              v-if="row.avatar"
              :src="row.avatar"
              style="width: 40px; height: 40px; border-radius: 50%;"
              fit="cover" />
            <div v-else style="width: 40px; height: 40px; background: #f5f7fa; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #c0c4cc; font-size: 12px;">
              {{ row.name?.charAt(0) || '会' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column label="会员等级" width="120">
          <template #default="{ row }">
            <el-tag :type="getLevelTagType(row.level)">{{ getLevelName(row.level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="points" label="积分" width="100" align="center">
          <template #default="{ row }">
            <span class="points">{{ row.points || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额(元)" width="120" align="center">
          <template #default="{ row }">
            <span class="balance">¥{{ (row.balance || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total_consume" label="累计消费(元)" width="140" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="160" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="warning" link size="small" @click="openPointsDialog(row)">积分调整</el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
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
      :title="isEdit ? '编辑会员' : '新增会员'"
      width="600px"
      :close-on-click-modal="false">
      <el-form
        ref="memberFormRef"
        :model="memberForm"
        :rules="memberRules"
        label-width="100px">
        <el-form-item label="头像">
          <el-input v-model="memberForm.avatar" placeholder="请输入头像URL" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="memberForm.name" placeholder="请输入姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="memberForm.phone" placeholder="请输入手机号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="性别">
              <el-radio-group v-model="memberForm.gender">
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
                <el-radio :value="0">未知</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="生日">
              <el-date-picker
                v-model="memberForm.birthday"
                type="date"
                placeholder="选择生日"
                value-format="YYYY-MM-DD"
                style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="会员等级" prop="level">
              <el-select v-model="memberForm.level" placeholder="请选择会员等级" style="width: 100%">
                <el-option label="普通会员" :value="1" />
                <el-option label="银卡会员" :value="2" />
                <el-option label="金卡会员" :value="3" />
                <el-option label="钻石会员" :value="4" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="初始余额">
              <el-input-number v-model="memberForm.balance" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="初始积分">
          <el-input-number v-model="memberForm.points" :min="0" style="width: 200px" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="memberForm.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="pointsDialogVisible" title="积分调整" width="500px">
      <el-form label-width="100px">
        <el-form-item label="当前会员">
          <span>{{ currentMember?.name }} - {{ currentMember?.phone }}</span>
        </el-form-item>
        <el-form-item label="当前积分">
          <span class="current-points">{{ currentMember?.points || 0 }}</span>
        </el-form-item>
        <el-form-item label="调整类型">
          <el-radio-group v-model="pointsForm.type">
            <el-radio value="add">增加积分</el-radio>
            <el-radio value="subtract">扣除积分</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="调整数量">
          <el-input-number v-model="pointsForm.points" :min="1" style="width: 200px" />
        </el-form-item>
        <el-form-item label="调整原因">
          <el-input
            v-model="pointsForm.reason"
            type="textarea"
            :rows="2"
            placeholder="请输入调整原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pointsDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="pointsLoading" @click="handlePointsAdjust">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  getMemberList,
  createMember,
  updateMember,
  deleteMember,
  adjustPoints
} from '@/api/members'

const loading = ref(false)
const submitLoading = ref(false)
const pointsLoading = ref(false)
const list = ref([])
const total = ref(0)
const currentMember = ref(null)

const query = reactive({
  keyword: '',
  level: null,
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const pointsDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const memberFormRef = ref()

const memberForm = reactive({
  avatar: '',
  name: '',
  phone: '',
  gender: 0,
  birthday: '',
  level: 1,
  balance: 0,
  points: 0,
  remark: ''
})

const pointsForm = reactive({
  type: 'add',
  points: 0,
  reason: ''
})

const memberRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  level: [{ required: true, message: '请选择会员等级', trigger: 'change' }]
}

const levelMap = {
  1: { name: '普通会员', type: 'info' },
  2: { name: '银卡会员', type: '' },
  3: { name: '金卡会员', type: 'warning' },
  4: { name: '钻石会员', type: 'danger' }
}

function getLevelName(level) {
  return levelMap[level]?.name || '普通会员'
}

function getLevelTagType(level) {
  return levelMap[level]?.type || 'info'
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getMemberList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.keyword = ''
  query.level = null
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(memberForm, {
      avatar: row.avatar || '',
      name: row.name,
      phone: row.phone,
      gender: row.gender || 0,
      birthday: row.birthday || '',
      level: row.level,
      balance: row.balance || 0,
      points: row.points || 0,
      remark: row.remark || ''
    })
  } else {
    memberForm.avatar = ''
    memberForm.name = ''
    memberForm.phone = ''
    memberForm.gender = 0
    memberForm.birthday = ''
    memberForm.level = 1
    memberForm.balance = 0
    memberForm.points = 0
    memberForm.remark = ''
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function openPointsDialog(row) {
  currentMember.value = row
  pointsForm.type = 'add'
  pointsForm.points = 0
  pointsForm.reason = ''
  pointsDialogVisible.value = true
}

async function handlePointsAdjust() {
  if (!pointsForm.points || pointsForm.points <= 0) {
    ElMessage.warning('请输入调整数量')
    return
  }
  if (!pointsForm.reason) {
    ElMessage.warning('请输入调整原因')
    return
  }

  try {
    pointsLoading.value = true
    const points = pointsForm.type === 'add' ? pointsForm.points : -pointsForm.points
    await adjustPoints(currentMember.value.id, {
      points,
      reason: pointsForm.reason
    })
    ElMessage.success('积分调整成功')
    pointsDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    pointsLoading.value = false
  }
}

async function handleStatusChange(row) {
  try {
    await updateMember(row.id, { status: row.status })
    ElMessage.success(row.status === 1 ? '会员已解冻' : '会员已冻结')
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除会员"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteMember(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await memberFormRef.value.validate()
    submitLoading.value = true

    if (isEdit.value) {
      await updateMember(editId.value, memberForm)
      ElMessage.success('更新成功')
    } else {
      await createMember(memberForm)
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

onMounted(() => {
  fetchList()
})
</script>

<style scoped lang="scss">
.members-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .points {
    color: #e6a23c;
    font-weight: 600;
  }

  .balance {
    color: #f56c6c;
    font-weight: 600;
  }

  .current-points {
    color: #e6a23c;
    font-weight: 600;
    font-size: 18px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
