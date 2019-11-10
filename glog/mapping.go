package glog

import (
	"github.com/mattes/log"
)

var (
	ErrorFunc      = log.Error
	ErrorfFunc     = log.Errorf
	ErrorlnFunc    = log.Error
	ErrorDepthFunc = log.Error // TODO depth

	FatalFunc      = log.Fatal
	FatalfFunc     = log.Fatalf
	FatallnFunc    = log.Fatal
	FatalDepthFunc = log.Fatal // TODO depth

	InfoFunc      = log.Info
	InfofFunc     = log.Infof
	InfolnFunc    = log.Info
	InfoDepthFunc = log.Info

	ExitFunc      = log.Fatal
	ExitfFunc     = log.Fatalf
	ExitlnFunc    = log.Fatal
	ExitDepthFunc = log.Fatal // TODO depth

	WarningFunc      = log.Warn
	WarningfFunc     = log.Warnf
	WarninglnFunc    = log.Warn
	WarningDepthFunc = log.Warn // TODO depth

	VerboseInfoFunc   = log.Debug
	VerboseInfofFunc  = log.Debugf
	VerboseInfolnFunc = log.Debug
)

// DiscardAll discards all messages and also won't exit control flow.
func DiscardAll() {
	ErrorFunc = nil
	ErrorfFunc = nil
	ErrorlnFunc = nil
	ErrorDepthFunc = nil

	FatalFunc = nil
	FatalfFunc = nil
	FatallnFunc = nil
	FatalDepthFunc = nil

	InfoFunc = nil
	InfofFunc = nil
	InfolnFunc = nil
	InfoDepthFunc = nil

	ExitFunc = nil
	ExitfFunc = nil
	ExitlnFunc = nil
	ExitDepthFunc = nil

	WarningFunc = nil
	WarningfFunc = nil
	WarninglnFunc = nil
	WarningDepthFunc = nil

	VerboseInfoFunc = nil
	VerboseInfofFunc = nil
	VerboseInfolnFunc = nil
}

// DiscardError discards Error messages.
func DiscardError() {
	ErrorFunc = nil
	ErrorfFunc = nil
	ErrorlnFunc = nil
	ErrorDepthFunc = nil
}

// DiscardFatal discards Fatal messages and also won't exit control flow.
func DiscardFatal() {
	FatalFunc = nil
	FatalfFunc = nil
	FatallnFunc = nil
	FatalDepthFunc = nil
}

// DiscardInfo discards Info messages.
func DiscardInfo() {
	InfoFunc = nil
	InfofFunc = nil
	InfolnFunc = nil
	InfoDepthFunc = nil
}

// DiscardExit discards Exit messages and also won't exit control flow.
func DiscardExit() {
	ExitFunc = nil
	ExitfFunc = nil
	ExitlnFunc = nil
	ExitDepthFunc = nil
}

// DiscardWarning discards Warning messages.
func DiscardWarning() {
	WarningFunc = nil
	WarningfFunc = nil
	WarninglnFunc = nil
	WarningDepthFunc = nil
}

// DiscardVerboseInfo discards Verbose messages.
func DiscardVerboseInfo() {
	VerboseInfoFunc = nil
	VerboseInfofFunc = nil
	VerboseInfolnFunc = nil
}

// RedirectToDebug redirects Error, Info, Warning, Verbose messages
// to Debug. Fatal messages continue to be logged as Fatal.
// Exit messages continue to be logged as Fatal.
func RedirectToDebug() {
	ErrorFunc = log.Debug
	ErrorfFunc = log.Debugf
	ErrorlnFunc = log.Debug
	ErrorDepthFunc = log.Debug

	// Fatal stays Fatal so it exits
	FatalFunc = log.Fatal
	FatalfFunc = log.Fatalf
	FatallnFunc = log.Fatal
	FatalDepthFunc = log.Fatal

	InfoFunc = log.Debug
	InfofFunc = log.Debugf
	InfolnFunc = log.Debug
	InfoDepthFunc = log.Debug

	// Exit stays Fatal so it exits
	ExitFunc = log.Fatal
	ExitfFunc = log.Fatalf
	ExitlnFunc = log.Fatal
	ExitDepthFunc = log.Fatal

	WarningFunc = log.Debug
	WarningfFunc = log.Debugf
	WarninglnFunc = log.Debug
	WarningDepthFunc = log.Debug

	VerboseInfoFunc = log.Debug
	VerboseInfofFunc = log.Debugf
	VerboseInfolnFunc = log.Debug
}
