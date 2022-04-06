package rabbitmq

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type RabbitMQPlugin struct {
}

const name = "rabbitmq"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(RabbitMQPlugin)
	})
}

func (mq *RabbitMQPlugin) Name() string {
	return name
}

func (mq *RabbitMQPlugin) Description() string {
	return "使用 AMQP 队列的长度（可从队列中检索的消息数）动态扩展 kubernetes 资源"
}

func (mq *RabbitMQPlugin) Run(plugin *k8sq1comv1.Plugin) (int64, error) {
	return getQueueLength(plugin.Url, plugin.Config["queue"])
}
