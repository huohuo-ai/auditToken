<template>
  <div class="models-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>模型管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>添加模型
          </el-button>
        </div>
      </template>
      
      <el-table :data="models" v-loading="loading" stripe>
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="display_name" label="显示名称" />
        <el-table-column prop="provider" label="提供商" />
        <el-table-column prop="model_id" label="模型ID" />
        <el-table-column prop="system_prompt" label="系统提示词" show-overflow-tooltip />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'active'" type="success">正常</el-tag>
            <el-tag v-else type="info">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_default" label="默认">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="primary">是</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
    
    <!-- 创建/编辑对话框 -->
    <el-dialog :title="dialogTitle" v-model="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：gpt-3.5-turbo" />
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="form.display_name" placeholder="如：GPT-3.5" />
        </el-form-item>
        <el-form-item label="提供商" prop="provider">
          <el-select v-model="form.provider" style="width: 100%">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Azure" value="azure" />
            <el-option label="Anthropic" value="anthropic" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="Base URL" prop="base_url">
          <el-input v-model="form.base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="API Key" prop="api_key">
          <el-input v-model="form.api_key" type="password" show-password />
        </el-form-item>
        <el-form-item label="模型ID" prop="model_id">
          <el-input v-model="form.model_id" placeholder="实际调用的模型ID" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="最大Token">
              <el-input-number v-model="form.max_tokens" :min="1" :max="32768" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="温度">
              <el-input-number v-model="form.temperature" :min="0" :max="2" :step="0.1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="默认模型">
          <el-switch v-model="form.is_default" />
        </el-form-item>
        <el-form-item label="系统提示词">
          <el-input 
            v-model="form.system_prompt" 
            type="textarea" 
            :rows="4" 
            placeholder="设置系统提示词，如：你是一位专业的企业助手，请用中文回答..."
          />
          <div class="form-hint">配置后，每次请求会自动添加此系统提示词到对话开头</div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getModelList, createModel, updateModel, deleteModel } from '@/api/model'

export default {
  name: 'Models',
  setup() {
    const loading = ref(false)
    const models = ref([])
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    
    const dialogVisible = ref(false)
    const dialogTitle = ref('')
    const isEdit = ref(false)
    const submitLoading = ref(false)
    const formRef = ref(null)
    
    const form = reactive({
      id: null,
      name: '',
      display_name: '',
      provider: 'openai',
      base_url: '',
      api_key: '',
      model_id: '',
      max_tokens: 4096,
      temperature: 0.7,
      is_default: false,
      system_prompt: '',
      description: ''
    })
    
    const rules = {
      name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
      display_name: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
      provider: [{ required: true, message: '请选择提供商', trigger: 'change' }],
      base_url: [{ required: true, message: '请输入Base URL', trigger: 'blur' }],
      api_key: [{ required: true, message: '请输入API Key', trigger: 'blur' }],
      model_id: [{ required: true, message: '请输入模型ID', trigger: 'blur' }]
    }
    
    const loadModels = async () => {
      loading.value = true
      try {
        const res = await getModelList({
          page: page.value,
          page_size: pageSize.value
        })
        models.value = res.data
        total.value = res.total
      } catch (error) {
        console.error('Failed to load models:', error)
      } finally {
        loading.value = false
      }
    }
    
    const handleCreate = () => {
      dialogTitle.value = '添加模型'
      isEdit.value = false
      Object.assign(form, {
        id: null,
        name: '',
        display_name: '',
        provider: 'openai',
        base_url: '',
        api_key: '',
        model_id: '',
        max_tokens: 4096,
        temperature: 0.7,
        is_default: false,
        system_prompt: '',
        description: ''
      })
      dialogVisible.value = true
    }
    
    const handleEdit = (row) => {
      dialogTitle.value = '编辑模型'
      isEdit.value = true
      Object.assign(form, row)
      dialogVisible.value = true
    }
    
    const handleSubmit = async () => {
      try {
        await formRef.value.validate()
        submitLoading.value = true
        
        if (isEdit.value) {
          await updateModel(form.id, form)
          ElMessage.success('更新成功')
        } else {
          await createModel(form)
          ElMessage.success('创建成功')
        }
        
        dialogVisible.value = false
        loadModels()
      } catch (error) {
        console.error('Submit error:', error)
      } finally {
        submitLoading.value = false
      }
    }
    
    const handleDelete = (row) => {
      ElMessageBox.confirm(`确定要删除模型 "${row.name}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        await deleteModel(row.id)
        ElMessage.success('删除成功')
        loadModels()
      })
    }
    
    const handleSizeChange = (val) => {
      pageSize.value = val
      loadModels()
    }
    
    const handlePageChange = (val) => {
      page.value = val
      loadModels()
    }
    
    onMounted(() => {
      loadModels()
    })
    
    return {
      loading,
      models,
      page,
      pageSize,
      total,
      dialogVisible,
      dialogTitle,
      isEdit,
      submitLoading,
      formRef,
      form,
      rules,
      loadModels,
      handleCreate,
      handleEdit,
      handleSubmit,
      handleDelete,
      handleSizeChange,
      handlePageChange
    }
  }
}
</script>

<style scoped>
.models-management {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .pagination {
    margin-top: 20px;
    text-align: right;
  }
  
  .form-hint {
    font-size: 12px;
    color: #909399;
    margin-top: 5px;
  }
}
</style>
