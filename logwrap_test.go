// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

package logwrap

import (
	"errors"
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

func TestSet(t *testing.T) {
	var pri Priority
	if err := (&pri).Set("debug"); err != nil || pri != Debug {
		t.Errorf("Set failed: %s, %v", pri, err)
	}
	err := pri.Set("fatal")
	confirmError(t, err, ErrInvalidPriority, "invalid priority: fatal")
}
