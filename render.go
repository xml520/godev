// Package godev
// Generate
// {{/* gotype: github.com/xml520/godev.Generate */}}
// {{define "-{path}"}}
package godev

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
)

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
