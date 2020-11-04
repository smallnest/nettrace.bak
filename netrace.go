package nettrace

import (
	"context"
	"net"
	"net/http/httptrace"
	"reflect"
)

// ClientTraceKey is a context.Context Value key. Its associated value should
// be a *Trace struct.
type ClientTraceKey struct{}

// ClientTrace contains a set of hooks for tracing events within
// the net package. Any specific hook may be nil.
type ClientTrace struct {
	// DNSStart is called with the hostname of a DNS lookup
	// before it begins.
	DNSStart func(name string)

	// DNSDone is called after a DNS lookup completes (or fails).
	// The coalesced parameter is whether singleflight de-dupped
	// the call. The addrs are of type net.IPAddr but can't
	// actually be for circular dependency reasons.
	DNSDone func(netIPs []net.IPAddr, coalesced bool, err error)

	// ConnectStart is called before a Dial, excluding Dials made
	// during DNS lookups. In the case of DualStack (Happy Eyeballs)
	// dialing, this may be called multiple times, from multiple
	// goroutines.
	ConnectStart func(network, addr string)

	// ConnectStart is called after a Dial with the results, excluding
	// Dials made during DNS lookups. It may also be called multiple
	// times, like ConnectStart.
	ConnectDone func(network, addr string, err error)
}

// ContextClientTrace returns the ClientTrace associated with the
// provided context. If none, it returns nil.
func ContextClientTrace(ctx context.Context) *ClientTrace {
	trace, _ := ctx.Value(ClientTraceKey{}).(*ClientTrace)
	return trace
}

// WithClientTrace returns a new context based on the provided parent
// ctx. HTTP client requests made with the returned context will use
// the provided trace hooks, in addition to any previous hooks
// registered with ctx. Any hooks defined in the provided trace will
// be called first.
func WithClientTrace(ctx context.Context, trace *ClientTrace) context.Context {
	if trace == nil {
		panic("nil trace")
	}
	old := ContextClientTrace(ctx)
	trace.compose(old)

	ctx = context.WithValue(ctx, ClientTraceKey{}, trace)

	httpClientTrace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			trace.DNSStart(info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			trace.DNSDone(info.Addrs, info.Coalesced, info.Err)
		},
		ConnectStart: trace.ConnectStart,
		ConnectDone:  trace.ConnectDone,
	}

	ctx = httptrace.WithClientTrace(ctx, httpClientTrace)

	return ctx
}

// compose modifies t such that it respects the previously-registered hooks in old,
// subject to the composition policy requested in t.Compose.
func (t *ClientTrace) compose(old *ClientTrace) {
	if old == nil {
		return
	}
	tv := reflect.ValueOf(t).Elem()
	ov := reflect.ValueOf(old).Elem()
	structType := tv.Type()
	for i := 0; i < structType.NumField(); i++ {
		tf := tv.Field(i)
		hookType := tf.Type()
		if hookType.Kind() != reflect.Func {
			continue
		}
		of := ov.Field(i)
		if of.IsNil() {
			continue
		}
		if tf.IsNil() {
			tf.Set(of)
			continue
		}

		// Make a copy of tf for tf to call. (Otherwise it
		// creates a recursive call cycle and stack overflows)
		tfCopy := reflect.ValueOf(tf.Interface())

		// We need to call both tf and of in some order.
		newFunc := reflect.MakeFunc(hookType, func(args []reflect.Value) []reflect.Value {
			tfCopy.Call(args)
			return of.Call(args)
		})
		tv.Field(i).Set(newFunc)
	}
}
