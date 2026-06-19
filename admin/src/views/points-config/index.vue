<template>
  <div class="points-config-page">
    <div class="page-header">
      <h2 class="page-title">积分配置</h2>
      <div class="header-actions">
        <el-button type="success" @click="openRuleDialog">
          <el-icon><Plus /></el-icon>新增规则
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="section-title">积分获取规则</div>
      <el-table :data="earnRules" v-loading="loading" empty-text="暂无积分获取规则">
        <el-table-column prop="rule_name" label="规则名称" min-width="150" />
        <el-table-column prop="rule_key" label="规则标识" width="150" />
        <el-table-column label="每元积分数" width="120" align="center">
          <template #default="{ row }">
            <span class="highlight">{{ row.points_per_yuan }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="min_consume_amount" label="最低消费金额" width="140" align="center">
          <template #default="{ row }">
            <span v-if="row.min_consume_amount > 0">¥{{ row.min_consume_amount }}</span>
            <span v-else>无限制</span>
          </template>
        </el-table-column>
        <el-table-column prop="bonus_points" label="赠送积分" width="120" align="center">
          <template #default="{ row }">
            <span v-if="row.bonus_points > 0" class="highlight">{{ row.bonus_points }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditRule(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="section-title" style="margin-top: 30px;">积分抵扣规则</div>
      <el-table :data="redeemRules" v-loading="loading" empty-text="暂无积分抵扣规则">
        <el-table-column prop="rule_name" label="规则名称" min-width="150" />
        <el-table-column prop="rule_key" label="规则标识" width="150" />
        <el-table-column label="抵扣比率" width="200" align="center">
          <template #default="{ row }">
            <span class="highlight">{{ Math.round(1 / row.redeem_rate) }}积分 = 1元</span>
          </template>
        </el-table-column>
        <el-table-column prop="min_redeem_points" label="最低抵扣积分" width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.min_redeem_points }} 积分</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditRule(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="section-title" style="margin-top: 30px;">注册赠送规则</div>
      <el-table :data="registerRules" v-loading="loading" empty-text="暂无注册赠送规则">
        <el-table-column prop="rule_name" label="规则名称" min-width="150" />
        <el-table-column prop="rule_key" label="规则标识" width="150" />
        <el-table-column prop="bonus_points" label="注册赠送积分" width="150" align="center">
          <template #default="{ row }">
            <span class="highlight">{{ row.bonus_points }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditRule(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="ruleDialogVisible"
      :title="isEdit ? '编辑积分规则' : '新增积分规则'"
      width="600px"
      :close-on-click-modal="false">
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleRules"
        label-width="120px">
        <el-form-item label="规则名称" prop="rule_name">
          <el-input v-model="ruleForm.rule_name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="规则标识" prop="rule_key">
          <el-input v-model="ruleForm.rule_key" placeholder="请输入规则标识" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="规则类型" prop="rule_type">
          <el-select v-model="ruleForm.rule_type" placeholder="请选择规则类型" style="width: 100%" :disabled="isEdit" @change="handleRuleTypeChange">
            <el-option label="积分获取" value="earn" />
            <el-option label="积分抵扣" value="redeem" />
            <el-option label="注册赠送" value="register" />
            <el-option label="活动赠送" value="bonus" />
          </el-select>
        </el-form-item>
        <template v-if="ruleForm.rule_type === 'earn'">
          <el-form-item label="每元积分数" prop="points_per_yuan">
            <el-input-number v-model="ruleForm.points_per_yuan" :min="0.01" :precision="2" :step="0.5" style="width: 200px" />
            <span style="margin-left: 8px; color: #909399;">积分/元</span>
          </el-form-item>
          <el-form-item label="最低消费金额">
            <el-input-number v-model="ruleForm.min_consume_amount" :min="0" :precision="2" style="width: 200px" />
            <span style="margin-left: 8px; color: #909399;">元 (0为无限制)</span>
          </el-form-item>
        </template>
        <template v-if="ruleForm.rule_type === 'redeem'">
          <el-form-item label="抵扣比率" prop="redeem_rate">
            <el-input-number v-model="redeemPointsPerYuan" :min="1" :step="10" style="width: 200px" />
            <span style="margin-left: 8px; color: #909399;">积分 = 1元</span>
          </el-form-item>
          <el-form-item label="最低抵扣积分">
            <el-input-number v-model="ruleForm.min_redeem_points" :min="1" :step="10" style="width: 200px" />
            <span style="margin-left: 8px; color: #909399;">积分</span>
          </el-form-item>
        </template>
        <template v-if="ruleForm.rule_type === 'register' || ruleForm.rule_type === 'bonus'">
          <el-form-item label="赠送积分" prop="bonus_points">
            <el-input-number v-model="ruleForm.bonus_points" :min="1" style="width: 200px" />
            <span style="margin-left: 8px; color: #909399;">积分</span>
          </el-form-item>
        </template>
        <el-form-item label="优先级">
          <el-input-number v-model="ruleForm.priority" :min="0" style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="ruleForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmitRule">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  getPointsRuleList,
  createPointsRule,
  updatePointsRule,
  deletePointsRule
} from '@/api/points-rules'

const loading = ref(false)
const submitLoading = ref(false)
const allRules = ref([])

const earnRules = computed(() => allRules.value.filter(r => r.rule_type === 'earn'))
const redeemRules = computed(() => allRules.value.filter(r => r.rule_type === 'redeem'))
const registerRules = computed(() => allRules.value.filter(r => r.rule_type === 'register' || r.rule_type === 'bonus'))

const ruleDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const ruleFormRef = ref()
const redeemPointsPerYuan = ref(100)

const ruleForm = reactive({
  rule_key: '',
  rule_name: '',
  rule_type: 'earn',
  points_per_yuan: 1,
  redeem_rate: 0.01,
  min_redeem_points: 100,
  bonus_points: 0,
  min_consume_amount: 0,
  priority: 0,
  status: 1
})

const ruleRules = {
  rule_name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  rule_key: [{ required: true, message: '请输入规则标识', trigger: 'blur' }],
  rule_type: [{ required: true, message: '请选择规则类型', trigger: 'change' }]
}

async function fetchRules() {
  loading.value = true
  try {
    const res = await getPointsRuleList({ page: 1, page_size: 100 })
    allRules.value = res.list || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function openRuleDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(ruleForm, {
      rule_key: row.rule_key,
      rule_name: row.rule_name,
      rule_type: row.rule_type,
      points_per_yuan: row.points_per_yuan,
      redeem_rate: row.redeem_rate,
      min_redeem_points: row.min_redeem_points,
      bonus_points: row.bonus_points,
      min_consume_amount: row.min_consume_amount,
      priority: row.priority,
      status: row.status
    })
    redeemPointsPerYuan.value = row.redeem_rate > 0 ? Math.round(1 / row.redeem_rate) : 100
  } else {
    ruleForm.rule_key = ''
    ruleForm.rule_name = ''
    ruleForm.rule_type = 'earn'
    ruleForm.points_per_yuan = 1
    ruleForm.redeem_rate = 0.01
    ruleForm.min_redeem_points = 100
    ruleForm.bonus_points = 0
    ruleForm.min_consume_amount = 0
    ruleForm.priority = 0
    ruleForm.status = 1
    redeemPointsPerYuan.value = 100
  }

  ruleDialogVisible.value = true
}

function handleRuleTypeChange(type) {
  if (type === 'earn') {
    ruleForm.points_per_yuan = 1
    ruleForm.min_consume_amount = 0
  } else if (type === 'redeem') {
    ruleForm.redeem_rate = 0.01
    ruleForm.min_redeem_points = 100
  } else if (type === 'register') {
    ruleForm.bonus_points = 100
  }
}

function handleEditRule(row) {
  openRuleDialog(row)
}

async function handleStatusChange(row) {
  try {
    await updatePointsRule(row.id, { status: row.status })
    ElMessage.success(row.status === 1 ? '规则已启用' : '规则已禁用')
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

function handleDeleteRule(row) {
  ElMessageBox.confirm(`确定删除规则"${row.rule_name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deletePointsRule(row.id)
    ElMessage.success('删除成功')
    fetchRules()
  })
}

async function handleSubmitRule() {
  try {
    await ruleFormRef.value.validate()
    submitLoading.value = true

    const data = { ...ruleForm }
    if (data.rule_type === 'redeem' && redeemPointsPerYuan.value > 0) {
      data.redeem_rate = parseFloat((1 / redeemPointsPerYuan.value).toFixed(6))
    }

    if (isEdit.value) {
      await updatePointsRule(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createPointsRule(data)
      ElMessage.success('创建成功')
    }

    ruleDialogVisible.value = false
    fetchRules()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchRules()
})
</script>

<style scoped lang="scss">
.points-config-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 16px;
    padding-left: 10px;
    border-left: 3px solid #409eff;
  }

  .highlight {
    color: #e6a23c;
    font-weight: 600;
  }
}
</style>
