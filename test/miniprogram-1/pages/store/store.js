const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    products: [],
    orders: [],
    selectedProduct: null,
    cartItems: {},
    currentPage: 1,
    totalPages: 1,
    loading: false,
    orderStatus: '',
    statusOptions: [
      { value: '', label: '全部状态' },
      { value: 'pending_payment', label: '待支付' },
      { value: 'processing', label: '处理中' },
      { value: 'shipped', label: '已发货' },
      { value: 'completed', label: '已完成' },
      { value: 'cancelled', label: '已取消' }
    ],
    orderData: {
      address_id: '',
      payment_method: '',
      referral_code: ''
    },
    showOrderForm: false
  },

  onLoad() {
    this.loadProducts()
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

  // 加载商品列表
  async loadProducts(page = 1) {
    try {
      this.setData({ loading: true })
      const result = await api.getProducts(page, 20)
      
      if (result.code === 200) {
        this.setData({
          products: result.data.products || [],
          currentPage: result.data.pagination.page,
          totalPages: result.data.pagination.total_pages
        })
      }
    } catch (error) {
      console.error('获取商品列表失败:', error)
      wx.showToast({
        title: '获取商品失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 查看商品详情
  async viewProduct(e) {
    const productId = e.currentTarget.dataset.productId
    
    try {
      wx.showLoading({ title: '加载中...' })
      const result = await api.getProduct(productId)
      
      if (result.code === 200) {
        this.setData({
          selectedProduct: result.data
        })
        wx.showToast({
          title: '获取商品详情成功',
          icon: 'success'
        })
      }
    } catch (error) {
      console.error('获取商品详情失败:', error)
      wx.showToast({
        title: '获取详情失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 添加到购物车
  async addToCart(e) {
    const productId = e.currentTarget.dataset.productId
    const productName = e.currentTarget.dataset.productName
    const productPrice = e.currentTarget.dataset.productPrice
    
    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    try {
      wx.showLoading({ title: '添加中...' })
      
      const result = await api.addToCart(userInfo.openID, {
        product_id: productId,
        quantity: 1
      })

      if (result.code === 200) {
        // 更新本地购物车状态
        const cartItems = { ...this.data.cartItems }
        cartItems[productId] = (cartItems[productId] || 0) + 1
        
        this.setData({ cartItems })
        
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        })
      } else {
        throw new Error(result.message)
      }
    } catch (error) {
      console.error('添加到购物车失败:', error)
      wx.showToast({
        title: error.message || '添加失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 更新购物车商品数量
  async updateCartItem(e) {
    const productId = e.currentTarget.dataset.productId
    const action = e.currentTarget.dataset.action // 'increase' or 'decrease'
    
    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    const currentQuantity = this.data.cartItems[productId] || 0
    let newQuantity = currentQuantity
    
    if (action === 'increase') {
      newQuantity += 1
    } else if (action === 'decrease') {
      newQuantity = Math.max(0, currentQuantity - 1)
    }

    if (newQuantity === 0) {
      this.removeFromCart(productId)
      return
    }

    try {
      const result = await api.updateCartItem(userInfo.openID, productId, newQuantity)
      
      if (result.code === 200) {
        const cartItems = { ...this.data.cartItems }
        cartItems[productId] = newQuantity
        this.setData({ cartItems })
      }
    } catch (error) {
      console.error('更新购物车失败:', error)
      wx.showToast({
        title: '更新失败',
        icon: 'none'
      })
    }
  },

  // 从购物车删除
  async removeFromCart(productId) {
    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) return

    try {
      await api.deleteCartItem(userInfo.openID, productId)
      
      const cartItems = { ...this.data.cartItems }
      delete cartItems[productId]
      this.setData({ cartItems })
      
      wx.showToast({
        title: '已移除',
        icon: 'success'
      })
    } catch (error) {
      console.error('删除购物车商品失败:', error)
    }
  },

  // 显示订单表单
  showCreateOrder() {
    // 检查是否有购物车商品
    const hasItems = Object.keys(this.data.cartItems).length > 0
    if (!hasItems) {
      wx.showToast({
        title: '购物车为空',
        icon: 'none'
      })
      return
    }
    
    this.setData({ showOrderForm: true })
  },

  // 隐藏订单表单
  hideOrderForm() {
    this.setData({ showOrderForm: false })
  },

  // 订单表单输入
  onAddressIdInput(e) {
    this.setData({
      'orderData.address_id': e.detail.value
    })
  },

  onPaymentMethodInput(e) {
    this.setData({
      'orderData.payment_method': e.detail.value
    })
  },

  onReferralCodeInput(e) {
    this.setData({
      'orderData.referral_code': e.detail.value
    })
  },

  // 创建订单
  async createOrder() {
    const { orderData } = this.data
    const userInfo = auth.getUserInfo()
    
    if (!userInfo || !userInfo.openID) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    if (!orderData.address_id || !orderData.payment_method) {
      wx.showToast({
        title: '请填写完整信息',
        icon: 'none'
      })
      return
    }

    try {
      wx.showLoading({ title: '创建中...' })
      
      const result = await api.createOrder(userInfo.openID, orderData)
      
      if (result.code === 201) {
        wx.showToast({
          title: '订单创建成功',
          icon: 'success'
        })
        
        // 清空购物车和表单
        this.setData({
          cartItems: {},
          showOrderForm: false,
          orderData: {
            address_id: '',
            payment_method: '',
            referral_code: ''
          }
        })
        
        // 刷新订单列表
        this.loadOrders()
      } else {
        throw new Error(result.message)
      }
    } catch (error) {
      console.error('创建订单失败:', error)
      wx.showToast({
        title: error.message || '创建失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 加载订单列表
  async loadOrders() {
    const userInfo = auth.getUserInfo()
    if (!userInfo || !userInfo.openID) return

    try {
      const result = await api.getOrders(userInfo.openID, this.data.orderStatus, 1, 20)
      
      if (result.code === 200) {
        this.setData({
          orders: result.data.orders || []
        })
      }
    } catch (error) {
      console.error('获取订单列表失败:', error)
      wx.showToast({
        title: '获取订单失败',
        icon: 'none'
      })
    }
  },

  // 订单状态筛选
  onOrderStatusChange(e) {
    const status = this.data.statusOptions[e.detail.value].value
    this.setData({ orderStatus: status })
    this.loadOrders()
  },

  // 切换到订单页面
  switchToOrders() {
    this.loadOrders()
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.loadProducts().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 获取购物车商品数量
  getCartQuantity(productId) {
    return this.data.cartItems[productId] || 0
  },

  // 格式化日期
  formatDate(dateString) {
    const date = new Date(dateString)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  }
})