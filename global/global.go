package global

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/utils/timer"
	"github.com/qiniu/qmgo"

	"github.com/songzhibin97/gkit/cache/local_cache"

	"golang.org/x/sync/singleflight"

	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	GVA_DB                  *gorm.DB
	GVA_DBList              map[string]*gorm.DB
	GVA_REDIS               redis.UniversalClient
	GVA_REDISList           map[string]redis.UniversalClient
	GVA_MONGO               *qmgo.QmgoClient
	GVA_CONFIG              config.Server
	GVA_VP                  *viper.Viper
	GVA_LOG                 *zap.Logger
	GVA_Timer               timer.Timer = timer.NewTimerTask()
	GVA_Concurrency_Control             = &singleflight.Group{}
	GVA_ROUTERS             gin.RoutesInfo
	GVA_ACTIVE_DBNAME       *string
	BlackCache              local_cache.Cache
	lock                    sync.RWMutex
)

// GetGlobalDBByDBName Retrieves databases from the database list by name
func GetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	return GVA_DBList[dbname]
}

// MustGetGlobalDBByDBName Retrieve the database by name; panic if it does not exist.
func MustGetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := GVA_DBList[dbname]
	if !ok || db == nil {
		panic("db no init")
	}
	return db
}

func GetRedis(name string) redis.UniversalClient {
	redis, ok := GVA_REDISList[name]
	if !ok || redis == nil {
		panic(fmt.Sprintf("redis `%s` no init", name))
	}
	return redis
}
