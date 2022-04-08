package kafka

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type KafkaPlugin struct {
}

const name = "kafka"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(KafkaPlugin)
	})
}

func (k *KafkaPlugin) Name() string {
	return name
}

func (k *KafkaPlugin) Description() string {
	return "从kafka获取阈值进行动态扩展 kubernetes 资源"
}

func (k *KafkaPlugin) Run(plugin *k8sq1comv1.Plugin) (int64, error) {
	return 0, nil
}
