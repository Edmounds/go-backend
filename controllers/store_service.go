package controllers

import (
	"errors"
	"miniprogram/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===== 商店服务层 =====

// ProductService 商品服务
type ProductService struct{}

// GetProductService 获取商品服务实例
func GetProductService() *ProductService {
	return &ProductService{}
}

// GetProductsList 获取商品列表
func (s *ProductService) GetProductsList(page, limit int) ([]models.Product, int, error) {
	collection := GetCollection("products")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 计算跳过的数量
	skip := (page - 1) * limit

	// 查询总数
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// 查询商品列表
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	return products, int(total), nil
}

// GetProductByID 根据ID获取商品详情
func (s *ProductService) GetProductByID(productID string) (*models.Product, error) {
	collection := GetCollection("products")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var product models.Product
	err := collection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// CartService 购物车服务
type CartService struct{}

// GetCartService 获取购物车服务实例
func GetCartService() *CartService {
	return &CartService{}
}

// GetUserCart 获取用户购物车
func (s *CartService) GetUserCart(userID string) (*models.Cart, error) {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var cart models.Cart
	err := collection.FindOne(ctx, bson.M{"user_openid": userID}).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新购物车
			return s.createNewCart(userID)
		}
		return nil, err
	}

	return &cart, nil
}

// createNewCart 创建新购物车
func (s *CartService) createNewCart(userID string) (*models.Cart, error) {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	cart := models.Cart{
		CartID:      generateCartID(),
		UserOpenID:  userID,
		Items:       []models.CartItem{},
		TotalAmount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := collection.InsertOne(ctx, cart)
	if err != nil {
		return nil, err
	}

	cart.ID = result.InsertedID.(primitive.ObjectID)
	return &cart, nil
}

// AddItemToCart 添加商品到购物车
func (s *CartService) AddItemToCart(userID, productID string, quantity int) (*models.Cart, error) {
	// 获取商品信息
	productService := GetProductService()
	product, err := productService.GetProductByID(productID)
	if err != nil {
		return nil, err
	}

	// 检查库存
	if product.Stock < quantity {
		return nil, errors.New("商品库存不足")
	}

	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// 检查商品是否已在购物车中
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = cart.Items[i].Price * float64(cart.Items[i].Quantity)
			found = true
			break
		}
	}

	// 如果商品不在购物车中，添加新项
	if !found {
		cartItem := models.CartItem{
			ProductID: productID,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  quantity,
			Subtotal:  product.Price * float64(quantity),
		}
		cart.Items = append(cart.Items, cartItem)
	}

	// 重新计算总金额
	s.recalculateCartTotal(cart)

	// 更新数据库
	return s.updateCart(cart)
}

// UpdateCartItem 更新购物车商品数量
func (s *CartService) UpdateCartItem(userID, productID string, quantity int) (*models.Cart, error) {
	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// 查找并更新商品
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			if quantity <= 0 {
				// 数量为0或负数，删除商品
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			} else {
				cart.Items[i].Quantity = quantity
				cart.Items[i].Subtotal = cart.Items[i].Price * float64(quantity)
			}
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("商品不在购物车中")
	}

	// 重新计算总金额
	s.recalculateCartTotal(cart)

	// 更新数据库
	return s.updateCart(cart)
}

// DeleteCartItem 删除购物车商品
func (s *CartService) DeleteCartItem(userID, productID string) (*models.Cart, error) {
	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// 查找并删除商品
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("商品不在购物车中")
	}

	// 重新计算总金额
	s.recalculateCartTotal(cart)

	// 更新数据库
	return s.updateCart(cart)
}

// recalculateCartTotal 重新计算购物车总金额
func (s *CartService) recalculateCartTotal(cart *models.Cart) {
	total := 0.0
	for _, item := range cart.Items {
		total += item.Subtotal
	}
	cart.TotalAmount = total
	cart.UpdatedAt = time.Now()
}

// updateCart 更新购物车到数据库
func (s *CartService) updateCart(cart *models.Cart) (*models.Cart, error) {
	collection := GetCollection("carts")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"user_openid": cart.UserOpenID}
	update := bson.M{
		"$set": bson.M{
			"items":        cart.Items,
			"total_amount": cart.TotalAmount,
			"updated_at":   cart.UpdatedAt,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// OrderService 订单服务
type OrderService struct{}

// GetOrderService 获取订单服务实例
func GetOrderService() *OrderService {
	return &OrderService{}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID string, req models.CreateOrderRequest) (*models.Order, error) {
	// 获取购物车
	cartService := GetCartService()
	cart, err := cartService.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("购物车为空")
	}

	// 验证地址
	userService := GetUserService()
	user, err := userService.FindUserByOpenID(userID)
	if err != nil {
		return nil, err
	}

	var selectedAddress *models.Address
	for _, addr := range user.Addresses {
		if addr.ID.Hex() == req.AddressID {
			selectedAddress = &addr
			break
		}
	}

	if selectedAddress == nil {
		return nil, errors.New("收货地址不存在")
	}

	// 转换购物车商品为订单商品
	var orderItems []models.OrderItem
	for _, cartItem := range cart.Items {
		orderItems = append(orderItems, models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
		})
	}

	// 计算折扣
	discountRate := 0.0
	subtotalAmount := cart.TotalAmount
	var referrerOpenID string

	if req.ReferralCode != "" {
		authService := GetAuthService()
		valid, err := authService.ValidateReferralCode(req.ReferralCode)
		if err == nil && valid {
			referralService := NewReferralCodeService()
			referrer, err := referralService.GetUserByReferralCode(req.ReferralCode)
			if err == nil {
				discountRate = referralService.CalculateDiscountRate(referrer.AgentLevel)
				referrerOpenID = referrer.OpenID
			}
		}
	}

	discountAmount := subtotalAmount * discountRate
	totalAmount := subtotalAmount - discountAmount

	// 创建订单
	order := models.Order{
		UserOpenID:     userID,
		Items:          orderItems,
		SubtotalAmount: subtotalAmount,
		DiscountAmount: discountAmount,
		DiscountRate:   discountRate,
		TotalAmount:    totalAmount,
		Status:         "pending",
		AddressID:      req.AddressID,
		PaymentMethod:  req.PaymentMethod,
		ReferralCode:   req.ReferralCode,
		ReferrerOpenID: referrerOpenID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存订单到数据库
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	result, err := collection.InsertOne(ctx, order)
	if err != nil {
		return nil, err
	}

	order.ID = result.InsertedID.(primitive.ObjectID)

	// 清空购物车
	cart.Items = []models.CartItem{}
	cart.TotalAmount = 0
	cartService.updateCart(cart)

	return &order, nil
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(userID string, page, limit int) ([]models.Order, int, error) {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 构建查询条件
	filter := bson.M{"user_openid": userID}

	// 计算跳过的数量
	skip := (page - 1) * limit

	// 查询总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询订单列表
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, int(total), nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID primitive.ObjectID, status string) error {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateOrderPayment 更新订单支付信息
func (s *OrderService) UpdateOrderPayment(orderID primitive.ObjectID, transactionID string) error {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"status":         "paid",
			"transaction_id": transactionID,
			"paid_at":        time.Now(),
			"updated_at":     time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// generateCartID 生成购物车ID
func generateCartID() string {
	return "CART" + time.Now().Format("20060102150405") + GenerateRandomString(4)
}

// ===== 向后兼容函数 =====

// GetProductsList 获取商品列表 (向后兼容)
func GetProductsList(page, limit int) ([]models.Product, int, error) {
	service := GetProductService()
	return service.GetProductsList(page, limit)
}

// GetProductByID 根据ID获取商品详情 (向后兼容)
func GetProductByID(productID string) (*models.Product, error) {
	service := GetProductService()
	return service.GetProductByID(productID)
}

// GetUserCart 获取用户购物车 (向后兼容)
func GetUserCart(userID string) (*models.Cart, error) {
	service := GetCartService()
	return service.GetUserCart(userID)
}

// AddItemToCart 添加商品到购物车 (向后兼容)
func AddItemToCart(userID, productID string, quantity int) (*models.Cart, error) {
	service := GetCartService()
	return service.AddItemToCart(userID, productID, quantity)
}

// UpdateCartItem 更新购物车商品数量 (向后兼容)
func UpdateCartItem(userID, productID string, quantity int) (*models.Cart, error) {
	service := GetCartService()
	return service.UpdateCartItem(userID, productID, quantity)
}

// DeleteCartItem 删除购物车商品 (向后兼容)
func DeleteCartItem(userID, productID string) (*models.Cart, error) {
	service := GetCartService()
	return service.DeleteCartItem(userID, productID)
}

// CreateOrder 创建订单 (向后兼容)
func CreateOrder(userID string, req models.CreateOrderRequest) (*models.Order, error) {
	service := GetOrderService()
	return service.CreateOrder(userID, req)
}

// GetUserOrders 获取用户订单列表 (向后兼容)
func GetUserOrders(userID string, page, limit int) ([]models.Order, int, error) {
	service := GetOrderService()
	return service.GetUserOrders(userID, page, limit)
}
