const config = require('./config.js')
const auth = require('./auth.js')

// 基础请求函数
function request(url, options = {}) {
  const token = auth.getToken()
  
  return new Promise((resolve, reject) => {
    wx.request({
      url: config.apiBaseUrl + url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : '',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode === 200) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          // Token过期，跳转到登录页
          auth.clearToken()
          wx.redirectTo({
            url: '/pages/login/login'
          })
          reject(new Error('登录已过期'))
        } else {
          reject(new Error(res.data.message || '请求失败'))
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

// API方法
const api = {
  // ===== 认证相关 =====
  // 微信登录
  wechatLogin(code) {
    return request('/auth', {
      method: 'POST',
      data: { code }
    })
  },

  // ===== 用户管理相关 =====
  // 创建用户
  createUser(userData) {
    return request('/users', {
      method: 'POST',
      data: userData
    })
  },

  // 获取用户信息
  getUserInfo(userId) {
    return request(`/users/${userId}`)
  },

  // 更新用户信息
  updateUserInfo(userId, userData) {
    return request(`/users/${userId}`, {
      method: 'PUT',
      data: userData
    })
  },

  // ===== 学习进度相关 =====
  // 获取书籍列表
  getBooks(page = 1, limit = 20) {
    return request(`/books?page=${page}&limit=${limit}`)
  },

  // 获取书籍单词
  getBookWords(bookId, unitId = '') {
    let url = `/books/${bookId}/words`
    if (unitId) url += `?unit_id=${unitId}`
    return request(url)
  },

  // 获取用户学习进度
  getUserProgress(userId) {
    return request(`/users/${userId}/progress`)
  },

  // 更新用户学习进度
  updateUserProgress(userId, progressData) {
    return request(`/users/${userId}/progress`, {
      method: 'PUT',
      data: progressData
    })
  },

  // ===== 单词卡片相关 =====
  // 根据单元ID获取单词列表
  getUnitWords(unitId) {
    return request(`/units/${unitId}/words`)
  },

  // 获取单词卡片详情
  getWordCard(wordName) {
    return request(`/words/${wordName}/card`)
  },

  // 通过单元名称获取单词列表
  getWordsByUnitName(unitName, bookName = '') {
    let url = `/words?unit_name=${unitName}`
    if (bookName) url += `&book_name=${bookName}`
    return request(url)
  },

  // ===== 商城相关 =====
  // 获取商品列表
  getProducts(page = 1, limit = 20) {
    return request(`/products?page=${page}&limit=${limit}`)
  },

  // 获取商品详情
  getProduct(productId) {
    return request(`/product/${productId}`)
  },

  // 添加到购物车
  addToCart(userId, productData) {
    return request(`/users/${userId}/cart`, {
      method: 'POST',
      data: productData
    })
  },

  // 更新购物车商品数量
  updateCartItem(userId, productId, quantity) {
    return request(`/users/${userId}/cart/items/${productId}`, {
      method: 'PUT',
      data: { quantity }
    })
  },

  // 删除购物车商品
  deleteCartItem(userId, productId) {
    return request(`/users/${userId}/cart/items/${productId}`, {
      method: 'DELETE'
    })
  },

  // 创建订单
  createOrder(userId, orderData) {
    return request(`/users/${userId}/orders`, {
      method: 'POST',
      data: orderData
    })
  },

  // 获取订单列表
  getOrders(userId, status = '', page = 1, limit = 10) {
    let url = `/users/${userId}/orders?page=${page}&limit=${limit}`
    if (status) url += `&status=${status}`
    return request(url)
  },

  // ===== 推荐系统相关 =====
  // 验证推荐码
  validateReferralCode(referralCode) {
    return request('/referrals/validate', {
      method: 'POST',
      data: { referral_code: referralCode }
    })
  },
  
  // 获取用户推荐信息
  getUserReferralInfo(userId) {
    return request(`/users/${userId}/referral`)
  },
  
  // 获取佣金记录
  getCommissions(userId, status = '', type = '') {
    let url = `/users/${userId}/referral/commissions`
    const params = []
    if (status) params.push(`status=${status}`)
    if (type) params.push(`type=${type}`)
    if (params.length > 0) {
      url += '?' + params.join('&')
    }
    return request(url)
  },
  
  // 跟踪推荐关系
  trackReferral(referralCode, referredUserId) {
    return request('/referrals', {
      method: 'POST',
      data: {
        referral_code: referralCode,
        referred_user_id: referredUserId
      }
    })
  },

  // ===== 代理系统相关 =====
  // 获取代理管理的用户列表
  getAgentUsers(userId, school = '', region = '') {
    let url = `/agents/${userId}/users`
    const params = []
    if (school) params.push(`school=${school}`)
    if (region) params.push(`region=${region}`)
    if (params.length > 0) {
      url += '?' + params.join('&')
    }
    return request(url)
  },

  // 获取代理销售数据
  getAgentSales(userId, startDate = '', endDate = '') {
    let url = `/agents/${userId}/sales`
    const params = []
    if (startDate) params.push(`start_date=${startDate}`)
    if (endDate) params.push(`end_date=${endDate}`)
    if (params.length > 0) {
      url += '?' + params.join('&')
    }
    return request(url)
  },

  // 提取佣金
  withdrawCommission(userId, withdrawData) {
    return request(`/agents/${userId}/withdraw`, {
      method: 'POST',
      data: withdrawData
    })
  },

  // ===== 管理员相关 =====
  // 更新用户代理等级
  updateAgentLevel(userId, agentLevel) {
    return request(`/admin/users/${userId}/agent-level`, {
      method: 'PUT',
      data: { agent_level: agentLevel }
    })
  },

  // ===== 二维码相关 =====
  // 生成不限制小程序码
  generateUnlimitedQRCode(qrCodeData) {
    return request('/wxacode/unlimited', {
      method: 'POST',
      data: qrCodeData
    })
  }
}

module.exports = api