package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DatabaseName = "miniprogram_db"

// DatabaseCreator 数据库创建器结构体
type DatabaseCreator struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewDatabaseCreator 创建数据库创建器实例
func NewDatabaseCreator(client *mongo.Client) *DatabaseCreator {
	return &DatabaseCreator{
		client: client,
		db:     client.Database(DatabaseName),
	}
}

// CreateAllCollections 创建所有集合
func (dc *DatabaseCreator) CreateAllCollections(ctx context.Context) error {
	log.Println("开始创建所有MongoDB集合...")

	// 创建所有集合
	if err := dc.CreateUsersCollection(ctx); err != nil {
		return fmt.Errorf("创建用户集合失败: %v", err)
	}

	if err := dc.CreateProductsCollection(ctx); err != nil {
		return fmt.Errorf("创建商品集合失败: %v", err)
	}

	if err := dc.CreateAddressesCollection(ctx); err != nil {
		return fmt.Errorf("创建地址集合失败: %v", err)
	}

	if err := dc.CreateCartsCollection(ctx); err != nil {
		return fmt.Errorf("创建购物车集合失败: %v", err)
	}

	if err := dc.CreateOrdersCollection(ctx); err != nil {
		return fmt.Errorf("创建订单集合失败: %v", err)
	}

	if err := dc.CreateReferralsCollection(ctx); err != nil {
		return fmt.Errorf("创建推荐码集合失败: %v", err)
	}

	if err := dc.CreateCommissionsCollection(ctx); err != nil {
		return fmt.Errorf("创建佣金集合失败: %v", err)
	}

	if err := dc.CreateBooksCollection(ctx); err != nil {
		return fmt.Errorf("创建书籍集合失败: %v", err)
	}

	if err := dc.CreateUnitsCollection(ctx); err != nil {
		return fmt.Errorf("创建单元集合失败: %v", err)
	}

	if err := dc.CreateWordsCollection(ctx); err != nil {
		return fmt.Errorf("创建单词集合失败: %v", err)
	}

	log.Println("所有MongoDB集合创建完成!")
	return nil
}

// CreateUsersCollection 创建用户集合
func (dc *DatabaseCreator) CreateUsersCollection(ctx context.Context) error {
	collectionName := "users"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "referral_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "referred_by", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "school", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "agent_level", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":             primitive.NewObjectID(),
		"user_name":       "john_doe",
		"user_password":   "$2b$10$hashed_password_example",
		"class":           "计算机科学与技术1班",
		"age":             20,
		"school":          "北京大学",
		"phone":           "13800138000",
		"agent_level":     0,
		"referral_code":   "JOHN123",
		"referred_by":     "",
		"collected_cards": bson.A{},
		"addresses":       bson.A{},
		"progress": bson.M{
			"current_unit":     "Unit 1",
			"current_sentence": "Hello, how are you?",
			"learned_words":    bson.A{"hello", "how", "are", "you"},
		},
		"is_agent":        false,
		"agent_type":      nil,
		"managed_schools": bson.A{},
		"managed_regions": bson.A{},
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateProductsCollection 创建商品集合
func (dc *DatabaseCreator) CreateProductsCollection(ctx context.Context) error {
	collectionName := "products"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "product_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "price", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":         primitive.NewObjectID(),
		"product_id":  "PROD001",
		"name":        "英语学习卡片套装",
		"price":       99.9,
		"description": "包含1000个常用英语单词卡片",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateAddressesCollection 创建地址集合
func (dc *DatabaseCreator) CreateAddressesCollection(ctx context.Context) error {
	collectionName := "addresses"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_default", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据 (需要先有用户数据)
	sampleData := bson.M{
		"_id":            primitive.NewObjectID(),
		"user_id":        primitive.NewObjectID(), // 实际使用时应该引用真实的用户ID
		"recipient_name": "张三",
		"phone":          "13800138001",
		"province":       "北京市",
		"city":           "北京市",
		"district":       "海淀区",
		"street":         "中关村大街1号",
		"postal_code":    "100000",
		"is_default":     true,
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateCartsCollection 创建购物车集合
func (dc *DatabaseCreator) CreateCartsCollection(ctx context.Context) error {
	collectionName := "carts"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "cart_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":     primitive.NewObjectID(),
		"cart_id": "CART001",
		"user_id": primitive.NewObjectID(), // 实际使用时应该引用真实的用户ID
		"items": bson.A{
			bson.M{
				"product_id": primitive.NewObjectID(), // 实际使用时应该引用真实的商品ID
				"quantity":   1,
				"price":      99.9,
			},
		},
		"total_price": 99.9,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateOrdersCollection 创建订单集合
func (dc *DatabaseCreator) CreateOrdersCollection(ctx context.Context) error {
	collectionName := "orders"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "order_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "referral_code", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":        primitive.NewObjectID(),
		"order_id":   "ORDER001",
		"user_id":    primitive.NewObjectID(), // 实际使用时应该引用真实的用户ID
		"address_id": primitive.NewObjectID(), // 实际使用时应该引用真实的地址ID
		"products": bson.A{
			bson.M{
				"product_id": primitive.NewObjectID(), // 实际使用时应该引用真实的商品ID
				"quantity":   1,
				"price":      99.9,
			},
		},
		"subtotal":          99.9,
		"total_price":       109.9,
		"discount_applied":  0,
		"referral_discount": 5,
		"shipping_fee":      10,
		"price_breakdown": bson.M{
			"subtotal": 99.9,
			"discount_details": bson.A{
				bson.M{
					"type":        "referral",
					"amount":      5,
					"description": "推荐码折扣",
				},
			},
			"shipping_fee": 10,
			"final_total":  109.9,
		},
		"status":                 "pending",
		"payment_method":         "wechat_pay",
		"referral_code":          "",
		"referral_bonus_applied": false,
		"created_at":             time.Now(),
		"updated_at":             time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateReferralsCollection 创建推荐码集合
func (dc *DatabaseCreator) CreateReferralsCollection(ctx context.Context) error {
	collectionName := "referrals"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "referral_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":           primitive.NewObjectID(),
		"referral_code": "JOHN123",
		"user_id":       primitive.NewObjectID(), // 实际使用时应该引用真实的用户ID
		"used_by":       bson.A{},
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateCommissionsCollection 创建佣金集合
func (dc *DatabaseCreator) CreateCommissionsCollection(ctx context.Context) error {
	collectionName := "commissions"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "commission_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "date", Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":           primitive.NewObjectID(),
		"commission_id": "COMM001",
		"user_id":       primitive.NewObjectID(), // 实际使用时应该引用真实的用户ID
		"amount":        10.5,
		"date":          time.Now(),
		"status":        "pending",
		"type":          "referral",
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateBooksCollection 创建书籍集合
func (dc *DatabaseCreator) CreateBooksCollection(ctx context.Context) error {
	collectionName := "books"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "book_name", Value: 1}, {Key: "book_version", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":          primitive.NewObjectID(),
		"book_name":    "新概念英语第一册",
		"book_version": "2024版",
		"units": bson.A{
			primitive.NewObjectID(), // 实际使用时应该引用真实的单元ID
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateUnitsCollection 创建单元集合
func (dc *DatabaseCreator) CreateUnitsCollection(ctx context.Context) error {
	collectionName := "units"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "book_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "unit_name", Value: 1}, {Key: "book_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":        primitive.NewObjectID(),
		"unit_name":  "Unit 1: Greetings",
		"book_id":    primitive.NewObjectID(), // 实际使用时应该引用真实的书籍ID
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}

// CreateWordsCollection 创建单词集合
func (dc *DatabaseCreator) CreateWordsCollection(ctx context.Context) error {
	collectionName := "words"
	log.Printf("创建集合: %s", collectionName)

	collection := dc.db.Collection(collectionName)

	// 创建索引
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "word_name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "unit_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "book_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "word_name", Value: 1}, {Key: "unit_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 插入示例数据
	sampleData := bson.M{
		"_id":               primitive.NewObjectID(),
		"word_name":         "hello",
		"word_meaning":      "你好",
		"pronunciation_url": "https://example.com/audio/hello.mp3",
		"img_url":           "https://example.com/images/hello.jpg",
		"unit_id":           primitive.NewObjectID(), // 实际使用时应该引用真实的单元ID
		"book_id":           primitive.NewObjectID(), // 实际使用时应该引用真实的书籍ID
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	}

	_, err = collection.InsertOne(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("插入示例数据失败: %v", err)
	}

	log.Printf("集合 %s 创建成功", collectionName)
	return nil
}
