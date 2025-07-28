# Go MiniProgram 后端指导

## 项目概览

这是一个使用 Go 和 MongoDB 构建的微型程序后端服务。项目结构采用模块化设计，分离了数据库交互、控制器逻辑、认证处理和路由处理。

## 架构和组件

### 目录结构
- `/controllers`: 包含与数据库交互的控制器和认证相关代码
- `/database`: 包含数据库初始化脚本和JSON模型模板
- `/routes`: 预留用于API路由定义（尚未实现）
- `/utils`: 预留用于通用工具函数（尚未实现）

### 核心组件

1. **数据库连接管理**
   - 位置: `controllers/database_controller.go`
   - 功能: 提供MongoDB连接初始化、关闭和集合访问
   - 关键函数: `InitMongoDB()`, `CloseMongoDB()`, `GetCollection()`

2. **认证系统**
   - 位置: `controllers/auth_controller.go`
   - 功能: JWT令牌生成、验证和刷新
   - 关键函数: `GenerateToken()`, `ValidateToken()`, `RefreshToken()`

3. **数据模型**
   - 数据使用JSON格式定义，位于 `database/` 目录下的JSON文件
   - 主要模型:
     - 用户 (`user.json`): 包含用户信息、收藏卡片和地址
     - 单词 (`words.json`): 包含单词、发音URL和图片URL
     - 书籍 (`books.json`): 包含书籍和单元的层次结构

4. **数据库初始化**
   - 位置: `database/database_creator.go`
   - 功能: 初始化MongoDB并插入示例数据
   - 关键函数: `CreateUser()`, `CreateWord()`, `CreateBook()`

## 数据流

1. 应用启动时，通过 `InitMongoDB()` 连接到本地MongoDB实例
2. 通过 `GetCollection(collectionName)` 获取特定集合的引用
3. 使用MongoDB驱动执行CRUD操作
4. 用户认证流程通过JWT令牌实现，包括生成、验证和刷新令牌

## 开发工作流

### 构建和运行

对于Linux环境构建:
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o myserver main.go
```

### 数据库连接

默认配置连接到本地MongoDB:
```go
client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
```

### 数据模型模式

数据模型在JSON文件中定义，表示MongoDB中的文档结构:
- 用户模型（`user.json`）: 包括个人信息和收藏内容
- 单词模型（`words.json`）: 包括单词、发音和图像URL
- 书籍模型（`books.json`）: 组织为书籍和单元的层次结构

## 项目约定

1. **输出要求**
   - 所有代码应符合Go语言的最佳实践
   - 使用 `go fmt` 格式化代码
   - 不要修改现有项目代码
   - 不要新增任何文件
   - 尽可能阅读更多的相关项目文件
   - 用简单易懂的中文与用户交流

2. **错误处理**
   - 使用 `log.Fatal` 处理关键错误，特别是数据库操作
   - 示例: 
     ```go
     if err != nil {
         log.Fatal(err)
     }
     ```

3. **上下文使用**
   - 数据库操作使用带超时的上下文:
     ```go
     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()
     ```

4. **认证处理**
   - 使用JWT实现认证，密钥存储在代码中:
     ```go
     var jwtKey = []byte("miniprogram_secret_key")
     ```

5. **模块化设计**
   - 数据库逻辑与业务逻辑分离
   - 控制器与路由分离



## 关键依赖

- **MongoDB驱动**: `go.mongodb.org/mongo-driver v1.17.4` - 用于MongoDB交互
- **JWT库**: `github.com/golang-jwt/jwt/v5 v5.2.3` - 用于认证令牌处理
- **Go版本**: 1.24.5

## 集成点

1. **MongoDB数据库**: 应用配置为连接本地MongoDB实例 (`mongodb://localhost:27017`)
2. **数据库名称**: 使用名为 "miniprogram" 的数据库
   ```go
   client.Database("miniprogram").Collection(collectionName)
   ```
