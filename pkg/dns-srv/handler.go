package dns_srv

import (
	"github.com/miekg/dns"
	"mapdns/pkg/log"
	"net"
)

type Handler struct {
	srv *Server
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	msg.Authoritative = true
	domain := msg.Question[0].Name
	qType := msg.Question[0].Qtype
	log.Logger().Debugf("qtype: %d, domain: %s", qType, domain)
	address, ok := h.srv.cache.GetRecord(qType, domain)
	if ok {
		switch qType {
		case dns.TypeA:
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: h.srv.cfg.DNS.TTL},
				A:   net.ParseIP(address),
			})
			break
		case dns.TypeAAAA:
			msg.Answer = append(msg.Answer, &dns.AAAA{
				Hdr:  dns.RR_Header{Name: domain, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: h.srv.cfg.DNS.TTL},
				AAAA: net.ParseIP(address),
			})
			break
		}
		log.Logger().Debugf("request: (%s, %d), response: %s", domain, qType, address)
	} else {
		log.Logger().Debugf("cloud not find domain: %d", domain)
	}
	if err := w.WriteMsg(&msg); err != nil {
		log.Logger().Errorf("write message error: %s", err)
	}
}
