package redis

import (
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
)

type RedisPlugin struct {
	Key string `json:"key"`
}

const name = "redis"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(RedisPlugin)
	})
}

func (r *RedisPlugin) Name() string {
	return name
}

func (r *RedisPlugin) Description() string {
	return "从redis获取阈值进行动态扩展 kubernetes 资源"
}

func (r *RedisPlugin) Init(config k8sq1comv1.Config) error {
	return nil
}

func (r *RedisPlugin) Run() (int64, error) {
	return 0, nil
}
