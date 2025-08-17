package controllers

import (
	"fmt"
	"miniprogram/middlewares"
	"miniprogram/models"
	"miniprogram/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetProgressHandler 获取用户学习进度处理器
func GetProgressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")

		// 初始化用户服务
		userService := GetUserService()

		// 根据openID查询用户信息
		user, err := userService.FindUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 返回用户的学习进度信息
		SuccessResponse(c, "获取学习进度成功", gin.H{
			"openID":        openID,
			"current_unit":  user.Progress.CurrentUnit,
			"learned_words": user.Progress.LearnedWords,
			"total_words":   len(user.Progress.LearnedWords),
		})
	}
}

// UpdateProgressHandler 更新用户学习进度处理器
func UpdateProgressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")
		var req models.UpdateProgressRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 验证用户是否存在
		_, err := GetUserByOpenID(openID)
		if err != nil {
			NotFoundResponse(c, "用户不存在", err)
			return
		}

		// 只更新progress字段
		collection := GetCollection("users")
		ctx, cancel := CreateDBContext()
		defer cancel()

		update := bson.M{
			"$set": bson.M{
				"progress.current_unit":  req.CurrentUnit,
				"progress.learned_words": req.LearnedWords,
				"updated_at":             utils.GetCurrentUTCTime(),
			},
		}

		filter := bson.M{"openID": openID}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			InternalServerErrorResponse(c, "更新学习进度失败", err)
			return
		}

		SuccessResponse(c, "学习进度更新成功", gin.H{
			"openID":        openID,
			"current_unit":  req.CurrentUnit,
			"learned_words": req.LearnedWords,
			"total_words":   len(req.LearnedWords),
		})
	}
}

// GetBooksHandler 获取书籍列表处理器
func GetBooksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		// 从数据库获取书籍列表
		books, total, err := GetBooksList(page, limit)
		if err != nil {
			if middlewares.HandleError(err, "获取书籍列表失败", false) {
				InternalServerErrorResponse(c, "获取书籍列表失败", err)
				return
			}
		}

		totalPages := (total + int64(limit) - 1) / int64(limit)

		SuccessResponse(c, "获取书籍列表成功", gin.H{
			"books": books,
			"pagination": gin.H{
				"current_page":   page,
				"total_pages":    totalPages,
				"total_items":    total,
				"items_per_page": limit,
			},
		})
	}
}

// GetBookWordsHandler 获取书籍单词处理器
func GetBookWordsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID（从token中获取）
		userID, exists := c.Get("user_openid")
		if !exists {
			UnauthorizedResponse(c, "未找到用户身份信息", nil)
			return
		}

		bookID := c.Param("book_id")
		unitID := c.Query("unit_id")

		// 验证书籍ID格式
		bookObjectID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			BadRequestResponse(c, "无效的书籍ID格式", err)
			return
		}

		// 检查用户是否有访问该书籍的权限
		hasPermission, accessType, err := CheckUserBookPermission(userID.(string), bookObjectID)
		if err != nil {
			InternalServerErrorResponse(c, "检查用户权限失败", err)
			return
		}

		if !hasPermission {
			ForbiddenResponse(c, "您没有访问该书籍的权限，请先购买相关单词卡", nil)
			return
		}

		// 从数据库获取书籍单词
		words, unitInfo, err := GetBookWords(bookID, unitID)
		if err != nil {
			if middlewares.HandleError(err, "获取单词列表失败", false) {
				InternalServerErrorResponse(c, "获取单词列表失败", err)
				return
			}
		}

		SuccessResponse(c, "获取单词列表成功", gin.H{
			"words":       words,
			"total_count": len(words),
			"unit_info":   unitInfo,
			"access_type": accessType,
		})
	}
}

// ===== 数据库操作函数 =====

// GetBooksList 获取书籍列表
func GetBooksList(page, limit int) ([]models.Book, int64, error) {
	collection := GetCollection("books")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 计算跳过的文档数
	skip := (page - 1) * limit

	// 设置查询选项
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间倒序

	// 执行查询
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var books []models.Book
	if err = cursor.All(ctx, &books); err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

// GetBookWords 获取书籍单词
func GetBookWords(bookID, unitID string) ([]models.Word, map[string]interface{}, error) {
	wordsCollection := GetCollection("words")
	unitsCollection := GetCollection("units")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 构建查询条件
	filter := bson.M{}

	// 添加书籍ID过滤条件
	if bookID != "" {
		bookObjectID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			return nil, nil, fmt.Errorf("无效的书籍ID: %v", err)
		}
		filter["book_id"] = bookObjectID
	}

	// 添加单元ID过滤条件
	var unitObjectID primitive.ObjectID
	if unitID != "" {
		var err error
		unitObjectID, err = primitive.ObjectIDFromHex(unitID)
		if err != nil {
			return nil, nil, fmt.Errorf("无效的单元ID: %v", err)
		}
		filter["unit_id"] = unitObjectID
	}

	// 查询单词
	cursor, err := wordsCollection.Find(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var words []models.Word
	if err = cursor.All(ctx, &words); err != nil {
		return nil, nil, err
	}

	// 获取单元信息
	unitInfo := map[string]interface{}{
		"unit_id":     unitID,
		"unit_name":   "",
		"total_words": len(words),
	}

	if unitID != "" {
		var unit models.Unit
		err = unitsCollection.FindOne(ctx, bson.M{"_id": unitObjectID}).Decode(&unit)
		if err == nil {
			unitInfo["unit_name"] = unit.UnitName
		}
	}

	return words, unitInfo, nil
}
