package dns

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/protocol/dns"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/common/signal/pubsub"
	"github.com/vmessocket/vmessocket/common/task"
	dns_feature "github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/routing"
	"github.com/vmessocket/vmessocket/transport/internet"
)

type DoHNameServer struct {
	sync.RWMutex
	ips        map[string]*record
	pub        *pubsub.Service
	cleanup    *task.Periodic
	reqID      uint32
	httpClient *http.Client
	dohURL     string
	name       string
}

func baseDOHNameServer(url *url.URL, prefix string) *DoHNameServer {
	s := &DoHNameServer{
		ips:    make(map[string]*record),
		pub:    pubsub.NewService(),
		name:   prefix + "//" + url.Host,
		dohURL: url.String(),
	}
	s.cleanup = &task.Periodic{
		Interval: time.Minute,
		Execute:  s.Cleanup,
	}
	return s
}

func NewDoHLocalNameServer(url *url.URL) *DoHNameServer {
	url.Scheme = "https"
	s := baseDOHNameServer(url, "DOHL")
	tr := &http.Transport{
		IdleConnTimeout:   90 * time.Second,
		ForceAttemptHTTP2: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := net.ParseDestination(network + ":" + addr)
			if err != nil {
				return nil, err
			}
			conn, err := internet.DialSystem(ctx, dest, nil)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	s.httpClient = &http.Client{
		Timeout:   time.Second * 180,
		Transport: tr,
	}
	newError("DNS: created Local DOH client for ", url.String()).AtInfo().WriteToLog()
	return s
}

func NewDoHNameServer(url *url.URL, dispatcher routing.Dispatcher) (*DoHNameServer, error) {
	newError("DNS: created Remote DOH client for ", url.String()).AtInfo().WriteToLog()
	s := baseDOHNameServer(url, "DOH")
	tr := &http.Transport{
		MaxIdleConns:        30,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
		ForceAttemptHTTP2:   true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := net.ParseDestination(network + ":" + addr)
			if err != nil {
				return nil, err
			}
			link, err := dispatcher.Dispatch(ctx, dest)
			if err != nil {
				return nil, err
			}
			return net.NewConnection(
				net.ConnectionInputMulti(link.Writer),
				net.ConnectionOutputMulti(link.Reader),
			), nil
		},
	}
	dispatchedClient := &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}
	s.httpClient = dispatchedClient
	return s, nil
}

func (s *DoHNameServer) Cleanup() error {
	now := time.Now()
	s.Lock()
	defer s.Unlock()
	if len(s.ips) == 0 {
		return newError("nothing to do. stopping...")
	}
	for domain, record := range s.ips {
		if record.A != nil && record.A.Expire.Before(now) {
			record.A = nil
		}
		if record.AAAA != nil && record.AAAA.Expire.Before(now) {
			record.AAAA = nil
		}
		if record.A == nil && record.AAAA == nil {
			newError(s.name, " cleanup ", domain).AtDebug().WriteToLog()
			delete(s.ips, domain)
		} else {
			s.ips[domain] = record
		}
	}
	if len(s.ips) == 0 {
		s.ips = make(map[string]*record)
	}
	return nil
}

func (s *DoHNameServer) dohHTTPSContext(ctx context.Context, b []byte) ([]byte, error) {
	body := bytes.NewBuffer(b)
	req, err := http.NewRequest("POST", s.dohURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/dns-message")
	req.Header.Add("Content-Type", "application/dns-message")
	resp, err := s.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("DOH server returned code %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (s *DoHNameServer) findIPsForDomain(domain string, option dns_feature.IPOption) ([]net.IP, error) {
	s.RLock()
	record, found := s.ips[domain]
	s.RUnlock()
	if !found {
		return nil, errRecordNotFound
	}
	var err4 error
	var err6 error
	var ips []net.Address
	var ip6 []net.Address
	if option.IPv4Enable {
		ips, err4 = record.A.getIPs()
	}
	if option.IPv6Enable {
		ip6, err6 = record.AAAA.getIPs()
		ips = append(ips, ip6...)
	}
	if len(ips) > 0 {
		return toNetIP(ips)
	}
	if err4 != nil {
		return nil, err4
	}
	if err6 != nil {
		return nil, err6
	}
	if (option.IPv4Enable && record.A != nil) || (option.IPv6Enable && record.AAAA != nil) {
		return nil, dns_feature.ErrEmptyResponse
	}
	return nil, errRecordNotFound
}

func (s *DoHNameServer) Name() string {
	return s.name
}

func (s *DoHNameServer) newReqID() uint16 {
	return uint16(atomic.AddUint32(&s.reqID, 1))
}

func (s *DoHNameServer) QueryIP(ctx context.Context, domain string, clientIP net.IP, option dns_feature.IPOption, disableCache bool) ([]net.IP, error) {
	fqdn := Fqdn(domain)
	if disableCache {
		newError("DNS cache is disabled. Querying IP for ", domain, " at ", s.name).AtDebug().WriteToLog()
	} else {
		ips, err := s.findIPsForDomain(fqdn, option)
		if err != errRecordNotFound {
			newError(s.name, " cache HIT ", domain, " -> ", ips).Base(err).AtDebug().WriteToLog()
			return ips, err
		}
	}
	var sub4, sub6 *pubsub.Subscriber
	if option.IPv4Enable {
		sub4 = s.pub.Subscribe(fqdn + "4")
		defer sub4.Close()
	}
	if option.IPv6Enable {
		sub6 = s.pub.Subscribe(fqdn + "6")
		defer sub6.Close()
	}
	done := make(chan interface{})
	go func() {
		if sub4 != nil {
			select {
			case <-sub4.Wait():
			case <-ctx.Done():
			}
		}
		if sub6 != nil {
			select {
			case <-sub6.Wait():
			case <-ctx.Done():
			}
		}
		close(done)
	}()
	s.sendQuery(ctx, fqdn, clientIP, option)
	for {
		ips, err := s.findIPsForDomain(fqdn, option)
		if err != errRecordNotFound {
			return ips, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
	}
}

func (s *DoHNameServer) sendQuery(ctx context.Context, domain string, clientIP net.IP, option dns_feature.IPOption) {
	newError(s.name, " querying: ", domain).AtInfo().WriteToLog(session.ExportIDToError(ctx))
	reqs := buildReqMsgs(domain, option, s.newReqID, genEDNS0Options(clientIP))
	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(time.Second * 5)
	}
	for _, req := range reqs {
		go func(r *dnsRequest) {
			dnsCtx := ctx
			if inbound := session.InboundFromContext(ctx); inbound != nil {
				dnsCtx = session.ContextWithInbound(dnsCtx, inbound)
			}
			dnsCtx = session.ContextWithContent(dnsCtx, &session.Content{
				Protocol:       "https",
				SkipDNSResolve: true,
			})
			var cancel context.CancelFunc
			dnsCtx, cancel = context.WithDeadline(dnsCtx, deadline)
			defer cancel()
			b, err := dns.PackMessage(r.msg)
			if err != nil {
				newError("failed to pack dns query").Base(err).AtError().WriteToLog()
				return
			}
			resp, err := s.dohHTTPSContext(dnsCtx, b.Bytes())
			if err != nil {
				newError("failed to retrieve response").Base(err).AtError().WriteToLog()
				return
			}
			rec, err := parseResponse(resp)
			if err != nil {
				newError("failed to handle DOH response").Base(err).AtError().WriteToLog()
				return
			}
			s.updateIP(r, rec)
		}(req)
	}
}

func (s *DoHNameServer) updateIP(req *dnsRequest, ipRec *IPRecord) {
	elapsed := time.Since(req.start)
	s.Lock()
	rec, found := s.ips[req.domain]
	if !found {
		rec = &record{}
	}
	updated := false
	switch req.reqType {
	case dnsmessage.TypeA:
		if isNewer(rec.A, ipRec) {
			rec.A = ipRec
			updated = true
		}
	case dnsmessage.TypeAAAA:
		addr := make([]net.Address, 0, len(ipRec.IP))
		for _, ip := range ipRec.IP {
			if len(ip.IP()) == net.IPv6len {
				addr = append(addr, ip)
			}
		}
		ipRec.IP = addr
		if isNewer(rec.AAAA, ipRec) {
			rec.AAAA = ipRec
			updated = true
		}
	}
	newError(s.name, " got answer: ", req.domain, " ", req.reqType, " -> ", ipRec.IP, " ", elapsed).AtInfo().WriteToLog()
	if updated {
		s.ips[req.domain] = rec
	}
	switch req.reqType {
	case dnsmessage.TypeA:
		s.pub.Publish(req.domain+"4", nil)
	case dnsmessage.TypeAAAA:
		s.pub.Publish(req.domain+"6", nil)
	}
	s.Unlock()
	common.Must(s.cleanup.Start())
}
