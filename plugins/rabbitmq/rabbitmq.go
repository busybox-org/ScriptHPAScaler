package rabbitmq

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
	"strings"
)

type RabbitMQPlugin struct {
	Url   string `json:"url,required"`
	Queue string `json:"queue,required"`
}

const name = "rabbitmq"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(RabbitMQPlugin)
	})
}

func (r *RabbitMQPlugin) Name() string {
	return name
}

func (r *RabbitMQPlugin) Description() string {
	return "使用 AMQP 队列的长度（可从队列中检索的消息数）动态扩展 kubernetes 资源"
}

func (r *RabbitMQPlugin) Init(config k8sq1comv1.Config) error {
	err := plugins.MapToStruct(config, r)
	if err != nil {
		return err
	}
	if r.Url == "" || r.Queue == "" {
		return plugins.ErrInvalidConfig
	}
	return nil
}

func (r *RabbitMQPlugin) Run() (int64, error) {
	if strings.HasPrefix(r.Url, "http") {
		return r.getQueueLengthFromAPI()
	}
	return 0, plugins.ErrProtocol
}
