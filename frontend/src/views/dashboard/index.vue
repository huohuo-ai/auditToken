<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon" style="background: #409EFF;">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.today_requests || 0 }}</div>
              <div class="stat-label">今日请求数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon" style="background: #67C23A;">
              <el-icon><DataLine /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatNumber(stats.today_tokens) || 0 }}</div>
              <div class="stat-label">今日Token使用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon" style="background: #E6A23C;">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.active_users || 0 }}</div>
              <div class="stat-label">活跃用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon" style="background: #F56C6C;">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.risk_events || 0 }}</div>
              <div class="stat-label">风险事件</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 趋势图表 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>使用趋势（近7天）</span>
            </div>
          </template>
          <div ref="trendChart" style="height: 350px;"></div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>模型使用排行</span>
            </div>
          </template>
          <div ref="modelChart" style="height: 350px;"></div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- API使用说明 -->
    <el-card class="api-doc">
      <template #header>
        <div class="card-header">
          <span>API使用说明</span>
        </div>
      </template>
      <div class="api-content">
        <h4>API Base URL</h4>
        <el-input v-model="apiBaseUrl" readonly>
          <template #append>
            <el-button @click="copyText(apiBaseUrl)">复制</el-button>
          </template>
        </el-input>
        
        <h4 style="margin-top: 20px;">您的 API Key</h4>
        <el-input v-model="apiKey" type="password" show-password readonly>
          <template #append>
            <el-button @click="copyText(apiKey)">复制</el-button>
          </template>
        </el-input>
        
        <h4 style="margin-top: 20px;">示例代码</h4>
        <pre class="code-block"><code>{{ exampleCode }}</code></pre>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { useStore } from 'vuex'
import * as echarts from 'echarts'
import { getDashboardStats } from '@/api/audit'
import { ElMessage } from 'element-plus'

export default {
  name: 'Dashboard',
  setup() {
    const store = useStore()
    const trendChart = ref(null)
    const modelChart = ref(null)
    
    const stats = ref({})
    const apiBaseUrl = window.location.origin + '/v1'
    const apiKey = computed(() => store.state.user.api_key || '')
    
    const exampleCode = `curl ${window.location.origin}/v1/chat/completions \\
  -H "Authorization: Bearer ${apiKey.value || 'YOUR_API_KEY'}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`
    
    const formatNumber = (num) => {
      if (!num) return 0
      if (num >= 10000) {
        return (num / 10000).toFixed(1) + '万'
      }
      return num.toLocaleString()
    }
    
    const copyText = (text) => {
      navigator.clipboard.writeText(text).then(() => {
        ElMessage.success('已复制到剪贴板')
      })
    }
    
    const initTrendChart = (data) => {
      if (!trendChart.value) return
      
      const chart = echarts.init(trendChart.value)
      const option = {
        tooltip: {
          trigger: 'axis'
        },
        legend: {
          data: ['请求数', 'Token使用量']
        },
        xAxis: {
          type: 'category',
          data: data.map(item => item.date)
        },
        yAxis: [
          {
            type: 'value',
            name: '请求数'
          },
          {
            type: 'value',
            name: 'Token数'
          }
        ],
        series: [
          {
            name: '请求数',
            type: 'line',
            data: data.map(item => item.requests),
            smooth: true
          },
          {
            name: 'Token使用量',
            type: 'line',
            yAxisIndex: 1,
            data: data.map(item => item.tokens),
            smooth: true
          }
        ]
      }
      chart.setOption(option)
    }
    
    const initModelChart = (data) => {
      if (!modelChart.value || !data) return
      
      const chart = echarts.init(modelChart.value)
      const option = {
        tooltip: {
          trigger: 'item'
        },
        legend: {
          orient: 'vertical',
          right: 10,
          top: 'center'
        },
        series: [
          {
            type: 'pie',
            radius: ['40%', '70%'],
            avoidLabelOverlap: false,
            itemStyle: {
              borderRadius: 10,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: false
            },
            data: data.map(item => ({
              name: item.model_name,
              value: item.requests
            }))
          }
        ]
      }
      chart.setOption(option)
    }
    
    const loadStats = async () => {
      try {
        const res = await getDashboardStats()
        stats.value = res
        
        if (res.trends) {
          initTrendChart(res.trends)
        }
        if (res.model_stats) {
          initModelChart(res.model_stats)
        }
      } catch (error) {
        console.error('Failed to load stats:', error)
      }
    }
    
    onMounted(() => {
      loadStats()
    })
    
    return {
      trendChart,
      modelChart,
      stats,
      apiBaseUrl,
      apiKey,
      exampleCode,
      formatNumber,
      copyText
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stat-row {
  margin-bottom: 20px;
}

.stat-card {
  .stat-item {
    display: flex;
    align-items: center;
  }
  
  .stat-icon {
    width: 60px;
    height: 60px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 30px;
    color: #fff;
    margin-right: 15px;
  }
  
  .stat-value {
    font-size: 24px;
    font-weight: bold;
    color: #303133;
  }
  
  .stat-label {
    font-size: 14px;
    color: #909399;
    margin-top: 5px;
  }
}

.chart-row {
  margin-bottom: 20px;
}

.card-header {
  font-weight: bold;
}

.api-doc {
  margin-top: 20px;
}

.api-content h4 {
  margin-bottom: 10px;
  color: #606266;
}

.code-block {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #333;
}
</style>
