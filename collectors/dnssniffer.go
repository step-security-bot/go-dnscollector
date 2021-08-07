package collectors

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/dmachard/go-dnscollector/dnsutils"
	"github.com/dmachard/go-dnscollector/processors"
	"github.com/dmachard/go-logger"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

// Convert a uint16 to host byte order (big endian)
func Htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}

func GetBpfFilter(port int) []bpf.Instruction {
	// bpf filter: (ip  or ip6 ) and (udp or tcp) and port 53
	// fragmented packets are ignored
	var filter = []bpf.Instruction{
		// Load eth.type (2 bytes at offset 12) and push-it in register A
		bpf.LoadAbsolute{Off: 12, Size: 2},
		// if eth.type == IPv4 continue with the next instruction
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x0800, SkipTrue: 0, SkipFalse: 10},
		// Load ip.proto (1 byte at offset 23) and push-it in register A
		bpf.LoadAbsolute{Off: 23, Size: 1},
		// ip.proto == UDP ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x11, SkipTrue: 1, SkipFalse: 0},
		// ip.proto == TCP ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x6, SkipTrue: 0, SkipFalse: 16},
		// load flags and fragment offset (2 bytes at offset 20) to ignore fragmented packet
		bpf.LoadAbsolute{Off: 20, Size: 2},
		// Only look at the last 13 bits of the data saved in regiter A
		//  0x1fff == 0001 1111 1111 1111 (fragment offset)
		// If any of the data in fragment offset is true, ignore the packet
		bpf.JumpIf{Cond: bpf.JumpBitsSet, Val: 0x1fff, SkipTrue: 14, SkipFalse: 0},
		// Load ip.length
		// Register X = ip header len * 4
		bpf.LoadMemShift{Off: 14},
		// Load source port in tcp or udp (2 bytes at offset x+14)
		bpf.LoadIndirect{Off: 14, Size: 2},
		// source port equal to 53 ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(port), SkipTrue: 10, SkipFalse: 0},
		// Load estination port in tcp or udp  (2 bytes at offset x+16)
		bpf.LoadIndirect{Off: 16, Size: 2},
		// destination port equal to 53 ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(port), SkipTrue: 8, SkipFalse: 9},
		// if eth.type == IPv6 continue with the next instruction
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x86dd, SkipTrue: 0, SkipFalse: 8},
		// Load ipv6.nxt (2 bytes at offset 12) and push-it in register A
		bpf.LoadAbsolute{Off: 20, Size: 1},
		// ip.proto == UDP ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x11, SkipTrue: 1, SkipFalse: 0},
		// ip.proto == TCP ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x6, SkipTrue: 0, SkipFalse: 5},
		// Load source port tcp or udp (2 bytes at offset 54)
		bpf.LoadAbsolute{Off: 54, Size: 2},
		// source port equal to 53 ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(port), SkipTrue: 2, SkipFalse: 0},
		// Load destination port tcp or udp (2 bytes at offset 56)
		bpf.LoadAbsolute{Off: 56, Size: 2},
		// destination port equal to 53 ?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(port), SkipTrue: 0, SkipFalse: 1},
		// Keep the packet and send up to 65k of the packet to userspace
		bpf.RetConstant{Val: 0xFFFF},
		// Ignore packet
		bpf.RetConstant{Val: 0},
	}
	return filter
}

func ApplyBpfFilter(filter []bpf.Instruction, fd int) (err error) {
	var assembled []bpf.RawInstruction
	if assembled, err = bpf.Assemble(filter); err != nil {
		return err
	}

	prog := &unix.SockFprog{
		Len:    uint16(len(assembled)),
		Filter: (*unix.SockFilter)(unsafe.Pointer(&assembled[0])),
	}

	return unix.SetsockoptSockFprog(fd, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, prog)
}

func RemoveBpfFilter(fd int) (err error) {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_DETACH_FILTER, 0)
}

type DnsSniffer struct {
	done       chan bool
	exit       chan bool
	device     string
	port       int
	identity   string
	generators []dnsutils.Worker
	config     *dnsutils.Config
	logger     *logger.Logger
}

func NewDnsSniffer(generators []dnsutils.Worker, config *dnsutils.Config, logger *logger.Logger) *DnsSniffer {
	logger.Info("collector dns sniffer - enabled")
	s := &DnsSniffer{
		done:       make(chan bool),
		exit:       make(chan bool),
		config:     config,
		generators: generators,
		logger:     logger,
	}
	s.ReadConfig()
	return s
}

func (c *DnsSniffer) LogInfo(msg string, v ...interface{}) {
	c.logger.Info("collector dns sniffer - "+msg, v...)
}

func (c *DnsSniffer) LogError(msg string, v ...interface{}) {
	c.logger.Error("collector dns sniffer - "+msg, v...)
}

func (c *DnsSniffer) Generators() []chan dnsutils.DnsMessage {
	channels := []chan dnsutils.DnsMessage{}
	for _, p := range c.generators {
		channels = append(channels, p.Channel())
	}
	return channels
}
func (c *DnsSniffer) ReadConfig() {
	c.device = c.config.Collectors.DnsSniffer.Device
	c.port = c.config.Collectors.DnsSniffer.Port
	c.identity = c.config.Collectors.DnsSniffer.Identity
}

func (c *DnsSniffer) Channel() chan dnsutils.DnsMessage {
	return nil
}

func (c *DnsSniffer) Stop() {
	c.LogInfo("stopping...")

	// exit to close properly
	c.exit <- true

	// read done channel and block until run is terminated
	<-c.done
	close(c.done)
}

func (c *DnsSniffer) Run() {
	dns_processor := processors.NewDnsProcessor(c.logger)
	go dns_processor.Run(c.Generators())

	// raw socket
	sd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, Htons(syscall.ETH_P_ALL))
	if err != nil {
		panic(err)
	}
	defer syscall.Close(sd)

	// set nano timestamp
	err = syscall.SetsockoptInt(sd, syscall.SOL_SOCKET, syscall.SO_TIMESTAMPNS, 1)
	if err != nil {
		panic(err)
	}

	filter := GetBpfFilter(c.port)
	err = ApplyBpfFilter(filter, sd)
	if err != nil {
		panic(err)
	}
	defer RemoveBpfFilter(sd)

	go func() {
		buf := make([]byte, 65536)
		oob := make([]byte, 100)
		for {
			//flags, from
			bufN, oobn, _, _, err := syscall.Recvmsg(sd, buf, oob, 0)
			if err != nil {
				panic(err)
			}
			if bufN == 0 {
				panic("buf empty")
			}
			if bufN > len(buf) {
				panic("buf overflow")
			}
			if oobn == 0 {
				panic("oob missing")
			}

			scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
			if err != nil {
				panic(err)
			}
			if len(scms) != 1 {
				continue
			}
			scm := scms[0]
			if scm.Header.Type != syscall.SCM_TIMESTAMPNS {
				panic("scm timestampns missing")
			}
			tsec := binary.LittleEndian.Uint32(scm.Data[:4])
			nsec := binary.LittleEndian.Uint32(scm.Data[8:12])

			var eth layers.Ethernet
			var ip4 layers.IPv4
			var ip6 layers.IPv6
			var tcp layers.TCP
			var udp layers.UDP
			parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp)
			decodedLayers := make([]gopacket.LayerType, 0, 10)

			// copy packet data from buffer
			pkt := make([]byte, bufN)
			copy(pkt, buf[:bufN])

			// decode-it
			parser.DecodeLayers(pkt, &decodedLayers)

			dm := dnsutils.DnsMessage{}
			dm.Init()

			for _, layertyp := range decodedLayers {
				switch layertyp {
				case layers.LayerTypeIPv4:
					dm.Family = "INET"
					dm.QueryIp = ip4.SrcIP.String()
					dm.ResponseIp = ip4.DstIP.String()
				case layers.LayerTypeIPv6:
					dm.QueryIp = ip6.SrcIP.String()
					dm.ResponseIp = ip6.DstIP.String()
					dm.Family = "INET6"
					fmt.Println(eth)
				case layers.LayerTypeUDP:
					dm.QueryPort = fmt.Sprint(int(udp.SrcPort))
					dm.ResponsePort = fmt.Sprint(int(udp.DstPort))
					dm.Payload = udp.Payload
					dm.Length = len(udp.Payload)
					dm.Protocol = "UDP"
				case layers.LayerTypeTCP:
					dm.QueryPort = fmt.Sprint(int(tcp.SrcPort))
					dm.ResponsePort = fmt.Sprint(int(tcp.DstPort))
					dm.Payload = tcp.Payload
					dm.Length = len(tcp.Payload)
					dm.Protocol = "TCP"
				}
			}

			dm.Identity = c.identity

			// set timestamp
			dm.TimeSec = int(tsec)
			dm.TimeNsec = int(nsec)

			dns_processor.GetChannel() <- dm

		}
	}()

	<-c.exit

	// stop dns processor
	dns_processor.Stop()

	c.LogInfo("run terminated")
	c.done <- true
}
