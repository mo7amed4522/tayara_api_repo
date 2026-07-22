package core

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/core/internal"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/mo7amed4522/tayara_api/global"
)

// Viper //
// Priority: Command line > Environment variables > Default value
// Author [mo7amed4522](https://github.com/mo7amed4522)
func Viper(path ...string) *viper.Viper {
	var config string

	if len(path) == 0 {
		flag.StringVar(&config, "c", "", "choose config file.")
		flag.Parse()
		if config == "" { // Check if command-line arguments are empty
			if configEnv := os.Getenv(internal.ConfigEnv); configEnv == "" { // Check if the environment variable stored in the internal.ConfigEnv constant is empty
				switch gin.Mode() {
				case gin.DebugMode:
					config = internal.ConfigDefaultFile
				case gin.ReleaseMode:
					config = internal.ConfigReleaseFile
				case gin.TestMode:
					config = internal.ConfigTestFile
				}
				fmt.Printf("You are using the %s environment name in gin mode, the path of config is %s\n", gin.Mode(), config)
			} else { // The environment variable stored in the internal.ConfigEnv constant is not empty, the value is assigned to config
				config = configEnv
				fmt.Printf("You are using the %s environment variable, the path of config is %s\n", internal.ConfigEnv, config)
			}
		} else { // The command line parameter is not empty, the value is assigned to config
			fmt.Printf("You are using the -c command line parameter, the path of config is %s\n", config)
		}
	} else { // The first value of the variable parameter passed by the function is assigned to config
		config = path[0]
		fmt.Printf("You are using the func Viper() passed value, the path of config is %s\n", config)
	}

	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
			fmt.Println(err)
		}
	})
	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(err)
	}

	// root adaptability, find the corresponding migration location based on the root location, ensure the root path is valid
	global.GVA_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	return v
}
