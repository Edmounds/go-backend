// 配置文件
const config = {
  // 后端API地址
  apiBaseUrl: 'http://localhost:8080/api',
  
  // 微信小程序配置
  appId: 'your_wechat_app_id_here',
  
  // API路径
  apiPaths: {
    auth: '/auth',
    validateReferral: '/referrals/validate',
    getUserReferral: '/users/{user_id}/referral',
    getCommissions: '/users/{user_id}/referral/commissions',
    trackReferral: '/referrals'
  }
}

module.exports = config