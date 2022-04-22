package shell

import (
	"context"
	"github.com/google/uuid"
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ShellPlugin struct {
	Envs   string `json:"envs"`
	Script string `json:"script"`
	envMap map[string]string
	id     string
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

func (s *ShellPlugin) Init(config k8sq1comv1.Config) error {
	s.envMap = make(map[string]string)
	err := plugins.MapToStruct(config, s)
	if err != nil {
		return err
	}
	if s.Script == "" {
		return plugins.ErrInvalidConfig
	}
	s.id = uuid.New().String()
	if s.Envs != "" {
		for _, env := range strings.Split(s.Envs, ",") {
			kv := strings.Split(env, "=")
			if len(kv) != 2 {
				return plugins.ErrInvalidConfig
			}
			s.envMap[kv[0]] = kv[1]
		}
	}
	return nil
}

func (s *ShellPlugin) writeScript(script string) (string, error) {
	fileName := "/tmp/supersetscalers_" + s.id + ".sh"
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

func (s *ShellPlugin) Run() (int64, error) {
	script, err := s.writeScript(s.Script)
	if err != nil {
		return 0, err
	}
	defer os.Remove(script)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	cmd.Env = os.Environ()
	for k, v := range s.envMap {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Env = append(cmd.Env, "SUPERSET_ID="+s.id)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	count, err := plugins.ParseInt64(string(out))
	if err != nil {
		return 0, err
	}
	return count, nil
}
