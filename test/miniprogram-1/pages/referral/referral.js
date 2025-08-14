const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    referralCode: '',
    validateResult: null,
    validating: false,
    userReferralInfo: null,
    loading: true,
    basePrice: '',
    finalPrice: '',
    discountRate: 0
  },

  onLoad() {
    this.loadUserReferralInfo()
  },

  onShow() {
    // 检查登录状态
    if (!auth.isLoggedIn()) {
      wx.redirectTo({
        url: '/pages/login/login'
      })
      return
    }
  },

  // 加载用户推荐信息
  async loadUserReferralInfo() {
    try {
      const userInfo = auth.getUserInfo()
      if (userInfo && userInfo.openID) {
        const result = await api.getUserReferralInfo(userInfo.openID)
        if (result.code === 200) {
          this.setData({
            userReferralInfo: result.data
          })
        }
      }
    } catch (error) {
      console.error('获取推荐信息失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  // 输入推荐码
  onReferralCodeInput(e) {
    const value = e.detail.value.trim()
    this.setData({
      referralCode: value,
      validateResult: null,
      finalPrice: '',  // 清除之前的价格计算结果
      discountRate: 0
    })
  },

  // 验证推荐码
  async validateReferralCode() {
    const { referralCode } = this.data
    
    if (!referralCode.trim()) {
      wx.showToast({
        title: '请输入推荐码',
        icon: 'none'
      })
      return
    }

    try {
      this.setData({ validating: true })
      
      const result = await api.validateReferralCode(referralCode)
      
      this.setData({
        validateResult: result,
        discountRate: result && result.data ? (result.data.discount_rate || 0) : 0
      })
      
      if (result.code === 200) {
        wx.showToast({
          title: '推荐码有效',
          icon: 'success'
        })
      } else {
        wx.showToast({
          title: result.message || '推荐码无效',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('验证推荐码失败:', error)
      wx.showToast({
        title: error.message || '验证失败',
        icon: 'none'
      })
    } finally {
      this.setData({ validating: false })
    }
  },

  // 复制推荐码
  copyReferralCode() {
    const { userReferralInfo } = this.data
    if (userReferralInfo && userReferralInfo.referral_code) {
      wx.setClipboardData({
        data: userReferralInfo.referral_code,
        success: () => {
          wx.showToast({
            title: '已复制推荐码',
            icon: 'success'
          })
        }
      })
    }
  },

  // 分享推荐码
  onShareAppMessage() {
    const { userReferralInfo } = this.data
    return {
      title: '邀请你使用我的推荐码',
      path: `/pages/referral/referral?referral_code=${userReferralInfo.referral_code}`,
      imageUrl: '/images/share.png'
    }
  },

  // 输入基础价格
  onBasePriceInput(e) {
    this.setData({ basePrice: e.detail.value })
  },

  // 模拟计算价格
  simulatePrice() {
    const { basePrice, discountRate } = this.data
    const price = parseFloat(basePrice)
    if (isNaN(price) || price <= 0) {
      wx.showToast({ title: '请输入有效价格', icon: 'none' })
      return
    }
    const finalPrice = (price * (1 - (discountRate || 0))).toFixed(2)
    this.setData({ finalPrice })
  },

  // 绑定推荐关系（测试）
  async bindReferral() {
    const userInfo = auth.getUserInfo()
    const { referralCode } = this.data
    if (!userInfo || !userInfo.openID) {
      wx.showToast({ title: '未登录', icon: 'none' })
      return
    }
    if (!referralCode) {
      wx.showToast({ title: '请先输入推荐码', icon: 'none' })
      return
    }
    try {
      const res = await api.trackReferral(referralCode, userInfo.openID)
      if (res && res.code === 200) {
        wx.showToast({ title: '绑定成功', icon: 'success' })
        this.loadUserReferralInfo()
      } else {
        wx.showToast({ title: res.message || '绑定失败', icon: 'none' })
      }
    } catch (error) {
      wx.showToast({ title: error.message || '绑定失败', icon: 'none' })
    }
  }
})