// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

// Package logwrap provides a very basic abstraction supporting syslog-style
// filterable prioritized string messages.  Logger instances can be created
// for specific objects or roles, and can specify an identifier for
// themselves.
//
// The use case is helper packages that should emit log messages with the
// same tool as the application itself.
package logwrap

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

// Priority distinguishes log message priority.  Higher priority messages have
// lower numeric value.  Priority levels derive from the classic syslog(3)
// taxonomy.
type Priority int32

const (
	// Emerg means the system is unusable
	Emerg Priority = iota
	// Crit identifies critical conditions
	Crit
	// Error conditions
	Error
	// Warning conditions
	Warning
	// Notice identifies normal but significant conditions
	Notice
	// Info identifies informational messages
	Info
	// Debug is used for debugging
	Debug
)

var (
	// ErrInvalidPriority indicates that Set() was invoked with an
	// incorrect text representation of a priority.
	ErrInvalidPriority = errors.New("invalid priority")
)

func (p Priority) String() string {
	switch p {
	case Emerg:
		return "Emerg"
	case Crit:
		return "Crit"
	case Error:
		return "Error"
	case Warning:
		return "Warning"
	case Notice:
		return "Notice"
	case Info:
		return "Info"
	case Debug:
		return "Debug"
	}
	panic("unhandled Priority")
}

// ParsePriority accepts strings of any case corresponding to Priority
// identifiers and returns the corresponding Priority value paired with true.
// If the string does not identify a priority the returned boolean will be
// false.
func ParsePriority(s string) (pri Priority, ok bool) {
	ok = true
	switch strings.ToLower(s) {
	default:
		ok = false
	case "emergency":
		fallthrough
	case "emerg":
		pri = Emerg
	case "critical":
		fallthrough
	case "crit":
		pri = Crit
	case "error":
		pri = Error
	case "warn":
		fallthrough
	case "warning":
		pri = Warning
	case "notice":
		pri = Notice
	case "info":
		pri = Info
	case "debug":
		pri = Debug
	}
	return
}

// Logf is the signature for a printf-like function.  Here it's one that's
// bound to a logger and a priority.
type Logf func(format string, args ...interface{})

// MakePriWrapper creates Logf functions bound to the given logger and
// priority.
func MakePriWrapper(lgr Logger, pri Priority) Logf {
	return func(format string, args ...interface{}) {
		lgr.F(pri, format, args...)
	}
}

// Logger provides the key functionality for filterable prioritized text log
// messages.  Types that implement this interface may provide an Instance()
// method that exposes the underlying log object for logger-specific
// configuration.
type Logger interface {
	// SetId adds an identification string to the start of each emitted
	// message.  By default the logger assumes it has no identifier
	// assigned.
	SetId(id string) Logger

	// Priority returns the priority of the lowest priority message that
	// will be emitted to the log.  E.g. if set to Warning, Error and
	// Warning messages will be logged, but Notice and Info messages will
	// be dropped.  The default Priority() shall be Warning.
	Priority() Priority

	// SetPriority specifies the priority used to filter emitted messages.
	SetPriority(pri Priority) Logger

	// F formats a message and emits it to the log, as long as the
	// provided priority is at or above Priority() in precedence.
	F(pri Priority, format string, args ...interface{})
}

// LogOwner indicates that the implementing object owns a Logger, and provides
// ways to access its priority.
type LogOwner interface {
	// LogPriority returns the priority of an owned Logger.
	LogPriority() Priority

	// LogSetPriority changes the priority of an owned Logger.
	LogSetPriority(pri Priority)
}

// A LogMaker is a factory function that constructs a logger instance for some
// object or operation.  It allows the selection of a log infrastructure to be
// injected into a package in a way that ensures active objects created by the
// package are provided with a custom log before any goroutines associated
// with the object are started.
//
// Unless a LogMaker specifies otherwise the defaults for created Logger
// instances should be the same: no identifier assigned, priority is Warning.
type LogMaker func(owner interface{}) Logger

// NullLogMaker returns a Logger that drops all messages sent to it.
func NullLogMaker(interface{}) Logger {
	var lgr = nullLogger(Warning)
	return &lgr
}

type nullLogger Priority

// SetId per Logger.
func (v *nullLogger) SetId(id string) Logger {
	return v
}

// SetPriority per Logger.
func (v *nullLogger) SetPriority(pri Priority) Logger {
	*v = nullLogger(pri)
	return v
}

// Priority per Logger.
func (v *nullLogger) Priority() Priority {
	return Priority(*v)
}

// F per Logger.
func (v *nullLogger) F(pri Priority, format string, args ...interface{}) {}

// LogLogger uses a dedicated instance of log.Logger.
type LogLogger struct {
	lgr *log.Logger
	pri Priority
}

// LogLogMaker returns a Logger that uses a dedicated instance of the core
// log.Logger type to emit messages via the Print API.  The initial priority
// is Warning.
func LogLogMaker(interface{}) Logger {
	return &LogLogger{
		lgr: log.New(os.Stderr, "", log.LstdFlags),
		pri: Warning,
	}
}

// Set a priority variable from a string.  This supports flag.Value.
func (p *Priority) Set(s string) (err error) {
	if pri, ok := ParsePriority(s); ok {
		*p = pri
	} else {
		err = fmt.Errorf("%w: %s", ErrInvalidPriority, s)
	}
	return
}

var priMap = map[Priority]string{
	Emerg:   "!",
	Crit:    "C",
	Error:   "E",
	Warning: "W",
	Notice:  "N",
	Info:    "I",
	Debug:   "D",
}

// SetId per Logger.  The provided id becomes the log.Logger prefix,
// and log.Lmsgprefix is applied to the flags.
func (v *LogLogger) SetId(id string) Logger {
	v.lgr.SetFlags(v.lgr.Flags() | log.Lmsgprefix)
	v.lgr.SetPrefix(id)
	return v
}

// SetPriority per Logger.
func (v *LogLogger) SetPriority(pri Priority) Logger {
	v.pri = pri
	return v
}

// Priority per Logger.
func (v *LogLogger) Priority() Priority {
	return v.pri
}

// F per Logger.  Priorities are represented in the messages as the first
// letter of the priority (or '!' for Emerg) within square brackets prefixing
// the formatted message.
func (v *LogLogger) F(pri Priority, format string, args ...interface{}) {
	if pri <= v.pri {
		s := fmt.Sprintf(format, args...)
		v.lgr.Printf("[%s] %s", priMap[pri], s)
	}
}

// Instance provides access to the underlying log.Logger to configure things
// that are not part of the logwrap API.
func (v *LogLogger) Instance() *log.Logger {
	return v.lgr
}
