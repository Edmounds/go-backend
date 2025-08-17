# 微信小程序 API 路由文档

## 概述

本文档描述了根据OpenAPI文档完善后的所有API路由端点。

## 路由分类

### 1. 微信小程序相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/sns/jscode2session` | 微信登录API | 否 |

### 2. 用户认证相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/auth` | 用户登录 | 否 |
| POST | `/api/auth/refresh` | 刷新Token | 是 |

### 3. 用户管理相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/users` | 填写用户信息 | 否 |
| PUT | `/api/users/:user_id` | 更新用户信息 | 是 |
| POST | `/api/users/:user_id/address` | 创建收货地址 | 是 |
| DELETE | `/api/users/:user_id/address/:address_id` | 删除收货地址 | 是 |

### 4. 商城相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/products` | 获取所有商品 | 否 |
| GET | `/api/product/:product_id` | 获取特定商品详情 | 否 |
| POST | `/api/users/:user_id/cart` | 添加商品到购物车 | 是 |
| PUT | `/api/users/:user_id/cart/items/:product_id` | 更新购物车商品数量 | 是 |
| DELETE | `/api/users/:user_id/cart/items/:product_id` | 删除购物车商品 | 是 |
| POST | `/api/users/:user_id/orders` | 从购物车创建订单 | 是 |
| GET | `/api/users/:user_id/orders` | 获取订单历史 | 是 |
| POST | `/api/users/:user_id/orders/pay` | 创建微信支付订单 | 是 |

### 6. 微信支付相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/users/:user_id/orders/pay` | 创建微信支付订单 | 是 |
| POST | `/api/wechat/pay/notify` | 微信支付回调通知 | 否 |

### 5. 学习进度相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/books` | 获取书籍列表 | 否 |
| GET | `/api/users/:user_id/progress` | 获取用户学习进度 | 是 |
| PUT | `/api/users/:user_id/progress` | 更新用户学习进度 | 是 |
| GET | `/api/books/:book_id/words` | 获取书籍单词 | 是 |

### 5.1. 单词卡片相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/units/:unit_id/words` | 根据单元ID获取该单元的所有单词列表 | 是 |
| GET | `/api/words/:word_id/card` | 根据单词ID获取单词卡片详细信息（包括图片） | 是 |
| GET | `/api/words?unit_name=xxx&book_name=xxx` | 通过单元名称和书籍名称获取单词列表（可选参数） | 是 |

### 5.1.1. 收藏功能相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/users/:user_id/collected-cards` | 获取用户收藏的单词卡列表 | 是 |
| POST | `/api/users/:user_id/collected-cards/:word_id` | 添加单词卡到收藏列表 | 是 |
| DELETE | `/api/users/:user_id/collected-cards/:word_id` | 从收藏列表中移除单词卡 | 是 |
| GET | `/api/users/:user_id/collected-cards/:word_id/status` | 检查单词卡是否已被收藏 | 是 |

### 5.2. 搜索相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/search` | 综合搜索（支持单词、课本、订单） | 否 |
| GET | `/api/search/words` | 单词模糊搜索 | 否 |
| GET | `/api/search/books` | 课本模糊搜索 | 否 |
| GET | `/api/search/orders` | 订单搜索（根据商品名称） | 是 |

### 6. 推荐系统相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/referrals/validate` | 验证推荐码 | 否 |
| GET | `/api/users/:user_id/referral` | 获取用户推荐信息 | 是 |
| POST | `/api/referrals` | 跟踪推荐关系 | 是 |
| GET | `/api/users/:user_id/referral/commissions` | 查看返现记录 | 是 |

### 7. 代理系统相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/agents/:user_id/users` | 列出代理管理的用户 | 是 |
| GET | `/api/agents/:user_id/sales` | 列出所有代理销售数据和佣金 | 是 |
| POST | `/api/agents/:user_id/withdraw` | 提取佣金 | 是 |

### 8. 系统路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 否 |

## 控制器文件说明

- `auth_controller.go` - 用户认证相关处理器
- `user_controller.go` - 用户管理相关处理器
- `store_controller.go` - 商城相关处理器
- `progress_controller.go` - 学习进度相关处理器
- `card_controller.go` - 单词卡片相关处理器
- `search_controller.go` - 搜索功能相关处理器
- `referral_controller.go` - 推荐系统相关处理器
- `agent_controller.go` - 代理系统相关处理器

## 注意事项

1. 所有带有"是"认证标记的路由都需要在请求头中包含有效的JWT Token
2. JWT Token格式：`Authorization: Bearer <token>`
3. 所有控制器函数都包含了基本的请求验证和错误处理
4. 控制器中的TODO注释标记了需要进一步实现的业务逻辑
5. 响应格式统一为JSON，包含code、message和data字段

## 下一步开发建议

1. 实现数据库模型和连接
2. 完成微信小程序登录逻辑
3. 实现JWT Token生成和验证
4. 添加数据库操作逻辑
5. 实现业务逻辑验证
6. 添加单元测试
7. 完善错误处理和日志记录