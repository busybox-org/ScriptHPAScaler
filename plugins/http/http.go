package http

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type HttpPlugin struct {
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

func (h *HttpPlugin) Run(plugin *k8sq1comv1.Plugin) (int64, error) {
	return 0, nil
}
