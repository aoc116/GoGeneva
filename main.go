package main

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

const (
	DAEMON  = "daemon"
	FOREVER = "forever"
)



func init() {
	if getProcessOwner() != "root\n" {
		log.Fatalln("Please run this program with root.")
	}

}


func startTask(poolName string, start int, end int, enable bool, wg *sync.WaitGroup) {
	if !enable {
		return
	}
	p, err := ants.NewPoolWithFunc(128, func(i interface{}) {
		packetHandle(i.(int))
		wg.Done()
	})
	if err != nil {
		log.Fatalf("Failed to create goroutine pool: %v", err)
	}
	defer p.Release()
	// Submit tasks one by one.
	log.Printf("Starting Task %s\n", poolName)
	for i := start; i < end; i++ {
		wg.Add(1)
		_ = p.Invoke(int(i))
	}
}

func main() {
	InitParams()

	if !Debug {
		log.SetOutput(ioutil.Discard)
	}
	if Daemon {
		SubProcess(StripSlice(os.Args, "-"+DAEMON))
		fmt.Printf("[*] Daemon running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
		os.Exit(0)
	} else if Forever {
		for {
			cmd := SubProcess(StripSlice(os.Args, "-"+FOREVER))
			fmt.Printf("[*] Forever running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
			cmd.Wait()
		}
		os.Exit(0)
	} else {
		UnsetIptable(Port)
		SetIptable(Port)
		var wg sync.WaitGroup
		startTask("p1", 1000, 1000+taskRange, SaEnable, &wg)
		startTask("p2", 2000, 2000+taskRange, AEnable, &wg)
		startTask("p3", 3000, 3000+taskRange, PaEnable, &wg)
		startTask("p4", 4000, 4000+taskRange, FaEnable, &wg)
		wg.Wait()
	}
}