// Copyright 2021 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

// Package logwrap provides a very basic abstraction supporting syslog-style
// filterable prioritized string messages.  Logger instances can be created
// for specific objects or roles, which can specify an identifier for
// themselves.
//
// The use case is helper packages which should emit log messages with the
// same tool as the application itself.
package logwrap

import (
	"fmt"
	"log"
	"os"
)

// Priority distinguishs log message priority.  Higher priority messages have
// lower numeric value.  Priority levels derive from the classic syslog(3)
// taxonomy.
type Priority int32

const (
	// The system is unusable
	Emerg Priority = iota
	// Critical conditions
	Crit
	// Error conditions
	Error
	// Warning conditions
	Warning
	// Normal but significate
	Notice
	// Informational
	Info
	// Debugging
	Debug
)

// Logger provides the key functionality for filterable prioritized text log
// messages.  Types that implement this interface may provide an Instance()
// method that exposes the underlying log object for logger-specific
// configuration.
type Logger interface {
	// SetId adds an identification string to the start of each emitted
	// message.
	SetId(id string) Logger

	// SetPriority specifies the priority used to filter emitted messages.
	SetPriority(pri Priority) Logger

	// Priority returns the priority of the lowest priority message that
	// will be emitted to the log.  E.g. if set to Warning, Error and
	// Warning messages will be logged, but Notice and Info messages will
	// be dropped.
	Priority() Priority

	// F formats a message and emits it to the log, as long as the
	// provided priority is at or above Priority() in precedence.
	F(pri Priority, format string, args ...interface{})
}

// A LogMaker is a factory function that constructs a logger instance for some
// object or operation.  It allows the selection of a log infrastructure to be
// injected into a package in a way that ensures active objects created by the
// package are provided with a custom log before any goroutines associated
// with the object are started.
type LogMaker func(owner interface{}) Logger

// NullLogMaker returns a Logger that drops all messages sent to it.
func NullLogMaker(interface{}) Logger {
	var lgr = nullLogger(Warning)
	return &lgr
}

type nullLogger Priority

// Implement Logger.
func (v *nullLogger) SetId(id string) Logger {
	return v
}

// Implement Logger.
func (v *nullLogger) SetPriority(pri Priority) Logger {
	*v = nullLogger(pri)
	return v
}

// Implement Logger.
func (v *nullLogger) Priority() Priority {
	return Priority(*v)
}

// Implement Logger.
func (v *nullLogger) F(pri Priority, format string, args ...interface{}) {}

// LogLogger uses a dedicated instance of log.Logger.
type LogLogger struct {
	lgr *log.Logger
	pri Priority
}

// LogLogMaker returns a Logger that uses the core log package to emit
// messages via the Print API.  The initial priority is Warning.
func LogLogMaker(interface{}) Logger {
	return &LogLogger{
		lgr: log.New(os.Stderr, "", log.LstdFlags),
		pri: Warning,
	}
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

// Implement Logger.  The provided id becomes the log.Logger prefix, and
// log.Lmsgprefix is applied to the flags.
func (v *LogLogger) SetId(id string) Logger {
	v.lgr.SetFlags(v.lgr.Flags() | log.Lmsgprefix)
	v.lgr.SetPrefix(id)
	return v
}

// Implement Logger.
func (v *LogLogger) SetPriority(pri Priority) Logger {
	v.pri = pri
	return v
}

// Implement Logger.
func (v *LogLogger) Priority() Priority {
	return v.pri
}

// Implement Logger.  Priorities are represented in the messages as the first
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
