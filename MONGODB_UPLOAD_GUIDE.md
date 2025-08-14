# MongoDB 数据上传指南（简化版）

## 概述

此指南适用于手动上传图片到 OBS 的情况，工具只负责：
1. 解析单词图片文件名导入到 MongoDB
2. 生成重命名后的图片文件夹（文件名为数据库ID）
3. 在 MongoDB 中存储 OBS 图片 URL

## 第一步：配置服务器 MongoDB 连接

### 方式1：使用环境变量（推荐）

创建 `.env` 文件：

```bash
# MongoDB 配置
MONGODB_URL=mongodb://username:password@your-server-ip:27017/miniprogram_db

# 或者使用 MongoDB Atlas 云服务
# MONGODB_URL=mongodb+srv://username:password@cluster.mongodb.net/miniprogram_db
```

### 方式2：修改配置文件

直接修改 `config/config.go` 中的默认值：

```go
MongoDBURL: getEnv("MONGODB_URL", "mongodb://your-server-url:27017"),
```

## 第二步：准备数据

确保您的单词卡图片文件命名格式正确：

**仅支持格式**：`单词 词性. 中文释义.jpg`
- `a art. 一个.png`
- `about prep. 关于；adv. 大约.png`
- `beautiful adj. 美丽的.png`

## 第三步：安装 Python 依赖

```bash
cd test/tools
pip install pymongo
```

## 第四步：执行数据导入和图片重命名

使用新的简化工具：

```bash
cd test/tools
python simple_word_importer.py 七上/ [OBS基础URL]
```

**示例**：
```bash
# 基本用法（不设置OBS URL）
python simple_word_importer.py 七上/

# 包含OBS基础URL，自动生成完整图片URL
python simple_word_importer.py 七上/ https://your-bucket.obs.cn-north-4.myhuaweicloud.com/word_images/
```

## 第五步：手动上传图片到 OBS

1. 工具会在原目录旁创建一个 `七上_renamed` 文件夹
2. 该文件夹中的图片文件名已重命名为数据库 ID（如：`65a1b2c3d4e5f6789abcdef0.jpg`）
3. 手动将 `七上_renamed` 文件夹中的图片上传到您的 OBS 存储桶

## 工作流程

```
原图片文件 → 解析文件名 → 导入MongoDB → 生成重命名文件夹 → 手动上传到OBS
```

**详细步骤**：

1. **运行导入工具**：
   ```bash
   python simple_word_importer.py 七上/ https://your-bucket.obs.cn-north-4.myhuaweicloud.com/word_images/
   ```

2. **检查结果**：
   - MongoDB 中生成单词记录，包含 OBS URL
   - 本地生成 `七上_renamed` 文件夹，包含重命名的图片

3. **手动上传图片**：
   - 将 `七上_renamed` 文件夹中的图片上传到 OBS 对应路径
   - 确保 OBS 中的路径与数据库中的 URL 一致

## 常用命令

### 检查数据库连接
```bash
cd test/tools
python -c "
from simple_word_importer import SimpleWordImporter
import os
mongo_url = os.getenv('MONGODB_URL', 'mongodb://localhost:27017')
try:
    importer = SimpleWordImporter(mongo_url)
    print('✅ MongoDB 连接成功')
    importer.close()
except Exception as e:
    print(f'❌ MongoDB 连接失败: {e}')
"
```

### 仅导入单词数据（不设置图片URL）
```bash
python simple_word_importer.py 七上/
```

### 导入数据并设置OBS图片URL
```bash
python simple_word_importer.py 七上/ https://your-bucket.obs.cn-north-4.myhuaweicloud.com/word_images/
```

## 注意事项

1. **文件格式**：仅支持 `单词 词性. 中文释义.jpg` 格式
2. **网络连接**：确保能够连接到您的服务器 MongoDB
3. **图片上传**：需要手动上传重命名后的图片到 OBS
4. **URL一致性**：确保 OBS 中的文件路径与数据库中的 URL 一致

## 生成的数据库结构

```json
{
  "_id": "65a1b2c3d4e5f6789abcdef0",
  "word_name": "hello",
  "word_meaning": "你好",
  "part_of_speech": "interj",
  "pronunciation_url": "",
  "img_url": "https://your-bucket.obs.cn-north-4.myhuaweicloud.com/word_images/65a1b2c3d4e5f6789abcdef0.jpg",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## 故障排除

### MongoDB 连接问题
- 检查服务器 IP 和端口
- 确认用户名密码正确
- 检查防火墙设置
- 确认 MongoDB 服务已启动

### 文件格式问题
- 确保文件名严格符合 `单词 词性. 中文释义.jpg` 格式
- 检查文件名中是否有特殊字符
- 确认文件扩展名正确