const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    testUserId: '',
    agentLevel: 0,
    updateResult: null,
    loading: false,
    agentLevels: [
      { value: 0, label: '普通用户' },
      { value: 1, label: '校代理' },
      { value: 2, label: '区域代理' }
    ],
    testUsers: [
      {
        openID: 'test_user_001',
        name: '测试用户1',
        currentLevel: 0,
        school: '北京大学'
      },
      {
        openID: 'test_user_002', 
        name: '测试用户2',
        currentLevel: 1,
        school: '清华大学'
      },
      {
        openID: 'test_user_003',
        name: '测试用户3', 
        currentLevel: 0,
        school: '复旦大学'
      }
    ],
    operationLog: []
  },

  onLoad() {
    // 页面加载完成
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

  // 输入用户ID
  onUserIdInput(e) {
    this.setData({
      testUserId: e.detail.value
    })
  },

  // 选择代理等级
  onAgentLevelChange(e) {
    this.setData({
      agentLevel: parseInt(e.detail.value)
    })
  },

  // 使用预设用户
  usePresetUser(e) {
    const user = e.currentTarget.dataset.user
    this.setData({
      testUserId: user.openID,
      agentLevel: user.currentLevel === 2 ? 0 : user.currentLevel + 1
    })
  },

  // 更新用户代理等级
  async updateAgentLevel() {
    const { testUserId, agentLevel } = this.data

    if (!testUserId.trim()) {
      wx.showToast({
        title: '请输入用户ID',
        icon: 'none'
      })
      return
    }

    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    try {
      this.setData({ loading: true })
      wx.showLoading({ title: '更新中...' })

      const result = await api.updateAgentLevel(testUserId, agentLevel)

      if (result.code === 200) {
        const updateResult = {
          success: true,
          message: '更新成功',
          data: result.data,
          timestamp: new Date().toLocaleString()
        }

        this.setData({
          updateResult
        })

        // 添加到操作日志
        this.addToOperationLog({
          action: '更新代理等级',
          userId: testUserId,
          level: agentLevel,
          result: '成功',
          timestamp: updateResult.timestamp
        })

        wx.showToast({
          title: '更新成功',
          icon: 'success'
        })

        // 更新预设用户的等级显示
        this.updatePresetUserLevel(testUserId, agentLevel)

      } else {
        throw new Error(result.message || '更新失败')
      }
    } catch (error) {
      console.error('更新代理等级失败:', error)
      
      const updateResult = {
        success: false,
        message: error.message || '更新失败',
        timestamp: new Date().toLocaleString()
      }

      this.setData({
        updateResult
      })

      // 添加到操作日志
      this.addToOperationLog({
        action: '更新代理等级',
        userId: testUserId,
        level: agentLevel,
        result: '失败: ' + updateResult.message,
        timestamp: updateResult.timestamp
      })

      wx.showToast({
        title: error.message || '更新失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
      wx.hideLoading()
    }
  },

  // 更新预设用户的等级显示
  updatePresetUserLevel(openID, newLevel) {
    const testUsers = this.data.testUsers.map(user => {
      if (user.openID === openID) {
        return { ...user, currentLevel: newLevel }
      }
      return user
    })
    this.setData({ testUsers })
  },

  // 添加到操作日志
  addToOperationLog(logEntry) {
    const operationLog = [logEntry, ...this.data.operationLog]
    // 最多保留20条记录
    if (operationLog.length > 20) {
      operationLog.splice(20)
    }
    this.setData({ operationLog })
  },

  // 清空操作日志
  clearOperationLog() {
    wx.showModal({
      title: '确认清空',
      content: '确定要清空所有操作日志吗？',
      success: (res) => {
        if (res.confirm) {
          this.setData({ operationLog: [] })
          wx.showToast({
            title: '已清空日志',
            icon: 'success'
          })
        }
      }
    })
  },

  // 清空表单
  clearForm() {
    this.setData({
      testUserId: '',
      agentLevel: 0,
      updateResult: null
    })
  },

  // 复制用户ID
  copyUserId(e) {
    const userId = e.currentTarget.dataset.userId
    wx.setClipboardData({
      data: userId,
      success: () => {
        wx.showToast({
          title: '已复制用户ID',
          icon: 'success'
        })
      }
    })
  },

  // 获取代理等级文本
  getAgentLevelText(level) {
    const levelMap = {
      0: '普通用户',
      1: '校代理',
      2: '区域代理'
    }
    return levelMap[level] || '未知'
  },

  // 批量测试
  async batchTest() {
    wx.showModal({
      title: '批量测试',
      content: '将对所有预设用户进行代理等级测试，确定继续吗？',
      success: async (res) => {
        if (res.confirm) {
          await this.performBatchTest()
        }
      }
    })
  },

  // 执行批量测试
  async performBatchTest() {
    wx.showLoading({ title: '批量测试中...' })
    
    let successCount = 0
    let failCount = 0

    for (const user of this.data.testUsers) {
      try {
        // 测试升级用户等级
        const newLevel = user.currentLevel < 2 ? user.currentLevel + 1 : 0
        const result = await api.updateAgentLevel(user.openID, newLevel)
        
        if (result.code === 200) {
          successCount++
          this.updatePresetUserLevel(user.openID, newLevel)
          this.addToOperationLog({
            action: '批量测试',
            userId: user.openID,
            level: newLevel,
            result: '成功',
            timestamp: new Date().toLocaleString()
          })
        } else {
          failCount++
        }
        
        // 添加延迟避免请求过快
        await new Promise(resolve => setTimeout(resolve, 1000))
        
      } catch (error) {
        failCount++
        this.addToOperationLog({
          action: '批量测试',
          userId: user.openID,
          level: 'N/A',
          result: '失败: ' + error.message,
          timestamp: new Date().toLocaleString()
        })
      }
    }

    wx.hideLoading()
    wx.showModal({
      title: '批量测试完成',
      content: `成功: ${successCount}个\n失败: ${failCount}个`,
      showCancel: false
    })
  }
})