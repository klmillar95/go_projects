package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"time"

	qt "querytool"
)

//receiver structure with slice of jobs
type Receiver struct {
	jobs []qt.Job
}

//allows the feeder to send a job over to the receiver
func (r *Receiver) SendJob(job *qt.Job, reply *bool) error {
	r.jobs = append(r.jobs,*job)
	*reply = true
	return nil
}

//loop to check for new jobs in the receiver and process them before discarding
func (r *Receiver) run() {
	for {
		if len(r.jobs) > 0 {
			for _,job := range r.jobs {
				fmt.Println(job)
			}
			r.jobs = []qt.Job{}
		}
		time.Sleep(100)
	}
}

func main(){
	fmt.Println("STARTING UP THE RECEIVER... ")
	receiver := new(Receiver)

	rpc.Register(receiver)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp",":8675")
	if err != nil {
		panic(err)
	}

	go http.Serve(l, nil)
	receiver.run()
	time.Sleep(10000)
}
