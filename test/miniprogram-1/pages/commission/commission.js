const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    commissions: [],
    summary: null,
    loading: true,
    selectedStatus: 0,
    selectedType: 0,
    statusOptions: [
      { value: '', label: '全部状态' },
      { value: 'pending', label: '待发放' },
      { value: 'paid', label: '已发放' },
      { value: 'cancelled', label: '已取消' }
    ],
    typeOptions: [
      { value: '', label: '全部类型' },
      { value: 'referral', label: '推荐佣金' },
      { value: 'agent', label: '代理佣金' }
    ]
  },

  onLoad() {
    this.loadCommissions()
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

  // 加载佣金记录
  async loadCommissions() {
    try {
      this.setData({ loading: true })
      
      const userInfo = auth.getUserInfo()
      if (userInfo && userInfo.openID) {
        const result = await api.getCommissions(
          userInfo.openID, 
          this.data.statusOptions[this.data.selectedStatus].value, 
          this.data.typeOptions[this.data.selectedType].value
        )
        
        if (result.code === 200) {
          this.setData({
            commissions: result.data.commissions || [],
            summary: result.data.summary || {}
          })
        }
      }
    } catch (error) {
      console.error('获取佣金记录失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 状态筛选
  onStatusChange(e) {
    this.setData({
      selectedStatus: e.detail.value
    })
    this.loadCommissions()
  },

  // 类型筛选
  onTypeChange(e) {
    this.setData({
      selectedType: e.detail.value
    })
    this.loadCommissions()
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.loadCommissions().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 格式化日期
  formatDate(dateString) {
    const date = new Date(dateString)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  },

  // 获取状态显示文本
  getStatusText(status) {
    const statusMap = {
      'pending': '待发放',
      'paid': '已发放',
      'cancelled': '已取消'
    }
    return statusMap[status] || status
  },

  // 获取类型显示文本
  getTypeText(type) {
    const typeMap = {
      'referral': '推荐佣金',
      'agent': '代理佣金'
    }
    return typeMap[type] || type
  }
})