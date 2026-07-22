package system

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system/response"
)

var AutoCodeSqlite = new(autoCodeSqlite)

type autoCodeSqlite struct{}

// GetDB Get all database names
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (a *autoCodeSqlite) GetDB(businessDB string) (data []response.Db, err error) {
	var entities []response.Db
	sql := "PRAGMA database_list;"
	var databaseList []struct {
		File string `gorm:"column:file"`
	}
	if businessDB == "" {
		err = global.GVA_DB.Raw(sql).Find(&databaseList).Error
	} else {
		err = global.GVA_DBList[businessDB].Raw(sql).Find(&databaseList).Error
	}
	for _, database := range databaseList {
		if database.File != "" {
			fileName := filepath.Base(database.File)
			fileExt := filepath.Ext(fileName)
			fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)

			entities = append(entities, response.Db{Database: fileNameWithoutExt})
		}
	}
	// entities = append(entities, response.Db{global.GVA_CONFIG.Sqlite.Dbname})
	return entities, err
}

// GetTables Get all table names
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (a *autoCodeSqlite) GetTables(businessDB string, dbName string) (data []response.Table, err error) {
	var entities []response.Table
	sql := `SELECT name FROM sqlite_master WHERE type='table'`
	tabelNames := []string{}
	if businessDB == "" {
		err = global.GVA_DB.Raw(sql).Find(&tabelNames).Error
	} else {
		err = global.GVA_DBList[businessDB].Raw(sql).Find(&tabelNames).Error
	}
	for _, tabelName := range tabelNames {
		entities = append(entities, response.Table{TableName: tabelName})
	}
	return entities, err
}

// GetColumn Get all column names, types, and values for a specified database and table
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (a *autoCodeSqlite) GetColumn(businessDB string, tableName string, dbName string) (data []response.Column, err error) {
	var entities []response.Column
	sql := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	var columnInfos []struct {
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
		Pk   int    `gorm:"column:pk"`
	}
	if businessDB == "" {
		err = global.GVA_DB.Raw(sql).Scan(&columnInfos).Error
	} else {
		err = global.GVA_DBList[businessDB].Raw(sql).Scan(&columnInfos).Error
	}
	for _, columnInfo := range columnInfos {
		entities = append(entities, response.Column{
			ColumnName: columnInfo.Name,
			DataType:   columnInfo.Type,
			PrimaryKey: columnInfo.Pk == 1,
		})
	}
	return entities, err
}
