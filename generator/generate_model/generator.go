// Package generate_model @Author  wangjian    2023/6/15 10:15 AM
package generate_model

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TableColumn struct {
	ColumnName    string `gorm:"column:COLUMN_NAME"`
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
	IsNullable    string `gorm:"column:IS_NULLABLE"`
	DataType      string `gorm:"column:DATA_TYPE"`
	ColumnKey     string `gorm:"column:COLUMN_KEY"`
}

type Config struct {
	*DBConfig
	ModuleName         string
	ModelPackagePath   string            // model文件包路径
	DaoPackagePath     string            // dao文件包路径
	DaoImplPackagePath string            // dao层impl文件路径
	FileNames          []string          // 文件名
	NeedImportPkgPaths []string          // 需要导入的包路径
	FieldTypeMap       map[string]string // 定制化字段类型，字段名到类型映射，例如可实现tinyint->bool
}

func (c *Config) Generate() error {
	// 1.获取所有表的列信息
	tablesColumns, err := c.GetDBTablesAndColumns()
	if err != nil {
		return err
	}

	// 2.构造Formatter对象
	c.UpdateFileNames()
	formatter := &ModelOutputFormatter{
		ModuleName:         c.ModuleName,
		ModelPackagePath:   c.ModelPackagePath,
		NeedImportPkgPaths: c.NeedImportPkgPaths,
		TableNames:         c.TableNames,
		TablesColumns:      tablesColumns,
		OutputBytes:        make(map[string][]byte),
	}

	err = formatter.Format()
	if err != nil {
		return err
	}

	// 3.写入文件
	writer := NewModelWriter(c, formatter.OutputBytes)
	err = writer.Write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetDBTablesAndColumns() (map[string][]*TableColumn, error) {
	// 连接数据库
	db, err := connectDB(c.DBConfig)
	if err != nil {
		return nil, err
	}

	// 获取每一张表的所有列信息
	var tablesColumns = make(map[string][]*TableColumn, len(c.TableNames))
	for _, t := range c.TableNames {
		var columns []*TableColumn
		db.Raw("SELECT COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, DATA_TYPE, COLUMN_KEY FROM columns WHERE table_schema=? AND table_name=? ORDER BY ORDINAL_POSITION", c.DBName,
			t).Scan(&columns)
		// 更新字段类型
		for _, col := range columns {
			if val, ok := c.FieldTypeMap[col.ColumnName]; ok {
				col.DataType = val
			}
		}
		tablesColumns[t] = columns
	}
	return tablesColumns, nil
}

func connectDB(config *DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

const GoFileSuffix = ".go"

// UpdateFileNames
//
//	@Description: 生成FileName，如果没有指定则根据tableName生成
//	@receiver c
//	@return []string
func (c *Config) UpdateFileNames() {
	var fileNames = make([]string, len(c.TableNames))
	for i, t := range c.TableNames {
		if c.FileNames != nil && len(c.FileNames) > i && c.FileNames[i] != "" {
			fileNames[i] = c.FileNames[i] + GoFileSuffix
			continue
		}
		fileNames[i] = GetFileNameByTableName(t)
	}
	c.FileNames = fileNames
}

func GetFileNameByTableName(tableName string) string {
	return tableName + GoFileSuffix
}
