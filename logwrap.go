// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

// Package logwrap provides a very basic logging abstraction supporting
// syslog-style filterable prioritized text messages.  The underlying log
// implementation is injected by providing a wrapper object that implements
// Logger.  Logger instances can be created for specific objects or roles, and
// can specify an identifier for themselves.
//
// Where the underlying log infrastructure is not safe for concurrent use,
// MakeChanLogger allows multiple goroutines to send messages through a
// channel to a goroutine that exclusively uses the logger.
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

// Enables returns true if and only if a logger set to the receiver's priority
// should emit log messages at priority p2.  For example Info.Enables(Crit) is
// true, but Warning.Enables(Debug) is false.
func (p Priority) Enables(p2 Priority) bool {
	return p2 <= p
}

// Logf is the signature for a printf-like function.  Here it's one that's
// bound to a logger and a priority.
type Logf func(format string, args ...interface{})

// MakePriWrapper creates Logf functions bound to the given logger and
// priority.
func MakePriWrapper(lgr ImmutableLogger, pri Priority) Logf {
	return func(format string, args ...interface{}) {
		lgr.F(pri, format, args...)
	}
}

// PriPr provides LogF implementations for each possible priority.
//
// This structure simplifies the common need for short-hand loggers at
// different priorities within a routine.  Instead of doing:
//
//    ...
//    fn(lgr)
//    ...
//
//  func fn(lgr lw.Logger) {
//    lprn := lw.MakePriWrapper(lgr, lw.Notice)
//    lpri := lw.MakePriWrapper(lgr, lw.Info)
//    lprd := lw.MakePriWrapper(lgr, lw.Debug)
//    ...
//    lprn("At notice")
//    lpri("At info")
//    ...
//  }
//
// the application can use:
//
//    ...
//    fn(MakePriPr(lgr))
//    ...
//
//  func fn(lpr *lw.PriPr) {
//    ...
//    lpr.N("At notice")
//    lpr.I("At info")
//    ...
//  }
//
// which avoids having to enable and disable creation of loggers based on
// which levels are used in the routine.
type PriPr struct {
	// Em logs its arguments at Emerg priority.
	Em Logf
	// C logs its arguments at Crit priority.
	C Logf
	// E logs its arguments at Error priority.
	E Logf
	// W logs its arguments at Warning priority.
	W Logf
	// N logs its arguments at Notice priority.
	N Logf
	// I logs its arguments at Info priority.
	I Logf
	// D logs its arguments at Debug priority.
	D Logf
}

// MakePriPri returns a PriPr structure that logs at each priority using lgr.
func MakePriPr(lgr ImmutableLogger) PriPr {
	return PriPr{
		Em: MakePriWrapper(lgr, Emerg),
		C:  MakePriWrapper(lgr, Crit),
		E:  MakePriWrapper(lgr, Error),
		W:  MakePriWrapper(lgr, Warning),
		N:  MakePriWrapper(lgr, Notice),
		I:  MakePriWrapper(lgr, Info),
		D:  MakePriWrapper(lgr, Debug),
	}
}

// ImmutableLogger provides the key functionality for emitting filterable
// prioritized text log messages.
type ImmutableLogger interface {
	// Priority returns the priority of the lowest priority message that
	// will be emitted to the log.  E.g. if set to Warning, Error and
	// Warning messages will be logged, but Notice and Info messages will
	// be dropped.  The default Priority() shall be Warning.
	Priority() Priority

	// F formats a message and emits it to the log, as long as the
	// provided priority is at or above Priority() in precedence.
	F(pri Priority, format string, args ...interface{})
}

// Logger extends ImmutableLogger with methods that can be used to change its
// priority and other behavior.  Types that implement this interface may
// provide an Instance() method that exposes the underlying log object for
// logger-specific configuration.
type Logger interface {
	ImmutableLogger

	// SetId adds an identification string to the start of each emitted
	// message.  By default the logger has no identifier assigned.
	SetId(id string) Logger

	// SetPriority specifies the priority used to filter emitted messages.
	SetPriority(pri Priority) Logger
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

// Priority per ImmutableLogger.
func (v *nullLogger) Priority() Priority {
	return Priority(*v)
}

// F per ImmutableLogger.
func (v *nullLogger) F(pri Priority, format string, args ...interface{}) {}

// SetId per Logger.
func (v *nullLogger) SetId(id string) Logger {
	return v
}

// SetPriority per Logger.
func (v *nullLogger) SetPriority(pri Priority) Logger {
	*v = nullLogger(pri)
	return v
}

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

// Priority per ImmutableLogger.
func (v *LogLogger) Priority() Priority {
	return v.pri
}

// F per ImmutableLogger.  Priorities are represented in the messages as the
// first letter of the priority (or '!' for Emerg) within square brackets
// prefixing the formatted message.
func (v *LogLogger) F(pri Priority, format string, args ...interface{}) {
	if v.pri.Enables(pri) {
		s := fmt.Sprintf(format, args...)
		v.lgr.Printf("[%s] %s", priMap[pri], s)
	}
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

// Instance provides access to the underlying log.Logger to configure things
// that are not part of the logwrap API.
func (v *LogLogger) Instance() *log.Logger {
	return v.lgr
}

// chanLogger is a ImmutableLogger that packages log messages and transmits them
// over a channel where they can be emitted in a different goroutine.
//
// This allows an active object that owns a Logger to spawn goroutines that
// can emit messages on that Logger even if the Logger is not safe for
// concurrent use.
//
// chanLogger's F() method is safe for concurrent use.  Its Priority() method
// is not safe for concurrent use.
type chanLogger struct {
	ech chan<- Emitter
	pfx string
	lgr ImmutableLogger
}

// Emitter is implemented by encapsulated log messages, e.g. those sent by a
// channel logger.
type Emitter interface {
	// Emit emits a log message based on information held by the
	// implementing object.
	Emit()
}

// MakeChanLogger constructs a channel and a ImmutableLogger such that
// messages emitted by the ImmutableLogger are processed and emitted via lgr
// in its native context.
//
// lgr can be any ImmutableLogger.  cap specifies the capacity of the channel
// used to communicate messages.  Values of cap less than 1 are replaced by 1.
//
// Be sure to set cap appropriately so routines that use the channel logger
// will not block because the routine responsible for processing messages from
// it is delayed.
//
// The F method of the returned logger is safe for concurrent use.  The
// returned channel is never closed.
func MakeChanLogger(lgr ImmutableLogger, cap int) (ImmutableLogger, <-chan Emitter) {
	if cap < 1 {
		cap = 1
	}
	ech := make(chan Emitter, cap)
	return &chanLogger{
		ech: ech,
		lgr: lgr,
	}, ech
}

// PrefixedChanLogger constructs a new ImmutableLogger that uses the same
// channel as lgr, but prepends pfx to all format strings passed to the
// returned logger's F function.  This simplifies ensuring that messages can
// be tracked back to the goroutine that produced them.
//
// The returned ImmutableLogger is nil if lgr was not constructed by
// MakeChanLogger.  Calls to the F method of the nil logger will silently drop
// all messages submitted to it.
func PrefixedChanLogger(lgr ImmutableLogger, pfx string) ImmutableLogger {
	var rv *chanLogger
	if cl, ok := lgr.(*chanLogger); ok {
		cl2 := *cl
		cl2.pfx = pfx
		rv = &cl2
	}
	return rv
}

// Priority per ImmutableLogger.
func (v *chanLogger) Priority() Priority {
	return v.lgr.Priority()
}

// F per ImmutableLogger.
func (v *chanLogger) F(pri Priority, format string, args ...interface{}) {
	if v != nil {
		v.ech <- &emittable{
			lgr:  v.lgr,
			pri:  pri,
			fmt:  v.pfx + format,
			args: args,
		}
	}
}

// emittable packages the log message parameters with the logger to be used to
// emit them.  It implements Emitter() to output the message.
type emittable struct {
	lgr  ImmutableLogger
	pri  Priority
	fmt  string
	args []interface{}
}

func (m *emittable) Emit() {
	m.lgr.F(m.pri, m.fmt, m.args...)
}
