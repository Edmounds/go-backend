// app.js
App({
  onLaunch() {
    // 展示本地存储能力
    const logs = wx.getStorageSync('logs') || []
    logs.unshift(Date.now())
    wx.setStorageSync('logs', logs)

    // 登录
    wx.login({
      success: res => {
        console.log(res)
        wx.request({
          url: 'https://backend.edmounds.top/api/auth', // 或你的服务器地址
          method: 'POST',
          header: {
            'Content-Type': 'application/json'
          },
          data: {
            code: res.code // 从wx.login获取的code
          },
          success: function(res) {
            console.log('登录响应:', res.data);
            // 这里应该能看到 session_key, openid, unionid
          },
          fail: function(err) {
            console.error('登录失败:', err);
          }
        });
        // 发送 res.code 到后台换取 openId, sessionKey, unionId
      }
    })
  },
  globalData: {
    userInfo: null
  }
})
