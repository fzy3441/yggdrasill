package main

import (
	"flag"
	"fmt"

	"github.com/lpxxn/yggdrasill/code_gen"
	"github.com/lpxxn/yggdrasill/code_gen/templates"
	"github.com/lpxxn/yggdrasill/config"
	"github.com/lpxxn/yggdrasill/db_meta_data"
	"github.com/lpxxn/yggdrasill/utils"
)

var (
	target       string
	dsn          string
	destDir      string
	packageName  string
	tableNames   strFlags
	templatePath string
	debug        bool
	isJsonTag    bool
	tagName      string
	prefix       string
)

type strFlags []string

func (i *strFlags) String() string {
	return "table names"
}

func (i *strFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	flag.StringVar(&target, "target", "postgresql", "mysql postgresql[pg]")
	flag.StringVar(&dsn, "dsn", "postgresql", "dsn")
	flag.StringVar(&destDir, "dir", "./model", "Destination dir for files generated.")
	flag.StringVar(&packageName, "package_name", "model", "package name default model.")
	flag.StringVar(&templatePath, "template_path", "", "custom template file path")
	flag.Var(&tableNames, "table_names", "if it is empty, will generate all tables in database")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&isJsonTag, "is_json_tag", true, "default is true and show json tag")
	flag.StringVar(&tagName, "tag", "db", "default is 'db' and show json tag is 'db' ")
	flag.StringVar(&prefix, "prefix", "", "set table prefix")

}

func main() {
	flag.Parse()
	config.Global.PackageName = packageName
	config.Global.DestDir = destDir
	config.Global.IsJsonTag = isJsonTag
	config.Global.TagName = tagName
	config.Global.Prefix = prefix
	utils.CleanUpGenFiles(config.Global.DestDir)
	utils.MkdirPathIfNotExist(config.Global.DestDir)
	dbMetaData, err := code_gen.NewDbCodeGen(target)
	if err != nil {
		fmt.Println("unsupported db type, please input mysql postgresql[pg]")
		return
	}
	if err := dbMetaData.ConnectionDB(dsn); err != nil {
		fmt.Printf("connection db error: %#v", err)
		return
	}
	tables := db_meta_data.TableMetaDataList{}
	if len(tableNames) == 0 {
		tables, err = dbMetaData.AllTableData()
	} else {
		tables, err = dbMetaData.SpecifiedTables(tableNames)
	}
	if err != nil {
		fmt.Printf("AllTableData err: %#v", err)
		return
	}
	if packageName == "" {
		packageName = "model"
	}
	temp := templates.NewTemplate()
	if err := temp.SetPath(templatePath); err != nil {
		fmt.Printf("template path error: %#v", err)
		return
	}
	for _, item := range tables {
		body, err := code_gen.GenerateTemplate(temp.GetTemplate(), item, map[string]interface{}{
			"packageName": config.Global.PackageName,
		})
		if debug {
			fmt.Println(string(body))
		}
		if err != nil {
			fmt.Printf("GenerateTemplate err: %#v", err.Error())
			return
		}
		if err := utils.SaveFile(destDir, item.StructName()+".go", body); err != nil {
			fmt.Printf("save file error: %#v", err)
			return
		}
	}
	fmt.Println("generate finished...")
}
