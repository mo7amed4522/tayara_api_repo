package initialize

import (
	"context"

	"sort"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/service/system"
	"go.uber.org/zap"
)

// InitializeData runs all registered data initializers
// This function is called from main.go to initialize data from source/system and source/example
func InitializeData() {
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("Database not initialized, skipping data initialization")
		return
	}

	// Get all registered initializers from the system service
	initializers := system.GetRegisteredInitializers()
	if len(initializers) == 0 {
		global.GVA_LOG.Info("No data initializers registered, skipping data initialization")
		return
	}

	sort.Sort(&initializers)
	// Create context with database
	ctx := context.Background()
	ctx = context.WithValue(ctx, "db", global.GVA_DB)
	ctx = context.WithValue(ctx, "dbtype", global.GVA_CONFIG.System.DbType)

	global.GVA_LOG.Info("Starting data initialization...")

	// Run data initialization for each initializer
	for _, init := range initializers {
		initName := init.InitializerName()

		// Check if data already exists
		if init.DataInserted(ctx) {
			global.GVA_LOG.Info("Data already exists for: " + initName)
			continue
		}

		// Initialize data
		global.GVA_LOG.Info("Initializing data for: " + initName)
		newCtx, err := init.InitializeData(ctx)
		if err != nil {
			global.GVA_LOG.Error("Failed to initialize data for: "+initName, zap.Error(err))
			continue
		}

		// Update context for dependent initializers
		ctx = newCtx
		global.GVA_LOG.Info("Successfully initialized data for: " + initName)
	}

	global.GVA_LOG.Info("Data initialization completed")
}
