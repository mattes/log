package googleErrorReporting

import (
	"strings"
	"testing"
)

func TestTrimStack(t *testing.T) {
	sugarStack := strings.TrimSpace(`
goroutine 1 [running]:
github.com/mattes/log/googleErrorReporting.(*core).Write(0xc00006ae00, 0x2, 0xbf696abc69fb3891, 0x31e1b5, 0xe61100, 0x0, 0x0, 0x9f10bb, 0x10, 0x0, ...)
        /home/mattes/go/src/github.com/mattes/log/googleErrorReporting/core.go:204 +0x207
go.uber.org/zap/zapcore.(*CheckedEntry).Write(0xc0000b1ef0, 0xc0000de380, 0x2, 0x2)
        /home/mattes/go/pkg/mod/go.uber.org/zap@v1.12.0/zapcore/entry.go:216 +0x117
go.uber.org/zap.(*SugaredLogger).log(0xc000211c80, 0xc000211d02, 0x9f10bb, 0x10, 0x0, 0x0, 0x0, 0xc000211c90, 0x2, 0x2)
        /home/mattes/go/pkg/mod/go.uber.org/zap@v1.12.0/sugar.go:234 +0x100
go.uber.org/zap.(*SugaredLogger).Errorw(...)
        /home/mattes/go/pkg/mod/go.uber.org/zap@v1.12.0/sugar.go:191
main.main()
        /home/mattes/go/src/github.com/mattes/log/abc.go:108 +0x775
`)

	normalStack := strings.TrimSpace(`
goroutine 1 [running]:
github.com/mattes/log/googleErrorReporting.(*core).Write(0xc00006ae00, 0x2, 0xbf696abc6a1be229, 0x528b57, 0xe61100, 0x0, 0x0, 0x9f10bb, 0x10, 0x0, ...)
        /home/mattes/go/src/github.com/mattes/log/googleErrorReporting/core.go:204 +0x207
go.uber.org/zap/zapcore.(*CheckedEntry).Write(0xc0000b1ef0, 0x0, 0x0, 0x0)
        /home/mattes/go/pkg/mod/go.uber.org/zap@v1.12.0/zapcore/entry.go:216 +0x117
go.uber.org/zap.(*Logger).Error(0xc00006cd20, 0x9f10bb, 0x10, 0x0, 0x0, 0x0)
        /home/mattes/go/pkg/mod/go.uber.org/zap@v1.12.0/logger.go:203 +0x7f
main.main()
        /home/mattes/go/src/github.com/mattes/log/abc.go:108 +0x775
`)

	expect := strings.TrimSpace(`
goroutine 1 [running]:
main.main()
        /home/mattes/go/src/github.com/mattes/log/abc.go:108 +0x775
`)

	{
		out := trimStack([]byte(sugarStack))
		if expect != string(out) {
			t.Errorf("sugar stack failed:\n%s", out)
		}
	}

	{
		out := trimStack([]byte(normalStack))
		if expect != string(out) {
			t.Errorf("normal stack failed:\n%s", out)
		}
	}
}
