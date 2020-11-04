package nettrace

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestClientTrace(t *testing.T) {
	var result = make(map[string]string)
	trace := &ClientTrace{
		DNSStart: func(name string) {
			result["DNSStart"] = name
		},
		DNSDone: func(netIPs []net.IPAddr, coalesced bool, err error) {
			var ips []string
			for _, ip := range netIPs {
				ips = append(ips, ip.String())
			}
			result["DNSDone"] = strings.Join(ips, ",")
		},
		ConnectStart: func(network, addr string) {
			result["ConnectStart"] = network + "@" + addr
		},
		ConnectDone: func(network, addr string, err error) {
			result["ConnectDone"] = fmt.Sprintf("%s@%s@%v", network, addr, err)
		},
	}

	ctx := WithClientTrace(context.Background(), trace)

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", "www.baidu.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	if result["DNSStart"] != "www.baidu.com" {
		t.Errorf("DNSStart failed. got %v", result["DNSStart"])
	}
	if !strings.Contains(result["DNSDone"], ",") {
		t.Errorf("DNSDone failed. got %v", result["DNSDone"])
	}
	if !strings.HasPrefix(result["ConnectStart"], "tcp@") {
		t.Errorf("ConnectStart failed. got %v", result["ConnectStart"])
	}
	if !strings.HasPrefix(result["ConnectDone"], "tcp@") {
		t.Errorf("ConnectDone failed. got %v", result["ConnectDone"])
	}
}
