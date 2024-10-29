package godev

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"
)

var (
	defaultLabelMap = map[string]string{
		"ID":         "ID",
		"CreateTIme": "创建时间",
		"DeleteTime": "删除时间",
		"UpdateTime": "更新时间",
	}
)

type Generate struct {
	val    reflect.Value
	Fields []Field
}

func (g *Generate) Names() Names {
	return Names(g.val.Type().Name())
}

// NewGenerate 模板生成器
func NewGenerate(model any) (*Generate, error) {
	var g = new(Generate)
	g.val = reflect.ValueOf(model)
	if g.val.Kind() == reflect.Ptr {
		g.val = g.val.Elem()
	}
	if g.val.Kind() != reflect.Struct {
		return nil, errors.New("不是有效的结构体")
	}
	for _, field := range reflect.VisibleFields(g.val.Type()) {
		g.Fields = append(g.Fields, newField(field))
	}
	return g, nil
}

type Field struct {
	field       reflect.StructField
	ValidateTag TagValue
	GormTag     TagValue
}

func newField(f reflect.StructField) Field {
	return Field{
		ValidateTag: newTagValue(f.Tag.Get("validate"), ",", "="),
		GormTag:     newTagValue(f.Tag.Get("gorm"), ";", ":"),
		field:       f,
	}
}

// Name 获取结构体名称
func (f *Field) Name() string {
	return f.field.Name
}

// Names
//
//	ToSnake(s)	any_kind_of_string
//	ToSnakeWithIgnore(s, '.')	any_kind.of_string
//	ToScreamingSnake(s)	ANY_KIND_OF_STRING
//	ToKebab(s)	any-kind-of-string
//	ToScreamingKebab(s)	ANY-KIND-OF-STRING
//	ToDelimited(s, '.')	any.kind.of.string
//	ToScreamingDelimited(s, '.', ”, true)	ANY.KIND.OF.STRING
//	ToScreamingDelimited(s, '.', ' ', true)	ANY.KIND OF.STRING
//	ToCamel(s)	AnyKindOfString
//	ToLowerCamel(s)	anyKindOfString
func (f *Field) Names() Names {
	return Names(f.field.Name)
}

// JsonName 获取json名称，如未设置返回Name
func (f *Field) JsonName() string {
	if name := f.field.Tag.Get("json"); name == "" || name == "-" {
		return f.Name()
	} else {
		return name
	}
}

// IsJson 是否设置json字段
func (f *Field) IsJson() bool {
	text := f.field.Tag.Get("json")
	return text != "" && text != "-"
}

// IsRequired 是否必填
func (f *Field) IsRequired() bool {
	//validate
	return f.field.Tag.Get("required") == "true"
}

func (f *Field) Label() string {
	var comment string
	if comment = f.field.Tag.Get("label"); comment != "" {
		return comment
	}
	comment = f.GormTag.Get("comment")
	if comment != "" {
		return comment
	}
	comment = f.defaultLabel()
	if comment != "" {
		return comment
	}
	return f.Name()
}

// Remark 字段备注信息
func (f *Field) Remark() string {
	return f.field.Tag.Get("remark")
}

// IsNumber 是否数字
func (f *Field) IsNumber() bool {
	switch f.typ().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.TypeFor[sql.NullInt32]().Kind(), reflect.TypeFor[sql.NullInt64]().Kind(), reflect.TypeFor[sql.NullInt16]().Kind():
		return true
	default:
		return false
	}
}

// IsString 是否字符串
func (f *Field) IsString() bool {

	switch f.typ().Kind() {
	case reflect.String, reflect.TypeFor[sql.NullString]().Kind():
		return true
	default:
		return false
	}
}

// IsBool 是否bool
func (f *Field) IsBool() bool {

	switch f.typ().Kind() {
	case reflect.Bool, reflect.TypeFor[sql.NullBool]().Kind():
		return true
	default:
		return false
	}
}

// IsTime 是否时间
func (f *Field) IsTime() bool {

	switch f.typ().Kind() {
	case reflect.TypeFor[time.Time]().Kind(), reflect.TypeFor[sql.NullTime]().Kind(), reflect.TypeFor[gorm.DeletedAt]().Kind():
		return true
	default:
		return false
	}
}

func (f *Field) typ() reflect.Type {
	typ := f.field.Type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}
func (f *Field) defaultLabel() string {
	return defaultLabelMap[f.field.Name]
}

type TagValue map[string]string

func newTagValue(str string, sep string, valuesep string) TagValue {
	val := make(TagValue)
	if str == "" {
		return val
	}
	for _, v := range strings.Split(str, sep) {
		if valuesep == "" {
			val[v] = v
		} else {
			var valueArr = strings.Split(v, valuesep)
			if len(valueArr) == 1 {
				val[v] = v
			} else if len(valueArr) == 2 {
				val[valueArr[0]] = valueArr[1]
			}
		}
	}
	return val
}
func (v TagValue) IsExist(key string) bool {
	_, ok := v[key]
	return ok
}
func (v TagValue) Get(key string) string {
	val, _ := v[key]
	return val
}

// Names
//
//	| Function                        | Result             |
//	|---------------------------------|--------------------|
//	| ToSnake(s)                      | any_kind_of_string |
//	| ToScreamingSnake(s)             | ANY_KIND_OF_STRING |
//	| ToKebab(s)                      | any-kind-of-string |
//	| ToScreamingKebab(s)             | ANY-KIND-OF-STRING |
//	| ToDelimited(s, '.')             | any. kind. of. string |
//	| ToScreamingDelimited(s, '.')    | ANY. KIND. OF. STRING |
//	| ToCamel(s)                      | AnyKindOfString    |
//	| ToLowerCamel(s)                 | anyKindOfString    |
type Names string

func (name Names) ToCamel() string {
	return strcase.ToCamel(string(name))
}
func (name Names) ToLowerCamel() string {
	return strcase.ToLowerCamel(string(name))
}
func (name Names) ToSnake() string {
	return strcase.ToSnake(string(name))
}
func (name Names) ToScreamingSnake() string {
	return strcase.ToScreamingSnake(string(name))
}

func (name Names) ToKebab() string {
	return strcase.ToKebab(string(name))
}

func (name Names) ToScreamingKebab() string {
	return strcase.ToScreamingKebab(string(name))
}

func (name Names) ToDelimited(delimiter uint8) string {
	return strcase.ToDelimited(string(name), delimiter)
}

func (name Names) ToScreamingDelimited(delimiter uint8, ignore string, screaming bool) string {
	return strcase.ToScreamingDelimited(string(name), delimiter, ignore, screaming)
}
func (name Names) String() string {
	return string(name)
}

func NewRender(glob string, data any, covered bool) error {
	tmp, err := template.New("").ParseGlob(glob)
	if err != nil {
		return errors.New("解析模板失败 " + err.Error())
	}
	var tmps = make(map[string][]byte)
	for _, t := range tmp.Templates() {
		if strings.HasPrefix(t.Name(), "-") {
			path, err := renderPath(t.Name(), data)
			if err != nil {
				return fmt.Errorf("渲染路径失败 " + err.Error())
			}
			buf, err := renderTemp(t, data)
			if err != nil {
				return fmt.Errorf("渲染模板失败 " + err.Error())
			}
			tmps[path] = buf
		}
	}
	if len(tmps) == 0 {
		return errors.New("没有解析到模板")
	}
	for path, buf := range tmps {
		if !covered {
			_, err = os.Stat(path)
			if err == nil {
				fmt.Printf("路径已存在 %s \n", path)
				continue
			}
		}
		if err = os.WriteFile(path, buf, 0644); err != nil {
			return fmt.Errorf("无法写入文件 %s"+err.Error(), path)
		} else {
			fmt.Printf("-- 生成成功 %s \n", path)
		}
	}

	return nil
}
func renderTemp(t *template.Template, data any) ([]byte, error) {
	var strs bytes.Buffer
	err := t.Execute(&strs, data)
	if err != nil {
		return nil, err
	}
	return strs.Bytes(), nil
}
func renderPath(str string, data interface{}) (string, error) {
	str = str[1:]
	var strs bytes.Buffer
	parse, err := template.New("").Parse(str)
	if err != nil {
		return "", err
	}
	err = parse.Execute(&strs, data)
	if err != nil {
		return "", err
	}
	return strs.String(), nil
}
