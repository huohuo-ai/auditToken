import request from '@/utils/request'

export function login(data) {
  return request({
    url: '/api/v1/auth/login',
    method: 'post',
    data
  })
}

export function register(data) {
  return request({
    url: '/api/v1/auth/register',
    method: 'post',
    data
  })
}

export function getProfile() {
  return request({
    url: '/api/v1/auth/profile',
    method: 'get'
  })
}

export function changePassword(data) {
  return request({
    url: '/api/v1/auth/change-password',
    method: 'post',
    data
  })
}

export function regenerateApiKey() {
  return request({
    url: '/api/v1/auth/regenerate-apikey',
    method: 'post'
  })
}
