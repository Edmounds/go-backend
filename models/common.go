package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户结构体（统一定义）
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	OpenID         string             `bson:"openID" json:"openID"` // 微信openID，作为主要标识符
	UserName       string             `bson:"user_name" json:"user_name"`
	UserPassword   string             `bson:"user_password" json:"-"` // 不在JSON中显示密码
	Class          string             `bson:"class" json:"class"`
	Age            int                `bson:"age" json:"age"`
	School         string             `bson:"school" json:"school"`
	Phone          string             `bson:"phone" json:"phone"`
	City           string             `bson:"city" json:"city"`
	AgentLevel     int                `bson:"agent_level" json:"agent_level"`
	ReferralCode   string             `bson:"referral_code" json:"referral_code"`
	ReferredBy     string             `bson:"referred_by" json:"referred_by"`
	CollectedCards []string           `bson:"collected_cards" json:"collected_cards"`
	Addresses      []Address          `bson:"addresses" json:"addresses"`
	Progress       Progress           `bson:"progress" json:"progress"`
	IsAgent        bool               `bson:"is_agent" json:"is_agent"`

	ManagedSchools []string  `bson:"managed_schools" json:"managed_schools"`
	ManagedRegions []string  `bson:"managed_regions" json:"managed_regions"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

// Progress 学习进度结构体
type Progress struct {
	CurrentUnit  string   `bson:"current_unit" json:"current_unit"`
	LearnedWords []string `bson:"learned_words" json:"learned_words"`
}

// Address 地址结构体
type Address struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserOpenID    string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	RecipientName string             `bson:"recipient_name" json:"recipient_name"`
	Phone         string             `bson:"phone" json:"phone"`
	Province      string             `bson:"province" json:"province"`
	City          string             `bson:"city" json:"city"`
	District      string             `bson:"district" json:"district"`
	Street        string             `bson:"street" json:"street"`
	PostalCode    string             `bson:"postal_code" json:"postal_code"`
	IsDefault     bool               `bson:"is_default" json:"is_default"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// Word 单词结构体
type Word struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	WordName         string             `bson:"word_name" json:"word_name"`
	WordMeaning      string             `bson:"word_meaning" json:"word_meaning"`
	PronunciationURL string             `bson:"pronunciation_url" json:"pronunciation_url"`
	ImgURL           string             `bson:"img_url" json:"img_url"`
	UnitID           primitive.ObjectID `bson:"unit_id" json:"unit_id"`
	BookID           primitive.ObjectID `bson:"book_id" json:"book_id"`
	CreatedAt        time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// Product 商品结构体
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProductID   string             `bson:"product_id" json:"product_id"`
	Name        string             `bson:"name" json:"name"`
	Price       float64            `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	Stock       int                `bson:"stock" json:"stock"`
	Images      []string           `bson:"images" json:"images"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Cart 购物车结构体
type Cart struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	CartID      string             `bson:"cart_id" json:"cart_id"`
	UserOpenID  string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	Items       []CartItem         `bson:"items" json:"items"`
	TotalAmount float64            `bson:"total_amount" json:"total_amount"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CartItem 购物车项结构体
type CartItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Name      string  `bson:"name" json:"name"`
	Price     float64 `bson:"price" json:"price"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Subtotal  float64 `bson:"subtotal" json:"subtotal"`
}

// Order 订单结构体
type Order struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserOpenID     string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	Items          []OrderItem        `bson:"items" json:"items"`
	SubtotalAmount float64            `bson:"subtotal_amount" json:"subtotal_amount"`
	DiscountAmount float64            `bson:"discount_amount" json:"discount_amount"`
	DiscountRate   float64            `bson:"discount_rate" json:"discount_rate"`
	TotalAmount    float64            `bson:"total_amount" json:"total_amount"`
	Status         string             `bson:"status" json:"status"`
	AddressID      string             `bson:"address_id" json:"address_id"`
	PaymentMethod  string             `bson:"payment_method" json:"payment_method"`
	ReferralCode   string             `bson:"referral_code" json:"referral_code"`
	ReferrerOpenID string             `bson:"referrer_openid,omitempty" json:"referrer_openid,omitempty"` // 推荐人OpenID
	TransactionID  string             `bson:"transaction_id,omitempty" json:"transaction_id,omitempty"`   // 微信支付交易ID
	PaidAt         time.Time          `bson:"paid_at,omitempty" json:"paid_at,omitempty"`                 // 支付时间
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// OrderItem 订单项结构体
type OrderItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Price     float64 `bson:"price" json:"price"`
}

// Commission 佣金记录结构体
type Commission struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	CommissionID string             `bson:"commission_id" json:"commission_id"`
	UserOpenID   string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	Amount       float64            `bson:"amount" json:"amount"`
	Date         time.Time          `bson:"date" json:"date"`
	Status       string             `bson:"status" json:"status"` // pending, paid, cancelled
	Type         string             `bson:"type" json:"type"`     // referral, agent
	Description  string             `bson:"description" json:"description"`
	OrderID      string             `bson:"order_id,omitempty" json:"order_id,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// Referral 推荐关系结构体
type Referral struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ReferralCode string             `bson:"referral_code" json:"referral_code"`
	UserOpenID   string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	UsedBy       []ReferralUsage    `bson:"used_by" json:"used_by"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// ReferralUsage 推荐码使用记录
type ReferralUsage struct {
	UserOpenID string    `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	UserName   string    `bson:"user_name" json:"user_name"`
	UsedAt     time.Time `bson:"used_at" json:"used_at"`
	OrderID    string    `bson:"order_id,omitempty" json:"order_id,omitempty"`
	Commission float64   `bson:"commission" json:"commission"`
	Status     string    `bson:"status" json:"status"` // pending, completed, cancelled
}

// WithdrawRecord 提现记录结构体
type WithdrawRecord struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	WithdrawID       string             `bson:"withdraw_id" json:"withdraw_id"`
	UserOpenID       string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	Amount           float64            `bson:"amount" json:"amount"`
	WithdrawMethod   string             `bson:"withdraw_method" json:"withdraw_method"`
	AccountInfo      AccountInfo        `bson:"account_info" json:"account_info"`
	Status           string             `bson:"status" json:"status"` // pending, processing, completed, rejected
	ProcessingFee    float64            `bson:"processing_fee" json:"processing_fee"`
	ActualAmount     float64            `bson:"actual_amount" json:"actual_amount"`
	EstimatedArrival time.Time          `bson:"estimated_arrival" json:"estimated_arrival"`
	CompletedAt      time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	RejectionReason  string             `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// AccountInfo 账户信息结构体
type AccountInfo struct {
	AccountName   string `json:"account_name" bson:"account_name"`
	AccountNumber string `json:"account_number" bson:"account_number"`
}

// UnlimitedQRCodeRequest 获取不限制小程序码请求结构体（服务端调用微信API）
type UnlimitedQRCodeRequest struct {
	Scene      string     `json:"scene" binding:"required"`
	Page       string     `json:"page,omitempty"`
	CheckPath  *bool      `json:"check_path,omitempty"`
	EnvVersion string     `json:"env_version,omitempty"`
	Width      int        `json:"width,omitempty"`
	AutoColor  *bool      `json:"auto_color,omitempty"`
	LineColor  *QRCodeRGB `json:"line_color,omitempty"`
	IsHyaline  *bool      `json:"is_hyaline,omitempty"`
}

// QRCodeRGB 线条颜色
type QRCodeRGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	Query string `json:"query" binding:"required"`                          // 搜索关键词
	Type  string `json:"type" binding:"required,oneof=word book order all"` // 搜索类型
	Page  int    `json:"page"`                                              // 页码
	Limit int    `json:"limit"`                                             // 每页数量
}

// SearchResponse 搜索响应结构体
type SearchResponse struct {
	Words  []Word  `json:"words,omitempty"`  // 单词搜索结果
	Books  []Book  `json:"books,omitempty"`  // 课本搜索结果
	Orders []Order `json:"orders,omitempty"` // 订单搜索结果
	Total  int64   `json:"total"`            // 总数量
	Page   int     `json:"page"`             // 当前页码
	Limit  int     `json:"limit"`            // 每页数量
}

// Book 课本结构体
type Book struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"_id,omitempty"`
	BookName        string               `bson:"book_name" json:"book_name"`
	BookVersion     string               `bson:"book_version" json:"book_version"`
	Description     string               `bson:"description,omitempty" json:"description,omitempty"`
	Level           string               `bson:"level,omitempty" json:"level,omitempty"`
	TotalWords      int                  `bson:"total_words,omitempty" json:"total_words,omitempty"`
	Units           []primitive.ObjectID `bson:"units,omitempty" json:"units,omitempty"`
	CoverImage      string               `bson:"cover_image,omitempty" json:"cover_image,omitempty"`
	Author          string               `bson:"author,omitempty" json:"author,omitempty"`
	Publisher       string               `bson:"publisher,omitempty" json:"publisher,omitempty"`
	PublicationDate time.Time            `bson:"publication_date,omitempty" json:"publication_date,omitempty"`
	CreatedAt       time.Time            `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt       time.Time            `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// Unit 单元结构体
type Unit struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UnitName  string             `bson:"unit_name" json:"unit_name"`
	BookID    primitive.ObjectID `bson:"book_id" json:"book_id"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
