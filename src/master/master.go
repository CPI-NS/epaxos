package main

import (
	"flag"
	"fmt"
	"github.com/efficient/epaxos/src/genericsmrproto"
	"github.com/efficient/epaxos/src/dlog"
	"log"
	"github.com/efficient/epaxos/src/masterproto"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

var portnum *int = flag.Int("port", 7087, "Port # to listen on. Defaults to 7087")
var numNodes *int = flag.Int("N", 3, "Number of replicas. Defaults to 3.")
var timeout *int = flag.Int("timeout", 2, "Timeout in seconds for nodes responding to ping" )

type Master struct {
	N        int
	nodeList []string
	addrList []string
	portList []int
	lock     *sync.Mutex
	nodes    []*rpc.Client
	leader   []bool
	alive    []bool
}

func main() {
	flag.Parse()

	log.Printf("Master starting on port %d\n", *portnum)
	log.Printf("...waiting for %d replicas\n", *numNodes)

	master := &Master{*numNodes,
		make([]string, 0, *numNodes),
		make([]string, 0, *numNodes),
		make([]int, 0, *numNodes),
		new(sync.Mutex),
		make([]*rpc.Client, *numNodes),
		make([]bool, *numNodes),
		make([]bool, *numNodes)}

	rpc.Register(master)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *portnum))
  fmt.Println("Master got connection")
	if err != nil {
		log.Fatal("Master listen error:", err)
	}

	go master.run()

	http.Serve(l, nil)
}

func (master *Master) run() {
	for true {
		master.lock.Lock()
		if len(master.nodeList) == master.N {
			master.lock.Unlock()
			break
		}
		master.lock.Unlock()
		time.Sleep(100000000)
	}
	time.Sleep(5000000000)

  // TODO: How does it get addr list?
	// connect to SMR servers
	for i := 0; i < master.N; i++ {
		var err error
		addr := fmt.Sprintf("%s:%d", master.addrList[i], master.portList[i]+1000)
		master.nodes[i], err = rpc.DialHTTP("tcp", addr)
		if err != nil {
      //log.Fatalf("Error connecting to replica %d at addr %s with error: %s\n", i, addr, err)
      fmt.Printf("Error connecting to replica %d at addr %s with error: %s\n", i, addr, err)
		} else {
      log.Println("Connected to replica ", i, "Addr: ", master.addrList[i], " Port: ", master.portList[i]+1000)
    }
    master.leader[i] = false
	}
  log.Println("Connected to all replicas")
	master.leader[0] = true
  log.Printf("Replica 0 is the leader")

	for true {
		time.Sleep(3000 * 1000 * 1000)
		new_leader := false
		for i, node := range master.nodes {
      pingChannel := make(chan int) 
      errorChannel := make(chan error)

      go func() {
        err := node.Call("Replica.Ping", new(genericsmrproto.PingArgs), new(genericsmrproto.PingReply))
        if err != nil {
          log.Printf("Replica %d has failed to reply to ping\n", i)
          errorChannel <- err
        } else {
          pingChannel <- 0
        }
      }()

      select {
      case <-pingChannel:
        dlog.Printf("Ping worked")
				master.alive[i] = true
      case <- time.After(time.Duration(*timeout)*time.Second):
        log.Printf("Timeout")
        new_leader = checkLeader(i, master)
      case <- errorChannel:
        log.Printf("Error from call")
        new_leader = checkLeader(i, master)
      }
		}
		if !new_leader {
      dlog.Printf("continuing")
			continue
		}
    log.Printf("Choosing new leader")
		for i, new_master := range master.nodes {
			if master.alive[i] {
				err := new_master.Call("Replica.BeTheLeader", new(genericsmrproto.BeTheLeaderArgs), new(genericsmrproto.BeTheLeaderReply))
				if err == nil {
					master.leader[i] = true
					log.Printf("Replica %d is the new leader.", i)
					break
				}
			}
		}
	}
}

func checkLeader(i int, master *Master) bool {
  log.Printf("checking leader")
  master.alive[i] = false
  new_leader := false
  if master.leader[i] {
    // neet to choose a new leader
    new_leader = true
    master.leader[i] = false
    log.Printf("Need to choose a new leader")
  }
  return new_leader
}

func (master *Master) Register(args *masterproto.RegisterArgs, reply *masterproto.RegisterReply) error {

	master.lock.Lock()
	defer master.lock.Unlock()

	nlen := len(master.nodeList)
	index := nlen

	addrPort := fmt.Sprintf("%s:%d", args.Addr, args.Port)

	for i, ap := range master.nodeList {
		if addrPort == ap {
			index = i
			break
		}
	}

	if index == nlen {
		master.nodeList = master.nodeList[0 : nlen+1]
		master.nodeList[nlen] = addrPort
		master.addrList = master.addrList[0 : nlen+1]
		master.addrList[nlen] = args.Addr
		master.portList = master.portList[0 : nlen+1]
		master.portList[nlen] = args.Port
		nlen++
	}

	if nlen == master.N {
		reply.Ready = true
		reply.ReplicaId = index
		reply.NodeList = master.nodeList
	} else {
		reply.Ready = false
	}

	return nil
}

func (master *Master) GetLeader(args *masterproto.GetLeaderArgs, reply *masterproto.GetLeaderReply) error {
	time.Sleep(4 * 1000 * 1000)
	for i, l := range master.leader {
		if l {
			*reply = masterproto.GetLeaderReply{i}
			break
		}
	}
	return nil
}

func (master *Master) GetReplicaList(args *masterproto.GetReplicaListArgs, reply *masterproto.GetReplicaListReply) error {
	master.lock.Lock()
	defer master.lock.Unlock()

	if len(master.nodeList) == master.N {
		reply.ReplicaList = master.nodeList
		reply.Ready = true
	} else {
		reply.Ready = false
	}
	return nil
}
