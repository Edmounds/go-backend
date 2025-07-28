package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"miniprogram/controllers"
	"time"
)

func CreateWord() {
	// 初始化MongoDB连接
	controllers.InitMongoDB()
	collection := controllers.GetCollection("words")

	fmt.Println("开始创建新单词...")

	// 创建一个新的单词文档
	word := bson.D{
		{"word_name", "banana"},
		{"word_meaning", "香蕉"},
		{"pronunciation_url", ""},
		{"img_url", ""},
		{"unit_id", ""},
		{"book_id", ""},
	}

	// 1. 插入新单词文档
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, word)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("新单词已成功插入到数据库中")
	controllers.CloseMongoDB()
}

func CreateBook() {
	controllers.InitMongoDB()
	collection := controllers.GetCollection("books")

	// 创建一个新的书籍文档
	book := bson.D{
		{"book_name", "七年级上册"},
		{"book_version", "人教版"},
		{"units", bson.A{
			bson.D{{"_id", primitive.NewObjectID()}, {"unit_name", "第一单元"}},
			bson.D{{"_id", primitive.NewObjectID()}, {"unit_name", "第二单元"}},
			bson.D{{"_id", primitive.NewObjectID()}, {"unit_name", "第三单元"}},
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, book)
	if err != nil {
		log.Fatal(err)
	}
	controllers.CloseMongoDB()
}

func CreateUser() {
	controllers.InitMongoDB()
	collection := controllers.GetCollection("users")

	// 创建一个新的用户文档
	user := bson.D{
		{"user_name", "cqc"},
		{"user_password", "123456"},
		{"class", "3"},
		{"age", 18},
		{"school", "cqup"},
		{"phone", "131234567890"},
		{"agent_level", 2},
		{"collected_cards", bson.A{
			bson.D{{"card_id", ""}},
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("新用户已成功插入到数据库中")
	controllers.CloseMongoDB()
}
