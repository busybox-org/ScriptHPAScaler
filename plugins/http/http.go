package http

import (
	"fmt"
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpPlugin struct {
	Url     string `json:"url,required"`
	Mode    string `json:"mode" default:"GET"`
	Headers string `json:"headers"`
	headers map[string]string
}

const name = "http"

func init() {
	plugins.Register(name, func() plugins.Plugin {
		return new(HttpPlugin)
	})
}

func (h *HttpPlugin) Name() string {
	return name
}

func (h *HttpPlugin) Description() string {
	return "使用http获取阈值进行动态扩展 kubernetes 资源"
}

func (h *HttpPlugin) Init(config k8sq1comv1.Config) error {
	err := plugins.MapToStruct(config, h)
	if err != nil {
		return err
	}
	if h.Url == "" {
		return plugins.ErrInvalidConfig
	}
	if h.Mode == "" {
		h.Mode = "GET"
	}
	h.headers = make(map[string]string)
	if h.Headers != "" {
		for _, header := range strings.Split(h.Headers, ",") {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid header %q", header)
			}
			h.headers[parts[0]] = parts[1]
		}
	}
	return nil
}

func (h *HttpPlugin) Run() (int64, error) {
	req, err := http.NewRequest(h.Mode, h.Url, nil)
	if err != nil {
		return 0, err
	}
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("http status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	count, err := plugins.ParseInt64(string(body))
	if err != nil {
		return 0, err
	}
	return count, nil
}
