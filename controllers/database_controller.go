package controllers

import (
	"context"
	"log"
	"miniprogram/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBQuery struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoDBQueryPractice 创建MongoDB查询练习实例
func NewMongoDBQuery(client *mongo.Client) *MongoDBQuery {
	return &MongoDBQuery{
		client: client,
		db:     client.Database("miniprogram_db"),
	}
}

// 全局变量，保存MongoDB客户端连接
var mongoClient *mongo.Client

// 初始化MongoDB连接
func InitMongoDB() {
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 从配置文件获取MongoDB连接字符串
	cfg := config.GetConfig()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURL))
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	mongoClient = client
	log.Println("成功连接到MongoDB")
}

// GetCollection 获取指定名称的集合
func GetCollection(collectionName string) *mongo.Collection {
	if mongoClient == nil {
		log.Fatal("MongoDB客户端未初始化")
	}
	return mongoClient.Database("miniprogram_db").Collection(collectionName)
}

// 关闭MongoDB连接
func CloseMongoDB() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("MongoDB连接已关闭")
	}
}
