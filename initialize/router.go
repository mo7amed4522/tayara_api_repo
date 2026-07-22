package initialize

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/docs"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if stat.IsDir() {
		return nil, os.ErrPermission
	}

	return f, nil
}

type ZapWriter struct {
}

func (z *ZapWriter) Write(p []byte) (n int, err error) {
	global.GVA_LOG.Error(string(p))
	return len(p), nil
}

// Initialize the main route
func Routers() *gin.Engine {
	Router := gin.New()
	gin.DefaultErrorWriter = io.MultiWriter(&ZapWriter{})
	Router.Use(gin.Recovery())
	if gin.Mode() == gin.DebugMode {
		Router.Use(gin.Logger())
	}
	// Enable proxy on port 8888 (frontend HTML)
	Router.Use(static.Serve("/qrcode", static.LocalFile("./web/payment/", false)))

	// If you want to avoid using nginx as a proxy for your front-end webpages, you can modify the configuration file under web/.env.production.
	// VUE_APP_BASE_API = /
	// VUE_APP_BASE_PATH = http://localhost
	// Then execute the packaging command. npm run build. (Uncomment the following 3 lines)
	// Router.Static("/favicon.ico", "./dist/favicon.ico")
	// Router.Static("/assets", "./dist/assets")   // Static resources in dist
	// Router.StaticFile("/", "./dist/index.html") // Front-end web entrance page

	Router.StaticFS(global.GVA_CONFIG.Local.StorePath, justFilesFilesystem{http.Dir(global.GVA_CONFIG.Local.StorePath)}) // Router.Use(middleware.LoadTls())  // 如果需要使用https 请打开此中间件 然后前往 core/server.go 将启动模式 更变为 Router.RunTLS("端口","你的cre/pem文件","你的key文件")
	// Cross-domain requests can be resolved by uncommenting the following text.
	// Router.Use(middleware.Cors()) // Allow all cross-origin requests directly
	Router.Use(middleware.CorsByRules()) // Allow cross-origin requests according to the configured rules.
	global.GVA_LOG.Info("use middleware cors")
	docs.SwaggerInfo.BasePath = global.GVA_CONFIG.System.RouterPrefix
	Router.GET(global.GVA_CONFIG.System.RouterPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	global.GVA_LOG.Info("register swagger handler")
	Router.GET("/", func(c *gin.Context) {
		address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)
		c.String(http.StatusOK, "Welcome to use Tayara API Server, server is running on %s\n", address)
	})
	// Convenient for adding routing group prefixes and enabling use across multiple servers.

	PublicGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	PrivateGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	BxPrivateGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)

	PrivateGroup.Use(middleware.JWTAuth())
	BxPrivateGroup.Use(middleware.BxJWTAuth())

	{
		// Health monitoring
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})
	}
	{

	}

	{

	}

	// Plugin routing installation
	InstallPlugin(PrivateGroup, PublicGroup, Router)

	// Register business routes
	initBizRouter(PrivateGroup, PublicGroup, BxPrivateGroup)

	global.GVA_ROUTERS = Router.Routes()

	global.GVA_LOG.Info("router register success")
	return Router
}
