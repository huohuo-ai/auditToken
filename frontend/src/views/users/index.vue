<template>
  <div class="users-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>新建用户
          </el-button>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部角色" clearable>
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable>
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadUsers">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 用户列表 -->
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="role" label="角色">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'admin'" type="danger">管理员</el-tag>
            <el-tag v-else>普通用户</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'active'" type="success">正常</el-tag>
            <el-tag v-else type="info">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="api_key" label="API Key" width="280">
          <template #default="{ row }">
            <code class="api-key">{{ maskApiKey(row.api_key) }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" @click="handleQuota(row)">配额</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
    
    <!-- 创建/编辑对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 配额对话框 -->
    <el-dialog
      title="设置配额"
      v-model="quotaDialogVisible"
      width="400px"
    >
      <el-form :model="quotaForm" label-width="100px">
        <el-form-item label="日限额">
          <el-input-number v-model="quotaForm.daily_limit" :min="0" :step="1000" style="width: 100%" />
          <span class="hint">0表示无限制</span>
        </el-form-item>
        <el-form-item label="周限额">
          <el-input-number v-model="quotaForm.weekly_limit" :min="0" :step="10000" style="width: 100%" />
          <span class="hint">0表示无限制</span>
        </el-form-item>
        <el-form-item label="月限额">
          <el-input-number v-model="quotaForm.monthly_limit" :min="0" :step="100000" style="width: 100%" />
          <span class="hint">0表示无限制</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="quotaDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleQuotaSubmit" :loading="quotaSubmitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, createUser, updateUser, deleteUser, getUserQuota, updateUserQuota } from '@/api/user'

export default {
  name: 'Users',
  setup() {
    const loading = ref(false)
    const users = ref([])
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    
    const searchForm = reactive({
      role: '',
      status: ''
    })
    
    const dialogVisible = ref(false)
    const dialogTitle = ref('')
    const isEdit = ref(false)
    const submitLoading = ref(false)
    const formRef = ref(null)
    
    const form = reactive({
      id: null,
      username: '',
      email: '',
      password: '',
      role: 'user',
      status: 'active'
    })
    
    const rules = {
      username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
      ],
      password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
      role: [{ required: true, message: '请选择角色', trigger: 'change' }]
    }
    
    const quotaDialogVisible = ref(false)
    const quotaSubmitLoading = ref(false)
    const currentUser = ref(null)
    const quotaForm = reactive({
      daily_limit: 0,
      weekly_limit: 0,
      monthly_limit: 0
    })
    
    const loadUsers = async () => {
      loading.value = true
      try {
        const res = await getUserList({
          page: page.value,
          page_size: pageSize.value,
          role: searchForm.role,
          status: searchForm.status
        })
        users.value = res.data
        total.value = res.total
      } catch (error) {
        console.error('Failed to load users:', error)
      } finally {
        loading.value = false
      }
    }
    
    const resetSearch = () => {
      searchForm.role = ''
      searchForm.status = ''
      loadUsers()
    }
    
    const handleCreate = () => {
      dialogTitle.value = '新建用户'
      isEdit.value = false
      Object.assign(form, {
        id: null,
        username: '',
        email: '',
        password: '',
        role: 'user',
        status: 'active'
      })
      dialogVisible.value = true
    }
    
    const handleEdit = (row) => {
      dialogTitle.value = '编辑用户'
      isEdit.value = true
      Object.assign(form, {
        id: row.id,
        username: row.username,
        email: row.email,
        role: row.role,
        status: row.status
      })
      dialogVisible.value = true
    }
    
    const handleSubmit = async () => {
      try {
        await formRef.value.validate()
        submitLoading.value = true
        
        if (isEdit.value) {
          await updateUser(form.id, {
            username: form.username,
            email: form.email,
            role: form.role,
            status: form.status
          })
          ElMessage.success('更新成功')
        } else {
          await createUser({
            username: form.username,
            email: form.email,
            password: form.password,
            role: form.role
          })
          ElMessage.success('创建成功')
        }
        
        dialogVisible.value = false
        loadUsers()
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitLoading.value = false
      }
    }
    
    const handleDelete = (row) => {
      ElMessageBox.confirm(`确定要删除用户 "${row.username}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        await deleteUser(row.id)
        ElMessage.success('删除成功')
        loadUsers()
      })
    }
    
    const handleQuota = async (row) => {
      currentUser.value = row
      try {
        const res = await getUserQuota(row.id)
        Object.assign(quotaForm, {
          daily_limit: res.daily_limit || 0,
          weekly_limit: res.weekly_limit || 0,
          monthly_limit: res.monthly_limit || 0
        })
        quotaDialogVisible.value = true
      } catch (error) {
        ElMessage.error('获取配额信息失败')
      }
    }
    
    const handleQuotaSubmit = async () => {
      quotaSubmitLoading.value = true
      try {
        await updateUserQuota(currentUser.value.id, quotaForm)
        ElMessage.success('配额设置成功')
        quotaDialogVisible.value = false
      } catch (error) {
        console.error('Failed to update quota:', error)
      } finally {
        quotaSubmitLoading.value = false
      }
    }
    
    const handleSizeChange = (val) => {
      pageSize.value = val
      loadUsers()
    }
    
    const handlePageChange = (val) => {
      page.value = val
      loadUsers()
    }
    
    const formatDate = (date) => {
      if (!date) return '-'
      return new Date(date).toLocaleString()
    }
    
    const maskApiKey = (key) => {
      if (!key || key.length < 12) return '****'
      return key.substring(0, 6) + '****' + key.substring(key.length - 4)
    }
    
    onMounted(() => {
      loadUsers()
    })
    
    return {
      loading,
      users,
      page,
      pageSize,
      total,
      searchForm,
      dialogVisible,
      dialogTitle,
      isEdit,
      submitLoading,
      formRef,
      form,
      rules,
      quotaDialogVisible,
      quotaSubmitLoading,
      quotaForm,
      loadUsers,
      resetSearch,
      handleCreate,
      handleEdit,
      handleSubmit,
      handleDelete,
      handleQuota,
      handleQuotaSubmit,
      handleSizeChange,
      handlePageChange,
      formatDate,
      maskApiKey
    }
  }
}
</script>

<style scoped>
.users-management {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-form {
    margin-bottom: 20px;
  }
  
  .api-key {
    font-size: 12px;
    background: #f5f7fa;
    padding: 2px 6px;
    border-radius: 3px;
  }
  
  .pagination {
    margin-top: 20px;
    text-align: right;
  }
  
  .hint {
    font-size: 12px;
    color: #909399;
    margin-left: 10px;
  }
}
</style>
