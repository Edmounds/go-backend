package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReferralError 推荐码相关错误
type ReferralError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ReferralError) Error() string {
	return e.Message
}

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
	Avatar         string             `bson:"avatar" json:"avatar"`                   // 用户头像路径
	QRCode         string             `bson:"qr_code" json:"qr_code"`                 // 用户推荐二维码base64数据
	CollectedCards []CollectedCard    `bson:"collected_cards" json:"collected_cards"` // 收藏的单词卡列表
	UnlockedBooks  []BookPermission   `bson:"unlocked_books" json:"unlocked_books"`   // 已解锁的书籍权限
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

// BookPermission 书籍权限结构体
type BookPermission struct {
	BookID     primitive.ObjectID `bson:"book_id" json:"book_id"`         // 书籍ID
	BookName   string             `bson:"book_name" json:"book_name"`     // 书籍名称（便于查询）
	AccessType string             `bson:"access_type" json:"access_type"` // "digital" 电子版权限, "physical" 实体版权限（包含电子版）
	OrderID    primitive.ObjectID `bson:"order_id" json:"order_id"`       // 购买订单ID
	UnlockedAt time.Time          `bson:"unlocked_at" json:"unlocked_at"` // 解锁时间
}

// CollectedCard 收藏的单词卡结构体
type CollectedCard struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	WordID      primitive.ObjectID `bson:"word_id" json:"word_id"`           // 单词ID
	WordName    string             `bson:"word_name" json:"word_name"`       // 单词名称（便于查询）
	CollectedAt time.Time          `bson:"collected_at" json:"collected_at"` // 收藏时间
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
	ProductType string             `bson:"product_type" json:"product_type"` // "physical" 实体卡, "digital" 电子卡
	BookID      primitive.ObjectID `bson:"book_id" json:"book_id"`           // 关联的书籍ID
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
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	CommissionID       string             `bson:"commission_id" json:"commission_id"`
	UserOpenID         string             `bson:"user_openid" json:"user_openid"` // 推荐人/代理人OpenID
	Amount             float64            `bson:"amount" json:"amount"`
	Date               time.Time          `bson:"date" json:"date"`
	Status             string             `bson:"status" json:"status"` // pending, paid, cancelled
	Type               string             `bson:"type" json:"type"`     // referral, agent
	Description        string             `bson:"description" json:"description"`
	OrderID            string             `bson:"order_id,omitempty" json:"order_id,omitempty"`
	ReferredUserOpenID string             `bson:"referred_user_openid,omitempty" json:"referred_user_openid,omitempty"` // 被推荐用户OpenID
	ReferredUserName   string             `bson:"referred_user_name,omitempty" json:"referred_user_name,omitempty"`     // 被推荐用户名
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
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

// ReferralUsage 推荐码使用记录（仅记录注册时的使用关系）
type ReferralUsage struct {
	UserOpenID string    `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	UserName   string    `bson:"user_name" json:"user_name"`
	UsedAt     time.Time `bson:"used_at" json:"used_at"`
}

// WithdrawRecord 提现记录结构体
type WithdrawRecord struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	WithdrawID       string             `bson:"withdraw_id" json:"withdraw_id"`
	UserOpenID       string             `bson:"user_openid" json:"user_openid"` // 使用OpenID而不是MongoDB的_id
	Amount           float64            `bson:"amount" json:"amount"`
	WithdrawMethod   string             `bson:"withdraw_method" json:"withdraw_method"`
	AccountInfo      AccountInfo        `bson:"account_info" json:"account_info"`
	Status           string             `bson:"status" json:"status"` // pending, processing, completed, rejected, failed
	ProcessingFee    float64            `bson:"processing_fee" json:"processing_fee"`
	ActualAmount     float64            `bson:"actual_amount" json:"actual_amount"`
	EstimatedArrival time.Time          `bson:"estimated_arrival" json:"estimated_arrival"`
	CompletedAt      time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	RejectionReason  string             `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	FailureReason    string             `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`   // 微信转账失败原因
	WechatBatchID    string             `bson:"wechat_batch_id,omitempty" json:"wechat_batch_id,omitempty"` // 微信转账批次ID
	OutBatchNo       string             `bson:"out_batch_no,omitempty" json:"out_batch_no,omitempty"`       // 商户批次号
	OutDetailNo      string             `bson:"out_detail_no,omitempty" json:"out_detail_no,omitempty"`     // 商户明细号
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

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	OpenID       string `json:"openID" binding:"required"`
	UserName     string `json:"user_name,omitempty"`
	UserPassword string `json:"user_password,omitempty"`
	Class        string `json:"class,omitempty"`
	Age          int    `json:"age,omitempty"`
	School       string `json:"school,omitempty"`
	Phone        string `json:"phone,omitempty"`
	City         string `json:"city,omitempty"`
	ReferredBy   string `json:"referred_by,omitempty"`
}

// AddressRequest 地址请求
type AddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Province      string `json:"province" binding:"required"`
	City          string `json:"city" binding:"required"`
	District      string `json:"district"`
	Street        string `json:"street" binding:"required"`
	PostalCode    string `json:"postal_code"`
	IsDefault     bool   `json:"is_default"`
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

// ===== 请求/响应结构体定义 =====

// 商店相关请求结构体
// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// UpdateCartRequest 更新购物车请求
type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AddressID     string `json:"address_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	ReferralCode  string `json:"referral_code"`
}

// 微信支付相关请求结构体
// CreateWechatPayOrderRequest 创建微信支付订单请求
type CreateWechatPayOrderRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

// WechatPayOrder 微信支付订单结构体
type WechatPayOrder struct {
	PrepayId  string `json:"prepayId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// 认证相关请求/响应结构体
// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	Code         string `json:"code" binding:"required"`
	ReferralCode string `json:"referral_code,omitempty"` // 可选的推荐码参数，用于扫码进入的场景
}

// WechatLoginResponse 微信登录响应
type WechatLoginResponse struct {
	SessionKey string `json:"session_key"`
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string      `json:"token"`
		User  interface{} `json:"user"`
	} `json:"data"`
}

// WechatAccessTokenResponse 微信访问令牌响应（内部使用）
type WechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// DevLoginRequest 开发登录请求
type DevLoginRequest struct {
	OpenID       string `json:"openID" binding:"required"`
	ReferralCode string `json:"referral_code,omitempty"`
}

// 推荐相关请求结构体
// TrackReferralRequest 跟踪推荐关系请求
type TrackReferralRequest struct {
	ReferralCode   string `json:"referral_code" binding:"required"`
	ReferredUserID string `json:"referred_user_id" binding:"required"`
}

// ValidateReferralRequest 验证推荐码请求
type ValidateReferralRequest struct {
	ReferralCode string `json:"referral_code" binding:"required"`
}

// 进度相关请求结构体
// UpdateProgressRequest 更新进度请求
type UpdateProgressRequest struct {
	CurrentUnit  string   `json:"current_unit" binding:"required"`
	LearnedWords []string `json:"learned_words"`
}

// 代理相关请求结构体
// WithdrawRequest 提取佣金请求
type WithdrawRequest struct {
	Amount         float64     `json:"amount" binding:"required,min=0.01"`
	WithdrawMethod string      `json:"withdraw_method" binding:"required"`
	AccountInfo    AccountInfo `json:"account_info"`
}

// 管理员相关请求结构体
// UpdateAgentLevelRequest 更新代理等级请求
type UpdateAgentLevelRequest struct {
	AgentLevel int `json:"agent_level" binding:"required"`
}
