package plugins

import k8sq1comV1 "github.com/xmapst/supersetscalers/api/v1"

type Plugin interface {
	Name() string
	Description() string
	Run(plugin *k8sq1comV1.Plugin) (int64, error)
}

type PluginsCreator func() Plugin

var Plugins = make(map[string]PluginsCreator)

func Register(name string, creator PluginsCreator) {
	Plugins[name] = creator
}
