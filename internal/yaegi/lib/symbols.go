package lib

import (
	"os"
	"reflect"

	_ "github.com/avast/retry-go/v4"
	_ "github.com/dlclark/regexp2"
	_ "github.com/dustin/go-humanize"
	_ "github.com/go-cmd/cmd"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/go-zookeeper/zk"
	_ "github.com/google/uuid"
	_ "github.com/gorilla/websocket"
	_ "github.com/imroc/req/v3"
	_ "github.com/json-iterator/go"
	_ "github.com/michaelklishin/rabbit-hole/v2"
	_ "github.com/pelletier/go-toml/v2"
	_ "github.com/pires/go-proxyproto"
	_ "github.com/pkg/errors"
	_ "github.com/rabbitmq/amqp091-go"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/segmentio/ksuid"
	_ "github.com/tidwall/gjson"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unrestricted"
	"github.com/traefik/yaegi/stdlib/unsafe"
	_ "github.com/xmapst/go-rabbitmq"
	_ "github.com/yargevad/filepathx"
	_ "google.golang.org/grpc"
	_ "gopkg.in/yaml.v3"
	_ "k8s.io/klog/v2"
)

var Symbols = map[string]map[string]reflect.Value{}

func init() {
	for _, symbols := range []interp.Exports{
		stdlib.Symbols,
		unsafe.Symbols,
		syscall.Symbols,
		unrestricted.Symbols,
		interp.Symbols,
	} {
		for name, value := range symbols {
			Symbols[name] = value
		}
	}

	// function, constant and variable definitions
	Symbols["os/os"]["Chown"] = reflect.ValueOf(os.Chown)
	Symbols["os/os"]["Chmod"] = reflect.ValueOf(os.Chmod)

	// type definitions
	Symbols["os/os"]["FileInfo"] = reflect.ValueOf((*os.FileInfo)(nil))
}

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@latest
//go:generate yaegi extract github.com/avast/retry-go/v4
//go:generate yaegi extract github.com/dlclark/regexp2
//go:generate yaegi extract github.com/dustin/go-humanize
//go:generate yaegi extract github.com/go-cmd/cmd
//go:generate yaegi extract github.com/go-sql-driver/mysql
//go:generate yaegi extract github.com/go-zookeeper/zk
//go:generate yaegi extract github.com/google/uuid
//go:generate yaegi extract github.com/gorilla/websocket
//go:generate yaegi extract github.com/imroc/req/v3
//go:generate yaegi extract github.com/json-iterator/go
//go:generate yaegi extract github.com/michaelklishin/rabbit-hole/v2
//go:generate yaegi extract github.com/pelletier/go-toml/v2
//go:generate yaegi extract github.com/pires/go-proxyproto
//go:generate yaegi extract github.com/pkg/errors
//go:generate yaegi extract github.com/rabbitmq/amqp091-go
//go:generate yaegi extract github.com/redis/go-redis/v9
//go:generate yaegi extract github.com/segmentio/ksuid
//go:generate yaegi extract github.com/tidwall/gjson
//go:generate yaegi extract github.com/xmapst/go-rabbitmq
//go:generate yaegi extract github.com/yargevad/filepathx
//go:generate yaegi extract google.golang.org/grpc
//go:generate yaegi extract gopkg.in/yaml.v3
//go:generate yaegi extract k8s.io/klog/v2
