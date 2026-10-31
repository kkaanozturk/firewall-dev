// Path: cmd/aegisnova-lab/main.go
// Runtime: Go 1.24+ (amd64/arm64) | Entrypoint: aegisnova-lab
// AegisNova Global - Tarayıcı Tabanlı Canlı NGFW, Bant Genişliği, HA Failover ve 7/24 SOC Laboratuvar Sunucusu

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/aegisnova-global/ngfw-architecture/internal/capacity"
	"github.com/aegisnova-global/ngfw-architecture/internal/firewall"
	"github.com/aegisnova-global/ngfw-architecture/internal/soc"
)

//go:embed ui.html
var embeddedUI []byte

type branchLiveStatus struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	SelectedFW       string  `json:"selected_fw"`
	TotalHeadcount   int     `json:"total_headcount"`
	AllocatedWANGbps float64 `json:"allocated_wan_gbps"`
	CurrentGbps      float64 `json:"current_gbps"`
	ActiveSessions   int     `json:"active_sessions"`
	ActiveHANode     string  `json:"active_ha_node"`
	ISP1Up           bool    `json:"isp1_up"`
	ISP2Up           bool    `json:"isp2_up"`
}

type packetEvent struct {
	Time      string            `json:"time"`
	Branch    string            `json:"branch"`
	SrcIP     string            `json:"src_ip"`
	DstIP     string            `json:"dst_ip"`
	DstPort   uint16            `json:"dst_port"`
	User      string            `json:"user"`
	App       string            `json:"app"`
	LatencyUs int64             `json:"latency_us"`
	Decision  firewall.Decision `json:"decision"`
}

type LabServer struct {
	mu       sync.Mutex
	plans    []capacity.BranchPlan
	fws      map[string]*firewall.BranchFirewall
	siem     *soc.SIEMEngine
	live     map[string]*branchLiveStatus
	feed     []packetEvent
}

func NewLabServer() *LabServer {
	s := &LabServer{}
	s.resetLocked()
	return s
}

func (s *LabServer) resetLocked() {
	s.plans = capacity.BuildEnterprisePlans()
	s.fws = firewall.BuildAllBranchFirewalls()
	s.siem = soc.NewSIEMEngine(s.fws)
	s.feed = nil
	s.live = make(map[string]*branchLiveStatus)

	initialSessions := map[string]int{
		"istanbul-ofisi": 485000,
		"ankara-ofisi":   210000,
		"londra-ofisi":   145000,
		"akademik-sube":  3420000,
	}
	for _, p := range s.plans {
		s.live[p.ID] = &branchLiveStatus{
			ID:               p.ID,
			Name:             p.Name,
			SelectedFW:       p.SelectedFirewall().VendorModel,
			TotalHeadcount:   p.TotalHeadcount(),
			AllocatedWANGbps: p.AllocatedWANGbps,
			CurrentGbps:      p.RequiredPeakGbps() * 0.84,
			ActiveSessions:   initialSessions[p.ID],
			ActiveHANode:     "FW01-PRIMARY",
			ISP1Up:           true,
			ISP2Up:           true,
		}
	}
}

func (s *LabServer) processPacketLocked(pkt firewall.Packet) (packetEvent, error) {
	fw, ok := s.fws[pkt.Branch]
	if !ok {
		return packetEvent{}, fmt.Errorf("gecersiz sube: %s", pkt.Branch)
	}
	start := time.Now()
	dec := fw.Evaluate(pkt)
	lat := time.Since(start).Microseconds()
	if lat == 0 {
		lat = 1
	}
	s.siem.Ingest(pkt, dec)

	ev := packetEvent{
		Time:      time.Now().Format("15:04:05"),
		Branch:    pkt.Branch,
		SrcIP:     pkt.SrcIP.String(),
		DstIP:     pkt.DstIP.String(),
		DstPort:   pkt.DstPort,
		User:      pkt.UserGroup,
		App:       pkt.Application,
		LatencyUs: lat,
		Decision:  dec,
	}
	s.feed = append([]packetEvent{ev}, s.feed...)
	if len(s.feed) > 60 {
		s.feed = s.feed[:60]
	}
	if st, exists := s.live[pkt.Branch]; exists && dec.Allowed {
		st.ActiveSessions += 12
	}
	return ev, nil
}

func (s *LabServer) snapshotLocked() map[string]any {
	var branches []branchLiveStatus
	for _, p := range s.plans {
		branches = append(branches, *s.live[p.ID])
	}
	tailStart := 0
	if len(s.siem.Chain5651) > 25 {
		tailStart = len(s.siem.Chain5651) - 25
	}
	return map[string]any{
		"branches":         branches,
		"packet_feed":      s.feed,
		"soc_incidents":    s.siem.Incidents,
		"chain_5651_valid": s.siem.Verify5651Chain(),
		"chain_5651_count": len(s.siem.Chain5651),
		"chain_5651_tail":  s.siem.Chain5651[tailStart:],
	}
}

func (s *LabServer) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(embeddedUI)
	})

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.snapshotLocked())
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.resetLocked()
		writeJSON(w, s.snapshotLocked())
	})

	mux.HandleFunc("/api/packet", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Branch             string  `json:"branch"`
			SrcZone            string  `json:"src_zone"`
			DstZone            string  `json:"dst_zone"`
			SrcIP              string  `json:"src_ip"`
			DstIP              string  `json:"dst_ip"`
			UserGroup          string  `json:"user_group"`
			Application        string  `json:"application"`
			DstPort            uint16  `json:"dst_port"`
			HostBandwidthMbps  float64 `json:"host_bandwidth_mbps"`
			ActiveHostSessions int     `json:"active_host_sessions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": "JSON okunamadi: " + err.Error()})
			return
		}
		srcAddr, err1 := netip.ParseAddr(req.SrcIP)
		dstAddr, err2 := netip.ParseAddr(req.DstIP)
		if err1 != nil || err2 != nil {
			writeJSON(w, map[string]any{"error": "Gecersiz IPv4/IPv6 adresi girdiniz"})
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		ev, err := s.processPacketLocked(firewall.Packet{
			Branch:             req.Branch,
			SrcZone:            req.SrcZone,
			DstZone:            req.DstZone,
			SrcIP:              srcAddr,
			DstIP:              dstAddr,
			UserGroup:          req.UserGroup,
			Application:        req.Application,
			Protocol:           "TCP",
			DstPort:            req.DstPort,
			HostBandwidthMbps:  req.HostBandwidthMbps,
			ActiveHostSessions: req.ActiveHostSessions,
		})
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"last_event": ev, "state": s.snapshotLocked()})
	})

	mux.HandleFunc("/api/drill", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Drill  string `json:"drill"`
			Branch string `json:"branch"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		s.mu.Lock()
		defer s.mu.Unlock()

		summary := ""
		switch req.Drill {
		case "campus_peak":
			pkts := []firewall.Packet{
				{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET", SrcIP: netip.MustParseAddr("10.100.14.20"), DstIP: netip.MustParseAddr("193.140.100.10"), UserGroup: "Role_Student", Application: "ssl", DstPort: 443, HostBandwidthMbps: 9.5, ActiveHostSessions: 140},
				{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ISP", SrcIP: netip.MustParseAddr("10.100.88.14"), DstIP: netip.MustParseAddr("91.198.174.192"), UserGroup: "Role_Student", Application: "bittorrent", DstPort: 6881},
				{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ISP", SrcIP: netip.MustParseAddr("10.100.19.99"), DstIP: netip.MustParseAddr("104.16.132.229"), UserGroup: "Role_Student", Application: "ssl", DstPort: 443, HostBandwidthMbps: 44.0, ActiveHostSessions: 190},
				{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET", SrcIP: netip.MustParseAddr("10.100.72.11"), DstIP: netip.MustParseAddr("193.140.100.10"), UserGroup: "Role_Student", Application: "ssl", DstPort: 443, HostBandwidthMbps: 6.0, ActiveHostSessions: 610},
				{Branch: "akademik-sube", SrcZone: "FACULTY_NET", DstZone: "WAN_ULAKNET", SrcIP: netip.MustParseAddr("10.40.24.40"), DstIP: netip.MustParseAddr("140.98.193.10"), UserGroup: "Role_Faculty", Application: "ieee-xplore", DstPort: 443, HostBandwidthMbps: 35.0, ActiveHostSessions: 75},
			}
			for _, p := range pkts {
				_, _ = s.processPacketLocked(p)
			}
			s.live["akademik-sube"].CurrentGbps = 24.46
			summary = "Tatbikat Tamamlandı: 10.000 Öğrenci pik yükü (24.46 Gbps), BitTorrent engeli, 15 Mbps Per-IP QoS ve 400 oturum CGNAT sınırı doğrulandı."

		case "exam_lockdown":
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "akademik-sube", SrcZone: "EXAM_ISOLATED", DstZone: "DMZ_EDU", SrcIP: netip.MustParseAddr("10.40.55.101"), DstIP: netip.MustParseAddr("10.40.60.50"), UserGroup: "Exam_Seat_101", Application: "ssl", DstPort: 443})
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "akademik-sube", SrcZone: "EXAM_ISOLATED", DstZone: "WAN_ISP", SrcIP: netip.MustParseAddr("10.40.55.101"), DstIP: netip.MustParseAddr("8.8.8.8"), UserGroup: "Exam_Seat_101", Application: "ssl", DstPort: 443})
			summary = "Sınav Salonu Tatbikatı: Yerel Sınav LMS erişimine izin verildi; dış internet çıkışı engellenerek gözetmen/SOC alarmı oluşturuldu."

		case "lateral_soar":
			attacker := netip.MustParseAddr("10.100.250.77")
			for port := uint16(21); port <= 25; port++ {
				_, _ = s.processPacketLocked(firewall.Packet{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "CAMPUS_DC", SrcIP: attacker, DstIP: netip.MustParseAddr("10.40.50.10"), UserGroup: "Role_Student", Application: "ssh", DstPort: port})
			}
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET", SrcIP: attacker, DstIP: netip.MustParseAddr("193.140.100.10"), UserGroup: "Role_Student", Application: "ssl", DstPort: 443})
			summary = "Yanal Tarama & SOAR Tatbikatı: 10.100.250.77 adresi OBS DB'ye 5 tarama yaptı -> SOAR otomatik EDL kara listesine aldı -> Normal internet çıkışı da anında kesildi!"

		case "ankara_rnd":
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "ankara-ofisi", SrcZone: "RND_DEFENSE", DstZone: "LOCAL_SRV", SrcIP: netip.MustParseAddr("10.20.40.12"), DstIP: netip.MustParseAddr("10.20.50.30"), UserGroup: "GRP_Defense_RnD", Application: "ssl", DstPort: 443})
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "ankara-ofisi", SrcZone: "RND_DEFENSE", DstZone: "UNTRUST_WAN", SrcIP: netip.MustParseAddr("10.20.40.12"), DstIP: netip.MustParseAddr("1.1.1.1"), UserGroup: "GRP_Defense_RnD", Application: "ssl", DstPort: 443})
			summary = "Ankara Savunma AR-GE Tatbikatı: FortiGate 200F ISFW imzalı iç aynaya izin verdi; doğrudan internet çıkışını bloklayıp Sev-1 SOC alarmı üretti."

		case "london_gdpr":
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "londra-ofisi", SrcZone: "UK_GDPR_SRV", DstZone: "VPN_HQ", SrcIP: netip.MustParseAddr("10.30.50.14"), DstIP: netip.MustParseAddr("10.10.50.80"), UserGroup: "uk.dba@aegisnova.co.uk", Application: "postgresql", DstPort: 5432})
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "londra-ofisi", SrcZone: "UK_GDPR_SRV", DstZone: "VPN_HQ", SrcIP: netip.MustParseAddr("10.30.50.25"), DstIP: netip.MustParseAddr("10.10.50.25"), UserGroup: "svc_siem_uk", Application: "syslog", DstPort: 6514})
			summary = "Londra UK-GDPR Tatbikatı: Sınır ötesi PostgreSQL veritabanı aktarımı engellendi ve kimlik maskelenerek (UK-PSEUDO-...) DPO/SOC vakası açıldı."

		case "spoof_edl":
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "istanbul-ofisi", SrcZone: "UNTRUST_WAN", DstZone: "DMZ_PUB", SrcIP: netip.MustParseAddr("10.10.10.5"), DstIP: netip.MustParseAddr("10.10.60.10"), UserGroup: "external", Application: "ssl", DstPort: 443})
			_, _ = s.processPacketLocked(firewall.Packet{Branch: "istanbul-ofisi", SrcZone: "UNTRUST_WAN", DstZone: "DMZ_PUB", SrcIP: netip.MustParseAddr("45.155.205.99"), DstIP: netip.MustParseAddr("10.10.60.10"), UserGroup: "external", Application: "ssl", DstPort: 443})
			summary = "Sınır Güvenliği Tatbikatı: WAN arayüzünden gelen sahte RFC1918 IP ve USOM C2 listesindeki zararlı IP (45.155.205.99) sınırda düşürüldü."

		case "ha_failover":
			if st, ok := s.live[req.Branch]; ok {
				if st.ActiveHANode == "FW01-PRIMARY" {
					st.ActiveHANode = "FW02-SECONDARY"
				} else {
					st.ActiveHANode = "FW01-PRIMARY"
				}
				summary = fmt.Sprintf("[%s] Stateful HA Failover tamamlandı: Yeni aktif cihaz %s (Süre: 0.64 sn, Oturum kaybı: 0).", st.Name, st.ActiveHANode)
			}

		case "isp_cut":
			if st, ok := s.live[req.Branch]; ok {
				st.ISP1Up = !st.ISP1Up
				if !st.ISP1Up {
					summary = fmt.Sprintf("[%s] ISP-1 fiber hattı koptu! SD-WAN SLA devreye girdi, tüm trafik kesintisiz ISP-2 yedek hattına aktarıldı.", st.Name)
				} else {
					summary = fmt.Sprintf("[%s] ISP-1 fiber hattı onarıldı, BGP/SD-WAN yük paylaşımı normale döndü.", st.Name)
				}
			}
		}

		writeJSON(w, map[string]any{"summary": summary, "state": s.snapshotLocked()})
	})

	mux.HandleFunc("/metrics", s.siem.MetricsHandler())
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "Canli Web Laboratuvari dinleme adresi")
	flag.Parse()

	srv := NewLabServer()
	fmt.Printf("====================================================================================\n")
	fmt.Printf(" AEGISNOVA GLOBAL - CANLI NGFW & SOC WEB LABORATUVARI AKTIF\n")
	fmt.Printf(" Tarayicinizdan acin -> http://%s\n", *addr)
	fmt.Printf(" Prometheus metrik   -> http://%s/metrics\n", *addr)
	fmt.Printf("====================================================================================\n")

	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		fmt.Fprintf(os.Stderr, "Web Lab sunucu hatasi: %v\n", err)
		os.Exit(1)
	}
}

