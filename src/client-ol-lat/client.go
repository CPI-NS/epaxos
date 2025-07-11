package main

import (
	"bufio"
//	"github.com/efficient/epaxos/src/dlog"
	"flag"
	"fmt"
	"github.com/efficient/epaxos/src/genericsmrproto"
	"log"
	"github.com/efficient/epaxos/src/masterproto"
	"math/rand"
	"net"
	"net/rpc"
	"runtime"
	"github.com/efficient/epaxos/src/state"
	"time"
  "os"
  "os/signal"
  "syscall"
)

var masterAddr *string = flag.String("maddr", "", "Master address. Defaults to localhost")
var masterPort *int = flag.Int("mport", 7087, "Master port.  Defaults to 7077.")
var reqsNb *int = flag.Int("q", 5000, "Total number of requests. Defaults to 5000.")
var noLeader *bool = flag.Bool("e", false, "Egalitarian (no leader). Defaults to false.")
var fast *bool = flag.Bool("f", false, "Fast Paxos: send message directly to all replicas. Defaults to false.")
var rounds *int = flag.Int("r", 1, "Split the total number of requests into this many rounds, and do rounds sequentially. Defaults to 1.")
var procs *int = flag.Int("p", 2, "GOMAXPROCS. Defaults to 2")
var check = flag.Bool("check", false, "Check that every expected reply was received exactly once.")
var eps *int = flag.Int("eps", 0, "Send eps more messages per round than the client will wait for (to discount stragglers). Defaults to 0.")
var conflicts *int = flag.Int("c", -1, "Percentage of conflicts. Defaults to 0%")
var s = flag.Float64("s", 2, "Zipfian s parameter")
var v = flag.Float64("v", 1, "Zipfian v parameter")
var nanosleep = flag.Int("ns", 0, "Amount of time (in ns) to sleep between two successive commands.")
var batch = flag.Int("batch", 100, "Commands to send before flush (and sleep).")
var timeout = flag.Int("t", 10, "How long to send requests for in seconds")

var N int

var successful []int

var rarray []int
var rsp []bool

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	randObj := rand.New(rand.NewSource(42))
	zipf := rand.NewZipf(randObj, *s, *v, uint64(*reqsNb / *rounds + *eps))

	if *conflicts > 100 {
		log.Fatalf("Conflicts percentage must be between 0 and 100.\n")
	}

	master, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", *masterAddr, *masterPort))
	if err != nil {
		log.Fatalf("Error connecting to master\n")
	}

	rlReply := new(masterproto.GetReplicaListReply)
	err = master.Call("Master.GetReplicaList", new(masterproto.GetReplicaListArgs), rlReply)
	if err != nil {
		log.Fatalf("Error making the GetReplicaList RPC")
	}

	N = len(rlReply.ReplicaList)
	servers := make([]net.Conn, N)
	readers := make([]*bufio.Reader, N)
	writers := make([]*bufio.Writer, N)

	rarray = make([]int, *reqsNb / *rounds + *eps)
	karray := make([]int64, *reqsNb / *rounds + *eps)
	perReplicaCount := make([]int, N)
	test := make([]int, *reqsNb / *rounds + *eps)
	for i := 0; i < len(rarray); i++ {
		r := rand.Intn(N)
		rarray[i] = r
		if i < *reqsNb / *rounds {
			perReplicaCount[r]++
		}

		if *conflicts >= 0 {
			r = rand.Intn(100)
			if r < *conflicts {
				karray[i] = 42
			} else {
				karray[i] = int64(43 + i)
			}
		} else {
			karray[i] = int64(zipf.Uint64())
			test[karray[i]]++
		}
	}
	if *conflicts >= 0 {
		fmt.Println("Uniform distribution")
	} else {
		fmt.Println("Zipfian distribution:")
		//fmt.Println(test[0:100])
	}

	for i := 0; i < N; i++ {
		var err error
		servers[i], err = net.Dial("tcp", rlReply.ReplicaList[i])
		if err != nil {
			log.Printf("Error connecting to replica %d\n", i)
		}
		readers[i] = bufio.NewReader(servers[i])
		writers[i] = bufio.NewWriter(servers[i])
	}

	successful = make([]int, N)
	leader := 0

	if *noLeader == false {
		reply := new(masterproto.GetLeaderReply)
		if err = master.Call("Master.GetLeader", new(masterproto.GetLeaderArgs), reply); err != nil {
			log.Fatalf("Error making the GetLeader RPC\n")
		}
		leader = reply.LeaderId
		log.Printf("The leader is replica %d\n", leader)
	}

	var id int32 = 0
	done := make(chan bool, N)
	args := genericsmrproto.Propose{id, state.Command{state.PUT, 0, 0}, 0}

	before_total := time.Now()

  reqSent := 0
	for j := 0; j < *rounds; j++ {

		n := *reqsNb / *rounds

		if *check {
			rsp = make([]bool, n)
			for j := 0; j < n; j++ {
				rsp[j] = false
			}
		}

		donePrinting := make(chan bool)
		readings := make(chan int64, n)
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go printer(readings, donePrinting)

    reqDone := make(chan bool)
		if *noLeader {
			for i := 0; i < N; i++ {
				//go waitReplies(readers, reqSent, perReplicaCount[reqSent], done, readings)
			}
		} else {
			//go  waitReplies(readers, leader, n + 100000, done, readings)
			go  waitReplies(readers, leader, reqDone, done, readings)
			//go  waitReplies(readers, leader, reqSent / 100, done, readings)
		}

		before := time.Now()

    expTimeout := make(chan bool, 1)
    go func() {
      time.Sleep(time.Duration(*timeout) * time.Second)
      expTimeout <- true
    }()

		// for i := 0; i < n+*eps; i++ {
    i := 0
RequestLoop:
    for {
      select {
      case <- expTimeout:
        reqDone <- true
        break RequestLoop
      case <-sigs:
        log.Printf("Termination signal recieved!")
        reqDone <- true
        break RequestLoop
      default:
        if id % 100 == 0 {
          log.Printf("Sending proposal %d\n", id)
        }
        args.CommandId = id
        args.Command.K = state.Key(i)
        args.Command.V = state.Value(i)
        if !*fast {
          if *noLeader {
            leader = rarray[i]
          }
          writers[leader].WriteByte(genericsmrproto.PROPOSE)
          args.Marshal(writers[leader])
         // if id % 100 == 0 {
          //  log.Printf("Sent Proposal %d\n", id)
          ///}
        } else {
          //send to everyone
          for rep := 0; rep < N; rep++ {
            writers[rep].WriteByte(genericsmrproto.PROPOSE)
            args.Marshal(writers[rep])
            writers[rep].Flush()
          }
        }
        id++
        //if i%*batch == 0 {
        //  for i := 0; i < N; i++ {
        //    writers[i].Flush()
        //  }
        //  if *nanosleep > 0 {
        //    time.Sleep(time.Duration(*nanosleep))
        //  }
        //}
        i++
        //if id % 100 == 0 {
        //  log.Printf("Flushing writers")
        //}
        for i := 0; i < N; i++ {
          err := writers[i].Flush()
         // log.Printf("Checking error")
          if err != nil {
            fmt.Println(err)
          }
        }
        //if id % 100 == 0 {
        //  log.Printf("Flushed writers")
        //}
      }
		}
    // Just because i is overloaded and might get confusing
    reqSent += i
    fmt.Println("After sending requests")

		for i := 0; i < N; i++ {
			writers[i].Flush()
		}
    fmt.Println("After flushing writers")

    //reqSent -= 10
		//if *noLeader {
		//	for i := 0; i < N; i++ {
		//		go waitReplies(readers, reqSent, perReplicaCount[reqSent], done, readings)
		//	}
		//} else {
		//	go  waitReplies(readers, leader, n + 100000, done, readings)
		//	//go  waitReplies(readers, leader, reqSent / 100, done, readings)
		//}

		err := false
		if *noLeader {
			for i := 0; i < N; i++ {
				e := <-done
				err = e || err
			}
		} else {
      // Done is written by get replies
      // It is waiting for replies to come in because it thinks there is only 5k req?
      fmt.Println("Before Error waiting")
			err = <-done
		}
    fmt.Println("After Error Waiting")

		after := time.Now()

		<-donePrinting

		fmt.Printf("Round took %v\n", after.Sub(before))

		if *check {
			for j := 0; j < n; j++ {
				if !rsp[j] {
					fmt.Println("Didn't receive", j)
				}
			}
		}

		if err {
			if *noLeader {
				N = N - 1
			} else {
				reply := new(masterproto.GetLeaderReply)
				master.Call("Master.GetLeader", new(masterproto.GetLeaderArgs), reply)
				leader = reply.LeaderId
				log.Printf("New leader is replica %d\n", leader)
			}
		}
	}

	after_total := time.Now()
	fmt.Printf("Test took %v\n", after_total.Sub(before_total))
  fmt.Printf("Sent requests  %v\n", reqSent)
  fmt.Printf("Throughput: %v\n", int64(reqSent)/after_total.Sub(before_total).Milliseconds())

	s := 0
	for _, succ := range successful {
		s += succ
	}

	fmt.Printf("Successful: %d\n", s)

	for _, client := range servers {
		if client != nil {
			client.Close()
		}
	}
	master.Close()
}

//func waitReplies(readers []*bufio.Reader, leader int, n int, done chan bool, readings chan int64) {
func waitReplies(readers []*bufio.Reader, leader int, reqDone chan bool, done chan bool, readings chan int64) {
	e := false

	//tss := make([]int64, n)

	reply := new(genericsmrproto.ProposeReplyTS)

  EndResp:
	for i := 0; ; i++ {
    select{
    case huh := <- reqDone:
      if huh {
        break EndResp
      }
    }
		/*if *noLeader {
		    leader = rarray[i]
		}*/
    // TODO: Eof when trying to unmarshall reply?
    //fmt.Println("Before ummarshall")
		if err := reply.Unmarshal(readers[leader]); err != nil {
			fmt.Println("Error when reading:", err)
			//e = true
			continue
		}
    //fmt.Println("After ummarshall")

	//	tss[i] = time.Now().UnixNano() - reply.Timestamp

		//if *check {
		//	if rsp[reply.Instance] {
		//		fmt.Println("Duplicate reply", reply.Instance)
		//	}
		//	rsp[reply.Instance] = true
		//}
		if reply.OK != 0 {
      //fmt.Println("Successfull ??? req ", i)
			successful[leader]++
		}
	}
  fmt.Printf("Done\n")
	done <- e

	//for i := 0; i < n; i++ {
	//	readings <- tss[i]
	//}
}

func printer(readings chan int64, done chan bool) {
	n := *reqsNb
	for i := 0; i < n; i++ {
		//lat := <-readings
		//fmt.Printf("%d\n", lat/1000)
	}
	done <- true
}
