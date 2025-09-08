# 微信小程序API路由文档 - 管理员后台开发指南

## 项目概述

本项目是一个微信小程序学习应用的后端API系统，提供用户管理、商城功能、学习进度跟踪、推荐系统、代理系统以及**管理员后台**等完整功能。

### 技术栈
- **后端框架**: Go + Gin 
- **数据库**: MongoDB
- **认证方式**: JWT Token + 微信登录
- **权限控制**: 基于用户角色的权限验证
- **API文档**: OpenAPI 3.0

### 核心功能模块
1. **用户认证系统** - 微信登录、JWT Token管理
2. **商城系统** - 商品展示、购物车、订单管理、支付
3. **学习系统** - 书籍管理、单词卡片、学习进度
4. **推荐系统** - 推荐码、佣金计算
5. **代理系统** - 校代理、区代理管理
6. **🎯 管理员后台** - 用户管理、商品管理、代理管理（**重点开发模块**）

### 管理员后台功能详述
管理员后台是为系统管理员设计的Web管理界面，具备以下核心功能：

#### 👥 用户管理
- 查看所有用户列表（支持分页、筛选）
- 查看用户详细信息和资料
- 设置和取消用户管理员权限
- 查看用户订单历史

#### 🤝 代理管理  
- 设置用户的代理等级（普通用户、校代理、区代理）
- 为校代理分配管理的学校
- 为区代理分配管理的区域
- 查看代理下属用户和销售统计

#### 🛍️ 商品管理
- 创建、编辑、删除商品
- 管理商品价格、库存、属性
- 商品上架和下架控制

### 用户权限体系
- **普通用户**: 基本功能访问权限
- **代理用户**: 代理功能 + 佣金管理权限  
- **管理员用户**: 全部功能 + 后台管理权限

## API路由端点

## 路由分类

### 1. 微信小程序相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/sns/jscode2session` | 微信登录API | 否 |
| POST | `/api/dev-login` | 开发环境登录（仅限开发） | 否 |

### 2. 用户认证相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/auth` | 用户登录 | 否 |
| POST | `/api/auth/refresh` | 刷新Token | 是 |

### 3. 用户管理相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/users/:user_id` | 获取用户信息 | 是 |
| POST | `/api/users/profile` | 更新用户资料 | 是 |
| GET | `/api/users/:user_id/avatar` | 获取用户头像 | 是 |
| POST | `/api/users/:user_id/avatar` | 上传用户头像 | 是 |
| GET | `/api/users/:user_id/qrcode` | 获取用户推荐二维码 | 是 |
| GET | `/api/users/:user_id/addresses` | 获取用户地址列表 | 是 |
| POST | `/api/users/:user_id/address` | 创建收货地址 | 是 |
| PUT | `/api/users/:user_id/address/:address_id` | 更新收货地址 | 是 |
| DELETE | `/api/users/:user_id/address/:address_id` | 删除收货地址 | 是 |

### 4. 商城相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/products` | 获取所有商品 | 否 |
| GET | `/api/product/:product_id` | 获取特定商品详情 | 否 |
| GET | `/api/users/:user_id/cart` | 获取用户购物车 | 是 |
| POST | `/api/users/:user_id/cart` | 添加商品到购物车 | 是 |
| PUT | `/api/users/:user_id/cart/items/:product_id` | 更新购物车商品数量 | 是 |
| DELETE | `/api/users/:user_id/cart/items/:product_id` | 删除购物车商品 | 是 |
| PUT | `/api/users/:user_id/cart/items/:product_id/select` | 选择/取消选择购物车商品 | 是 |
| PUT | `/api/users/:user_id/cart/select-all` | 全选/反选购物车商品 | 是 |
| GET | `/api/users/:user_id/cart/selected` | 获取选中的购物车商品 | 是 |
| POST | `/api/users/:user_id/orders` | 从购物车创建订单 | 是 |
| GET | `/api/users/:user_id/orders` | 获取订单历史 | 是 |
| GET | `/api/users/:user_id/orders/:order_id` | 获取订单详情 | 是 |
| POST | `/api/users/:user_id/direct-purchase` | 直接购买商品 | 是 |

### 5. 微信支付相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/users/:user_id/orders/pay` | 创建微信支付订单 | 是 |
| POST | `/api/wechat/pay/notify` | 微信支付回调通知 | 否 |
| POST | `/api/users/:user_id/refunds` | 申请退款 | 是 |
| GET | `/api/users/:user_id/refunds` | 获取退款记录 | 是 |
| GET | `/api/users/:user_id/refunds/:refund_id` | 获取退款详情 | 是 |
| GET | `/api/users/:user_id/transfer-bills/:transfer_bill_no` | 查询微信转账单详情 | 是 |

### 6. 学习进度相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/books` | 获取书籍列表 | 否 |
| GET | `/api/users/:user_id/progress` | 获取用户学习进度 | 是 |
| PUT | `/api/users/:user_id/progress` | 更新用户学习进度 | 是 |
| GET | `/api/books/:book_id/words` | 获取书籍单词 | 是 |

### 6.1. 单词卡片相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/units/:unit_id/words` | 根据单元ID获取该单元的所有单词列表 | 是 |
| GET | `/api/words/:word_name/card` | 根据单词名称获取单词卡片详细信息 | 是 |
| GET | `/api/words/:word_id/card` | 根据单词ID获取单词卡片详细信息（包括图片） | 是 |
| GET | `/api/words?unit_name=xxx&book_name=xxx` | 通过单元名称和书籍名称获取单词列表（可选参数） | 是 |

### 6.2. 收藏功能相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/users/:user_id/collected-cards` | 获取用户收藏的单词卡列表 | 是 |
| POST | `/api/users/:user_id/collected-cards/:word_id` | 添加单词卡到收藏列表 | 是 |
| DELETE | `/api/users/:user_id/collected-cards/:word_id` | 从收藏列表中移除单词卡 | 是 |
| GET | `/api/users/:user_id/collected-cards/:word_id/status` | 检查单词卡是否已被收藏 | 是 |

### 7. 搜索相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/search` | 综合搜索（支持单词、课本、订单） | 否 |
| GET | `/api/search/words` | 单词模糊搜索 | 否 |
| GET | `/api/search/books` | 课本模糊搜索 | 否 |
| GET | `/api/search/orders` | 订单搜索（根据商品名称） | 是 |

### 8. 推荐系统相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/referrals/validate` | 验证推荐码 | 否 |
| GET | `/api/users/:user_id/referral` | 获取用户推荐信息 | 是 |
| POST | `/api/referrals` | 跟踪推荐关系 | 是 |
| GET | `/api/users/:user_id/referral/commissions` | 查看返现记录 | 是 |
| POST | `/api/wxacode/unlimited` | 生成不限制小程序码 | 是 |

### 9. 代理系统相关路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/agents/:user_id/users` | 列出代理管理的用户 | 是 |
| GET | `/api/agents/:user_id/sales` | 列出所有代理销售数据和佣金 | 是 |
| POST | `/api/agents/:user_id/withdraw` | 提取佣金 | 是 |
| GET | `/api/agents/:user_id/commission/dashboard` | 获取代理佣金仪表板 | 是 |
| GET | `/api/agents/:user_id/commission/details` | 获取代理佣金明细 | 是 |

### 10. 🎯 管理员后台相关路由 (重点开发)

| 方法 | 路径 | 描述 | 认证 | 功能模块 |
|------|------|------|------|----------|
| GET | `/api/admin/users` | 获取所有用户列表（支持分页、筛选） | 是（管理员） | 👥 用户管理 |
| GET | `/api/admin/users/:user_id` | 获取用户详细信息 | 是（管理员） | 👥 用户管理 |
| PUT | `/api/admin/users/:user_id/admin` | 设置/取消用户管理员权限 | 是（管理员） | 👥 用户管理 |
| GET | `/api/admin/users/:user_id/orders` | 获取用户订单列表 | 是（管理员） | 👥 用户管理 |
| PUT | `/api/admin/users/:user_id/agent-level` | 更新用户代理等级 | 是（管理员） | 🤝 代理管理 |
| PUT | `/api/admin/agents/:user_id/schools` | 设置校代理管理的学校 | 是（管理员） | 🤝 代理管理 |
| PUT | `/api/admin/agents/:user_id/regions` | 设置区代理管理的区域 | 是（管理员） | 🤝 代理管理 |
| GET | `/api/admin/agents/:user_id/stats` | 获取代理统计信息 | 是（管理员） | 🤝 代理管理 |
| POST | `/api/admin/products` | 创建商品 | 是（管理员） | 🛍️ 商品管理 |
| PUT | `/api/admin/products/:product_id` | 更新商品信息 | 是（管理员） | 🛍️ 商品管理 |
| DELETE | `/api/admin/products/:product_id` | 删除商品 | 是（管理员） | 🛍️ 商品管理 |
| PUT | `/api/admin/products/:product_id/status` | 更新商品上下架状态 | 是（管理员） | 🛍️ 商品管理 |

> **注意**: 管理员后台API是前端Web管理界面的核心，需要特别关注这些接口的对接和测试。

### 11. 系统路由

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 否 |

## 🏗️ 项目架构说明

### 控制器文件说明
- `auth_controller.go` - 用户认证相关处理器
- `user_controller.go` - 用户管理相关处理器  
- `store_controller.go` - 商城相关处理器
- `progress_controller.go` - 学习进度相关处理器
- `card_controller.go` - 单词卡片相关处理器
- `search_controller.go` - 搜索功能相关处理器
- `referral_controller.go` - 推荐系统相关处理器
- `agent_controller.go` - 代理系统相关处理器
- **`admin_controller.go` - 🎯 管理员后台相关处理器（重点文件）**

### 中间件说明
- `middlewares/token_handler.go` - JWT Token认证中间件
- **`middlewares/admin_middleware.go` - 🎯 管理员权限验证中间件（重点文件）**
- `middlewares/error_handler.go` - 统一错误处理中间件

### 数据模型说明
- `models/common.go` - 所有数据结构体定义
- `models/微信小程序.openapi.json` - 完整的API文档

## 📝 核心数据模型

### 用户模型 (User)
```json
{
  "_id": "string",
  "openID": "string",           // 微信openID，作为主要标识
  "user_name": "string",
  "school": "string",
  "age": 20,
  "agent_level": 0,            // 0:普通用户, 1:校代理, 2:区代理
  "is_admin": false,           // 🎯 管理员标识（重要字段）
  "is_agent": false,
  "managed_schools": [],       // 校代理管理的学校
  "managed_regions": [],       // 区代理管理的区域
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 商品模型 (Product)
```json
{
  "_id": "string",
  "product_id": "string",
  "name": "string",
  "price": 99.99,
  "description": "string",
  "stock": 100,
  "product_type": "physical|digital",
  "is_active": true,           // 🎯 上下架状态（管理员可控制）
  "images": ["url1", "url2"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 订单模型 (Order)
```json
{
  "_id": "string", 
  "user_openid": "string",
  "items": [{"product_id": "string", "quantity": 1, "price": 99.99}],
  "total_amount": 99.99,
  "status": "pending|paid|shipped|delivered|cancelled",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 🔐 认证与权限

### JWT Token 使用
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 权限等级
1. **无认证**: 公开API（如获取商品列表）
2. **用户认证**: 需要有效JWT Token
3. **🎯 管理员认证**: 需要JWT Token + `is_admin=true`

### 管理员权限验证流程
1. JWT Token认证 → 2. 获取用户信息 → 3. 检查 `is_admin` 字段 → 4. 通过/拒绝

## 📡 API响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    // 具体数据内容
  }
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "VALIDATION_ERROR"
}
```

### 分页响应
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

## ⚠️ 重要注意事项

### 认证相关
1. 所有带有"是"认证标记的路由都需要在请求头中包含有效的JWT Token
2. **🎯 所有带有"是（管理员）"认证标记的路由需要JWT Token且用户具备管理员权限（is_admin=true）**
3. JWT Token格式：`Authorization: Bearer <token>`
4. Token过期时需要使用refresh接口刷新

### 管理员后台特殊说明
5. **🎯 管理员后台API统一使用`/api/admin`前缀**
6. **🎯 所有admin路由都通过AdminAuthMiddleware进行权限验证**
7. **🎯 前端需要先检查用户的`is_admin`字段，只有管理员才能访问后台界面**

### 数据处理
8. 所有响应格式统一为JSON，包含code、message和data字段
9. 用户标识统一使用openID，禁止使用MongoDB的_id作为业务标识
10. 时间格式统一使用UTC时间：`2024-01-01T00:00:00Z`

### 错误处理
11. 统一错误代码体系（详见下方错误代码表）
12. 所有异常都会被error_handler中间件捕获并统一处理

## 📊 错误代码表

| 代码 | 含义 | 常见场景 |
|------|------|----------|
| 200 | 成功 | 请求处理成功 |
| 201 | 创建成功 | 资源创建成功 |
| 400 | 请求参数错误 | 参数缺失、格式错误 |
| 401 | 未授权 | Token无效、Token过期 |
| 403 | 无权限 | **🎯 非管理员访问admin接口** |
| 404 | 资源不存在 | 用户、商品、订单不存在 |
| 500 | 服务器内部错误 | 数据库连接失败、系统异常 |

## 🔥 管理员后台API示例

### 获取用户列表
```http
GET /api/admin/users?page=1&limit=20&agent_level=1
Authorization: Bearer <admin_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取用户列表成功",
  "data": {
    "users": [
      {
        "_id": "507f1f77bcf86cd799439011",
        "openID": "oGZUI0egBJY1zhBY2E5hrWOjn_fs",
        "user_name": "张三",
        "school": "北京大学",
        "agent_level": 1,
        "is_admin": false,
        "is_agent": true,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### 设置用户管理员权限
```http
PUT /api/admin/users/oGZUI0egBJY1zhBY2E5hrWOjn_fs/admin
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "is_admin": true
}
```

### 创建商品
```http
POST /api/admin/products
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "英语学习卡片套装",
  "price": 99.99,
  "description": "包含1000个核心单词",
  "stock": 100,
  "product_type": "physical",
  "images": ["https://example.com/image1.jpg"]
}
```

## 🚀 前端开发指南

### 环境配置
```javascript
// 配置API基础URL
const API_BASE_URL = 'http://localhost:8080/api'
const ADMIN_API_BASE_URL = 'http://localhost:8080/api/admin'

// 开发环境登录（仅限开发）
const DEV_LOGIN_URL = 'http://localhost:8080/api/dev-login'
```

### 权限验证流程
```javascript
// 1. 登录获取Token
async function login(openid) {
  const response = await fetch(`${API_BASE_URL}/dev-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ openid })
  })
  const data = await response.json()
  
  // 保存Token和用户信息
  localStorage.setItem('token', data.data.token)
  localStorage.setItem('user', JSON.stringify(data.data.user))
  
  return data.data.user
}

// 2. 检查管理员权限
function isAdmin() {
  const user = JSON.parse(localStorage.getItem('user') || '{}')
  return user.is_admin === true
}

// 3. 管理员路由守卫
function requireAdmin() {
  if (!isAdmin()) {
    alert('需要管理员权限才能访问此页面')
    window.location.href = '/login'
    return false
  }
  return true
}
```

### API请求封装
```javascript
// 通用API请求函数
async function apiRequest(url, options = {}) {
  const token = localStorage.getItem('token')
  
  const config = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` }),
      ...options.headers,
    },
  }
  
  const response = await fetch(url, config)
  const data = await response.json()
  
  // 统一错误处理
  if (data.code !== 200 && data.code !== 201) {
    if (data.code === 401) {
      // Token过期，跳转登录
      localStorage.removeItem('token')
      window.location.href = '/login'
    } else if (data.code === 403) {
      // 权限不足
      alert('权限不足：' + data.message)
    } else {
      // 其他错误
      alert('错误：' + data.message)
    }
    throw new Error(data.message)
  }
  
  return data
}

// 管理员API专用函数
async function adminApiRequest(endpoint, options = {}) {
  if (!requireAdmin()) return null
  
  return await apiRequest(`${ADMIN_API_BASE_URL}${endpoint}`, options)
}
```

### 核心管理功能示例
```javascript
// 获取用户列表
async function getUserList(page = 1, limit = 20, filters = {}) {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    ...filters
  })
  
  return await adminApiRequest(`/users?${params}`)
}

// 更新用户管理员权限
async function updateUserAdminStatus(userId, isAdmin) {
  return await adminApiRequest(`/users/${userId}/admin`, {
    method: 'PUT',
    body: JSON.stringify({ is_admin: isAdmin })
  })
}

// 创建商品
async function createProduct(productData) {
  return await adminApiRequest('/products', {
    method: 'POST',
    body: JSON.stringify(productData)
  })
}
```

## 🎯 管理员后台开发重点

### 必须实现的页面
1. **🏠 仪表板** - 系统概览、统计数据
2. **👥 用户管理** - 用户列表、详情、权限设置
3. **🤝 代理管理** - 代理等级设置、区域分配、统计查看
4. **🛍️ 商品管理** - 商品CRUD、上下架控制
5. **📊 订单管理** - 订单查看、状态管理、退款审核
6. **📚 内容管理** - 书籍和单词管理（待开发）

### 技术建议
- **前端框架**: React/Vue.js + Ant Design/Element UI
- **状态管理**: Redux/Vuex + 持久化存储
- **路由守卫**: 基于`is_admin`字段的权限控制
- **API管理**: Axios + 统一错误处理
- **表格组件**: 支持分页、筛选、排序的数据表格

## 📋 下一步开发计划

### 后端任务
1. ✅ 实现数据库模型和连接
2. ✅ 完成微信小程序登录逻辑
3. ✅ 实现JWT Token生成和验证
4. ✅ 实现管理员权限验证中间件
5. ✅ 完成用户管理、代理管理、商品管理API
6. 🔲 完成订单管理API（审核退款等）
7. 🔲 完成书籍和单词管理API
8. 🔲 添加单元测试

### 前端任务  
1. 🔲 **🎯 搭建管理员后台Web界面**
2. 🔲 **🎯 实现用户管理页面**
3. 🔲 **🎯 实现代理管理页面**
4. 🔲 **🎯 实现商品管理页面**
5. 🔲 实现登录和权限验证
6. 🔲 实现统一的API调用封装
7. 🔲 完善错误处理和用户体验

> **🚨 重要提示**: 管理员后台的前端开发应该优先关注第10节列出的admin API，这些是核心管理功能，已完全实现并可以直接使用。