package dns

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS = "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) {
	if err := handlePacket(pc, addr, buf); err != nil {
		fmt.Printf("handlePacket error %s: %s\n", addr.String(), err)
	}
}

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	parse := dnsmessage.Parser{}
	header, err := parse.Start(buf)
	if err != nil {
		return err
	}
	question, err := parse.Question()
	if err != nil {
		return err
	}

	response, err := dnsQuery(getRootServers(), question)
	if err != nil {
		return err
	}

	response.Header.ID = header.ID
	responseBuffer, err := response.Pack()
	if err != nil {
		return err
	}

	_, err = pc.WriteTo(responseBuffer, addr)
	if err != nil {
		return err
	}

	return fmt.Errorf("not implemented yet")
}

func dnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Message, error) {
	fmt.Printf("Question: %+v\n", question)
	return &dnsmessage.Message{Header: dnsmessage.Header{RCode: dnsmessage.RCodeServerFailure}}, nil
}

func getRootServers() []net.IP {
	rootServers := []net.IP{}
	for _, server := range strings.Split(ROOT_SERVERS, ",") {
		rootServers = append(rootServers, net.ParseIP(server))
	}
	return rootServers
}

func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	// preparing the dns message to send to the servers
	max_value := ^uint16(0)
	randomID, err := rand.Int(rand.Reader, big.NewInt(int64(max_value)))
	if err != nil {
		return nil, nil, err
	}
	message := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:       uint16(randomID.Int64()),
			Response: false,
			OpCode:   dnsmessage.OpCode(0),
		},
		Questions: []dnsmessage.Question{question},
	}
	buf, err := message.Pack()
	if err != nil {
		return nil, nil, err
	}
	// establishing a connection with the servers
	var conn net.Conn
	for _, server := range servers {
		conn, err = net.Dial("udp", server.String()+":53")
		if err == nil {
			break
		}
	}
	if conn == nil {
		return nil, nil, fmt.Errorf("failed to make connection to servers: %s", err)
	}
	// sending the message encoded to the server
	_, err = conn.Write(buf)
	if err != nil {
		return nil, nil, err
	}

	// reading the answer
	answer := make([]byte, 512)
	lenOfMessage, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, nil, err
	}

	conn.Close()

	var parse dnsmessage.Parser

	header, err := parse.Start(answer[:lenOfMessage])
	if err != nil {
		return nil, nil, err
	}

	// Basic sanity checks
	questions, err := parse.AllQuestions()
	if err != nil {
		return nil, nil, err
	}

	if len(questions) != len(message.Questions) {
		return nil, nil, fmt.Errorf("answer packet doesn't have the same amount of questions")
	}

	err = parse.SkipAllQuestions()
	if err != nil {
		return nil, nil, err
	}

	return &parse, &header, nil
}
