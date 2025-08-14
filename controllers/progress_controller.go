package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateProgressRequest 更新学习进度请求
type UpdateProgressRequest struct {
	CurrentUnit  string   `json:"current_unit"`
	LearnedWords []string `json:"learned_words"`
}

// GetProgressHandler 获取用户学习进度处理器
func GetProgressHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 user_id 参数，注意：这里的 user_id 实际上是微信的 openID，不是 MongoDB 的 _id
		openID := c.Param("user_id")

		// 根据openID查询用户信息（只需要progress字段）
		user, err := GetUserByOpenID(openID)
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
		var req UpdateProgressRequest

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
				"updated_at":             time.Now(),
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

		// TODO: 实现书籍列表查询逻辑
		// 1. 从数据库获取书籍列表
		// 2. 实现分页
		// 3. 返回书籍信息

		SuccessResponse(c, "获取书籍列表成功", gin.H{
			"books": []gin.H{
				{
					"_id":              "BOOK001",
					"title":            "大学英语四级词汇",
					"description":      "涵盖大学英语四级考试所需的核心词汇，包含详细释义和例句",
					"level":            "intermediate",
					"total_words":      2000,
					"units":            20,
					"cover_image":      "https://example.com/book1_cover.jpg",
					"author":           "英语教学专家组",
					"publisher":        "教育出版社",
					"publication_date": "2023-01-01",
				},
				{
					"_id":              "BOOK002",
					"title":            "高中英语核心词汇",
					"description":      "高中阶段必备英语词汇，按主题分类学习",
					"level":            "beginner",
					"total_words":      1500,
					"units":            15,
					"cover_image":      "https://example.com/book2_cover.jpg",
					"author":           "高中英语教研组",
					"publisher":        "学习出版社",
					"publication_date": "2023-03-01",
				},
			},
			"pagination": gin.H{
				"current_page":   page,
				"total_pages":    3,
				"total_items":    5,
				"items_per_page": limit,
			},
		})
	}
}

// GetBookWordsHandler 获取书籍单词处理器
func GetBookWordsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookID := c.Param("book_id")
		unitID := c.Query("unit_id")

		// TODO: 实现书籍单词查询逻辑
		// 1. 根据书籍ID查询单词
		// 2. 根据单元ID筛选（如果提供）
		// 3. 返回单词列表

		// 构建响应数据，根据是否提供unitID来筛选结果
		var unitFilter string
		if unitID != "" {
			unitFilter = unitID
		} else {
			unitFilter = "UNIT001" // 默认单元
		}

		SuccessResponse(c, "获取单词列表成功", gin.H{
			"words": []gin.H{
				{
					"_id":           "WORD001",
					"word":          "computer",
					"pronunciation": "/kəmˈpjuːtər/",
					"definition":    "电子计算机，电脑",
					"example_sentences": []string{
						"I use my computer every day for work.",
						"The computer is running slowly today.",
					},
					"difficulty": "basic",
					"unit_id":    unitFilter,
					"book_id":    bookID,
				},
				{
					"_id":           "WORD002",
					"word":          "technology",
					"pronunciation": "/tekˈnɒlədʒi/",
					"definition":    "技术，科技",
					"example_sentences": []string{
						"Technology has changed our lives.",
						"Modern technology is advancing rapidly.",
					},
					"difficulty": "intermediate",
					"unit_id":    unitFilter,
					"book_id":    bookID,
				},
			},
			"total_count": 50,
			"unit_info": gin.H{
				"unit_id":     unitFilter,
				"unit_name":   "科技与生活",
				"total_words": 50,
			},
		})
	}
}
