import request from '@/utils/request'

export function getUserList(params) {
  return request({
    url: '/api/v1/admin/users',
    method: 'get',
    params
  })
}

export function createUser(data) {
  return request({
    url: '/api/v1/admin/users',
    method: 'post',
    data
  })
}

export function updateUser(id, data) {
  return request({
    url: `/api/v1/admin/users/${id}`,
    method: 'put',
    data
  })
}

export function deleteUser(id) {
  return request({
    url: `/api/v1/admin/users/${id}`,
    method: 'delete'
  })
}

export function resetPassword(id, newPassword) {
  return request({
    url: `/api/v1/admin/users/${id}/reset-password`,
    method: 'post',
    data: { new_password: newPassword }
  })
}

export function getUserQuota(id) {
  return request({
    url: `/api/v1/admin/users/${id}/quota`,
    method: 'get'
  })
}

export function updateUserQuota(id, data) {
  return request({
    url: `/api/v1/admin/users/${id}/quota`,
    method: 'put',
    data
  })
}
