package model

import (
	"RunnerGo-engine/log"
	"RunnerGo-engine/middlewares"
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"sync"
)

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Paths         []FileList     `json:"paths"` // 文件地址
	RealPaths     []string       `json:"real_paths"`
	VariableNames *VariableNames `json:"variable_names"` // 存储变量及数据的map
}

type FileList struct {
	IsChecked int64  `json:"is_checked"` // 1 开， 2： 关
	Path      string `json:"path"`
}

type VariableNames struct {
	VarMapList map[string][]string `json:"var_map_list"`
	Index      int                 `json:"index"`
	Mu         sync.Mutex          `json:"mu"`
}

func (p *ParameterizedFile) UseFile() {
	if p.Paths == nil || len(p.Paths) == 0 {
		return
	}
	fc := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	// set url
	req.Header.SetMethod("GET")
	resp := fasthttp.AcquireResponse()
	defer req.ConnectionClose()
	defer resp.ConnectionClose()
	if p.VariableNames == nil {
		p.VariableNames = new(VariableNames)
	}
	p.VariableNames.VarMapList = make(map[string][]string)
	for _, path := range p.Paths {
		if path.IsChecked != Open {
			continue
		}
		req.Header.SetRequestURI(path.Path)
		if err := fc.Do(req, resp); err != nil {
			log.Logger.Error(fmt.Sprintf("机器ip:%s, 下载参数化文件错误：", middlewares.LocalIp), err)
			continue
		}
		strs := strings.Split(string(resp.Body()), "\n")
		index := 0
		var keys []string
		for _, str := range strs {
			str = strings.TrimSpace(str)
			if index == 0 {
				keys = strings.Split(str, ",")
				for _, data := range keys {
					data = strings.TrimSpace(data)
					if _, ok := p.VariableNames.VarMapList[data]; !ok {
						p.VariableNames.VarMapList[data] = []string{}
					}
					if len(strs) <= 1 {
						p.VariableNames.VarMapList[data] = append(p.VariableNames.VarMapList[data], "")
					}

				}

			} else {
				dataList := strings.Split(str, ",")
				for i := 0; i < len(keys); i++ {
					data := ""
					if len(dataList)-1 >= i {
						data = strings.TrimSpace(dataList[i])
					}
					p.VariableNames.VarMapList[keys[i]] = append(p.VariableNames.VarMapList[keys[i]], data)
				}
			}
			index++
		}
	}
	p.VariableNames.Index = 0
}

// UseVar 使用数据
func (p *ParameterizedFile) UseVar(key string) (value string) {
	if values, ok := p.VariableNames.VarMapList[key]; ok {
		if p.VariableNames.Index >= len(p.VariableNames.VarMapList[key]) {
			p.VariableNames.Index = 0
		}
		value = values[p.VariableNames.Index]
		p.VariableNames.Index++

	}
	return
}
