package tag

import (
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/gobeam/stringy"
)

type BeegoTag struct{}

func (*BeegoTag) GetFieldName(totalTagValue, defaultName string) string {
	return GetOrmName(totalTagValue, defaultName)
}

func (*BeegoTag) GetStructTag(tag *ast.BasicLit) (result string, err error) {

	if tag == nil {
		result = reflect.StructTag("").Get("orm")
		return
	}
	tagValue, err := strconv.Unquote(tag.Value)

	if err != nil {
		return
	}

	result = reflect.StructTag(tagValue).Get("orm")
	return
}

// sf.Name = defaultName
func GetOrmName(totalTagValue, defaultName string) (ormName string) {
	ormName = defaultName
	if totalTagValue == "" {
		// snake로 변경 후 return
		ormName = stringy.New(defaultName).SnakeCase().ToLower()
		return
	}

	tags := make(map[string]string)

	// ex) column(id);auto
	for _, value := range strings.Split(totalTagValue, ";") {
		if value == "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.ToLower(value)

		if i := strings.Index(value, "("); i > 0 && strings.Index(value, ")") == len(value)-1 {
			tagName := value[:i]
			tagValue := value[i+1 : len(value)-1]

			tags[tagName] = tagValue

		}
	}

	for _, tagName := range []string{"rel", "reverse", "column"} {
		tagValue := tags[tagName]

		if tagValue == "" {
			continue
		}
		switch tagName {
		case "column":
			ormName = tagValue
		case "rel": // 이거 먼저 체크
			switch tagValue {
			case "fk", "one":
				//+_id
				ormName += "_id"
				return
			case "m2m":
				//sf.Name
				return
			}
		case "reverse": // 그 다음 이거
			switch tagValue {
			case "one", "many":
				//sf.Name
				return
			}
		}
	}
	return
}
