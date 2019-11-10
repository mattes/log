package log

// funcs from std/log package

func Print(v ...interface{}) {
	defaultSugarLogger.Info(v)
}

func Printf(format string, v ...interface{}) {
	defaultSugarLogger.Infof(format, v)
}

func Println(v ...interface{}) {
	defaultSugarLogger.Info(v)
}

func Panicln(v ...interface{}) {
	defaultSugarLogger.Panic(v)
}

func Fatalln(v ...interface{}) {
	defaultSugarLogger.Fatal(v)
}
