package kafka

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type RocketMQPlugin struct {
}

const name = "rocketmq"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(RocketMQPlugin)
	})
}

func (r *RocketMQPlugin) Name() string {
	return name
}

func (r *RocketMQPlugin) Description() string {
	return "从rocketmq获取阈值进行动态扩展 kubernetes 资源"
}

func (r *RocketMQPlugin) Init(uri string, config k8sq1comv1.Config) error {
	return nil
}

func (r *RocketMQPlugin) Run() (int64, error) {
	return 0, nil
}
