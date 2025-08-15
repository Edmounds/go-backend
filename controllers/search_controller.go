package controllers

import (
	"context"
	"miniprogram/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SearchHandler 综合搜索处理器 - 支持单词、课本、订单的模糊搜索
func SearchHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SearchRequest

		// 绑定JSON请求体
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数错误", err)
			return
		}

		// 设置默认分页参数
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Limit <= 0 || req.Limit > 100 {
			req.Limit = 20
		}

		// 创建数据库上下文
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建响应结构
		var response models.SearchResponse
		response.Page = req.Page
		response.Limit = req.Limit

		// 计算分页偏移量
		skip := (req.Page - 1) * req.Limit

		// 根据搜索类型执行不同的搜索逻辑
		switch strings.ToLower(req.Type) {
		case "word":
			searchWords(ctx, req.Query, skip, req.Limit, &response)
		case "book":
			searchBooks(ctx, req.Query, skip, req.Limit, &response)
		case "order":
			searchOrders(ctx, req.Query, skip, req.Limit, &response, c)
		case "all":
			// 搜索所有类型，但限制每种类型的结果数量
			searchWords(ctx, req.Query, 0, req.Limit/3, &response)
			searchBooks(ctx, req.Query, 0, req.Limit/3, &response)
			searchOrders(ctx, req.Query, 0, req.Limit/3, &response, c)
		default:
			BadRequestResponse(c, "不支持的搜索类型", nil)
			return
		}

		SuccessResponse(c, "搜索完成", response)
	}
}

// searchWords 搜索单词
func searchWords(ctx context.Context, query string, skip, limit int, response *models.SearchResponse) {
	collection := GetCollection("words")

	// 构建模糊搜索过滤器 - 支持单词名称和含义的模糊搜索
	filter := bson.M{
		"$or": []bson.M{
			{"word_name": bson.M{"$regex": query, "$options": "i"}},
			{"word_meaning": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// 设置查询选项
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	// 执行查询
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	// 解析结果
	var words []models.Word
	if err := cursor.All(ctx, &words); err == nil {
		response.Words = words

		// 获取总数量
		if count, err := collection.CountDocuments(ctx, filter); err == nil {
			response.Total += count
		}
	}
}

// searchBooks 搜索课本
func searchBooks(ctx context.Context, query string, skip, limit int, response *models.SearchResponse) {
	collection := GetCollection("books")

	// 构建模糊搜索过滤器 - 支持书名、版本、描述、作者、出版社的模糊搜索
	filter := bson.M{
		"$or": []bson.M{
			{"book_name": bson.M{"$regex": query, "$options": "i"}},
			{"book_version": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"author": bson.M{"$regex": query, "$options": "i"}},
			{"publisher": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// 设置查询选项
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	// 执行查询
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	// 解析结果
	var books []models.Book
	if err := cursor.All(ctx, &books); err == nil {
		response.Books = books

		// 获取总数量
		if count, err := collection.CountDocuments(ctx, filter); err == nil {
			response.Total += count
		}
	}
}

// searchOrders 搜索订单
func searchOrders(ctx context.Context, query string, skip, limit int, response *models.SearchResponse, c *gin.Context) {
	collection := GetCollection("orders")

	// 获取用户信息 - 只能搜索当前用户的订单
	userID := c.Param("user_id")
	if userID == "" {
		// 如果没有用户ID，尝试从查询参数获取
		userID = c.Query("user_id")
	}

	// 构建基础过滤器
	baseFilter := bson.M{}
	if userID != "" {
		baseFilter["user_openid"] = userID
	}

	// 首先搜索商品名称匹配的订单
	// 这需要通过订单中的商品ID来查找商品名称
	productsCollection := GetCollection("products")
	productFilter := bson.M{
		"name": bson.M{"$regex": query, "$options": "i"},
	}

	productCursor, err := productsCollection.Find(ctx, productFilter)
	if err == nil {
		var products []models.Product
		if err := productCursor.All(ctx, &products); err == nil {
			// 获取匹配的商品ID列表
			var productIDs []string
			for _, product := range products {
				productIDs = append(productIDs, product.ProductID)
			}

			if len(productIDs) > 0 {
				// 构建订单搜索过滤器 - 根据商品ID搜索订单
				filter := bson.M{
					"$and": []bson.M{
						baseFilter,
						{"items.product_id": bson.M{"$in": productIDs}},
					},
				}

				// 设置查询选项
				opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"created_at": -1})

				// 执行查询
				cursor, err := collection.Find(ctx, filter, opts)
				if err == nil {
					defer cursor.Close(ctx)

					// 解析结果
					var orders []models.Order
					if err := cursor.All(ctx, &orders); err == nil {
						response.Orders = orders

						// 获取总数量
						if count, err := collection.CountDocuments(ctx, filter); err == nil {
							response.Total += count
						}
					}
				}
			}
		}
	}
	productCursor.Close(ctx)
}

// SearchWordsHandler 单独的单词搜索处理器
func SearchWordsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			BadRequestResponse(c, "搜索关键词不能为空", nil)
			return
		}

		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		// 创建数据库上下文
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建响应结构
		var response models.SearchResponse
		response.Page = page
		response.Limit = limit

		// 计算分页偏移量
		skip := (page - 1) * limit

		// 搜索单词
		searchWords(ctx, query, skip, limit, &response)

		SuccessResponse(c, "单词搜索完成", response)
	}
}

// SearchBooksHandler 单独的课本搜索处理器
func SearchBooksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			BadRequestResponse(c, "搜索关键词不能为空", nil)
			return
		}

		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		// 创建数据库上下文
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建响应结构
		var response models.SearchResponse
		response.Page = page
		response.Limit = limit

		// 计算分页偏移量
		skip := (page - 1) * limit

		// 搜索课本
		searchBooks(ctx, query, skip, limit, &response)

		SuccessResponse(c, "课本搜索完成", response)
	}
}

// SearchOrdersHandler 单独的订单搜索处理器
func SearchOrdersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			BadRequestResponse(c, "搜索关键词不能为空", nil)
			return
		}

		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		// 创建数据库上下文
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建响应结构
		var response models.SearchResponse
		response.Page = page
		response.Limit = limit

		// 计算分页偏移量
		skip := (page - 1) * limit

		// 搜索订单
		searchOrders(ctx, query, skip, limit, &response, c)

		SuccessResponse(c, "订单搜索完成", response)
	}
}
