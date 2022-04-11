package http

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type HttpPlugin struct {
	uri          string
	Address      string            `json:"address,required"`
	Mode         string            `json:"mode" default:"GET"`
	ReadTimeout  int               `json:"read_timeout" default:"5"`
	WriteTimeout int               `json:"write_timeout" default:"5"`
	Headers      map[string]string `json:"headers"`
}

const name = "http"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(HttpPlugin)
	})
}

func (h *HttpPlugin) Name() string {
	return name
}

func (h *HttpPlugin) Description() string {
	return "使用http获取阈值进行动态扩展 kubernetes 资源"
}

func (h *HttpPlugin) Init(uri string, config k8sq1comv1.Config) error {
	h.uri = uri
	err := plugins.MapToStruct(config, h)
	if err != nil {
		return err
	}
	return nil
}

func (h *HttpPlugin) Run() (int64, error) {
	return 0, nil
}
