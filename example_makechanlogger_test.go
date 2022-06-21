// Copyright 2021-2022 Peter Bigot Consulting, LLC
// SPDX-License-Identifier: Apache-2.0

package logwrap_test

import (
	"fmt"
	"log"
	"strings"
	"sync"

	lw "github.com/pabigot/logwrap"
)

type ChanService struct {
	id  string
	lgr lw.Logger
	sb  strings.Builder
	wg  sync.WaitGroup
}

func NewChanService(id string) *ChanService {
	s := &ChanService{
		id: id,
	}

	s.lgr = lw.LogLogMaker(&s.sb)
	s.lgr.SetPriority(lw.Debug)
	s.lgr.SetId(id)
	ll := s.lgr.(*lw.LogLogger).Instance()
	ll.SetFlags(ll.Flags() & ^(log.Ldate | log.Ltime))
	ll.SetOutput(&s.sb)

	s.lgr.F(lw.Notice, "Constructed")
	s.wg.Add(1)
	go s.main()
	return s
}

func (s *ChanService) main() {
	lpr := lw.MakePriPr(s.lgr)
	lpr.N("Entered")

	// Create a channel to allow the goroutine of this call to emit log
	// messages submitted by goroutines started in this call.
	chlgr, lch := lw.MakeChanLogger(s.lgr, 2)

	// 2-deep channel for receiving exit notifications
	ich := make(chan struct{}, 2)
	go subMain(&s.wg, ich, lw.MakePriPr(lw.PrefixedChanLogger(chlgr, "s1: ")))
	lpr.I("Started s1")
	go subMain(&s.wg, ich, lw.MakePriPr(lw.PrefixedChanLogger(chlgr, "s2: ")))
	lpr.I("Started s2")

	loop := true
	cnt := 0
	for loop {
		select {
		case lm := <-lch:
			lm.Emit()
		case <-ich:
			cnt++
			lpr.I("%d signalled done", cnt)
			loop = cnt < 2
		}
	}

	// Flush out any remaining messages.
	loop = true
	for loop {
		select {
		case lm := <-lch:
			lm.Emit()
		default:
			loop = false
		}
	}

	lpr.N("Gone")
	s.wg.Done()
}

func subMain(wg *sync.WaitGroup, och chan<- struct{}, lpr lw.PriPr) {
	lpr.N("Started subservice")
	lpr.I("Notified service")
	lpr.D("Debug")
	lpr.N("Gone")
	och <- struct{}{}
}

func ExampleMakeChanLogger() {
	s := NewChanService("S1")
	s.wg.Wait()
	fmt.Printf("Service:\n%s", s.sb.String())

	// Unordered output: Service:
	// S1[N] Constructed
	// S1[N] Entered
	// S1[I] Started s1
	// S1[N] s1: Started subservice
	// S1[I] s1: Notified service
	// S1[D] s1: Debug
	// S1[N] s1: Gone
	// S1[I] 1 signalled done
	// S1[I] Started s2
	// S1[N] s2: Started subservice
	// S1[I] s2: Notified service
	// S1[D] s2: Debug
	// S1[N] s2: Gone
	// S1[I] 2 signalled done
	// S1[N] Gone
}
