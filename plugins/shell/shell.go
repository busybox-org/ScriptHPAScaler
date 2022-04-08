package shell

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type ShellPlugin struct {
}

const name = "shell"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(ShellPlugin)
	})
}

func (s *ShellPlugin) Name() string {
	return name
}

func (s *ShellPlugin) Description() string {
	return "从shell脚本获取阈值进行动态扩展 kubernetes 资源"
}

func (s *ShellPlugin) Run(plugin *k8sq1comv1.Plugin) (int64, error) {
	return 0, nil
}
