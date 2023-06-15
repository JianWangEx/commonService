// Package generate_model @Author  wangjian    2023/6/15 4:15 PM
package generate_model

import (
	"fmt"
	"testing"
)

func TestGenerate(t *testing.T) {
	config := &Config{
		DBConfig: &DBConfig{
			DSN:    "root:Daren718299.@tcp(175.178.156.113:3306)/information_schema?timeout=10s",
			DBName: "yoga_go_test",
			TableNames: []string{
				"private_training_course_tab",
				"private_training_course_record_tab",
			},
		},
		ModuleName:       "yoga-api-go",
		ModelPackagePath: "internal/data/model",
		FileNames:        nil,
		NeedImportPkgPaths: []string{
			"time",
			"yoga-api-go/internal/config",
		},
	}
	err := config.Generate()
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully generated")
}
