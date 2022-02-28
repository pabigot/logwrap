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

const (
	svcPri    = lw.Info
	subSvcPri = lw.Debug
)

type Service struct {
	id  string
	lgr lw.Logger
	sb  strings.Builder
	ss  *SubService
	ch  chan *SubService
	wg  sync.WaitGroup
}

type SubService struct {
	id  string
	sb  strings.Builder
	lgr lw.Logger
}

func logMaker(owner interface{}) lw.Logger {
	lgr := lw.LogLogMaker(owner)
	ll := lgr.(*lw.LogLogger).Instance()
	ll.SetFlags(ll.Flags() & ^(log.Ldate | log.Ltime))
	switch v := owner.(type) {
	case *Service:
		lgr.SetPriority(svcPri)
		lgr.SetId(v.id)
		ll.SetOutput(&v.sb)
	case *SubService:
		lgr.SetPriority(subSvcPri)
		lgr.SetId(v.id)
		ll.SetOutput(&v.sb)
	default:
		lgr.SetPriority(lw.Notice)
	}
	// TEST: mask variable output
	return lgr
}

func NewService(id string, newLog lw.LogMaker) *Service {
	rv := &Service{
		id: id,
		ss: newSubService(id+".sub", newLog),
		ch: make(chan *SubService),
	}
	rv.lgr = newLog(rv)
	rv.lgr.F(lw.Notice, "Constructed")
	rv.wg.Add(1)
	go rv.main()
	return rv
}

func newSubService(id string, newLog lw.LogMaker) *SubService {
	rv := &SubService{
		id: id,
	}
	rv.lgr = newLog(rv)
	rv.lgr.F(lw.Notice, "Constructed")
	return rv
}

func (s *Service) main() {
	lprn := lw.MakePriWrapper(s.lgr, lw.Notice)
	lpri := lw.MakePriWrapper(s.lgr, lw.Info)
	lprd := lw.MakePriWrapper(s.lgr, lw.Debug)
	lprn("Starting subservice %s", s.ss.id)
	s.wg.Add(1)
	go s.ss.main(s)
	lpri("Started, waiting for sync")
	ss := <-s.ch
	lpri("Sync from %s\n", ss.id)
	lprd("Debug")
	lprn("Gone")
	s.wg.Done()
	close(s.ch)
}

func (ss *SubService) main(s *Service) {
	lprn := lw.MakePriWrapper(ss.lgr, lw.Notice)
	lpri := lw.MakePriWrapper(ss.lgr, lw.Info)
	lprd := lw.MakePriWrapper(ss.lgr, lw.Debug)
	lprn("Started subservice")
	s.ch <- ss
	<-s.ch
	lpri("Notified service")
	lprd("Debug")
	lprn("Gone")
	s.wg.Done()
}

func ExampleLogMaker() {
	s := NewService("S1", logMaker)
	s.wg.Wait()
	fmt.Printf("Service:\n%s", s.sb.String())
	fmt.Printf("Subservice:\n%s", s.ss.sb.String())
	// Output: Service:
	// S1[N] Constructed
	// S1[N] Starting subservice S1.sub
	// S1[I] Started, waiting for sync
	// S1[I] Sync from S1.sub
	// S1[N] Gone
	// Subservice:
	// S1.sub[N] Constructed
	// S1.sub[N] Started subservice
	// S1.sub[I] Notified service
	// S1.sub[D] Debug
	// S1.sub[N] Gone
}
