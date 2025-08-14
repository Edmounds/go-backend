// pages/qrcode/qrcode.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');

Page({
  data: {
    // 基础数据
    userInfo: null,
    isLoggedIn: false,
    
    // 二维码相关
    qrCodeImage: '',
    qrCodeVisible: false,
    loading: false,
    
    // 表单数据
    scene: '',
    page: 'pages/index/index',
    width: 280,
    checkPath: true,
    envVersion: 'release',
    autoColor: false,
    isHyaline: false,
    lineColor: {
      r: 0,
      g: 0,
      b: 0
    },
    
    // 预设场景
    presetScenes: [
      {
        name: '首页',
        scene: 'home',
        page: 'pages/index/index'
      },
      {
        name: '推荐页',
        scene: 'referral_test',
        page: 'pages/referral/referral'
      },
      {
        name: '学习页',
        scene: 'learning_test',
        page: 'pages/learning/learning'
      },
      {
        name: '商店页',
        scene: 'store_test',
        page: 'pages/store/store'
      }
    ],
    
    // 环境版本选项
    envVersionOptions: [
      { value: 'release', name: '正式版' },
      { value: 'trial', name: '体验版' },
      { value: 'develop', name: '开发版' }
    ]
  },

  onLoad: function(options) {
    console.log('QRCode页面加载');
    this.checkLoginStatus();
  },

  onShow: function() {
    this.checkLoginStatus();
  },

  // 检查登录状态
  checkLoginStatus: function() {
    const userInfo = auth.getUserInfo();
    this.setData({
      isLoggedIn: !!userInfo,
      userInfo: userInfo
    });
  },

  // 登录
  doLogin: function() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  },

  // 输入框变化处理
  onSceneInput: function(e) {
    this.setData({
      scene: e.detail.value
    });
  },

  onPageInput: function(e) {
    this.setData({
      page: e.detail.value
    });
  },

  onWidthInput: function(e) {
    const width = parseInt(e.detail.value) || 280;
    this.setData({
      width: Math.min(Math.max(width, 100), 1280) // 限制在合理范围内
    });
  },

  // 开关变化处理
  onCheckPathChange: function(e) {
    this.setData({
      checkPath: e.detail.value
    });
  },

  onAutoColorChange: function(e) {
    this.setData({
      autoColor: e.detail.value
    });
  },

  onIsHyalineChange: function(e) {
    this.setData({
      isHyaline: e.detail.value
    });
  },

  // 环境版本选择
  onEnvVersionChange: function(e) {
    const envVersionOptions = this.data.envVersionOptions;
    this.setData({
      envVersion: envVersionOptions[e.detail.value].value
    });
  },

  // 颜色输入处理
  onRedInput: function(e) {
    const r = Math.min(Math.max(parseInt(e.detail.value) || 0, 0), 255);
    this.setData({
      [`lineColor.r`]: r
    });
  },

  onGreenInput: function(e) {
    const g = Math.min(Math.max(parseInt(e.detail.value) || 0, 0), 255);
    this.setData({
      [`lineColor.g`]: g
    });
  },

  onBlueInput: function(e) {
    const b = Math.min(Math.max(parseInt(e.detail.value) || 0, 0), 255);
    this.setData({
      [`lineColor.b`]: b
    });
  },

  // 使用预设场景
  usePresetScene: function(e) {
    const index = e.currentTarget.dataset.index;
    const preset = this.data.presetScenes[index];
    this.setData({
      scene: preset.scene,
      page: preset.page
    });
  },

  // 生成二维码
  generateQRCode: function() {
    if (!this.data.isLoggedIn) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      return;
    }

    if (!this.data.scene.trim()) {
      wx.showToast({
        title: '请输入场景值',
        icon: 'none'
      });
      return;
    }

    this.setData({ loading: true });

    const requestData = {
      scene: this.data.scene.trim(),
      page: this.data.page || 'pages/index/index',
      width: this.data.width,
      check_path: this.data.checkPath,
      env_version: this.data.envVersion
    };

    // 根据设置添加可选参数
    if (!this.data.autoColor) {
      requestData.auto_color = false;
      requestData.line_color = this.data.lineColor;
    } else {
      requestData.auto_color = true;
    }

    if (this.data.isHyaline) {
      requestData.is_hyaline = true;
    }

    console.log('生成二维码请求:', requestData);

    api.generateUnlimitedQRCode(requestData)
      .then(response => {
        console.log('二维码生成成功:', response);
        if (response.data && response.data.image_base64) {
          this.setData({
            qrCodeImage: 'data:image/png;base64,' + response.data.image_base64,
            qrCodeVisible: true,
            loading: false
          });
          
          wx.showToast({
            title: '二维码生成成功',
            icon: 'success'
          });
        } else {
          throw new Error('响应数据格式错误');
        }
      })
      .catch(error => {
        console.error('生成二维码失败:', error);
        this.setData({ loading: false });
        
        wx.showModal({
          title: '生成失败',
          content: error.message || '生成二维码时出现错误，请检查参数后重试',
          showCancel: false
        });
      });
  },

  // 保存二维码到相册
  saveQRCode: function() {
    if (!this.data.qrCodeImage) {
      wx.showToast({
        title: '没有可保存的二维码',
        icon: 'none'
      });
      return;
    }

    // 将base64转换为临时文件
    const base64Data = this.data.qrCodeImage.replace('data:image/png;base64,', '');
    const arrayBuffer = wx.base64ToArrayBuffer(base64Data);
    
    const fs = wx.getFileSystemManager();
    const filePath = wx.env.USER_DATA_PATH + '/qrcode_' + Date.now() + '.png';
    
    fs.writeFile({
      filePath: filePath,
      data: arrayBuffer,
      success: () => {
        wx.saveImageToPhotosAlbum({
          filePath: filePath,
          success: () => {
            wx.showToast({
              title: '保存成功',
              icon: 'success'
            });
          },
          fail: (err) => {
            console.error('保存到相册失败:', err);
            if (err.errMsg.includes('auth deny')) {
              wx.showModal({
                title: '需要授权',
                content: '保存图片需要访问您的相册权限，请在设置中开启',
                confirmText: '去设置',
                success: (res) => {
                  if (res.confirm) {
                    wx.openSetting();
                  }
                }
              });
            } else {
              wx.showToast({
                title: '保存失败',
                icon: 'none'
              });
            }
          }
        });
      },
      fail: (err) => {
        console.error('写入文件失败:', err);
        wx.showToast({
          title: '保存失败',
          icon: 'none'
        });
      }
    });
  },

  // 预览二维码
  previewQRCode: function() {
    if (!this.data.qrCodeImage) {
      return;
    }

    // 临时保存图片用于预览
    const base64Data = this.data.qrCodeImage.replace('data:image/png;base64,', '');
    const arrayBuffer = wx.base64ToArrayBuffer(base64Data);
    
    const fs = wx.getFileSystemManager();
    const filePath = wx.env.USER_DATA_PATH + '/preview_qrcode.png';
    
    fs.writeFile({
      filePath: filePath,
      data: arrayBuffer,
      success: () => {
        wx.previewImage({
          urls: [filePath],
          current: filePath
        });
      },
      fail: (err) => {
        console.error('预览失败:', err);
        wx.showToast({
          title: '预览失败',
          icon: 'none'
        });
      }
    });
  },

  // 隐藏二维码
  hideQRCode: function() {
    this.setData({
      qrCodeVisible: false
    });
  },

  // 清空表单
  clearForm: function() {
    this.setData({
      scene: '',
      page: 'pages/index/index',
      width: 280,
      checkPath: true,
      envVersion: 'release',
      autoColor: false,
      isHyaline: false,
      lineColor: {
        r: 0,
        g: 0,
        b: 0
      },
      qrCodeImage: '',
      qrCodeVisible: false
    });
  },

  // 复制场景值
  copyScene: function() {
    if (!this.data.scene) {
      wx.showToast({
        title: '没有可复制的内容',
        icon: 'none'
      });
      return;
    }

    wx.setClipboardData({
      data: this.data.scene,
      success: () => {
        wx.showToast({
          title: '已复制到剪贴板',
          icon: 'success'
        });
      }
    });
  }
});