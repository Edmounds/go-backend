package routes

import (
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
			public.POST("/users", controllers.CreateUserHandler())

			// 商城相关公开路由
			public.GET("/products", controllers.GetProductsHandler())
			public.GET("/product/:product_id", controllers.GetProductHandler())

			// 学习相关公开路由
			public.GET("/books", controllers.GetBooksHandler())

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
			protected.PUT("/users/:user_id", controllers.UpdateUserHandler())
			protected.PUT("/users/:user_id/agent", controllers.UpdateAgentHandler())
			protected.POST("/users/:user_id/address", controllers.CreateAddressHandler())
			protected.DELETE("/:user_id/address/:address_id", controllers.DeleteAddressHandler())

			// 商城相关路由
			protected.POST("/users/:user_id/cart", controllers.AddToCartHandler())
			protected.PUT("/users/:user_id/cart/items/:product_id", controllers.UpdateCartItemHandler())
			protected.DELETE("/users/:user_id/cart/items/:product_id", controllers.DeleteCartItemHandler())
			protected.POST("/users/:user_id/orders", controllers.CreateOrderHandler())
			protected.GET("/users/:user_id/orders", controllers.GetOrdersHandler())

			// 学习进度相关路由
			protected.GET("/users/:user_id/progress", controllers.GetProgressHandler())
			protected.PUT("/users/:user_id/progress", controllers.UpdateProgressHandler())
			protected.GET("/books/:book_id/words", controllers.GetBookWordsHandler())

			// 推荐系统相关路由
			protected.GET("/users/:user_id/referral", controllers.GetReferralInfoHandler())
			protected.POST("/referrals", controllers.TrackReferralHandler())
			protected.GET("/users/:user_id/referral/commissions", controllers.GetCommissionsHandler())

			// 代理系统相关路由
			protected.GET("/agents/:user_id/users", controllers.GetAgentUsersHandler())
			protected.GET("/agents/:user_id/sales", controllers.GetAgentSalesHandler())
			protected.POST("/agents/:user_id/withdraw", controllers.WithdrawCommissionHandler())
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "服务运行正常",
		})
	})
}
