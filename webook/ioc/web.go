package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
	myjwt "webook/internal/web/jwt"
	"webook/pkg/ginx/middlewares/ratelimit"
	"webook/pkg/limiter"
)

func InitGin(mdls []gin.HandlerFunc, hdl *web.UserHandler, oauth2WechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisclient redis.Cmdable, handler *myjwt.JwtHandler) []gin.HandlerFunc {

	return []gin.HandlerFunc{
		// 限流
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisclient, time.Minute, 10000)).Build(),

		// 跨域
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000"},          // 允许请求的路由来源，* 为全部来源（一般不写）
			AllowMethods:     []string{"POST", "GET"},                    // 允许请求的方法，不写为全部方法
			AllowHeaders:     []string{"Content-Type", "Authorization"},  //
			AllowCredentials: true,                                       // 是否允许携带 cookie 之类的用户认证信息
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"}, // 响应头里携带的东西
			// 判断来源的函数
			AllowOriginFunc: func(origin string) bool {
				if strings.Contains(origin, "http://localhost") {
					// 开发环境
					return true
				}
				return strings.Contains(origin, "youcompany.com")
			},
			MaxAge: 12 * time.Hour,
		}),

		myjwt.NewJwtServiceBuilder(handler).
			AddIgnorePath("/users/signup"). // 忽略路径
			AddIgnorePath("/users/login").
			AddIgnorePath("/users/login_sms/code/send").
			AddIgnorePath("/users/login_sms").
			AddIgnorePath("/users/refresh_token").
			AddIgnorePath("/oauth2/wechat/authurl").
			AddIgnorePath("/oauth2/wechat/callback").
			Build(),
	}
}

//// session基于redis实现
//// 最大空闲连接、tcp（不太可能用到udp）、连接信息和密码、两个key
//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//	// authentication key, encryption key（身份认证、数据加密）外加授权认证————安全三大概念
//	[]byte("YTsKHvuxjcQ3jGXrSXH27JvnA3XTkJ6T"),
//	[]byte("e5Z7W4YbVcerrtjEA77eT5J6hShjjNTp"))
//if err != nil {
//	panic(err)
//}
//store := memstore.NewStore(
//	[]byte("YTsKHvuxjcQ3jGXrSXH27JvnA3XTkJ6T"),
//	[]byte("e5Z7W4YbVcerrtjEA77eT5J6hShjjNTp"))
//server.Use(sessions.Sessions("mysession", store)) // cookie的name和值
//server.Use(
//	middleware.
//		NewLoginMiddlewareBuilder().
//		IgnorePaths("/users/signup"). // 忽略路径
//		IgnorePaths("/users/login").
//		Build())
