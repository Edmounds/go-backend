package routes

import (
	"miniprogram/config"
	"miniprogram/controllers"
	"miniprogram/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有API路由
func SetupRoutes(r *gin.Engine) {

	// API v1 组
	v1 := r.Group("/api")
	{
		// 公开路由（不需要认证）
		public := v1.Group("/")
		{
			// 用户认证相关路由
			public.POST("/auth", controllers.WechatAuthHandler())

			// 开发环境专用登录接口（可配置启用/禁用）
			cfg := config.GetConfig()
			if cfg.EnableDevLogin {
				public.POST("/dev-login", controllers.DevLoginHandler())
			}
			public.POST("/users/profile", controllers.UpdateUserProfileHandler())

			// 商城相关公开路由
			public.GET("/products", controllers.GetProductsHandler())
			public.GET("/product/:product_id", controllers.GetProductHandler())

			// 学习相关公开路由
			public.GET("/books", controllers.GetBooksHandler())

			// 搜索相关公开路由
			public.POST("/search", controllers.SearchHandler())
			public.GET("/search/words", controllers.SearchWordsHandler())
			public.GET("/search/books", controllers.SearchBooksHandler())

			// 推荐相关公开路由
			public.POST("/referrals/validate", controllers.ValidateReferralCodeHandler())
		}

		// 受保护的路由（需要JWT认证）
		protected := v1.Group("/")
		protected.Use(middlewares.JWTAuthMiddleware())
		{
			// Token相关路由
			protected.POST("/auth/refresh", middlewares.RefreshTokenMiddleware())

			// 用户管理路由
			protected.GET("/users/:user_id", controllers.GetUserHandler())
			protected.POST("/users/:user_id/avatar", controllers.UploadAvatarHandler())
			protected.GET("/users/:user_id/avatar", controllers.GetAvatarHandler())
			protected.GET("/users/:user_id/qrcode", controllers.GetUserQRCodeHandler())

			// 地址管理路由
			protected.POST("/users/:user_id/address", controllers.CreateAddressHandler())
			protected.GET("/users/:user_id/addresses", controllers.GetUserAddressesHandler())
			protected.PUT("/users/:user_id/address/:address_id", controllers.UpdateAddressHandler())
			protected.DELETE("/users/:user_id/address/:address_id", controllers.DeleteAddressHandler())

			// 收藏功能路由
			protected.GET("/users/:user_id/collected-cards", controllers.GetCollectedCardsHandler())
			protected.POST("/users/:user_id/collected-cards/:word_id", controllers.AddToCollectedCardsHandler())
			protected.DELETE("/users/:user_id/collected-cards/:word_id", controllers.RemoveFromCollectedCardsHandler())
			protected.GET("/users/:user_id/collected-cards/:word_id/status", controllers.CheckCardCollectedHandler())

			// 管理员路由
			protected.PUT("/admin/users/:user_id/agent-level", controllers.UpdateAgentLevelHandler())

			// 学习进度相关路由
			protected.GET("/users/:user_id/progress", controllers.GetProgressHandler())
			protected.PUT("/users/:user_id/progress", controllers.UpdateProgressHandler())
			protected.GET("/books/:book_id/words", controllers.GetBookWordsHandler())

			// 单词卡片相关路由
			protected.GET("/units/:unit_id/words", controllers.GetUnitWordsHandler())
			protected.GET("/words/:word_id/card", controllers.GetWordCardHandler())
			protected.GET("/words", controllers.GetWordsByUnitNameHandler()) // 通过查询参数获取单词

			// 推荐系统相关路由
			protected.GET("/users/:user_id/referral", controllers.GetReferralInfoHandler())
			protected.POST("/referrals", controllers.TrackReferralHandler())
			protected.GET("/users/:user_id/referral/commissions", controllers.GetCommissionsHandler())
			// 微信小程序码相关路由（服务端代理获取不限制小程序码）
			protected.POST("/wxacode/unlimited", controllers.GenerateUnlimitedQRCodeHandler())

			// 代理系统相关路由
			protected.GET("/agents/:user_id/users", controllers.GetAgentUsersHandler())
			protected.GET("/agents/:user_id/sales", controllers.GetAgentSalesHandler())
			protected.GET("/agents/:user_id/commission/dashboard", controllers.GetAgentCommissionDashboardHandler())
			protected.GET("/agents/:user_id/commission/details", controllers.GetAgentCommissionDetailsHandler())
			protected.POST("/agents/:user_id/withdraw", controllers.WithdrawCommissionHandler())

			// 商城相关路由
			// 购物车路由
			protected.GET("/users/:user_id/cart", controllers.GetCartHandler())
			protected.POST("/users/:user_id/cart", controllers.AddToCartHandler())
			protected.PUT("/users/:user_id/cart/items/:product_id", controllers.UpdateCartItemHandler())
			protected.DELETE("/users/:user_id/cart/items/:product_id", controllers.DeleteCartItemHandler())

			// 购物车选择功能路由
			protected.PUT("/users/:user_id/cart/items/:product_id/select", controllers.SelectCartItemHandler())
			protected.PUT("/users/:user_id/cart/select-all", controllers.SelectAllCartItemsHandler())
			protected.GET("/users/:user_id/cart/selected", controllers.GetSelectedCartItemsHandler())

			// 订单路由
			protected.POST("/users/:user_id/orders", controllers.CreateOrderHandler())
			protected.POST("/users/:user_id/direct-purchase", controllers.DirectPurchaseHandler())
			protected.GET("/users/:user_id/orders", controllers.GetOrdersHandler())
			protected.GET("/users/:user_id/orders/:order_id", controllers.GetOrderHandler())
			// 微信支付相关路由
			protected.POST("/users/:user_id/orders/pay", controllers.CreateWechatPayOrderHandler())

			// 退款相关路由
			protected.POST("/users/:user_id/refunds", controllers.CreateRefundHandler())
			protected.GET("/users/:user_id/refunds", controllers.GetRefundRecordsHandler())
			protected.GET("/users/:user_id/refunds/:refund_id", controllers.GetRefundHandler())

			// 受保护的搜索路由（需要用户身份）
			protected.GET("/search/orders", controllers.SearchOrdersHandler())
		}

		// 微信支付回调路由（不需要JWT认证）
		v1.POST("/wechat/pay/notify", controllers.WechatPayNotifyHandler())

		// 测试路由（仅用于调试）
		v1.POST("/test/update-order-status", controllers.TestUpdateOrderStatusHandler())
	}

	// 健康检查
	r.GET("/health", controllers.HealthCheckHandler())
}
