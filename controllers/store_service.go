package controllers

import (
	"errors"
	"fmt"
	"log"
	"miniprogram/models"
	"miniprogram/utils"
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
		CreatedAt:   utils.GetCurrentUTCTime(),
		UpdatedAt:   utils.GetCurrentUTCTime(),
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
			Selected:  true, // 新添加的商品默认选中
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
	cart.UpdatedAt = utils.GetCurrentUTCTime()
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

// SelectCartItem 选择/取消选择购物车商品
func (s *CartService) SelectCartItem(userID, productID string, selected bool) (*models.Cart, error) {
	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// 查找并更新商品选择状态
	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Selected = selected
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("商品不在购物车中")
	}

	// 重新计算总金额（只计算选中的商品）
	s.recalculateSelectedCartTotal(cart)

	// 更新数据库
	return s.updateCart(cart)
}

// SelectAllCartItems 全选/反选购物车商品
func (s *CartService) SelectAllCartItems(userID string, selected bool) (*models.Cart, error) {
	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// 更新所有商品的选择状态
	for i := range cart.Items {
		cart.Items[i].Selected = selected
	}

	// 重新计算总金额（只计算选中的商品）
	s.recalculateSelectedCartTotal(cart)

	// 更新数据库
	return s.updateCart(cart)
}

// GetSelectedCartItems 获取选中的购物车商品
func (s *CartService) GetSelectedCartItems(userID string) ([]models.CartItem, error) {
	// 获取购物车
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	var selectedItems []models.CartItem
	for _, item := range cart.Items {
		if item.Selected {
			selectedItems = append(selectedItems, item)
		}
	}

	return selectedItems, nil
}

// recalculateSelectedCartTotal 重新计算选中商品的总金额
func (s *CartService) recalculateSelectedCartTotal(cart *models.Cart) {
	total := 0.0
	for _, item := range cart.Items {
		if item.Selected {
			total += item.Subtotal
		}
	}
	cart.TotalAmount = total
	cart.UpdatedAt = utils.GetCurrentUTCTime()
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

	// 获取选中的商品
	var selectedItems []models.CartItem
	var selectedCartItemIDs []string
	for _, item := range cart.Items {
		if item.Selected {
			selectedItems = append(selectedItems, item)
			selectedCartItemIDs = append(selectedCartItemIDs, item.ProductID)
		}
	}

	if len(selectedItems) == 0 {
		return nil, errors.New("请选择要购买的商品")
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

	// 转换选中的购物车商品为订单商品
	var orderItems []models.OrderItem
	subtotalAmount := 0.0
	for _, cartItem := range selectedItems {
		orderItems = append(orderItems, models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
		})
		subtotalAmount += cartItem.Subtotal
	}

	// 计算折扣
	discountRate := 0.0
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
	totalAmount := utils.FormatMoneyForWechatPay(subtotalAmount - discountAmount)

	// 创建订单
	order := models.Order{
		UserOpenID:        userID,
		Items:             orderItems,
		SelectedCartItems: selectedCartItemIDs, // 记录选中的商品ID，用于支付回调清空
		SubtotalAmount:    subtotalAmount,
		DiscountAmount:    discountAmount,
		DiscountRate:      discountRate,
		TotalAmount:       totalAmount,
		Status:            "pending",
		OrderSource:       "cart", // 标记为购物车订单
		AddressID:         req.AddressID,
		PaymentMethod:     req.PaymentMethod,
		ReferralCode:      req.ReferralCode,
		ReferrerOpenID:    referrerOpenID,
		CreatedAt:         utils.GetCurrentUTCTime(),
		UpdatedAt:         utils.GetCurrentUTCTime(),
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

	// 注意：不在此处清空购物车，而是在支付成功回调时统一处理
	// 这样避免了订单创建成功但支付失败时购物车被误清空的问题

	return &order, nil
}

// CreateDirectOrder 直接购买创建订单
func (s *OrderService) CreateDirectOrder(userID string, req models.DirectPurchaseRequest) (*models.Order, error) {
	// 获取商品信息
	productService := GetProductService()
	product, err := productService.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	// 检查库存
	if product.Stock < req.Quantity {
		return nil, errors.New("库存不足")
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

	// 创建订单项
	orderItems := []models.OrderItem{
		{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     product.Price,
		},
	}

	// 计算小计金额
	subtotalAmount := product.Price * float64(req.Quantity)

	// 计算折扣 - 从用户的referred_by字段读取推荐码
	discountRate := 0.0
	var referrerOpenID string
	var referralCode string

	if user.ReferredBy != "" {
		// 根据推荐码获取推荐人信息
		referralService := NewReferralCodeService()
		referrer, err := referralService.GetUserByReferralCode(user.ReferredBy)
		if err == nil {
			discountRate = referralService.CalculateDiscountRate(referrer.AgentLevel)
			referrerOpenID = referrer.OpenID
			referralCode = user.ReferredBy
		}
	}

	discountAmount := subtotalAmount * discountRate
	totalAmount := utils.FormatMoneyForWechatPay(subtotalAmount - discountAmount)

	// 创建订单
	order := models.Order{
		UserOpenID:     userID,
		Items:          orderItems,
		SubtotalAmount: subtotalAmount,
		DiscountAmount: discountAmount,
		DiscountRate:   discountRate,
		TotalAmount:    totalAmount,
		Status:         "pending",
		OrderSource:    "direct", // 标记为直接购买订单
		AddressID:      req.AddressID,
		PaymentMethod:  req.PaymentMethod,
		ReferralCode:   referralCode,
		ReferrerOpenID: referrerOpenID,
		CreatedAt:      utils.GetCurrentUTCTime(),
		UpdatedAt:      utils.GetCurrentUTCTime(),
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
			"updated_at": utils.GetCurrentUTCTime(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateOrderPayment 更新订单支付信息
func (s *OrderService) UpdateOrderPayment(orderID primitive.ObjectID, transactionID string) error {
	log.Printf("[订单状态更新] 开始更新订单支付状态 - 订单ID: %s, 交易ID: %s", orderID.Hex(), transactionID)

	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	filter := bson.M{"_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"status":         "paid",
			"transaction_id": transactionID,
			"paid_at":        utils.GetCurrentUTCTime(),
			"updated_at":     utils.GetCurrentUTCTime(), // 修复时间函数不一致问题
		},
	}

	log.Printf("[订单状态更新] 执行数据库更新操作...")
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("[订单状态更新] 数据库更新失败: %v", err)
		return err
	}

	log.Printf("[订单状态更新] 数据库更新成功 - 匹配数量: %d, 修改数量: %d", result.MatchedCount, result.ModifiedCount)

	if result.MatchedCount == 0 {
		log.Printf("[订单状态更新] 警告: 未找到匹配的订单 ID: %s", orderID.Hex())
		return fmt.Errorf("未找到订单 ID: %s", orderID.Hex())
	}

	if result.ModifiedCount == 0 {
		log.Printf("[订单状态更新] 警告: 订单状态未发生变化，可能已经是paid状态")
	}

	log.Printf("[订单状态更新] 订单 %s 状态更新完成", orderID.Hex())
	return nil
}

// ProcessOrderUnlockBooks 处理订单完成后的书籍权限解锁
func (s *OrderService) ProcessOrderUnlockBooks(orderID primitive.ObjectID) error {
	collection := GetCollection("orders")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 获取订单信息
	var order models.Order
	err := collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return err
	}

	// 只处理已支付的订单
	if order.Status != "paid" {
		return errors.New("订单状态不是已支付，无法解锁权限")
	}

	// 获取订单中所有商品的书籍权限信息
	bookPermissions := make(map[primitive.ObjectID]string) // bookID -> productType

	productCollection := GetCollection("products")
	for _, item := range order.Items {
		var product models.Product
		err := productCollection.FindOne(ctx, bson.M{"product_id": item.ProductID}).Decode(&product)
		if err != nil {
			continue // 跳过无法找到的商品
		}

		// 如果是实体卡，提供完整权限（包含电子版）
		// 如果是电子卡，只提供电子版权限
		if product.ProductType == "physical" {
			bookPermissions[product.BookID] = "physical"
		} else if product.ProductType == "digital" {
			// 如果已经有实体权限，保持实体权限
			if existingType, exists := bookPermissions[product.BookID]; !exists || existingType != "physical" {
				bookPermissions[product.BookID] = "digital"
			}
		}
	}

	// 解锁用户的书籍权限
	for bookID, accessType := range bookPermissions {
		err := s.unlockBookForUser(order.UserOpenID, bookID, accessType, orderID)
		if err != nil {
			return err
		}
	}

	return nil
}

// unlockBookForUser 为用户解锁指定书籍的权限
func (s *OrderService) unlockBookForUser(userOpenID string, bookID primitive.ObjectID, accessType string, orderID primitive.ObjectID) error {
	userCollection := GetCollection("users")
	bookCollection := GetCollection("books")
	ctx, cancel := CreateDBContext()
	defer cancel()

	// 获取书籍信息
	var book models.Book
	err := bookCollection.FindOne(ctx, bson.M{"_id": bookID}).Decode(&book)
	if err != nil {
		return err
	}

	// 检查用户是否已经有此书的权限
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"openID": userOpenID}).Decode(&user)
	if err != nil {
		return err
	}

	// 查找是否已存在该书籍的权限
	hasPermission := false
	needUpdate := false
	for i, permission := range user.UnlockedBooks {
		if permission.BookID == bookID {
			hasPermission = true
			// 如果当前是电子权限，但新购买的是实体权限，则升级
			if permission.AccessType == "digital" && accessType == "physical" {
				user.UnlockedBooks[i].AccessType = "physical"
				user.UnlockedBooks[i].OrderID = orderID
				user.UnlockedBooks[i].UnlockedAt = utils.GetCurrentUTCTime()
				needUpdate = true
			}
			break
		}
	}

	// 如果没有权限，添加新权限
	if !hasPermission {
		newPermission := models.BookPermission{
			BookID:     bookID,
			BookName:   book.BookName,
			AccessType: accessType,
			OrderID:    orderID,
			UnlockedAt: utils.GetCurrentUTCTime(),
		}
		user.UnlockedBooks = append(user.UnlockedBooks, newPermission)
		needUpdate = true
	}

	// 更新用户权限
	if needUpdate {
		filter := bson.M{"openID": userOpenID}
		update := bson.M{
			"$set": bson.M{
				"unlocked_books": user.UnlockedBooks,
				"updated_at":     time.Now(),
			},
		}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		return err
	}

	return nil
}

// CheckUserBookPermission 检查用户是否有访问指定书籍的权限
func CheckUserBookPermission(userOpenID string, bookID primitive.ObjectID) (bool, string, error) {
	collection := GetCollection("users")
	ctx, cancel := CreateDBContext()
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"openID": userOpenID}).Decode(&user)
	if err != nil {
		return false, "", err
	}

	// 检查用户是否有该书籍的权限
	for _, permission := range user.UnlockedBooks {
		if permission.BookID == bookID {
			return true, permission.AccessType, nil
		}
	}

	return false, "", nil
}

// generateCartID 生成购物车ID
func generateCartID() string {
	return "CART" + utils.GetCurrentUTCTime().Format("20060102150405") + GenerateRandomString(4)
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

// SelectCartItem 选择/取消选择购物车商品 (向后兼容)
func SelectCartItem(userID, productID string, selected bool) (*models.Cart, error) {
	service := GetCartService()
	return service.SelectCartItem(userID, productID, selected)
}

// SelectAllCartItems 全选/反选购物车商品 (向后兼容)
func SelectAllCartItems(userID string, selected bool) (*models.Cart, error) {
	service := GetCartService()
	return service.SelectAllCartItems(userID, selected)
}

// GetSelectedCartItems 获取选中的购物车商品 (向后兼容)
func GetSelectedCartItems(userID string) ([]models.CartItem, error) {
	service := GetCartService()
	return service.GetSelectedCartItems(userID)
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

// CreateDirectOrder 直接购买创建订单 (向后兼容)
func CreateDirectOrder(userID string, req models.DirectPurchaseRequest) (*models.Order, error) {
	service := GetOrderService()
	return service.CreateDirectOrder(userID, req)
}
