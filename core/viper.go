package core

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/core/internal"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/mo7amed4522/tayara_api/global"
)

// expandEnv expands ${VAR} and ${VAR:-default} references in the config text from the
// process environment. A missing variable with no default expands to "". This lets secrets
// (JWT key, DB passwords, R2 keys, Stripe keys, ...) live in environment variables instead of
// being committed to the repo. Example: ${DB_PASSWORD:-} -> env value or ""; ${MYSQL_HOST:-db} -> env or "db".
func expandEnv(s string) string {
	return os.Expand(s, func(k string) string {
		if name, def, ok := strings.Cut(k, ":-"); ok {
			if v, found := os.LookupEnv(name); found {
				return v
			}
			return def
		}
		return os.Getenv(k)
	})
}

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

	// Read the raw file and expand ${VAR}/${VAR:-default} from the environment before
	// unmarshalling, so secrets can be injected via env vars without committing them.
	data, err := os.ReadFile(config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if err := v.ReadConfig(bytes.NewReader([]byte(expandEnv(string(data))))); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		// Re-expand env vars on hot-reload so the change reflects the current environment.
		if d, rerr := os.ReadFile(config); rerr == nil {
			_ = v.ReadConfig(bytes.NewReader([]byte(expandEnv(string(d)))))
		}
		if uerr := v.Unmarshal(&global.GVA_CONFIG); uerr != nil {
			fmt.Println(uerr)
		}
	})
	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(err)
	}

	// root adaptability, find the corresponding migration location based on the root location, ensure the root path is valid
	global.GVA_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	return v
}
