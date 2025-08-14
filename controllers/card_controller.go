package controllers

import (
	"miniprogram/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUnitWordsHandler 获取指定单元的所有单词列表
func GetUnitWordsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取unit_id参数
		unitIDStr := c.Param("unit_id")

		// 将字符串转换为ObjectID
		unitID, err := primitive.ObjectIDFromHex(unitIDStr)
		if err != nil {
			BadRequestResponse(c, "无效的单元ID格式", err)
			return
		}

		// 查询该单元的所有单词
		collection := GetCollection("words")
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建查询条件
		filter := bson.M{"unit_id": unitID}

		// 执行查询
		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			InternalServerErrorResponse(c, "查询单词列表失败", err)
			return
		}
		defer cursor.Close(ctx)

		// 解析结果
		var words []models.Word
		if err = cursor.All(ctx, &words); err != nil {
			InternalServerErrorResponse(c, "解析单词数据失败", err)
			return
		}

		// 返回结果
		SuccessResponse(c, "获取单词列表成功", gin.H{
			"unit_id":    unitIDStr,
			"words":      words,
			"word_count": len(words),
		})
	}
}

// GetWordCardHandler 获取指定单词的详细信息（包括图片）
func GetWordCardHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取word_name参数
		wordName := c.Param("word_name")

		if wordName == "" {
			BadRequestResponse(c, "单词名称不能为空", nil)
			return
		}

		// 查询指定单词的详细信息
		collection := GetCollection("words")
		ctx, cancel := CreateDBContext()
		defer cancel()

		// 构建查询条件
		filter := bson.M{"word_name": wordName}

		var word models.Word
		err := collection.FindOne(ctx, filter).Decode(&word)
		if err != nil {
			NotFoundResponse(c, "未找到指定单词", err)
			return
		}

		// 返回单词详细信息，特别是图片URL
		SuccessResponse(c, "获取单词信息成功", gin.H{
			"word_name":         word.WordName,
			"word_meaning":      word.WordMeaning,
			"pronunciation_url": word.PronunciationURL,
			"img_url":           word.ImgURL,
			"unit_id":           word.UnitID.Hex(),
			"book_id":           word.BookID.Hex(),
		})
	}
}

// GetWordsByUnitNameHandler 通过单元名称获取单词列表（备用方案）
func GetWordsByUnitNameHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取unit_name参数
		unitName := c.Query("unit_name")
		bookName := c.Query("book_name")

		if unitName == "" {
			BadRequestResponse(c, "单元名称不能为空", nil)
			return
		}

		// 首先根据单元名称查找unit_id
		unitCollection := GetCollection("units")
		ctx, cancel := CreateDBContext()
		defer cancel()

		unitFilter := bson.M{"unit_name": unitName}
		if bookName != "" {
			// 如果提供了书籍名称，先查找book_id
			bookCollection := GetCollection("books")
			bookFilter := bson.M{"book_name": bookName}

			var book struct {
				ID primitive.ObjectID `bson:"_id"`
			}

			err := bookCollection.FindOne(ctx, bookFilter).Decode(&book)
			if err != nil {
				NotFoundResponse(c, "未找到指定书籍", err)
				return
			}

			unitFilter["book_id"] = book.ID
		}

		var unit struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		err := unitCollection.FindOne(ctx, unitFilter).Decode(&unit)
		if err != nil {
			NotFoundResponse(c, "未找到指定单元", err)
			return
		}

		// 然后查询该单元的所有单词
		wordsCollection := GetCollection("words")
		wordsFilter := bson.M{"unit_id": unit.ID}

		cursor, err := wordsCollection.Find(ctx, wordsFilter)
		if err != nil {
			InternalServerErrorResponse(c, "查询单词列表失败", err)
			return
		}
		defer cursor.Close(ctx)

		var words []models.Word
		if err = cursor.All(ctx, &words); err != nil {
			InternalServerErrorResponse(c, "解析单词数据失败", err)
			return
		}

		// 返回结果
		SuccessResponse(c, "获取单词列表成功", gin.H{
			"unit_name":  unitName,
			"book_name":  bookName,
			"unit_id":    unit.ID.Hex(),
			"words":      words,
			"word_count": len(words),
		})
	}
}
