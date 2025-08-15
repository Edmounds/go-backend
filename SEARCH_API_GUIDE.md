# 搜索API使用指南

本文档介绍新增的搜索API功能，支持对数据库中的单词、课本和订单进行模糊搜索。

## API端点

### 1. 综合搜索 API
- **URL**: `POST /api/search`
- **描述**: 支持多种类型的综合搜索
- **需要认证**: 否

#### 请求参数
```json
{
  "query": "搜索关键词",
  "type": "word|book|order|all",
  "page": 1,
  "limit": 20
}
```

#### 参数说明
- `query`: 搜索关键词（必填）
- `type`: 搜索类型（必填）
  - `word`: 仅搜索单词
  - `book`: 仅搜索课本
  - `order`: 仅搜索订单
  - `all`: 搜索所有类型
- `page`: 页码，默认为1
- `limit`: 每页结果数量，默认为20，最大100

#### 响应格式
```json
{
  "code": 200,
  "message": "搜索完成",
  "data": {
    "words": [...],    // 单词搜索结果（如果type为word或all）
    "books": [...],    // 课本搜索结果（如果type为book或all）
    "orders": [...],   // 订单搜索结果（如果type为order或all）
    "total": 25,       // 总结果数量
    "page": 1,         // 当前页码
    "limit": 20        // 每页数量
  }
}
```

### 2. 单词搜索 API
- **URL**: `GET /api/search/words?q=关键词&page=1&limit=20`
- **描述**: 专门用于单词的模糊搜索
- **需要认证**: 否

#### 搜索范围
- 单词名称 (`word_name`)
- 单词含义 (`word_meaning`)

#### 示例请求
```
GET /api/search/words?q=hello&page=1&limit=10
```

### 3. 课本搜索 API
- **URL**: `GET /api/search/books?q=关键词&page=1&limit=20`
- **描述**: 专门用于课本的模糊搜索
- **需要认证**: 否

#### 搜索范围
- 课本名称 (`book_name`)
- 课本版本 (`book_version`)
- 课本描述 (`description`)
- 作者 (`author`)
- 出版社 (`publisher`)

#### 示例请求
```
GET /api/search/books?q=英语&page=1&limit=10
```

### 4. 订单搜索 API
- **URL**: `GET /api/search/orders?q=关键词&page=1&limit=20&user_id=用户openid`
- **描述**: 根据商品名称搜索订单
- **需要认证**: 是

#### 搜索范围
- 订单中商品的名称 (`products.name`)

#### 查询参数
- `q`: 搜索关键词（必填）
- `user_id`: 用户openid（可选，如果提供则只搜索该用户的订单）
- `page`: 页码，默认为1
- `limit`: 每页结果数量，默认为20

#### 示例请求
```
GET /api/search/orders?q=学习卡&user_id=wx_openid_123&page=1&limit=10
```

## 搜索特性

### 模糊搜索
所有搜索都支持模糊匹配，不区分大小写。使用MongoDB的正则表达式功能实现。

### 分页支持
所有搜索API都支持分页，通过`page`和`limit`参数控制。

### 多字段搜索
- **单词搜索**: 在单词名称和含义中搜索
- **课本搜索**: 在课本名称、版本、描述、作者、出版社中搜索
- **订单搜索**: 通过商品名称查找相关订单

## 使用示例

### 1. 搜索包含"hello"的单词
```bash
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "hello",
    "type": "word",
    "page": 1,
    "limit": 10
  }'
```

### 2. 搜索英语相关的课本
```bash
curl -X GET "http://localhost:8080/api/search/books?q=英语&page=1&limit=5"
```

### 3. 搜索包含"学习卡"的订单
```bash
curl -X GET "http://localhost:8080/api/search/orders?q=学习卡&user_id=wx_openid_123" \
  -H "Authorization: Bearer your_jwt_token"
```

### 4. 综合搜索所有类型
```bash
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "英语",
    "type": "all",
    "page": 1,
    "limit": 30
  }'
```

## 注意事项

1. **订单搜索权限**: 订单搜索需要JWT认证，用户只能搜索自己的订单
2. **搜索性能**: 建议在数据库中为搜索字段创建文本索引以提高性能
3. **结果限制**: 单次查询最多返回100条结果
4. **字符编码**: 支持中文和英文搜索
5. **空结果**: 如果没有匹配结果，相应的数组将为空但不会报错

## 数据库索引建议

为了提高搜索性能，建议在MongoDB中创建以下索引：

```javascript
// 单词搜索索引
db.words.createIndex({ "word_name": "text", "word_meaning": "text" })

// 课本搜索索引
db.books.createIndex({ 
  "book_name": "text", 
  "book_version": "text", 
  "description": "text", 
  "author": "text", 
  "publisher": "text" 
})

// 商品名称索引（用于订单搜索）
db.products.createIndex({ "name": "text" })

// 订单用户索引
db.orders.createIndex({ "user_openid": 1 })
```