// 认证相关工具函数
const auth = {
  // 保存token
  setToken(token) {
    wx.setStorageSync('access_token', token)
  },
  
  // 获取token
  getToken() {
    return wx.getStorageSync('access_token')
  },
  
  // 清除token
  clearToken() {
    wx.removeStorageSync('access_token')
    wx.removeStorageSync('user_info')
  },
  
  // 保存用户信息
  setUserInfo(userInfo) {
    wx.setStorageSync('user_info', userInfo)
  },
  
  // 获取用户信息
  getUserInfo() {
    return wx.getStorageSync('user_info')
  },
  
  // 检查是否已登录
  isLoggedIn() {
    return !!this.getToken()
  },
  
  // 微信登录
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            resolve(res.code)
          } else {
            reject(new Error('获取微信登录code失败'))
          }
        },
        fail: reject
      })
    })
  }
}

module.exports = auth