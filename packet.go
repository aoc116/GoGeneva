package main

import (
	"context"
	"fmt"
	"github.com/florianl/go-nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)


func handlePacket(a nfqueue.Attribute, nf *nfqueue.Nfqueue, packet gopacket.Packet) {
	id := *a.PacketID
	var srcIP, dstIP net.IP

	packet = gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.Default)

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		srcIP = ip.SrcIP
		dstIP = ip.DstIP
		log.Printf("%15s > %-15s \n", srcIP, dstIP)
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		ports := strings.Split(Port, ",")
		sport := strconv.Itoa(int(tcp.SrcPort))
		var matchedPort bool = false

		for _, port := range ports {
			if port == sport {
				matchedPort = true
				break
			}
		}

		if matchedPort {
			var ok1 bool = SaEnable && tcp.SYN && tcp.ACK
			var ok2 bool = AEnable && tcp.ACK && !tcp.PSH && !tcp.FIN && !tcp.SYN && !tcp.RST
			var ok3 bool = PaEnable && tcp.PSH && tcp.ACK
			var ok4 bool = FaEnable && tcp.FIN && tcp.ACK
			var windowSize uint16

			if ok1 || ok2 || ok3 || ok4 {
				if ok1 {
					windowSize = WindowSa
					log.Println("Handle SYN=1 and ACK=1")
				}
				if ok2 {
					windowSize = WindowA
					log.Println("Handle ACK=1")
				}
				if ok3 {
					windowSize = WindowPa
					log.Println("Handle PSH=1 and ACK=1")
				}
				if ok4 {
					windowSize = WindowFa
					log.Println("Handle FIN=1 and ACK=1")
				}
				packet.TransportLayer().(*layers.TCP).Window = windowSize
				err := packet.TransportLayer().(*layers.TCP).SetNetworkLayerForChecksum(packet.NetworkLayer())
				if err != nil {
					log.Fatalf("SetNetworkLayerForChecksum error: %v", err)
				}
				buffer := gopacket.NewSerializeBuffer()
				options := gopacket.SerializeOptions{
					ComputeChecksums: true,
					FixLengths:       true,
				}
				if err := gopacket.SerializePacket(buffer, options, packet); err != nil {
					log.Fatalf("SerializePacket error: %v", err)
				}
				packetBytes := buffer.Bytes()
				log.Printf("Set TCP window size to %d", windowSize)
				err = nf.SetVerdictModPacket(id, nfqueue.NfAccept, packetBytes)
				if err != nil {
					log.Fatalf("SetVerdictModified error: %v", err)
				}
				return
			}
		}
	}

	err := nf.SetVerdict(id, nfqueue.NfAccept)
	if err != nil {
		log.Fatalf("SetVerdictModified error: %v", err)
	}
}

func packetHandle(queueNum int) {
	config := nfqueue.Config{
		NfQueue:      uint16(queueNum),
		MaxPacketLen: 0xFFFF,
		MaxQueueLen:  0xFF,
		Copymode:     nfqueue.NfQnlCopyPacket,
		WriteTimeout: time.Duration(timeout) * time.Millisecond,
	}

	nf, err := nfqueue.Open(&config)
	if err != nil {
		fmt.Println("could not open nfqueue socket:", err)
		return
	}
	defer nf.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	fn := func(a nfqueue.Attribute) int {
		go handlePacket(a, nf, gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.Default))
		return 0
	}

	err = nf.RegisterWithErrorFunc(ctx, fn, func(e error) int {
		if e != nil {
			log.Fatalln("RegisterWithErrorFunc Error:", e)
		}
		return 0
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	<-ctx.Done()
}