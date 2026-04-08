import request from '@/utils/request'

export function getModelList(params) {
  return request({
    url: '/api/v1/admin/models',
    method: 'get',
    params
  })
}

export function createModel(data) {
  return request({
    url: '/api/v1/admin/models',
    method: 'post',
    data
  })
}

export function updateModel(id, data) {
  return request({
    url: `/api/v1/admin/models/${id}`,
    method: 'put',
    data
  })
}

export function deleteModel(id) {
  return request({
    url: `/api/v1/admin/models/${id}`,
    method: 'delete'
  })
}

export function getActiveModels() {
  return request({
    url: '/api/v1/models',
    method: 'get'
  })
}
