package python

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type PythonPlugin struct {
}

const name = "python"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(PythonPlugin)
	})
}

func (p *PythonPlugin) Name() string {
	return name
}

func (p *PythonPlugin) Description() string {
	return "从python脚本获取阈值进行动态扩展 kubernetes 资源"
}

func (p *PythonPlugin) Init(config k8sq1comv1.Config) error {
	return nil
}

func (p *PythonPlugin) Run() (int64, error) {
	return 0, nil
}
