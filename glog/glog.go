package glog

var MaxSize uint64 = 0

var Stats struct {
	Info, Warning, Error OutputStats
}

func CopyStandardLogTo(name string) {}

func Error(args ...interface{}) {
	if ErrorFunc != nil {
		ErrorFunc(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if ErrorfFunc != nil {
		ErrorfFunc(format, args...)
	}
}

func ErrorDepth(depth int, args ...interface{}) {
	// TODO depth
	if ErrorDepthFunc != nil {
		ErrorDepthFunc(args...)
	}
}

func Errorln(args ...interface{}) {
	if ErrorlnFunc != nil {
		ErrorlnFunc(args...)
	}
}

func Fatal(args ...interface{}) {
	if FatalFunc != nil {
		FatalFunc(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if FatalfFunc != nil {
		FatalfFunc(format, args...)
	}
}

func Fatalln(args ...interface{}) {
	if FatallnFunc != nil {
		FatallnFunc(args...)
	}
}

func Info(args ...interface{}) {
	if InfoFunc != nil {
		InfoFunc(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if InfofFunc != nil {
		InfofFunc(format, args...)
	}
}

func Exit(args ...interface{}) {
	if ExitFunc != nil {
		ExitFunc(args...)
	}
}

func ExitDepth(depth int, args ...interface{}) {
	// TODO depth
	if ExitDepthFunc != nil {
		ExitDepthFunc(args...)
	}
}

func Exitf(format string, args ...interface{}) {
	if ExitfFunc != nil {
		ExitfFunc(format, args...)
	}
}

func Exitln(args ...interface{}) {
	if ExitlnFunc != nil {
		ExitlnFunc(args...)
	}
}

func FatalDepth(depth int, args ...interface{}) {
	// TODO depth
	if FatalDepthFunc != nil {
		FatalDepthFunc(args...)
	}
}

func Flush() {}

func InfoDepth(depth int, args ...interface{}) {
	// TODO depth
	if InfoDepthFunc != nil {
		InfoDepthFunc(args...)
	}
}

func Infoln(args ...interface{}) {
	if InfolnFunc != nil {
		InfolnFunc(args...)
	}
}

func Warning(args ...interface{}) {
	if WarningFunc != nil {
		WarningFunc(args...)
	}
}

func WarningDepth(depth int, args ...interface{}) {
	// TODO depth
	if WarningDepthFunc != nil {
		WarningDepthFunc(args...)
	}
}

func Warningf(format string, args ...interface{}) {
	if WarningfFunc != nil {
		WarningfFunc(format, args...)
	}
}

func Warningln(args ...interface{}) {
	if WarninglnFunc != nil {
		WarninglnFunc(args...)
	}
}

type Level int32

func (l *Level) Get() interface{} {
	return 0
}

func (l *Level) Set(value string) error {
	return nil
}

func (l *Level) String() string {
	return ""
}

type OutputStats struct{}

func (s *OutputStats) Bytes() int64 {
	return 0
}

func (s *OutputStats) Lines() int64 {
	return 0
}

type Verbose bool

func V(level Level) Verbose {
	return false
}

func (v Verbose) Info(args ...interface{}) {
	if VerboseInfoFunc != nil {
		VerboseInfoFunc(args...)
	}
}

func (v Verbose) Infof(format string, args ...interface{}) {
	if VerboseInfofFunc != nil {
		VerboseInfofFunc(format, args...)
	}
}

func (v Verbose) Infoln(args ...interface{}) {
	if VerboseInfolnFunc != nil {
		VerboseInfolnFunc(args...)
	}
}
