import Cookies from 'js-cookie'

const TokenKey = 'ai_gateway_token'

export function getToken() {
  return Cookies.get(TokenKey) || localStorage.getItem(TokenKey)
}

export function setToken(token) {
  // 同时存到cookie和localStorage
  Cookies.set(TokenKey, token, { expires: 1 })
  localStorage.setItem(TokenKey, token)
}

export function removeToken() {
  Cookies.remove(TokenKey)
  localStorage.removeItem(TokenKey)
}
