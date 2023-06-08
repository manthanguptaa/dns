package dns

import (
	"crypto/rand"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type MockPacketConn struct{}

func (m *MockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}

func (m *MockPacketConn) Close() error {
	return nil
}

func (m *MockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

func (m *MockPacketConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestHandlePacket(t *testing.T) {
	urls := []string{"www.google.com.", "www.amazon.com."}
	for _, url := range urls {
		max := ^uint16(0)
		randomID, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
		if err != nil {
			t.Fatalf("randomID error %s", err)
		}
		message := dnsmessage.Message{
			Header: dnsmessage.Header{
				RCode:            dnsmessage.RCode(0),
				ID:               uint16(randomID.Int64()),
				OpCode:           dnsmessage.OpCode(0),
				Response:         false,
				AuthenticData:    false,
				RecursionDesired: false,
			},
			Questions: []dnsmessage.Question{
				{
					Name:  dnsmessage.MustNewName(url),
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
			},
		}

		buf, err := message.Pack()
		if err != nil {
			t.Fatalf("Message pack error %s", err)
		}
		err = handlePacket(&MockPacketConn{}, &net.IPAddr{IP: net.ParseIP("127.0.0.1")}, buf)
		if err != nil {
			t.Fatalf("Server error %s", err)
		}

	}
}

func TestOutgoingDnsQuery(t *testing.T) {
	question := dnsmessage.Question{
		Name:  dnsmessage.MustNewName("com."),
		Type:  dnsmessage.TypeNS,
		Class: dnsmessage.ClassINET,
	}
	rootServers := strings.Split(ROOT_SERVERS, ",")
	if len(rootServers) == 0 {
		t.Fatalf("no root servers found")
	}

	server := []net.IP{net.ParseIP(rootServers[0])}
	answer, header, err := outgoingDnsQuery(server, question)
	if err != nil {
		t.Fatalf("outgoingDnsQuery error %s", err)
	}
	if header == nil {
		t.Fatalf("no header found")
	}
	if answer == nil {
		t.Fatalf("no answer found")
	}
	if header.RCode != dnsmessage.RCodeSuccess {
		t.Fatalf("response wasn't successful")
	}
	err = answer.SkipAllAnswers()
	if err != nil {
		t.Fatalf("SkipAllAnswers error %s", err)
	}

	authorities, err := answer.AllAuthorities()
	if err != nil {
		t.Fatalf("error getting answers")
	}
	if len(authorities) == 0 {
		t.Fatalf("No answers received")
	}
}
