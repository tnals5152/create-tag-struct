package tag

import "go/ast"

type Tag interface {
	GetFieldName(totalTagValue, defaultName string) string
	GetStructTag(tag *ast.BasicLit) (result string, err error)
}

func NewTag(tag string) (tagInterface Tag, err error) {
	switch tag {
	case "beego":
		return &BeegoTag{}, nil
	}
	return
}
