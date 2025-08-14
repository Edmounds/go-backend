const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    unitId: '',
    unitName: '',
    bookName: '',
    wordName: '',
    unitWords: [],
    wordCard: null,
    wordsByName: [],
    loading: false,
    searchType: 0, // 0: 按单元ID, 1: 按单元名称, 2: 按单词名称
    searchOptions: [
      { label: '按单元ID获取单词', value: 0 },
      { label: '按单元名称获取单词', value: 1 },
      { label: '按单词名称获取卡片', value: 2 }
    ]
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

  // 切换搜索类型
  onSearchTypeChange(e) {
    this.setData({
      searchType: parseInt(e.detail.value),
      unitWords: [],
      wordCard: null,
      wordsByName: []
    })
  },

  // 输入单元ID
  onUnitIdInput(e) {
    this.setData({
      unitId: e.detail.value
    })
  },

  // 输入单元名称
  onUnitNameInput(e) {
    this.setData({
      unitName: e.detail.value
    })
  },

  // 输入书籍名称
  onBookNameInput(e) {
    this.setData({
      bookName: e.detail.value
    })
  },

  // 输入单词名称
  onWordNameInput(e) {
    this.setData({
      wordName: e.detail.value
    })
  },

  // 按单元ID获取单词列表
  async getUnitWordsById() {
    const { unitId } = this.data
    
    if (!unitId.trim()) {
      wx.showToast({
        title: '请输入单元ID',
        icon: 'none'
      })
      return
    }

    try {
      this.setData({ loading: true })
      wx.showLoading({ title: '加载中...' })
      
      const result = await api.getUnitWords(unitId)
      
      if (result.code === 200) {
        this.setData({
          unitWords: result.data.words || []
        })
        wx.showToast({
          title: `获取到${result.data.words?.length || 0}个单词`,
          icon: 'success'
        })
      } else {
        throw new Error(result.message || '获取失败')
      }
    } catch (error) {
      console.error('获取单元单词失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
      wx.hideLoading()
    }
  },

  // 按单元名称获取单词列表
  async getWordsByUnitName() {
    const { unitName, bookName } = this.data
    
    if (!unitName.trim()) {
      wx.showToast({
        title: '请输入单元名称',
        icon: 'none'
      })
      return
    }

    try {
      this.setData({ loading: true })
      wx.showLoading({ title: '加载中...' })
      
      const result = await api.getWordsByUnitName(unitName, bookName)
      
      if (result.code === 200) {
        this.setData({
          wordsByName: result.data.words || []
        })
        wx.showToast({
          title: `获取到${result.data.words?.length || 0}个单词`,
          icon: 'success'
        })
      } else {
        throw new Error(result.message || '获取失败')
      }
    } catch (error) {
      console.error('按名称获取单词失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
      wx.hideLoading()
    }
  },

  // 获取单词卡片详情
  async getWordCard() {
    const { wordName } = this.data
    
    if (!wordName.trim()) {
      wx.showToast({
        title: '请输入单词名称',
        icon: 'none'
      })
      return
    }

    try {
      this.setData({ loading: true })
      wx.showLoading({ title: '加载中...' })
      
      const result = await api.getWordCard(wordName)
      
      if (result.code === 200) {
        this.setData({
          wordCard: result.data
        })
        wx.showToast({
          title: '获取单词卡片成功',
          icon: 'success'
        })
      } else {
        throw new Error(result.message || '获取失败')
      }
    } catch (error) {
      console.error('获取单词卡片失败:', error)
      wx.showToast({
        title: error.message || '获取失败',
        icon: 'none'
      })
      this.setData({ wordCard: null })
    } finally {
      this.setData({ loading: false })
      wx.hideLoading()
    }
  },

  // 执行搜索
  performSearch() {
    const { searchType } = this.data
    
    switch (searchType) {
      case 0:
        this.getUnitWordsById()
        break
      case 1:
        this.getWordsByUnitName()
        break
      case 2:
        this.getWordCard()
        break
      default:
        break
    }
  },

  // 预览图片
  previewImage(e) {
    const { url } = e.currentTarget.dataset
    wx.previewImage({
      urls: [url],
      current: url
    })
  },

  // 播放发音 (模拟)
  playPronunciation(e) {
    const { url } = e.currentTarget.dataset
    if (url) {
      wx.showToast({
        title: '播放发音 (模拟)',
        icon: 'none'
      })
      // 实际项目中可以使用 wx.createInnerAudioContext() 播放音频
    } else {
      wx.showToast({
        title: '无发音文件',
        icon: 'none'
      })
    }
  },

  // 复制单词
  copyWord(e) {
    const { word } = e.currentTarget.dataset
    wx.setClipboardData({
      data: word,
      success: () => {
        wx.showToast({
          title: '已复制单词',
          icon: 'success'
        })
      }
    })
  },

  // 清空结果
  clearResults() {
    this.setData({
      unitWords: [],
      wordCard: null,
      wordsByName: []
    })
  }
})