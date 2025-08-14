const api = require('../../utils/api.js')
const auth = require('../../utils/auth.js')

Page({
  data: {
    books: [],
    userProgress: null,
    selectedBook: null,
    bookWords: [],
    loading: true,
    progressData: {
      currentUnit: '',
      currentSentence: '',
      learnedWords: []
    }
  },

  onLoad() {
    this.loadData()
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

  // 加载所有数据
  async loadData() {
    try {
      this.setData({ loading: true })
      await Promise.all([
        this.loadBooks(),
        this.loadUserProgress()
      ])
    } catch (error) {
      console.error('加载数据失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  // 加载书籍列表
  async loadBooks() {
    try {
      const result = await api.getBooks(1, 50)
      if (result.code === 200) {
        this.setData({
          books: result.data.books || []
        })
      }
    } catch (error) {
      console.error('获取书籍列表失败:', error)
      wx.showToast({
        title: '获取书籍失败',
        icon: 'none'
      })
    }
  },

  // 加载用户学习进度
  async loadUserProgress() {
    try {
      const userInfo = auth.getUserInfo()
      if (userInfo && userInfo.openID) {
        const result = await api.getUserProgress(userInfo.openID)
        if (result.code === 200) {
          this.setData({
            userProgress: result.data,
            progressData: {
              currentUnit: result.data.current_unit || '',
              currentSentence: result.data.current_sentence || '',
              learnedWords: result.data.learned_words || []
            }
          })
        }
      }
    } catch (error) {
      console.error('获取学习进度失败:', error)
    }
  },

  // 选择书籍
  async selectBook(e) {
    const bookId = e.currentTarget.dataset.bookId
    const bookName = e.currentTarget.dataset.bookName
    
    try {
      wx.showLoading({ title: '加载中...' })
      const result = await api.getBookWords(bookId)
      
      if (result.code === 200) {
        this.setData({
          selectedBook: { id: bookId, name: bookName },
          bookWords: result.data.words || []
        })
        wx.showToast({
          title: '书籍加载成功',
          icon: 'success'
        })
      }
    } catch (error) {
      console.error('获取书籍单词失败:', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 更新当前单元
  onCurrentUnitInput(e) {
    this.setData({
      'progressData.currentUnit': e.detail.value
    })
  },

  // 更新当前句子
  onCurrentSentenceInput(e) {
    this.setData({
      'progressData.currentSentence': e.detail.value
    })
  },

  // 添加学习的单词
  addLearnedWord() {
    wx.showModal({
      title: '添加单词',
      editable: true,
      placeholderText: '请输入单词',
      success: (res) => {
        if (res.confirm && res.content) {
          const learnedWords = [...this.data.progressData.learnedWords]
          if (!learnedWords.includes(res.content)) {
            learnedWords.push(res.content)
            this.setData({
              'progressData.learnedWords': learnedWords
            })
          } else {
            wx.showToast({
              title: '单词已存在',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  // 删除学习的单词
  removeLearnedWord(e) {
    const index = e.currentTarget.dataset.index
    const learnedWords = [...this.data.progressData.learnedWords]
    learnedWords.splice(index, 1)
    this.setData({
      'progressData.learnedWords': learnedWords
    })
  },

  // 保存学习进度
  async saveProgress() {
    try {
      const userInfo = auth.getUserInfo()
      if (!userInfo || !userInfo.openID) {
        wx.showToast({
          title: '请先登录',
          icon: 'none'
        })
        return
      }

      wx.showLoading({ title: '保存中...' })
      
      const result = await api.updateUserProgress(userInfo.openID, {
        current_unit: this.data.progressData.currentUnit,
        current_sentence: this.data.progressData.currentSentence,
        learned_words: this.data.progressData.learnedWords
      })

      if (result.code === 200) {
        wx.showToast({
          title: '保存成功',
          icon: 'success'
        })
        this.loadUserProgress() // 重新加载进度
      } else {
        throw new Error(result.message)
      }
    } catch (error) {
      console.error('保存学习进度失败:', error)
      wx.showToast({
        title: error.message || '保存失败',
        icon: 'none'
      })
    } finally {
      wx.hideLoading()
    }
  },

  // 清空进度数据
  clearProgress() {
    wx.showModal({
      title: '确认清空',
      content: '确定要清空所有学习进度吗？',
      success: (res) => {
        if (res.confirm) {
          this.setData({
            progressData: {
              currentUnit: '',
              currentSentence: '',
              learnedWords: []
            }
          })
        }
      }
    })
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.loadData().then(() => {
      wx.stopPullDownRefresh()
    })
  }
})