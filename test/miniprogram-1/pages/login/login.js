const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    loading: false,
    userInfo: null,
    hasUserInfo: false
  },

  onLoad() {
    // 检查是否已登录
    if (auth.isLoggedIn()) {
      wx.switchTab({
        url: '/pages/profile/profile'
      })
    }
  },

  // 微信一键登录
  async wxLogin() {
    try {
      this.setData({ loading: true })
      
      // 获取微信登录code
      const code = await auth.wxLogin()
      
      // 调用后端登录接口
      const result = await api.wechatLogin(code)
      
      if (result.code === 200) {
        // 保存token和用户信息
        auth.setToken(result.data.token)
        auth.setUserInfo(result.data.user)
        
        wx.showToast({
          title: '登录成功',
          icon: 'success'
        })
        
        // 跳转到主页
        setTimeout(() => {
          wx.switchTab({
            url: '/pages/profile/profile'
          })
        }, 1000)
      } else {
        throw new Error(result.message || '登录失败')
      }
    } catch (error) {
      console.error('登录失败:', error)
      wx.showToast({
        title: error.message || '登录失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 获取用户信息授权
  getUserProfile() {
    wx.getUserProfile({
      desc: '用于完善用户资料',
      success: (res) => {
        this.setData({
          userInfo: res.userInfo,
          hasUserInfo: true
        })
      },
      fail: (err) => {
        console.error('获取用户信息失败:', err)
      }
    })
  }
})