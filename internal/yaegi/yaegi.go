package yaegi

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/tidwall/gjson"
	"github.com/traefik/yaegi/interp"
	"k8s.io/apimachinery/pkg/util/json"
	log "k8s.io/klog/v2"

	"github.com/busybox-org/scripthpascaler/internal/yaegi/lib"
)

func Eval(ctx context.Context, script string, params map[string]any) (int64, error) {
	defer func() {
		if _r := recover(); _r != nil {
			stack := debug.Stack()
			log.Errorln("panic during execution", _r, string(stack))
		}
	}()
	data, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}

	vm, err := newVM(ctx, script)
	if err != nil {
		return 0, err
	}
	evalFnval, err := vm.EvalWithContext(ctx, "EvalCall")
	if err != nil {
		log.Errorln("failed to get EvalCall function from yaegi script", err)
		return 0, err
	}
	evalFn, ok := evalFnval.Interface().(func(ctx context.Context, params gjson.Result) (int64, error))
	if !ok {
		return 0, fmt.Errorf("failed to get EvalCall function from yaegi script")
	}
	return evalFn(ctx, gjson.ParseBytes(data))
}

func newVM(ctx context.Context, script string) (*interp.Interpreter, error) {
	vm := interp.New(interp.Options{
		Stdout: os.Stdout,
		Stderr: os.Stdin,
	})
	if err := vm.Use(lib.Symbols); err != nil {
		log.Errorln("use lib error", err)
		return nil, err
	}
	vm.ImportUsed()
	if _, err := vm.EvalWithContext(ctx, script); err != nil {
		log.Errorln("eval script error", err)
		return nil, err
	}
	return vm, nil
}
