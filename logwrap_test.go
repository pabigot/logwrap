// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

package logwrap

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
)

// Run standard verification of expected errors, i.e. that err is an
// error and its text contains errstr.
func confirmError(t *testing.T, err error, base error, errstr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("succeed, expected error %s", errstr)
	}
	if base != nil && !errors.Is(err, base) {
		t.Fatalf("err not from %s: %s", base, err)
	}
	if testing.Verbose() {
		t.Logf("Error=`%v`", err.Error())
	}
	if !strings.Contains(err.Error(), errstr) {
		t.Fatalf("failed, missing %s: %v", errstr, err)
	}
}

func TestLogLogger(t *testing.T) {
	var sb strings.Builder
	lgr := LogLogMaker(nil)

	wrapped, ok := lgr.(*LogLogger)
	if !ok {
		t.Errorf("failed to get wrapped implementation")
	}
	inst := wrapped.Instance()

	if v := inst.Flags(); v != log.LstdFlags {
		t.Errorf("Init flags %x not %x", v, log.LstdFlags)
	}

	inst.SetOutput(&sb)

	lgr.SetId("TestLogLogger ")
	if v := inst.Flags(); v != log.LstdFlags|log.Lmsgprefix {
		t.Errorf("SetId did not enable Lmsgprefix: %x", v)
	}

	lgr.F(Warning, "with prefix")
	if lv := sb.String(); !strings.HasSuffix(lv, "TestLogLogger [W] with prefix\n") {
		t.Errorf("bad warning: %s", lv)
	}
	sb.Reset()

	if p := lgr.Priority(); p != Warning {
		t.Errorf("unexpected init priority: %d", int(p))
	}
	lgr.F(Debug, "debug at Warning priority")
	if lv := sb.String(); lv != "" {
		t.Errorf("bad filtered debug: %s", lv)
	}
	sb.Reset()

	// The null logger retains its configured priority even though it
	// isn't used, for consistent behavior with other loggers.
	if p := lgr.SetPriority(Debug).Priority(); p != Debug {
		t.Errorf("failed to set priority: %d", int(p))
	}
	lgr.F(Debug, "debug at debug priority")
	if lv := sb.String(); !strings.HasSuffix(lv, "TestLogLogger [D] debug at debug priority\n") {
		t.Errorf("bad warning: %s", lv)
	}
	sb.Reset()
}

func TestNullLogger(t *testing.T) {
	lgr := NullLogMaker(nil)
	lgr.F(Emerg, "made it this far")
	if p := lgr.Priority(); p != Warning {
		t.Errorf("unexpected init priority: %d", int(p))
	}
	// The null logger retains its configured priority even though it
	// isn't used, for consistent behavior with other loggers.
	if p := lgr.SetPriority(Debug).Priority(); p != Debug {
		t.Errorf("failed to set priority: %d", int(p))
	}
	// SetId should work but have no effect.
	lgr.SetId("id")
}

func TestParsePriority(t *testing.T) {
	type testCase struct {
		pri    Priority
		inputs []string
	}
	testCases := []testCase{
		{Emerg, []string{Emerg.String(), "EmeRgenCY", "emerg"}},
		{Crit, []string{Crit.String(), "critical", "CRIT"}},
		{Error, []string{Error.String(), "error"}},
		{Warning, []string{Warning.String(), "wARN", "Warning"}},
		{Notice, []string{Notice.String(), "Notice"}},
		{Info, []string{Info.String(), "info"}},
		{Debug, []string{Debug.String(), "DeBug"}},
	}

	for _, tc := range testCases {
		for _, s := range tc.inputs {
			if pri, ok := ParsePriority(s); pri != tc.pri || !ok {
				t.Errorf("Failed %s = %s: %s %t",
					s, tc.pri, pri, ok)
			}
		}
	}
	if _, ok := ParsePriority("wrn"); ok {
		t.Error("Improper success")
	}
}

func TestEnables(t *testing.T) {
	if !Info.Enables(Crit) {
		t.Errorf("enables wrong for Info.Crit")
	}
	if Warning.Enables(Debug) {
		t.Errorf("enables wrong for Warning.Debug")
	}
}

func TestSet(t *testing.T) {
	var pri Priority
	if err := (&pri).Set("debug"); err != nil || pri != Debug {
		t.Errorf("Set failed: %s, %v", pri, err)
	}
	err := pri.Set("fatal")
	confirmError(t, err, ErrInvalidPriority, "invalid priority: fatal")
}

func TestMakePriWrapper(t *testing.T) {
	var sb strings.Builder
	lgr := LogLogMaker(nil)
	lgr.SetId("ID ")
	priorities := []Priority{
		Emerg, Debug, Crit, Info, Error, Notice, Warning,
	}

	lgr.(*LogLogger).Instance().SetOutput(&sb)
	lgr.SetPriority(Debug)

	for i, pri := range priorities {
		plgr := MakePriWrapper(lgr, pri)
		plgr("Test %d", i)
		exp := fmt.Sprintf("ID [%s] Test %d\n", priMap[pri], i)
		out := sb.String()
		sb.Reset()
		t.Logf("%s => %s", pri, out)
		if !strings.HasSuffix(out, exp) {
			t.Errorf("%s failed: %s", pri, out)
		}
		sb.Reset()
	}
}

func TestMakePriPr(t *testing.T) {
	var sb strings.Builder
	lgr := LogLogMaker(nil)
	lgr.SetId("ID ")

	lgr.(*LogLogger).Instance().SetOutput(&sb)
	lgr.SetPriority(Debug)
	lpr := MakePriPr(lgr)

	ck := func(t *testing.T, pri Priority) {
		t.Helper()
		exp := fmt.Sprintf("ID [%s] Test\n", priMap[pri])
		out := sb.String()
		sb.Reset()
		t.Logf("%s => %s", pri, out)
		if !strings.HasSuffix(out, exp) {
			t.Errorf("%s failed: %s", pri, out)
		}
		sb.Reset()
	}

	lpr.Em("Test")
	ck(t, Emerg)
	lpr.D("Test")
	ck(t, Debug)
	lpr.C("Test")
	ck(t, Crit)
	lpr.I("Test")
	ck(t, Info)
	lpr.E("Test")
	ck(t, Error)
	lpr.N("Test")
	ck(t, Notice)
	lpr.W("Test")
	ck(t, Warning)
}

type logOwner struct {
	lgr Logger
}

func (lo *logOwner) LogPriority() Priority {
	return lo.lgr.Priority()
}

func (lo *logOwner) LogSetPriority(pri Priority) {
	lo.lgr.SetPriority(pri)
}

func TestLogOwner(t *testing.T) {
	lo := &logOwner{
		lgr: LogLogMaker(nil),
	}
	lo.lgr.SetId("owned")

	var ilo LogOwner = lo
	if lp := ilo.LogPriority(); lp != Warning {
		t.Fatalf("bad init prio: %s", lp)
	}
	lo.LogSetPriority(Debug)
	if lp := ilo.LogPriority(); lp != Debug {
		t.Fatalf("bad changed prio: %s", lp)
	}
}

func TestChanLogger(t *testing.T) {
	var sb strings.Builder
	blgr := LogLogMaker(nil)
	blgr.(*LogLogger).Instance().SetOutput(&sb)

	lgr, lch := MakeChanLogger(blgr, -1)
	if v := cap(lch); v != 1 {
		t.Errorf("cap not updated to 1: %d", v)
	}
	if lgr.Priority() != blgr.Priority() {
		t.Errorf("priority not forwarded")
	}

	fmt := "format: %s %d"
	lgr.F(Warning, fmt, "arg", 2)
	if sb.Len() != 0 {
		t.Error("premature log")
	}
	m := <-lch
	if e, ok := m.(*emittable); !ok || e.lgr != blgr || e.pri != Warning || e.fmt != fmt || len(e.args) != 2 {
		t.Error("wrong emittable content")
	}
	m.Emit()
	if s := sb.String(); !strings.HasSuffix(s, " [W] format: arg 2\n") {
		t.Errorf("wrong content: %s", s)
	}
}
