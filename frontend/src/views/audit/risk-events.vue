<template>
  <div class="risk-events">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>风险事件</span>
          <el-radio-group v-model="filterLevel" size="small" @change="loadEvents">
            <el-radio-button label="">全部</el-radio-button>
            <el-radio-button label="critical">严重</el-radio-button>
            <el-radio-button label="high">高危</el-radio-button>
            <el-radio-button label="medium">中危</el-radio-button>
            <el-radio-button label="low">低危</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      
      <el-table :data="events" v-loading="loading" stripe>
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="risk_level" label="风险等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getRiskLevelType(row.risk_level)">{{ getRiskLevelText(row.risk_level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="risk_type" label="风险类型" width="120">
          <template #default="{ row }">
            {{ getRiskTypeText(row.risk_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="user_name" label="用户" width="100" />
        <el-table-column prop="request_ip" label="IP地址" width="120" />
        <el-table-column prop="risk_reason" label="风险原因" show-overflow-tooltip />
        <el-table-column prop="is_resolved" label="状态" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.is_resolved" type="success">已处理</el-tag>
            <el-tag v-else type="warning">待处理</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleView(row)">详情</el-button>
            <el-button v-if="!row.is_resolved" size="small" type="primary" @click="handleResolve(row)">处理</el-button>
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
    <el-dialog title="风险事件详情" v-model="detailVisible" width="700px">
      <el-alert
        :title="currentEvent.risk_reason"
        :type="getRiskLevelType(currentEvent.risk_level)"
        :closable="false"
        show-icon
        style="margin-bottom: 20px;"
      />
      
      <el-descriptions :column="2" border>
        <el-descriptions-item label="事件ID">{{ currentEvent.event_id }}</el-descriptions-item>
        <el-descriptions-item label="请求ID">{{ currentEvent.request_id }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatDate(currentEvent.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="风险评分">{{ currentEvent.risk_score }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ currentEvent.user_name }} (ID: {{ currentEvent.user_id }})</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentEvent.request_ip }}</el-descriptions-item>
        <el-descriptions-item label="模型">{{ currentEvent.model_name }}</el-descriptions-item>
      </el-descriptions>
      
      <div class="detail-section">
        <h4>风险描述</h4>
        <p>{{ currentEvent.description }}</p>
      </div>
      
      <div class="detail-section">
        <h4>证据</h4>
        <pre class="code-block">{{ formatJSON(currentEvent.evidence) }}</pre>
      </div>
    </el-dialog>
    
    <!-- 处理对话框 -->
    <el-dialog title="处理风险事件" v-model="resolveVisible" width="400px">
      <el-form :model="resolveForm">
        <el-form-item label="处理备注">
          <el-input v-model="resolveForm.note" type="textarea" :rows="4" placeholder="请输入处理备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resolveVisible = false">取消</el-button>
        <el-button type="primary" @click="handleResolveSubmit" :loading="resolveLoading">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getRiskEvents, resolveRiskEvent } from '@/api/audit'

export default {
  name: 'RiskEvents',
  setup() {
    const loading = ref(false)
    const events = ref([])
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    const filterLevel = ref('')
    
    const detailVisible = ref(false)
    const currentEvent = ref({})
    
    const resolveVisible = ref(false)
    const resolveLoading = ref(false)
    const currentEventId = ref('')
    const resolveForm = reactive({
      note: ''
    })
    
    const riskLevelMap = {
      critical: { text: '严重', type: 'danger' },
      high: { text: '高危', type: 'danger' },
      medium: { text: '中危', type: 'warning' },
      low: { text: '低危', type: 'info' }
    }
    
    const riskTypeMap = {
      token_abuse: 'Token滥用',
      off_hours_access: '非工作时间访问',
      sensitive_info: '敏感信息获取',
      abnormal_frequency: '异常请求频率',
      ip_anomaly: 'IP异常',
      abnormal_pattern: '异常请求模式',
      model_abuse: '模型滥用',
      multiple: '多种风险'
    }
    
    const loadEvents = async () => {
      loading.value = true
      try {
        const params = {
          page: page.value,
          page_size: pageSize.value,
          risk_level: filterLevel.value
        }
        
        const res = await getRiskEvents(params)
        events.value = res.data
        total.value = res.total
      } catch (error) {
        console.error('Failed to load events:', error)
      } finally {
        loading.value = false
      }
    }
    
    const handleView = (row) => {
      currentEvent.value = row
      detailVisible.value = true
    }
    
    const handleResolve = (row) => {
      currentEventId.value = row.event_id
      resolveForm.note = ''
      resolveVisible.value = true
    }
    
    const handleResolveSubmit = async () => {
      resolveLoading.value = true
      try {
        await resolveRiskEvent(currentEventId.value, { note: resolveForm.note })
        ElMessage.success('处理成功')
        resolveVisible.value = false
        loadEvents()
      } catch (error) {
        console.error('Failed to resolve event:', error)
      } finally {
        resolveLoading.value = false
      }
    }
    
    const handleSizeChange = (val) => {
      pageSize.value = val
      loadEvents()
    }
    
    const handlePageChange = (val) => {
      page.value = val
      loadEvents()
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
    
    const getRiskLevelType = (level) => {
      return riskLevelMap[level]?.type || 'info'
    }
    
    const getRiskLevelText = (level) => {
      return riskLevelMap[level]?.text || level
    }
    
    const getRiskTypeText = (type) => {
      return riskTypeMap[type] || type
    }
    
    onMounted(() => {
      loadEvents()
    })
    
    return {
      loading,
      events,
      page,
      pageSize,
      total,
      filterLevel,
      detailVisible,
      currentEvent,
      resolveVisible,
      resolveLoading,
      resolveForm,
      loadEvents,
      handleView,
      handleResolve,
      handleResolveSubmit,
      handleSizeChange,
      handlePageChange,
      formatDate,
      formatJSON,
      getRiskLevelType,
      getRiskLevelText,
      getRiskTypeText
    }
  }
}
</script>

<style scoped>
.risk-events {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
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
