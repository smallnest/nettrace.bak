# nettrace

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/smallnest/nettrace?status.png)](http://godoc.org/github.com/smallnest/nettrace)  [![travis](https://travis-ci.org/smallnest/nettrace.svg?branch=master)](https://travis-ci.org/smallnest/nettrace) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/netrace)](https://goreportcard.com/report/github.com/smallnest/nettrace) 

There is [nettrace](https://github.com/golang/go/blob/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/src/internal/nettrace/nettrace.go) in go source, but unfortunately it is internal has not been explored and be used.

[httptrace](https://github.com/golang/go/blob/b963149d4eddaf92d9e2a9d3bf5474c2d0a3b55d/src/net/http/httptrace/trace.go) uses nettrace to trace dns lookup and connecting, so we wrap it as a higher nettrace.

Finally the solution is like below:

 our nettrace --> httptrace --> internal trace.


 ## How to use it?


 ```go
    var result = make(map[string]string)
    
    // create the nettrace instance and set its hooks
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

    // create a context to use this trace
	ctx := WithClientTrace(context.Background(), trace)

    d := net.Dialer{Timeout: 5 * time.Second}
    // dial with this context
	conn, err := d.DialContext(ctx, "tcp", "www.baidu.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
 ```
