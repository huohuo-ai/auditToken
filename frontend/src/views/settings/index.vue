<template>
  <div class="settings">
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>个人信息</span>
          </template>
          
          <el-form :model="userForm" label-width="100px">
            <el-form-item label="用户名">
              <el-input v-model="userForm.username" disabled />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="userForm.email" />
            </el-form-item>
            <el-form-item label="角色">
              <el-tag v-if="userForm.role === 'admin'" type="danger">管理员</el-tag>
              <el-tag v-else>普通用户</el-tag>
            </el-form-item>
            <el-form-item label="API Key">
              <el-input v-model="userForm.api_key" type="password" show-password readonly>
                <template #append>
                  <el-button @click="copyApiKey">复制</el-button>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleRegenerateKey">重新生成 API Key</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>修改密码</span>
          </template>
          
          <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="120px">
            <el-form-item label="旧密码" prop="old_password">
              <el-input v-model="passwordForm.old_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="passwordForm.new_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="确认新密码" prop="confirm_password">
              <el-input v-model="passwordForm.confirm_password" type="password" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="passwordLoading">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>
        
        <el-card style="margin-top: 20px;">
          <template #header>
            <span>使用配额</span>
          </template>
          
          <div v-if="quota" class="quota-info">
            <div class="quota-item">
              <span class="quota-label">日限额:</span>
              <el-progress 
                :percentage="getQuotaPercentage(quota.daily_used, quota.daily_limit)" 
                :status="getQuotaStatus(quota.daily_used, quota.daily_limit)"
              />
              <span class="quota-value">{{ formatNumber(quota.daily_used) }} / {{ formatNumber(quota.daily_limit) }}</span>
            </div>
            <div class="quota-item">
              <span class="quota-label">周限额:</span>
              <el-progress 
                :percentage="getQuotaPercentage(quota.weekly_used, quota.weekly_limit)" 
                :status="getQuotaStatus(quota.weekly_used, quota.weekly_limit)"
              />
              <span class="quota-value">{{ formatNumber(quota.weekly_used) }} / {{ formatNumber(quota.weekly_limit) }}</span>
            </div>
            <div class="quota-item">
              <span class="quota-label">月限额:</span>
              <el-progress 
                :percentage="getQuotaPercentage(quota.monthly_used, quota.monthly_limit)" 
                :status="getQuotaStatus(quota.monthly_used, quota.monthly_limit)"
              />
              <span class="quota-value">{{ formatNumber(quota.monthly_used) }} / {{ formatNumber(quota.monthly_limit) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useStore } from 'vuex'
import { ElMessage } from 'element-plus'
import { changePassword, regenerateApiKey, getProfile } from '@/api/auth'

export default {
  name: 'Settings',
  setup() {
    const store = useStore()
    const passwordFormRef = ref(null)
    const passwordLoading = ref(false)
    const quota = ref(null)
    
    const userForm = computed(() => store.state.user)
    
    const passwordForm = reactive({
      old_password: '',
      new_password: '',
      confirm_password: ''
    })
    
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== passwordForm.new_password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }
    
    const passwordRules = {
      old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
      new_password: [
        { required: true, message: '请输入新密码', trigger: 'blur' },
        { min: 6, message: '密码长度至少6位', trigger: 'blur' }
      ],
      confirm_password: [
        { required: true, message: '请确认新密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ]
    }
    
    const loadQuota = async () => {
      try {
        const res = await getProfile()
        quota.value = res.quota
      } catch (error) {
        console.error('Failed to load quota:', error)
      }
    }
    
    const handleChangePassword = async () => {
      try {
        await passwordFormRef.value.validate()
        passwordLoading.value = true
        
        await changePassword({
          old_password: passwordForm.old_password,
          new_password: passwordForm.new_password
        })
        
        ElMessage.success('密码修改成功')
        passwordForm.old_password = ''
        passwordForm.new_password = ''
        passwordForm.confirm_password = ''
      } catch (error) {
        console.error('Change password error:', error)
      } finally {
        passwordLoading.value = false
      }
    }
    
    const handleRegenerateKey = () => {
      ElMessageBox.confirm('重新生成API Key后，旧Key将立即失效，确定继续吗？', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        const res = await regenerateApiKey()
        store.commit('SET_USER', { ...store.state.user, api_key: res.api_key })
        ElMessage.success('API Key已重新生成')
      })
    }
    
    const copyApiKey = () => {
      navigator.clipboard.writeText(userForm.value.api_key).then(() => {
        ElMessage.success('已复制到剪贴板')
      })
    }
    
    const getQuotaPercentage = (used, limit) => {
      if (!limit || limit === 0) return 0
      return Math.min(Math.round((used / limit) * 100), 100)
    }
    
    const getQuotaStatus = (used, limit) => {
      if (!limit || limit === 0) return ''
      const percentage = (used / limit) * 100
      if (percentage >= 90) return 'exception'
      if (percentage >= 70) return 'warning'
      return ''
    }
    
    const formatNumber = (num) => {
      if (!num || num === 0) return '无限制'
      if (num >= 10000) {
        return (num / 10000).toFixed(1) + '万'
      }
      return num.toLocaleString()
    }
    
    onMounted(() => {
      loadQuota()
    })
    
    return {
      userForm,
      passwordForm,
      passwordRules,
      passwordFormRef,
      passwordLoading,
      quota,
      handleChangePassword,
      handleRegenerateKey,
      copyApiKey,
      getQuotaPercentage,
      getQuotaStatus,
      formatNumber
    }
  }
}
</script>

<style scoped>
.settings {
  .quota-info {
    .quota-item {
      display: flex;
      align-items: center;
      margin-bottom: 20px;
      
      .quota-label {
        width: 80px;
        flex-shrink: 0;
      }
      
      .el-progress {
        flex: 1;
        margin: 0 15px;
      }
      
      .quota-value {
        width: 150px;
        text-align: right;
        color: #606266;
        font-size: 14px;
      }
    }
  }
}
</style>
