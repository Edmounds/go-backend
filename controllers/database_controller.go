package controllers

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 全局变量，保存MongoDB客户端连接
var mongoClient *mongo.Client

// 初始化MongoDB连接
func InitMongoDB() {
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接到本地 MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
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

// GetCollection 根据表名获取MongoDB集合
// 参数：collectionName - 集合名称（表名）
// 返回：对应的MongoDB集合对象
func GetCollection(collectionName string) *mongo.Collection {
	if mongoClient == nil {
		log.Fatal("MongoDB未初始化，请先调用InitMongoDB()")
	}

	// 返回指定数据库中的指定集合
	return mongoClient.Database("miniprogram").Collection(collectionName)
}

// GetUser 根据用户名获取用户信息
// 参数：username - 用户名

func GetUser(username string) (bson.M, error) {
	if mongoClient == nil {
		log.Fatal("MongoDB未初始化，请先调用InitMongoDB()")
	}

	collection := GetCollection("users")
	filter := bson.M{"user_name": username}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user bson.M
	err := collection.FindOne(ctx, filter).Decode(&user)
	return user, err
}

// 示例用法
func main() {
	// 初始化MongoDB连接
	InitMongoDB()
	defer CloseMongoDB()

	// 获取指定的集合
	collection := GetCollection("books")

	// 这里可以使用collection进行操作
	log.Println("成功连接到集合:", collection.Name())
}
