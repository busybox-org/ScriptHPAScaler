// Code generated by 'yaegi extract k8s.io/klog/v2'. DO NOT EDIT.

package lib

import (
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"reflect"
)

func init() {
	Symbols["k8s.io/klog/v2/klog"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Background":              reflect.ValueOf(klog.Background),
		"CalculateMaxSize":        reflect.ValueOf(klog.CalculateMaxSize),
		"CaptureState":            reflect.ValueOf(klog.CaptureState),
		"ClearLogger":             reflect.ValueOf(klog.ClearLogger),
		"ContextualLogger":        reflect.ValueOf(klog.ContextualLogger),
		"CopyStandardLogTo":       reflect.ValueOf(klog.CopyStandardLogTo),
		"EnableContextualLogging": reflect.ValueOf(klog.EnableContextualLogging),
		"Error":                   reflect.ValueOf(klog.Error),
		"ErrorDepth":              reflect.ValueOf(klog.ErrorDepth),
		"ErrorS":                  reflect.ValueOf(klog.ErrorS),
		"ErrorSDepth":             reflect.ValueOf(klog.ErrorSDepth),
		"Errorf":                  reflect.ValueOf(klog.Errorf),
		"ErrorfDepth":             reflect.ValueOf(klog.ErrorfDepth),
		"Errorln":                 reflect.ValueOf(klog.Errorln),
		"ErrorlnDepth":            reflect.ValueOf(klog.ErrorlnDepth),
		"Exit":                    reflect.ValueOf(klog.Exit),
		"ExitDepth":               reflect.ValueOf(klog.ExitDepth),
		"ExitFlushTimeout":        reflect.ValueOf(&klog.ExitFlushTimeout).Elem(),
		"Exitf":                   reflect.ValueOf(klog.Exitf),
		"ExitfDepth":              reflect.ValueOf(klog.ExitfDepth),
		"Exitln":                  reflect.ValueOf(klog.Exitln),
		"ExitlnDepth":             reflect.ValueOf(klog.ExitlnDepth),
		"Fatal":                   reflect.ValueOf(klog.Fatal),
		"FatalDepth":              reflect.ValueOf(klog.FatalDepth),
		"Fatalf":                  reflect.ValueOf(klog.Fatalf),
		"FatalfDepth":             reflect.ValueOf(klog.FatalfDepth),
		"Fatalln":                 reflect.ValueOf(klog.Fatalln),
		"FatallnDepth":            reflect.ValueOf(klog.FatallnDepth),
		"Flush":                   reflect.ValueOf(klog.Flush),
		"FlushAndExit":            reflect.ValueOf(klog.FlushAndExit),
		"FlushLogger":             reflect.ValueOf(klog.FlushLogger),
		"Format":                  reflect.ValueOf(klog.Format),
		"FromContext":             reflect.ValueOf(klog.FromContext),
		"Info":                    reflect.ValueOf(klog.Info),
		"InfoDepth":               reflect.ValueOf(klog.InfoDepth),
		"InfoS":                   reflect.ValueOf(klog.InfoS),
		"InfoSDepth":              reflect.ValueOf(klog.InfoSDepth),
		"Infof":                   reflect.ValueOf(klog.Infof),
		"InfofDepth":              reflect.ValueOf(klog.InfofDepth),
		"Infoln":                  reflect.ValueOf(klog.Infoln),
		"InfolnDepth":             reflect.ValueOf(klog.InfolnDepth),
		"InitFlags":               reflect.ValueOf(klog.InitFlags),
		"KObj":                    reflect.ValueOf(klog.KObj),
		"KObjSlice":               reflect.ValueOf(klog.KObjSlice),
		"KObjs":                   reflect.ValueOf(klog.KObjs),
		"KRef":                    reflect.ValueOf(klog.KRef),
		"LogToStderr":             reflect.ValueOf(klog.LogToStderr),
		"LoggerWithName":          reflect.ValueOf(klog.LoggerWithName),
		"LoggerWithValues":        reflect.ValueOf(klog.LoggerWithValues),
		"MaxSize":                 reflect.ValueOf(&klog.MaxSize).Elem(),
		"New":                     reflect.ValueOf(&klog.New).Elem(),
		"NewContext":              reflect.ValueOf(klog.NewContext),
		"NewKlogr":                reflect.ValueOf(klog.NewKlogr),
		"NewStandardLogger":       reflect.ValueOf(klog.NewStandardLogger),
		"OsExit":                  reflect.ValueOf(&klog.OsExit).Elem(),
		"SetLogFilter":            reflect.ValueOf(klog.SetLogFilter),
		"SetLogger":               reflect.ValueOf(klog.SetLogger),
		"SetLoggerWithOptions":    reflect.ValueOf(klog.SetLoggerWithOptions),
		"SetOutput":               reflect.ValueOf(klog.SetOutput),
		"SetOutputBySeverity":     reflect.ValueOf(klog.SetOutputBySeverity),
		"SetSlogLogger":           reflect.ValueOf(klog.SetSlogLogger),
		"StartFlushDaemon":        reflect.ValueOf(klog.StartFlushDaemon),
		"Stats":                   reflect.ValueOf(&klog.Stats).Elem(),
		"StopFlushDaemon":         reflect.ValueOf(klog.StopFlushDaemon),
		"TODO":                    reflect.ValueOf(klog.TODO),
		"V":                       reflect.ValueOf(klog.V),
		"VDepth":                  reflect.ValueOf(klog.VDepth),
		"Warning":                 reflect.ValueOf(klog.Warning),
		"WarningDepth":            reflect.ValueOf(klog.WarningDepth),
		"Warningf":                reflect.ValueOf(klog.Warningf),
		"WarningfDepth":           reflect.ValueOf(klog.WarningfDepth),
		"Warningln":               reflect.ValueOf(klog.Warningln),
		"WarninglnDepth":          reflect.ValueOf(klog.WarninglnDepth),
		"WriteKlogBuffer":         reflect.ValueOf(klog.WriteKlogBuffer),

		// type definitions
		"KMetadata":    reflect.ValueOf((*klog.KMetadata)(nil)),
		"Level":        reflect.ValueOf((*klog.Level)(nil)),
		"LogFilter":    reflect.ValueOf((*klog.LogFilter)(nil)),
		"LogSink":      reflect.ValueOf((*klog.LogSink)(nil)),
		"Logger":       reflect.ValueOf((*klog.Logger)(nil)),
		"LoggerOption": reflect.ValueOf((*klog.LoggerOption)(nil)),
		"ObjectRef":    reflect.ValueOf((*klog.ObjectRef)(nil)),
		"OutputStats":  reflect.ValueOf((*klog.OutputStats)(nil)),
		"RuntimeInfo":  reflect.ValueOf((*klog.RuntimeInfo)(nil)),
		"State":        reflect.ValueOf((*klog.State)(nil)),
		"Verbose":      reflect.ValueOf((*klog.Verbose)(nil)),

		// interface wrapper definitions
		"_KMetadata": reflect.ValueOf((*_k8s_io_klog_v2_KMetadata)(nil)),
		"_LogFilter": reflect.ValueOf((*_k8s_io_klog_v2_LogFilter)(nil)),
		"_LogSink":   reflect.ValueOf((*_k8s_io_klog_v2_LogSink)(nil)),
		"_State":     reflect.ValueOf((*_k8s_io_klog_v2_State)(nil)),
	}
}

// _k8s_io_klog_v2_KMetadata is an interface wrapper for KMetadata type
type _k8s_io_klog_v2_KMetadata struct {
	IValue        interface{}
	WGetName      func() string
	WGetNamespace func() string
}

func (W _k8s_io_klog_v2_KMetadata) GetName() string {
	return W.WGetName()
}
func (W _k8s_io_klog_v2_KMetadata) GetNamespace() string {
	return W.WGetNamespace()
}

// _k8s_io_klog_v2_LogFilter is an interface wrapper for LogFilter type
type _k8s_io_klog_v2_LogFilter struct {
	IValue   interface{}
	WFilter  func(args []interface{}) []interface{}
	WFilterF func(format string, args []interface{}) (string, []interface{})
	WFilterS func(msg string, keysAndValues []interface{}) (string, []interface{})
}

func (W _k8s_io_klog_v2_LogFilter) Filter(args []interface{}) []interface{} {
	return W.WFilter(args)
}
func (W _k8s_io_klog_v2_LogFilter) FilterF(format string, args []interface{}) (string, []interface{}) {
	return W.WFilterF(format, args)
}
func (W _k8s_io_klog_v2_LogFilter) FilterS(msg string, keysAndValues []interface{}) (string, []interface{}) {
	return W.WFilterS(msg, keysAndValues)
}

// _k8s_io_klog_v2_LogSink is an interface wrapper for LogSink type
type _k8s_io_klog_v2_LogSink struct {
	IValue      interface{}
	WEnabled    func(level int) bool
	WError      func(err error, msg string, keysAndValues ...any)
	WInfo       func(level int, msg string, keysAndValues ...any)
	WInit       func(info logr.RuntimeInfo)
	WWithName   func(name string) logr.LogSink
	WWithValues func(keysAndValues ...any) logr.LogSink
}

func (W _k8s_io_klog_v2_LogSink) Enabled(level int) bool {
	return W.WEnabled(level)
}
func (W _k8s_io_klog_v2_LogSink) Error(err error, msg string, keysAndValues ...any) {
	W.WError(err, msg, keysAndValues...)
}
func (W _k8s_io_klog_v2_LogSink) Info(level int, msg string, keysAndValues ...any) {
	W.WInfo(level, msg, keysAndValues...)
}
func (W _k8s_io_klog_v2_LogSink) Init(info logr.RuntimeInfo) {
	W.WInit(info)
}
func (W _k8s_io_klog_v2_LogSink) WithName(name string) logr.LogSink {
	return W.WWithName(name)
}
func (W _k8s_io_klog_v2_LogSink) WithValues(keysAndValues ...any) logr.LogSink {
	return W.WWithValues(keysAndValues...)
}

// _k8s_io_klog_v2_State is an interface wrapper for State type
type _k8s_io_klog_v2_State struct {
	IValue   interface{}
	WRestore func()
}

func (W _k8s_io_klog_v2_State) Restore() {
	W.WRestore()
}
