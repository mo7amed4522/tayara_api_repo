package request

import (
	"fmt"
	"os"

	"github.com/mo7amed4522/tayara_api/config"
)

type InitDB struct {
	AdminPassword string `json:"adminPassword" binding:"required"` // Admin password
	DBType        string `json:"dbType"`                           // Database type
	Host          string `json:"host"`                             // Server address
	Port          string `json:"port"`                             // Database connection port
	UserName      string `json:"userName"`                         // Database username
	Password      string `json:"password"`                         // Database password
	DBName        string `json:"dbName" binding:"required"`        // Database name
	DBPath        string `json:"dbPath"`                           // SQLite database file path
	Template      string `json:"template"`                         // PostgreSQL specified template
}

// MysqlEmptyDsn Empty database link for MySQL
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) MysqlEmptyDsn() string {
	if i.Host == "" {
		i.Host = "127.0.0.1"
	}
	if i.Port == "" {
		i.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", i.UserName, i.Password, i.Host, i.Port)
}

// PgsqlEmptyDsn Empty database link for PostgreSQL
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) PgsqlEmptyDsn() string {
	if i.Host == "" {
		i.Host = "127.0.0.1"
	}
	if i.Port == "" {
		i.Port = "5432"
	}
	return "host=" + i.Host + " user=" + i.UserName + " password=" + i.Password + " port=" + i.Port + " dbname=" + "postgres" + " " + "sslmode=disable TimeZone=Asia/Shanghai"
}

// SqliteEmptyDsn Empty database link for SQLite
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) SqliteEmptyDsn() string {
	separator := string(os.PathSeparator)
	return i.DBPath + separator + i.DBName + ".db"
}

// MssqlEmptyDsn Empty database link for Microsoft SQL Server
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) MssqlEmptyDsn() string {
	return "sqlserver://" + i.UserName + ":" + i.Password + "@" + i.Host + ":" + i.Port + "?database=" + i.DBName + "&encrypt=disable"
}

// ToMysqlConfig Convert config.Mysql
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) ToMysqlConfig() config.Mysql {
	return config.Mysql{
		GeneralDB: config.GeneralDB{
			Path:         i.Host,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
}

// ToPgsqlConfig Convert config.Pgsql
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) ToPgsqlConfig() config.Pgsql {
	return config.Pgsql{
		GeneralDB: config.GeneralDB{
			Path:         i.Host,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "sslmode=disable TimeZone=Asia/Shanghai",
		},
	}
}

// ToSqliteConfig Convert config.Sqlite
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) ToSqliteConfig() config.Sqlite {
	return config.Sqlite{
		GeneralDB: config.GeneralDB{
			Path:         i.DBPath,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "",
		},
	}
}

// ToMssqlConfig Convert config.Mssql
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (i *InitDB) ToMssqlConfig() config.Mssql {
	return config.Mssql{
		GeneralDB: config.GeneralDB{
			Path:         i.DBPath,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "",
		},
	}
}
