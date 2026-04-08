<template>
  <div class="audit-logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>审计日志</span>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input v-model="searchForm.user_id" placeholder="用户ID" />
        </el-form-item>
        <el-form-item label="模型">
          <el-input v-model="searchForm.model_name" placeholder="模型名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadLogs">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 日志列表 -->
      <el-table :data="logs" v-loading="loading" stripe>
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="request_id" label="请求ID" width="220" />
        <el-table-column prop="user_name" label="用户" width="100" />
        <el-table-column prop="model_name" label="模型" width="120" />
        <el-table-column prop="request_ip" label="IP地址" width="120" />
        <el-table-column prop="total_tokens" label="Token数" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.total_tokens }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="latency_ms" label="延迟" width="80">
          <template #default="{ row }">
            {{ row.latency_ms }}ms
          </template>
        </el-table-column>
        <el-table-column prop="has_error" label="状态" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.has_error" type="danger">失败</el-tag>
            <el-tag v-else type="success">成功</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleView(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
    
    <!-- 详情对话框 -->
    <el-dialog title="请求详情" v-model="detailVisible" width="800px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="请求ID">{{ currentLog.request_id }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatDate(currentLog.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ currentLog.user_name }} (ID: {{ currentLog.user_id }})</el-descriptions-item>
        <el-descriptions-item label="模型">{{ currentLog.model_name }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog.request_ip }}</el-descriptions-item>
        <el-descriptions-item label="延迟">{{ currentLog.latency_ms }}ms</el-descriptions-item>
        <el-descriptions-item label="Prompt Tokens">{{ currentLog.prompt_tokens }}</el-descriptions-item>
        <el-descriptions-item label="Completion Tokens">{{ currentLog.completion_tokens }}</el-descriptions-item>
        <el-descriptions-item label="Total Tokens">{{ currentLog.total_tokens }}</el-descriptions-item>
      </el-descriptions>
      
      <div class="detail-section">
        <h4>请求内容</h4>
        <pre class="code-block">{{ formatJSON(currentLog.request_body) }}</pre>
      </div>
      
      <div class="detail-section">
        <h4>响应内容</h4>
        <pre class="code-block">{{ formatJSON(currentLog.response_body) }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { getAuditLogs } from '@/api/audit'

export default {
  name: 'Audit',
  setup() {
    const loading = ref(false)
    const logs = ref([])
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    
    const searchForm = reactive({
      dateRange: [],
      user_id: '',
      model_name: ''
    })
    
    const detailVisible = ref(false)
    const currentLog = ref({})
    
    const loadLogs = async () => {
      loading.value = true
      try {
        const params = {
          page: page.value,
          page_size: pageSize.value
        }
        
        if (searchForm.dateRange && searchForm.dateRange.length === 2) {
          params.start_time = searchForm.dateRange[0]
          params.end_time = searchForm.dateRange[1]
        }
        if (searchForm.user_id) {
          params.user_id = searchForm.user_id
        }
        if (searchForm.model_name) {
          params.model_name = searchForm.model_name
        }
        
        const res = await getAuditLogs(params)
        logs.value = res.data
        total.value = res.total
      } catch (error) {
        console.error('Failed to load logs:', error)
      } finally {
        loading.value = false
      }
    }
    
    const resetSearch = () => {
      searchForm.dateRange = []
      searchForm.user_id = ''
      searchForm.model_name = ''
      loadLogs()
    }
    
    const handleView = (row) => {
      currentLog.value = row
      detailVisible.value = true
    }
    
    const handleSizeChange = (val) => {
      pageSize.value = val
      loadLogs()
    }
    
    const handlePageChange = (val) => {
      page.value = val
      loadLogs()
    }
    
    const formatDate = (date) => {
      if (!date) return '-'
      return new Date(date).toLocaleString()
    }
    
    const formatJSON = (str) => {
      if (!str) return ''
      try {
        return JSON.stringify(JSON.parse(str), null, 2)
      } catch {
        return str
      }
    }
    
    onMounted(() => {
      loadLogs()
    })
    
    return {
      loading,
      logs,
      page,
      pageSize,
      total,
      searchForm,
      detailVisible,
      currentLog,
      loadLogs,
      resetSearch,
      handleView,
      handleSizeChange,
      handlePageChange,
      formatDate,
      formatJSON
    }
  }
}
</script>

<style scoped>
.audit-logs {
  .card-header {
    font-weight: bold;
  }
  
  .search-form {
    margin-bottom: 20px;
  }
  
  .pagination {
    margin-top: 20px;
    text-align: right;
  }
  
  .detail-section {
    margin-top: 20px;
    
    h4 {
      margin-bottom: 10px;
      color: #606266;
    }
  }
  
  .code-block {
    background: #f5f7fa;
    padding: 15px;
    border-radius: 4px;
    overflow-x: auto;
    font-family: 'Courier New', monospace;
    font-size: 12px;
    max-height: 300px;
    overflow-y: auto;
  }
}
</style>
