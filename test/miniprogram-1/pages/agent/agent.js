const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    agentUsers: [],
    salesData: null,
    withdrawHistory: [],
    loading: false,
    currentTab: 0, // 0: 用户管理, 1: 销售数据, 2: 佣金提取
    tabList: ['用户管理', '销售数据', '佣金提取'],
    filters: {
      school: '',
      region: '',
      startDate: '',
      endDate: ''
    },
    withdrawForm: {
      amount: '',
      withdrawMethod: 'wechat',
      accountInfo: {
        accountName: '',
        accountNumber: ''
      }
    },
    withdrawMethods: [
      { value: 'wechat', label: '微信' },
      { value: 'alipay', label: '支付宝' },
      { value: 'bank_transfer', label: '银行卡' }
    ]
  },

  onLoad() {
    this.loadAgentUsers()
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

  // 切换标签
  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({ currentTab: tab })
    
    switch (tab) {
      case 0:
        this.loadAgentUsers()
        break
      case 1:
        this.loadSalesData()
        break
      case 2:
        this.loadWithdrawHistory()
        break
    }
  },

  // 加载代理管理的用户
  async loadAgentUsers() {
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
      
      const result = await api.getAgentUsers(
        userInfo.openID, 
        this.data.filters.school, 
        this.data.filters.region
      )
      
      if (result.code === 200) {
        this.setData({
          agentUsers: result.data.users || []
        })
      } else {
        wx.showToast({
          title: result.message || '获取用户失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('获取代理用户失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 加载销售数据
  async loadSalesData() {
    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) return

    try {
      this.setData({ loading: true })
      
      const result = await api.getAgentSales(
        userInfo.openID,
        this.data.filters.startDate,
        this.data.filters.endDate
      )
      
      if (result.code === 200) {
        this.setData({
          salesData: result.data
        })
      } else {
        wx.showToast({
          title: result.message || '获取销售数据失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('获取销售数据失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 加载提现历史 (模拟数据)
  loadWithdrawHistory() {
    // 由于后端没有提供获取提现历史的接口，这里使用模拟数据
    const mockHistory = [
      {
        id: 'WD20241201001',
        amount: 500.00,
        method: '微信',
        status: 'completed',
        createdAt: '2024-12-01',
        completedAt: '2024-12-02'
      },
      {
        id: 'WD20241115002',
        amount: 300.00,
        method: '支付宝',
        status: 'processing',
        createdAt: '2024-11-15'
      }
    ]
    
    this.setData({
      withdrawHistory: mockHistory
    })
  },

  // 筛选条件输入
  onSchoolInput(e) {
    this.setData({
      'filters.school': e.detail.value
    })
  },

  onRegionInput(e) {
    this.setData({
      'filters.region': e.detail.value
    })
  },

  onStartDateChange(e) {
    this.setData({
      'filters.startDate': e.detail.value
    })
  },

  onEndDateChange(e) {
    this.setData({
      'filters.endDate': e.detail.value
    })
  },

  // 应用筛选
  applyFilters() {
    if (this.data.currentTab === 0) {
      this.loadAgentUsers()
    } else if (this.data.currentTab === 1) {
      this.loadSalesData()
    }
  },

  // 清空筛选
  clearFilters() {
    this.setData({
      filters: {
        school: '',
        region: '',
        startDate: '',
        endDate: ''
      }
    })
    this.applyFilters()
  },

  // 提现表单输入
  onWithdrawAmountInput(e) {
    this.setData({
      'withdrawForm.amount': e.detail.value
    })
  },

  onWithdrawMethodChange(e) {
    const method = this.data.withdrawMethods[e.detail.value].value
    this.setData({
      'withdrawForm.withdrawMethod': method
    })
  },

  onAccountNameInput(e) {
    this.setData({
      'withdrawForm.accountInfo.accountName': e.detail.value
    })
  },

  onAccountNumberInput(e) {
    this.setData({
      'withdrawForm.accountInfo.accountNumber': e.detail.value
    })
  },

  // 提交提现申请
  async submitWithdraw() {
    const { withdrawForm } = this.data
    const userInfo = auth.getUserInfo()
    
    if (!userInfo || !userInfo.openID) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    // 验证表单
    if (!withdrawForm.amount || parseFloat(withdrawForm.amount) <= 0) {
      wx.showToast({
        title: '请输入有效金额',
        icon: 'none'
      })
      return
    }

    if (!withdrawForm.accountInfo.accountName || !withdrawForm.accountInfo.accountNumber) {
      wx.showToast({
        title: '请填写完整账户信息',
        icon: 'none'
      })
      return
    }

    try {
      wx.showLoading({ title: '提交中...' })
      
      const result = await api.withdrawCommission(userInfo.openID, {
        amount: parseFloat(withdrawForm.amount),
        withdraw_method: withdrawForm.withdrawMethod,
        account_info: withdrawForm.accountInfo
      })

      if (result.code === 200) {
        wx.showToast({
          title: '提现申请提交成功',
          icon: 'success'
        })
        
        // 清空表单
        this.setData({
          withdrawForm: {
            amount: '',
            withdrawMethod: 'wechat',
            accountInfo: {
              accountName: '',
              accountNumber: ''
            }
          }
        })
        
        // 刷新提现历史
        this.loadWithdrawHistory()
      } else {
        throw new Error(result.message)
      }
    } catch (error) {
      console.error('提现申请失败:', error)
      wx.showToast({
        title: error.message || '提交失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 格式化日期
  formatDate(dateString) {
    const date = new Date(dateString)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  },

  // 获取状态文本
  getStatusText(status) {
    const statusMap = {
      'pending': '待处理',
      'processing': '处理中',
      'completed': '已完成',
      'rejected': '已拒绝'
    }
    return statusMap[status] || status
  },

  // 下拉刷新
  onPullDownRefresh() {
    const refreshActions = [
      () => this.loadAgentUsers(),
      () => this.loadSalesData(),
      () => this.loadWithdrawHistory()
    ]
    
    refreshActions[this.data.currentTab]().then(() => {
      wx.stopPullDownRefresh()
    })
  }
})