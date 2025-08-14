const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    userInfo: null,
    isLoggedIn: false,
    testModules: [
      {
        name: '用户认证测试',
        desc: '微信登录、用户信息管理',
        icon: '👤',
        page: '/pages/login/login',
        requireAuth: false
      },
      {
        name: '学习进度测试',
        desc: '书籍、进度管理、单词学习',
        icon: '📚',
        page: '/pages/learning/learning',
        requireAuth: true
      },
      {
        name: '单词卡片测试',
        desc: '单词查询、卡片详情、图片发音',
        icon: '🃏',
        page: '/pages/wordcard/wordcard',
        requireAuth: true
      },
      {
        name: '商店功能测试',
        desc: '商品列表、购物车、订单管理',
        icon: '🛒',
        page: '/pages/store/store',
        requireAuth: true
      },
      {
        name: '推荐系统测试',
        desc: '推荐码验证、推荐关系管理',
        icon: '🎯',
        page: '/pages/referral/referral',
        requireAuth: true
      },
      {
        name: '佣金管理测试',
        desc: '佣金记录、返现统计',
        icon: '💰',
        page: '/pages/commission/commission',
        requireAuth: true
      },
      {
        name: '代理管理测试',
        desc: '用户管理、销售数据、佣金提取',
        icon: '👥',
        page: '/pages/agent/agent',
        requireAuth: true
      },
      {
        name: '管理员功能测试',
        desc: '用户等级管理、代理权限设置',
        icon: '⚙️',
        page: '/pages/admin/admin',
        requireAuth: true
      },
      {
        name: '二维码生成测试',
        desc: '小程序二维码生成、参数配置',
        icon: '📱',
        page: '/pages/qrcode/qrcode',
        requireAuth: true
      }
    ]
  },

  onLoad() {
    this.checkAuthStatus()
  },

  onShow() {
    this.checkAuthStatus()
  },

  // 检查登录状态
  checkAuthStatus() {
    const isLoggedIn = auth.isLoggedIn()
    const userInfo = auth.getUserInfo()
    
    this.setData({
      isLoggedIn,
      userInfo
    })
  },

  // 导航到功能页面
  navigateToTest(e) {
    const { page, requireAuth, name } = e.currentTarget.dataset.module
    
    // 检查是否需要登录
    if (requireAuth && !this.data.isLoggedIn) {
      wx.showModal({
        title: '需要登录',
        content: `访问"${name}"需要先登录，是否前往登录页面？`,
        success: (res) => {
          if (res.confirm) {
            wx.navigateTo({
              url: '/pages/login/login'
            })
          }
        }
      })
      return
    }

    // 导航到对应页面
    if (page.includes('/pages/login/') || page.includes('/pages/referral/') || page.includes('/pages/commission/')) {
      // 这些页面已在tabBar中，使用switchTab
      wx.switchTab({
        url: page
      })
    } else {
      wx.navigateTo({
        url: page
      })
    }
  },

  // 用户信息管理
  manageUserInfo() {
    if (!this.data.isLoggedIn) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }
    
    wx.showActionSheet({
      itemList: ['查看用户信息', '更新用户信息', '退出登录'],
      success: (res) => {
        switch (res.tapIndex) {
          case 0:
            this.showUserInfo()
            break
          case 1:
            this.updateUserInfo()
            break
          case 2:
            this.logout()
            break
        }
      }
    })
  },

  // 显示用户信息
  async showUserInfo() {
    try {
      const userInfo = auth.getUserInfo()
      const result = await api.getUserInfo(userInfo.openID)
      
      if (result.code === 200) {
        const user = result.data
        wx.showModal({
          title: '用户信息',
          content: `姓名: ${user.user_name || '未设置'}\n学校: ${user.school || '未设置'}\n城市: ${user.city || '未设置'}\n代理等级: ${user.agent_level || 0}\n推荐码: ${user.referral_code || '无'}`,
          showCancel: false
        })
      }
    } catch (error) {
      wx.showToast({
        title: '获取信息失败',
        icon: 'none'
      })
    }
  },

  // 更新用户信息
  updateUserInfo() {
    wx.showModal({
      title: '更新用户信息',
      content: '用户信息更新功能需要额外的表单页面，当前演示版本暂未实现完整的用户信息编辑界面。可以通过用户认证测试模块进行相关测试。',
      showCancel: false
    })
  },

  // 退出登录
  logout() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          auth.clearToken()
          this.setData({
            isLoggedIn: false,
            userInfo: null
          })
          wx.showToast({
            title: '已退出登录',
            icon: 'success'
          })
        }
      }
    })
  },

  // 快速登录
  quickLogin() {
    wx.switchTab({
      url: '/pages/login/login'
    })
  },

  // 查看API文档
  viewApiDocs() {
    wx.showModal({
      title: 'API接口文档',
      content: '本测试小程序包含完整的后端API测试功能，涵盖用户管理、学习系统、商城、推荐、代理等模块。详细接口文档请查看项目根目录的API_ROUTES.md文件。',
      showCancel: false
    })
  },

  // 关于应用
  aboutApp() {
    wx.showModal({
      title: '关于应用',
      content: '这是一个微信小程序英语学习平台的完整测试工具，包含前端界面和后端API的全面测试功能。\n\n主要功能：\n• 用户认证与管理\n• 学习进度跟踪\n• 单词卡片系统\n• 商城购物功能\n• 推荐返现系统\n• 代理管理系统\n• 管理员后台',
      showCancel: false
    })
  }
})