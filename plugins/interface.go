package plugins

import k8sq1comV1 "github.com/xmapst/supersetscalers/api/v1"

type Plugin interface {
	Name() string
	Description() string
	Init(uri string, config k8sq1comV1.Config) error
	Run() (int64, error)
}

type PluginsCreator func() Plugin

var Plugins = make(map[string]PluginsCreator)

func Register(name string, creator PluginsCreator) {
	Plugins[name] = creator
}

func GetPlugin(name string) Plugin {
	creator, ok := Plugins[name]
	if !ok {
		return nil
	}
	return creator()
}
