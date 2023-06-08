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

	// Getting the question from the packet
	question, err := parse.Question()
	if err != nil {
		return err
	}

	response, err := dnsQuery(getRootServers(), question)
	if err != nil {
		return err
	}

	response.Header.ID = header.ID
	// Encoding the message
	responseBuffer, err := response.Pack()
	if err != nil {
		return err
	}

	// Writing back to the client
	_, err = pc.WriteTo(responseBuffer, addr)
	if err != nil {
		return err
	}

	return nil
}

func dnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Message, error) {
	fmt.Printf("Question: %+v\n", question)

	for i := 0; i < 3; i++ {
		// getting the answer and header from the server
		dnsAnswer, header, err := outgoingDnsQuery(servers, question)
		if err != nil {
			return nil, err
		}
		// parsing the answers
		parsedAnswers, err := dnsAnswer.AllAnswers()
		if err != nil {
			return nil, err
		}

		// checking if we have received the final IP address of the question
		if header.Authoritative {
			return &dnsmessage.Message{
				Header:  dnsmessage.Header{Response: true},
				Answers: parsedAnswers,
			}, nil
		}
		// getting the authorities attached to the answer packet
		authorities, err := dnsAnswer.AllAuthorities()
		if err != nil {
			return nil, err
		}
		if len(authorities) == 0 {
			return &dnsmessage.Message{
					Header: dnsmessage.Header{RCode: dnsmessage.RCodeNameError}},
				nil
		}

		// getting the nameservers from the authorities that could give us the IP address for the question
		nameservers := make([]string, len(authorities))
		for k, authority := range authorities {
			if authority.Header.Type == dnsmessage.TypeNS {
				nameservers[k] = authority.Body.(*dnsmessage.NSResource).NS.String()
			}
		}

		// getting the additionals attached to the answer packet
		additionals, err := dnsAnswer.AllAdditionals()
		if err != nil {
			return nil, err
		}

		// We try to find next set of ipv4 addresses to ask the question
		newResolverServersFound := false
		// We edit the server param given to dnsQuery function so in the next loop we have new set of
		// servers to ask the question
		servers = []net.IP{}
		for _, additional := range additionals {
			if additional.Header.Type == dnsmessage.TypeA {
				for _, nameserver := range nameservers {
					if additional.Header.Name.String() == nameserver {
						newResolverServersFound = true
						servers = append(servers, additional.Body.(*dnsmessage.AResource).A[:])
					}
				}
			}
		}

		// if no new ipv4 address is found then we start querying for the nameserver as the question
		// and find its ipv4 address
		if !newResolverServersFound {
			for _, nameserver := range nameservers {
				if !newResolverServersFound {
					response, err := dnsQuery(getRootServers(),
						dnsmessage.Question{
							Name:  dnsmessage.MustNewName(nameserver),
							Type:  dnsmessage.TypeA,
							Class: dnsmessage.ClassINET})
					if err != nil {
						fmt.Printf("Warning: lookup of nameserver %s failed: %s\n", nameserver, err)
					} else {
						newResolverServersFound = true
						for _, answer := range response.Answers {
							if answer.Header.Type == dnsmessage.TypeA {
								servers = append(servers, answer.Body.(*dnsmessage.AResource).A[:])
							}
						}
					}
				}
			}
		}
	}

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
	fmt.Printf("New outgoing dns query for %s, servers: %+v\n", question.Name.String(), servers)
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
