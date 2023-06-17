package write

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	tags "tnals5152/create/tag/write/tag"
)

/*
gorm: 기본 스네이크, gorm:column:{value}
beego: 기본 스네이크, orm:column({value})
*/

func GetStructModel(tag string, files []*FileStructs) (err error) {
	tagModel, err := tags.NewTag(tag)

	if err != nil {
		return
	}

	for _, file := range files {
		var data []byte
		data, err = os.ReadFile(file.TotalPath)
		if err != nil {
			return
		}

		fset := token.NewFileSet()
		var f *ast.File
		f, err = parser.ParseFile(fset, "", string(data), 0)
		if err != nil {
			return
		}

		for _, decl := range f.Decls {

			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structModel := &StructModel{
					StructName: ts.Name.Name,
					Tag:        tag,
				}

				fieldStructType, ok := ts.Type.(*ast.StructType)

				if !ok {
					continue
				}

				for _, tsField := range fieldStructType.Fields.List {
					if len(tsField.Names) != 1 {
						continue
					}

					var tagBasicLit string
					tagBasicLit, err = tagModel.GetStructTag(tsField.Tag)

					if err != nil {
						return
					}

					fieldName := tagModel.GetFieldName(tagBasicLit, tsField.Names[0].Name)

					structModel.Fields = append(structModel.Fields, fieldName)

					if structModel.MaxLength < len(fieldName) {
						structModel.MaxLength = len(fieldName)
					}
				}

				file.StructModels = append(file.StructModels, structModel)
			}

		}

	}
	return
}

// 해당 디렉토리 위치에 있는 모든 파일 이름 가져오기
func GetFiles(files *[]*FileStructs, dir string) (err error) {
	dirFiles, err := os.ReadDir(dir)

	if err != nil {
		return
	}

	path, err := filepath.Abs(dir)

	if err != nil {
		return
	}

	for _, file := range dirFiles {

		if file.IsDir() {
			GetFiles(files, dir+"/"+file.Name())
			continue
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		*files = append(*files, &FileStructs{
			TotalPath: path + "/" + file.Name(),
			MainPath:  dir + "/" + file.Name(),
		})

	}

	return
}

func Start(dir string, tag string, writePath string) {
	var files []*FileStructs = make([]*FileStructs, 0)
	GetFiles(&files, dir)

	fmt.Println(files)

	err := GetStructModel(tag, files)

	fmt.Println(err)

	if err != nil {
		panic(err)
	}

	WriteFile(files, writePath)

}

type FileStructs struct {
	MainPath     string
	TotalPath    string
	StructModels []*StructModel
}

type StructModel struct {
	StructName string
	Tag        string
	Fields     []string
	MaxLength  int
}

func WriteFile(files []*FileStructs, writePath string) {

	for _, file := range files {
		if len(file.StructModels) == 0 {
			continue
		}

		paths := strings.Split(writePath+"/"+file.MainPath, "/")

		if len(paths) < 2 {
			return
		}

		dir := strings.Join(paths[:len(paths)-1], "/")

		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}

		f, err := os.OpenFile(writePath+"/"+file.MainPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		_, err = f.WriteString(MakeCode(file.StructModels, getPackageName(file.MainPath)))

		fmt.Println(err)

	}

}

func getPackageName(path string) string {
	paths := strings.Split(path, "/")

	if len(paths) == 0 {
		return ""
	}

	file := paths[len(paths)-1]

	packageName := strings.TrimSuffix(file, ".go")
	return packageName
}

func MakeCode(structModels []*StructModel, packageName string) (code string) {
	code += WriteCode(`package ` + packageName)

	for _, structModel := range structModels {

		var fieldsCode []string

		fieldsCode = append(fieldsCode, fmt.Sprintf(
			Join(
				"// structName: %s, tag: %s\n",
				"var %s struct {\n",
			),
			structModel.StructName, structModel.Tag, structModel.StructName))

		for _, field := range structModel.Fields {
			emptyLength := structModel.MaxLength - len(field) + 1
			fieldsCode = append(fieldsCode, fmt.Sprintf(
				// "\t%s = \"%s\"\n", getFieldName(field), field,
				"\t%s%sstring\n", getFieldName(field), getSpace(emptyLength),
			))
		}

		fieldsCode = append(fieldsCode, "}")

		code += WriteCode(fieldsCode...)

		fieldsCode = make([]string, 0)

		fieldsCode = append(fieldsCode, fmt.Sprintf("func %sSet() {\n", structModel.StructName))

		for _, field := range structModel.Fields {

			fieldsCode = append(fieldsCode, fmt.Sprintf(
				"\t%s.%s = \"%s\"\n",
				structModel.StructName, getFieldName(field), field,
			))
		}

		fieldsCode = append(fieldsCode, "}")
		code += WriteCode(fieldsCode...)
	}

	return
}

func getSpace(length int) string {
	var result []string

	for i := 0; i < length; i++ {
		result = append(result, " ")
	}
	return Join(result...)
}

func getFieldName(field string) string {
	return strings.ToUpper(field)
}

func WriteCode(code ...string) string {
	return strings.Join(code, "") + "\n\n"
}

func Join(str ...string) string {
	return strings.Join(str, "")
}
