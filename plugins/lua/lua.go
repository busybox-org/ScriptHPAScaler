package lua

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type LuaPlugin struct {
}

const name = "lua"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(LuaPlugin)
	})
}

func (l *LuaPlugin) Name() string {
	return name
}

func (l *LuaPlugin) Description() string {
	return "从lua脚本获取阈值进行动态扩展 kubernetes 资源"
}

func (l *LuaPlugin) Init(uri string, config k8sq1comv1.Config) error {
	return nil
}

func (l *LuaPlugin) Run() (int64, error) {
	return 0, nil
}
