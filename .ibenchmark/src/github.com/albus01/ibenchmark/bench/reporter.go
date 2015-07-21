/*
   Copyright 2015 Albus <albus@shaheng.me>.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package ibench

import (
	"fmt"
	"time"
)

type Reporter struct {
	Server              string
	Hostname            string
	Port                string
	Path                string
	Headers             string
	ContentLength       int64
	Concurrency         int
	TimeTaken           int64
	TimeDur             int64
	TotalRequest        int
	FailedRequest       int
	RequestPerSecond    int
	ConnectionPerSecond int
	Non2XXCode          int
}

func (r *Reporter) Print() {
	var avgT int64
	if r.TotalRequest == 0 {
		avgT = 0
	} else {
		avgT = r.TimeTaken / 1000 / int64(r.TotalRequest)
	}
	report := fmt.Sprintf("Server Software:%s\nServer Hostname:%s\nServer Port:%s\n\nRequest Headers:\n%s\n\nDocument Path:%s\nDocument Length:%d\n\nConcurrency:%d\nTime Duration:%dms\nAvg Time Taken:%dms\n\nComplete Requests:%d\nFailed Request:%d\n\nRequest Per Second:%d\nConnections Per Second:%d\n\nNon2XXCode:%d\n\n", r.Server, r.Hostname, r.Port, r.Headers, r.Path, r.ContentLength, r.Concurrency, r.TimeDur, avgT, r.TotalRequest, r.FailedRequest, r.RequestPerSecond, r.ConnectionPerSecond, r.Non2XXCode)
	fmt.Println(report)
}

func (r *Reporter) report(dur int) {
	if dur != 0 {
		report := fmt.Sprintf("\nFinished Request Numbers:%d\nFailed Request Numbers:%d\nTime Consume:%d\nDone Per Second:%d\n", r.TotalRequest, r.FailedRequest, dur, r.TotalRequest/dur)
		fmt.Println(report)
	}
}

func (r *Reporter) Reporter() {
	report := time.Tick(6 * time.Second)
	dur := 6
	for {
		select {
		case <-report:
			r.report(dur)
			dur += 6
		default:
			fmt.Print(".")
			time.Sleep(1000 * time.Millisecond)
		}
	}
}
