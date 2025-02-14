package netlib

import (
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestNetDecoder_Decode_IPv4_UDP(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x08, 0x00,
		// ipv4
		0x45, 0x00, 0x00, 0x44, 0xe5, 0x6a, 0x00, 0x00, 0x6f, 0x11,
		0xec, 0x11, 0xac, 0xd9, 0x28, 0x4c, 0xc1, 0x18, 0xe3, 0xee,
		// udp
		0xdd, 0x68, 0x00, 0x35, 0x00, 0x30, 0x0c, 0x33,
		// udp payload (dns)
		0xd4, 0x3f, 0x00, 0x10, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x08, 0x77,
		0x65, 0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02, 0x64, 0x65, 0x00, 0x00, 0x30, 0x00,
		0x01, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv4); !ok {
		t.Errorf("Expected IPv4 layer, got %T", packetLayers[1])
	}
	ip4 := packetLayers[1].(*layers.IPv4)
	if ip4.Flags&layers.IPv4MoreFragments > 0 {
		t.Errorf("Expected more fragment")
	}
	if _, ok := packetLayers[2].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv4_TCP(t *testing.T) {
	pkt := []byte{
		//ethernet
		0xb0, 0xbb, 0xe5, 0xb2, 0x46, 0x4c, 0xb0, 0x35, 0x9f, 0xd4, 0x03, 0x91, 0x08, 0x00,
		// ipv4
		0x45, 0x00, 0x00, 0x69, 0xb7, 0x65, 0x40, 0x00, 0x40, 0x06, 0xbf,
		0x6e, 0xc0, 0xa8, 0x01, 0x11, 0x01, 0x01, 0x01, 0x01,
		// tcp
		0x8d, 0xcd, 0x00, 0x35, 0x39, 0x4f, 0x0c, 0xbb, 0xcf, 0x72, 0x32, 0xb3, 0x80, 0x18,
		0x01, 0xf6, 0x38, 0xc2, 0x00, 0x00, 0x01, 0x01, 0x08, 0x0a, 0x09, 0x5d, 0x2c, 0x7a, 0x65, 0xe0,
		0x63, 0x90, 0x00, 0x33, 0x85, 0x9f, 0x01, 0x20, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x06, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01,
		0x00, 0x00, 0x29, 0x04, 0xd0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x0a, 0x00, 0x08, 0xdf,
		0x41, 0x92, 0x72, 0x53, 0xf5, 0x1b, 0x48,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[2].(*layers.TCP); !ok {
		t.Errorf("Expected TCP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv4_MoreFragment(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x08, 0x00,
		// ipv4
		0x45, 0x00, 0x00, 0x44, 0xd0, 0xfe, 0x20, 0x00, 0x40, 0x11,
		0x09, 0xe6, 0xc1, 0x18, 0xe3, 0xee, 0xac, 0xd9, 0x28, 0x4c,
		// udp
		0x00, 0x35, 0xdd, 0x68, 0x06, 0xae, 0xb4, 0x63, 0xd4, 0x3f, 0x84, 0x10, 0x00, 0x01,
		0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x08, 0x77, 0x65, 0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02,
		0x64, 0x65, 0x00, 0x00, 0x30, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x30, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x3c, 0x02, 0x08, 0x01, 0x01, 0x03, 0x0a, 0x03, 0x01, 0x00, 0x01, 0xdd, 0xef, 0xfd, 0xed, 0x22,
		0xad, 0x76, 0x0a, 0x3b, 0x0b, 0x58, 0x10, 0x1d, 0xd5, 0x3d, 0xee, 0xf3, 0xf7, 0xda, 0xaf, 0x8b,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)
	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv4); !ok {
		t.Errorf("Expected IPv4 layer, got %T", packetLayers[1])
	}

	ip4 := packetLayers[1].(*layers.IPv4)
	if ip4.Flags&layers.IPv4MoreFragments != 1 {
		t.Errorf("Expected more fragment flag")
	}
	if _, ok := packetLayers[2].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv4_FragmentOffset(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x08, 0x00,
		// ipv4
		0x45, 0x00, 0x00, 0xfa, 0xd0, 0xfe, 0x00, 0xb9, 0x40, 0x11, 0x2e, 0x0f, 0xc1, 0x18, 0xe3, 0xee, 0xac, 0xd9,
		0x28, 0x4c,
		// udp
		0x92, 0x56, 0x69, 0x0f, 0x05, 0x4b, 0xdb, 0x48, 0x1e, 0x8f, 0xa8, 0x56, 0x36, 0x39,
		0xd5, 0xcc, 0xba, 0xf9, 0xf8, 0x22, 0x24, 0xd0, 0x76, 0xcc, 0x24, 0x9b, 0xda, 0x1d, 0x49, 0xf0,
		0x3e, 0x34, 0x44, 0x9c, 0x94, 0x65, 0x87, 0x34, 0x96, 0x0b, 0x8d, 0x1a, 0xb3, 0x33, 0xbe, 0x88,
		0x01, 0x62, 0x76, 0xf1, 0x22, 0x7b, 0x83, 0x28, 0x3d, 0x81, 0xf1, 0x21, 0x9a, 0xba, 0x6c, 0x6c,
		0xca, 0x72, 0x6e, 0x94, 0x14, 0x99, 0x4d, 0xd7, 0xbb, 0xe2, 0x49, 0xee, 0x72, 0x69, 0x3e, 0xee,
		0x0e, 0x03, 0x6c, 0xcd, 0x33, 0xc9, 0xf4, 0x43, 0xd1, 0x6d, 0xd1, 0x84, 0x3d, 0xee, 0xd0, 0xd1,
		0x5d, 0x8e, 0x2f, 0xf4, 0xce, 0x68, 0x88, 0xf3, 0x5e, 0xd5, 0x90, 0x21, 0x36, 0x1a, 0x95, 0x6f,
		0xb8, 0xbd, 0xc5, 0xf0, 0xa0, 0xc2, 0x0b, 0xe1, 0x0c, 0x62, 0x32, 0x65, 0x38, 0x7a, 0x8c, 0xf9,
		0x24, 0xc9, 0xc4, 0xfa, 0xbd, 0x64, 0x5f, 0x31, 0x25, 0xc5, 0x48, 0x4e, 0x40, 0xba, 0x11, 0x8e,
		0x82, 0x75, 0x19, 0x98, 0x99, 0x07, 0x6a, 0xbd, 0x16, 0x16, 0xcc, 0x35, 0xcf, 0x8c, 0x6b, 0x72,
		0xbb, 0x95, 0xd3, 0xd7, 0x71, 0xf5, 0x54, 0x2f, 0x08, 0x26, 0x2b, 0x0d, 0x51, 0xe8, 0x41, 0x0e,
		0xbd, 0x8f, 0x7a, 0x9a, 0x40, 0x35, 0x47, 0x57, 0x16, 0x5c, 0xaa, 0x55, 0x0e, 0xa6, 0x01, 0x12,
		0xfa, 0x52, 0x74, 0xc1, 0x4f, 0x4c, 0x5a, 0x9b, 0xb0, 0xe9, 0x9a, 0xec, 0x72, 0x70, 0xee, 0xc1,
		0x3a, 0xa9, 0x76, 0xac, 0x2e, 0xca, 0x04, 0x96, 0xf8, 0x97, 0x29, 0x20, 0xf4, 0x00, 0x00, 0x29,
		0x10, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)
	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv4); !ok {
		t.Errorf("Expected IPv4 layer, got %T", packetLayers[1])
	}

	ip4 := packetLayers[1].(*layers.IPv4)
	if ip4.FragOffset == 1480 {
		t.Errorf("Expected fragment offset equal to 1480")
	}
	if ip4.Flags&layers.IPv4MoreFragments != 0 {
		t.Errorf("Expected no flag for more fragment")
	}

	if _, ok := packetLayers[2].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv6_UDP(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x86, 0xdd,
		// ipv6
		0x60, 0x02, 0xb8, 0xfc, 0x00, 0x42, 0x11, 0x6b, 0x2a, 0x00, 0x14, 0x50, 0x40, 0x13, 0x0c, 0x03, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x0a, 0x20, 0x01, 0x04, 0x70, 0x76, 0x5b, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0a, 0x25, 0x00, 0x53,
		// udp
		0xb5, 0x61, 0x00, 0x35, 0x00, 0x42, 0xec, 0x92, 0xe9, 0xc4,
		0x00, 0x10, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x70, 0x61, 0x08, 0x77, 0x65,
		0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02, 0x64, 0x65, 0x00, 0x00, 0x1c, 0x00, 0x01, 0x00, 0x00,
		0x29, 0x10, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x0f, 0x00, 0x08, 0x00, 0x0b, 0x00, 0x02, 0x38,
		0x00, 0x20, 0x01, 0x04, 0x70, 0x1f, 0x0b, 0x16,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv6); !ok {
		t.Errorf("Expected IPv6 layer, got %T", packetLayers[1])
	}
	if _, ok := packetLayers[2].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv6_TCP(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x0c, 0x29, 0x62, 0x31, 0x2a, 0x00, 0x0c, 0x29, 0x7c, 0xa4, 0xcb, 0x86, 0xdd,
		// ipv6
		0x60, 0x0f, 0x4e, 0xd4, 0x00, 0x56, 0x06, 0x40, 0x20, 0x01, 0x04, 0x70, 0x1f, 0x0b, 0x16, 0xb0, 0x02, 0x0c,
		0x29, 0xff, 0xfe, 0x7c, 0xa4, 0xcb, 0x20, 0x01, 0x04, 0x70, 0x1f, 0x0b, 0x16, 0xb0, 0x00, 0x00,
		0x00, 0x00, 0x0a, 0x26, 0x00, 0x53,
		// tcp
		0xdf, 0x01, 0x00, 0x35, 0x21, 0xcd, 0x16, 0x09, 0x5c, 0x07,
		0xf0, 0xa9, 0x80, 0x18, 0x00, 0xbf, 0x8e, 0x81, 0x00, 0x00, 0x01, 0x01, 0x08, 0x0a, 0x84, 0x45,
		0xdf, 0x3b, 0x12, 0x7c, 0xd3, 0xd2, 0x00, 0x34, 0x80, 0xe4, 0x01, 0x20, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x08, 0x77, 0x65, 0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02, 0x64, 0x65,
		0x00, 0x00, 0x30, 0x00, 0x01, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x0c,
		0x00, 0x0a, 0x00, 0x08, 0x1b, 0x9a, 0xf6, 0x22, 0xab, 0x2c, 0x97, 0x40,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 3 {
		t.Fatalf("Unexpected number of layers: expected 3, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv6); !ok {
		t.Errorf("Expected IPv6 layer, got %T", packetLayers[1])
	}
	if _, ok := packetLayers[2].(*layers.TCP); !ok {
		t.Errorf("Expected TCP layer, got %T", packetLayers[2])
	}
}

func TestNetDecoder_Decode_IPv6_Fragment(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x86, 0xdd,
		// ipv6
		0x60, 0x07, 0x87, 0xfd, 0x00, 0x28, 0x2c, 0x40, 0x20, 0x01, 0x04, 0x70, 0x76, 0x5b, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0a, 0x25, 0x00, 0x53, 0x2a, 0x00, 0x14, 0x50, 0x40, 0x13, 0x0c, 0x03, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x0a,
		// data fragment
		0x11, 0x00, 0x00, 0x01, 0x28, 0x40, 0x3c, 0x0b, 0x00, 0x35,
		0xb5, 0x61, 0x05, 0xe5, 0x14, 0x8e, 0xe9, 0xc4, 0x84, 0x10, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03,
		0x00, 0x09, 0x02, 0x70, 0x61, 0x08, 0x77, 0x65, 0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02, 0x64,
		0x65, 0x00, 0x00, 0x1c, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x1c, 0x00, 0x01, 0x00, 0x00, 0x00, 0x3c,
		0x00, 0x10, 0x20, 0x01, 0x04, 0x70, 0x1f, 0x0b, 0x10, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x02, 0xc0, 0x0c, 0x00, 0x2e, 0x00, 0x01, 0x00, 0x00, 0x00, 0x3c, 0x01, 0x1f, 0x00, 0x1c,
		0x0a, 0x03, 0x00, 0x00, 0x00, 0x3c, 0x5d, 0x06, 0x59, 0xfc, 0x5c, 0xde, 0xbe, 0xec, 0x90, 0x47,
		0x08, 0x77, 0x65, 0x62, 0x65, 0x72, 0x6c, 0x61, 0x62, 0x02, 0x64, 0x65, 0x00, 0xb5, 0xa6, 0x75,
		0xcd, 0xf5, 0xa2, 0x41, 0xe3, 0xbc, 0x5c, 0x12, 0x5d, 0x2d, 0xf9, 0x1c, 0x89, 0x3e, 0xbf, 0xe9,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 4 {
		t.Fatalf("Unexpected number of layers: expected 4, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv6); !ok {
		t.Errorf("Expected IPv6 layer, got %T", packetLayers[1])
	}
	if _, ok := packetLayers[2].(*layers.IPv6Fragment); !ok {
		t.Errorf("Expected IPv6 framgment layer, got %T", packetLayers[2])
	}
	if _, ok := packetLayers[3].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[3])
	}
}

func TestNetDecoder_Decode_IPv6_EndFragment(t *testing.T) {
	pkt := []byte{
		// ethernet
		0x00, 0x86, 0x9c, 0xe7, 0x55, 0x14, 0x00, 0x0c, 0x29, 0x8a, 0x5d, 0xd7, 0x86, 0xdd,
		// ipv6
		0x60, 0x07, 0x87, 0xfd, 0x00, 0x45, 0x2c, 0x40, 0x20, 0x01, 0x04, 0x70, 0x76, 0x5b, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0a, 0x25, 0x00, 0x53, 0x2a, 0x00, 0x14, 0x50, 0x40, 0x13, 0x0c, 0x03, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x0a, 0x11, 0x00, 0x05, 0xa8, 0x28, 0x40, 0x3c, 0x0b,
		// udp payload
		0x5d, 0x7a, 0xb6, 0x6a, 0x1c, 0xea, 0x61, 0x8d, 0x79, 0x65, 0x32, 0x4f, 0x2c, 0x1e, 0xcc, 0x06, 0x91, 0x26,
		0x9a, 0x0e, 0x84, 0x7f, 0x00, 0xbf, 0x5b, 0xa9, 0x29, 0xc8, 0x49, 0x05, 0xca, 0x72, 0x79, 0xec,
		0xe6, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x0f, 0x00, 0x08, 0x00, 0x0b,
		0x00, 0x02, 0x38, 0x00, 0x20, 0x01, 0x04, 0x70, 0x1f, 0x0b, 0x16,
	}

	decoder := &NetDecoder{}

	packet := gopacket.NewPacket(pkt, decoder, gopacket.NoCopy)

	packetLayers := packet.Layers()
	if len(packetLayers) != 4 {
		t.Fatalf("Unexpected number of layers: expected 4, got %d", len(packetLayers))
	}

	if _, ok := packetLayers[0].(*layers.Ethernet); !ok {
		t.Errorf("Expected Ethernet layer, got %T", packetLayers[0])
	}
	if _, ok := packetLayers[1].(*layers.IPv6); !ok {
		t.Errorf("Expected IPv6 layer, got %T", packetLayers[1])
	}
	if _, ok := packetLayers[2].(*layers.IPv6Fragment); !ok {
		t.Errorf("Expected IPv6 framgment layer, got %T", packetLayers[2])
	}
	if _, ok := packetLayers[3].(*layers.UDP); !ok {
		t.Errorf("Expected UDP layer, got %T", packetLayers[3])
	}
}
