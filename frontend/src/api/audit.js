import request from '@/utils/request'

export function getAuditLogs(params) {
  return request({
    url: '/api/v1/admin/audit/logs',
    method: 'get',
    params
  })
}

export function getRiskEvents(params) {
  return request({
    url: '/api/v1/admin/audit/risk-events',
    method: 'get',
    params
  })
}

export function resolveRiskEvent(eventId, data) {
  return request({
    url: `/api/v1/admin/audit/risk-events/${eventId}/resolve`,
    method: 'post',
    data
  })
}

export function getUserStatistics(userId, params) {
  return request({
    url: `/api/v1/admin/audit/users/${userId}/statistics`,
    method: 'get',
    params
  })
}

export function getDashboardStats() {
  return request({
    url: '/api/v1/admin/audit/dashboard',
    method: 'get'
  })
}
