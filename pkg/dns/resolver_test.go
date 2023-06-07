package dns

import (
	"net"
	"strings"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

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
