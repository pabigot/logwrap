// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

package logwrap

import (
	"log"
	"testing"
)

func TestLogLogger(t *testing.T) {
	lgr := LogLogMaker(nil)
	lgr.F(Emerg, "made it this far")

	wrapped, ok := lgr.(*LogLogger)
	if !ok {
		t.Errorf("failed to get wrapped implementation")
	}
	inst := wrapped.Instance()

	if v := inst.Flags(); v != log.LstdFlags {
		t.Errorf("Init flags %x not %x", v, log.LstdFlags)
	}

	lgr.SetId("TestLogLogger ")
	if v := inst.Flags(); v != log.LstdFlags|log.Lmsgprefix {
		t.Errorf("SetId did not enable Lmsgprefix: %x", v)
	}

	lgr.F(Warning, "with prefix")

	if p := lgr.Priority(); p != Warning {
		t.Errorf("unexpected init priority: %d", int(p))
	}
	lgr.F(Debug, "debug at Warning priority: you don't see this, right?")

	// The null logger retains its configured priority even though it
	// isn't used, for consistent behavior with other loggers.
	if p := lgr.SetPriority(Debug).Priority(); p != Debug {
		t.Errorf("failed to set priority: %d", int(p))
	}
	lgr.F(Debug, "debug at debug priority: you see this, right?")
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
		testCase{Emerg, []string{Emerg.String(), "EmeRgenCY", "emerg"}},
		testCase{Crit, []string{Crit.String(), "critical", "CRIT"}},
		testCase{Error, []string{Error.String(), "error"}},
		testCase{Warning, []string{Warning.String(), "wARN", "Warning"}},
		testCase{Notice, []string{Notice.String(), "Notice"}},
		testCase{Info, []string{Info.String(), "info"}},
		testCase{Debug, []string{Debug.String(), "DeBug"}},
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
