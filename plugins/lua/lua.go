package lua

import (
	"github.com/google/uuid"
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
	"github.com/yuin/gopher-lua"
	"os"
)

type LuaPlugin struct {
	Script string `json:"script"`
	id     string
}

const name = "lua"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(LuaPlugin)
	})
}

func (p *LuaPlugin) Name() string {
	return name
}

func (p *LuaPlugin) Description() string {
	return "从lua脚本获取阈值进行动态扩展 kubernetes 资源"
}

func (p *LuaPlugin) Init(config k8sq1comv1.Config) error {
	err := plugins.MapToStruct(config, p)
	if err != nil {
		return err
	}
	if p.Script == "" {
		return plugins.ErrInvalidConfig
	}
	p.id = uuid.New().String()
	return nil
}

func (p *LuaPlugin) writeScript(script string) (string, error) {
	fileName := "/tmp/supersetscalers_" + p.id + ".lua"
	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.WriteString(script)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func (p *LuaPlugin) Run() (int64, error) {
	l := lua.NewState()
	defer l.Close()
	script, err := p.writeScript(p.Script)
	if err != nil {
		return 0, err
	}
	defer os.Remove(script)
	if err = l.DoFile(script); err != nil {
		return 0, err
	}
	err = l.CallByParam(lua.P{
		Fn:      l.GetGlobal("main"),
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return 0, err
	}
	ret := l.Get(-1)
	l.Pop(1)
	return plugins.ParseInt64(ret.String())
}
