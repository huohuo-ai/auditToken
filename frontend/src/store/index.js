import { createStore } from 'vuex'
import { login, getProfile } from '@/api/auth'
import { setToken, removeToken } from '@/utils/auth'

export default createStore({
  state: {
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || '{}'),
    sidebarCollapsed: false
  },
  
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    SET_USER(state, user) {
      state.user = user
      localStorage.setItem('user', JSON.stringify(user))
      localStorage.setItem('userRole', user.role)
    },
    CLEAR_USER(state) {
      state.token = ''
      state.user = {}
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('userRole')
    },
    TOGGLE_SIDEBAR(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed
    }
  },
  
  actions: {
    // 登录
    async login({ commit }, credentials) {
      const response = await login(credentials)
      const { token, user } = response.data
      setToken(token)
      commit('SET_TOKEN', token)
      commit('SET_USER', user)
      return response
    },
    
    // 获取用户信息
    async getProfile({ commit }) {
      const response = await getProfile()
      commit('SET_USER', response.data)
      return response
    },
    
    // 退出登录
    logout({ commit }) {
      removeToken()
      commit('CLEAR_USER')
    }
  },
  
  getters: {
    isLoggedIn: state => !!state.token,
    isAdmin: state => state.user.role === 'admin',
    userName: state => state.user.username || '',
    userRole: state => state.user.role || ''
  }
})
