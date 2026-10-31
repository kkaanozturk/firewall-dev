// Path: cmd/aegisnova-sim/main.go
// Runtime: Go 1.24+ (amd64/arm64) | Entrypoint: aegisnova-sim
// AegisNova Global - Çok Şubeli Bant Genişliği, Top 5 NGFW, Kural Motoru ve SOC/SIEM Doğrulama Aracı

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/aegisnova-global/ngfw-architecture/internal/capacity"
	"github.com/aegisnova-global/ngfw-architecture/internal/firewall"
	"github.com/aegisnova-global/ngfw-architecture/internal/soc"
)

type scenarioTest struct {
	Title       string
	ExpectAllow bool
	Packet      firewall.Packet
}

func main() {
	jsonOutput := flag.Bool("json", false, "Tum kapasite, kural ve SOC raporunu JSON olarak uretir")
	serveAddr := flag.String("serve-metrics", "", "Prometheus/Grafana metrik sunucusunu baslatir (orn. :9115)")
	flag.Parse()

	plans := capacity.BuildEnterprisePlans()
	fws := firewall.BuildAllBranchFirewalls()
	siem := soc.NewSIEMEngine(fws)

	scenarios := []scenarioTest{
		{
			Title:       "İstanbul Yazılım Ekibi (140 Kişi) -> İç GitLab CI/CD Erişimi",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "istanbul-ofisi", SrcZone: "SW_DEV", DstZone: "CORE_DC",
				SrcIP: netip.MustParseAddr("10.10.30.25"), DstIP: netip.MustParseAddr("10.10.50.40"),
				UserGroup: "GRP_Software_Eng", Application: "git", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "İstanbul AR-GE Ekibi (65 Kişi) -> Doğrudan Ham İnternet Çıkış Denemesi (Engellenmeli)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "istanbul-ofisi", SrcZone: "RND_SEC", DstZone: "UNTRUST_WAN",
				SrcIP: netip.MustParseAddr("10.10.40.19"), DstIP: netip.MustParseAddr("1.1.1.1"),
				UserGroup: "GRP_RnD_Researchers", Application: "ssl", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "Ankara Savunma AR-GE (55 Kişi) -> FortiGate 200F ISFW Üzerinden İmzalı Paket Aynası",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "ankara-ofisi", SrcZone: "RND_DEFENSE", DstZone: "LOCAL_SRV",
				SrcIP: netip.MustParseAddr("10.20.40.12"), DstIP: netip.MustParseAddr("10.20.50.30"),
				UserGroup: "GRP_Defense_RnD", Application: "ssl", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "Ankara Savunma AR-GE -> Dış İnternet Sızıntı Girişimi (ISFW + PA-1420 Çift Engel)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "ankara-ofisi", SrcZone: "RND_DEFENSE", DstZone: "UNTRUST_WAN",
				SrcIP: netip.MustParseAddr("10.20.40.12"), DstIP: netip.MustParseAddr("8.8.8.8"),
				UserGroup: "GRP_Defense_RnD", Application: "ssl", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "Londra UK-GDPR Sunucusu -> Türkiye'ye Ham PostgreSQL Veritabanı Aktarımı (UK-GDPR Kilidi)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "londra-ofisi", SrcZone: "UK_GDPR_SRV", DstZone: "VPN_HQ",
				SrcIP: netip.MustParseAddr("10.30.50.14"), DstIP: netip.MustParseAddr("10.10.50.80"),
				UserGroup: "uk.dba@aegisnova.co.uk", Application: "postgresql", Protocol: "TCP", DstPort: 5432,
			},
		},
		{
			Title:       "Londra Maskeli Log Sunucusu -> İstanbul Merkez SIEM TLS 6514 Log Gönderimi",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "londra-ofisi", SrcZone: "UK_GDPR_SRV", DstZone: "VPN_HQ",
				SrcIP: netip.MustParseAddr("10.30.50.25"), DstIP: netip.MustParseAddr("10.10.50.25"),
				UserGroup: "svc_siem_forwarder", Application: "syslog", Protocol: "TCP", DstPort: 6514,
			},
		},
		{
			Title:       "Akademik Şube Öğrenci (10.000 Havuz) -> Eduroam 802.1X ile Normal HTTPS İnternet (8 Mbps, 120 Oturum)",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET",
				SrcIP: netip.MustParseAddr("10.100.42.19"), DstIP: netip.MustParseAddr("193.140.100.10"),
				UserGroup: "Role_Student", Application: "ssl", Protocol: "TCP", DstPort: 443,
				ActiveHostSessions: 120, HostBandwidthMbps: 8.0,
			},
		},
		{
			Title:       "Akademik Şube Öğrenci -> BitTorrent / P2P İndirme Girişimi (App-Control Engeli)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ISP",
				SrcIP: netip.MustParseAddr("10.100.88.14"), DstIP: netip.MustParseAddr("91.198.174.192"),
				UserGroup: "Role_Student", Application: "bittorrent", Protocol: "TCP", DstPort: 6881,
			},
		},
		{
			Title:       "Akademik Şube Öğrenci -> 15 Mbps Per-IP QoS Sınırını Aşma Girişimi (42 Mbps Anlık Çekim)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ISP",
				SrcIP: netip.MustParseAddr("10.100.12.77"), DstIP: netip.MustParseAddr("104.16.132.229"),
				UserGroup: "Role_Student", Application: "ssl", Protocol: "TCP", DstPort: 443,
				ActiveHostSessions: 210, HostBandwidthMbps: 42.0,
			},
		},
		{
			Title:       "Akademik Şube Dijital Sınav Salonu -> Yerel LMS Sınav Sunucusu Erişimi",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "EXAM_ISOLATED", DstZone: "DMZ_EDU",
				SrcIP: netip.MustParseAddr("10.40.55.101"), DstIP: netip.MustParseAddr("10.40.60.50"),
				UserGroup: "Exam_Seat_101", Application: "ssl", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "Akademik Şube Dijital Sınav Salonu -> Dış İnternet Kopya/Erişim Girişimi (Sınav Kilidi)",
			ExpectAllow: false,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "EXAM_ISOLATED", DstZone: "WAN_ISP",
				SrcIP: netip.MustParseAddr("10.40.55.101"), DstIP: netip.MustParseAddr("8.8.4.4"),
				UserGroup: "Exam_Seat_101", Application: "ssl", Protocol: "TCP", DstPort: 443,
			},
		},
		{
			Title:       "Akademik Şube Hoca (320 Akademisyen) -> ULAKNET Üzerinden IEEE Xplore Araştırma Erişimi",
			ExpectAllow: true,
			Packet: firewall.Packet{
				Branch: "akademik-sube", SrcZone: "FACULTY_NET", DstZone: "WAN_ULAKNET",
				SrcIP: netip.MustParseAddr("10.40.24.55"), DstIP: netip.MustParseAddr("140.98.193.10"),
				UserGroup: "Role_Faculty", Application: "ieee-xplore", Protocol: "TCP", DstPort: 443,
				ActiveHostSessions: 80, HostBandwidthMbps: 25.0,
			},
		},
	}

	type scenarioResult struct {
		Title    string            `json:"title"`
		Passed   bool              `json:"passed"`
		Decision firewall.Decision `json:"decision"`
	}
	var results []scenarioResult
	allPassed := true

	for _, sc := range scenarios {
		fw := fws[sc.Packet.Branch]
		dec := fw.Evaluate(sc.Packet)
		siem.Ingest(sc.Packet, dec)
		passed := (dec.Allowed == sc.ExpectAllow)
		if !passed {
			allPassed = false
		}
		results = append(results, scenarioResult{Title: sc.Title, Passed: passed, Decision: dec})
	}

	// Yanal Hareket / Port Tarama Simülasyonu: Öğrenci Wi-Fi içinden OBS Veritabanına 5 ardışık tarama -> SOAR Karantina!
	rogueStudentIP := netip.MustParseAddr("10.100.250.66")
	for port := uint16(22); port <= 26; port++ {
		scanPkt := firewall.Packet{
			Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "CAMPUS_DC",
			SrcIP: rogueStudentIP, DstIP: netip.MustParseAddr("10.40.50.10"),
			UserGroup: "Role_Student", Application: "ssh", Protocol: "TCP", DstPort: port,
		}
		dec := fws["akademik-sube"].Evaluate(scanPkt)
		siem.Ingest(scanPkt, dec)
	}
	// SOAR otomatik karantina sonrası aynı öğrencinin normal internet paketi de artık EDL tarafından engellenmelidir!
	postQuarantineDec := fws["akademik-sube"].Evaluate(firewall.Packet{
		Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET",
		SrcIP: rogueStudentIP, DstIP: netip.MustParseAddr("193.140.100.10"),
		UserGroup: "Role_Student", Application: "ssl", Protocol: "TCP", DstPort: 443,
	})

	// Performans Ölçümü: 100.000 Paket Değerlendirme Hızı
	const benchPackets = 100000
	benchPkt := scenarios[6].Packet
	campusFW := fws["akademik-sube"]
	start := time.Now()
	for i := 0; i < benchPackets; i++ {
		_ = campusFW.Evaluate(benchPkt)
	}
	elapsed := time.Since(start)
	pps := float64(benchPackets) / elapsed.Seconds()

	if *jsonOutput {
		payload := map[string]any{
			"enterprise_plans":           plans,
			"policy_verification_passed": allPassed,
			"scenario_results":           results,
			"soar_post_quarantine_block": postQuarantineDec,
			"soc_incidents":              siem.Incidents,
			"law_5651_chain_valid":       siem.Verify5651Chain(),
			"law_5651_records_count":     len(siem.Chain5651),
			"engine_packets_per_sec":     int64(pps),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return
	}

	fmt.Println(strings.Repeat("=", 108))
	fmt.Println(" AEGISNOVA GLOBAL - COK SUBELI NGFW, BANT GENISLIGI VE 7/24 SOC SIMULASYON RAPORU")
	fmt.Println(strings.Repeat("=", 108))

	fmt.Println("\n[1] SUBE KADRO, BANT GENISLIGI HESABI VE SECILEN #1 GUVENLIK DUVARI MATRISI")
	fmt.Println(strings.Repeat("-", 108))
	fmt.Printf("%-42s | %7s | %9s | %9s | %s\n", "Sube Adi", "Toplam", "Pik (Gbps)", "WAN (Gbps)", "Secilen Birincil NGFW")
	fmt.Println(strings.Repeat("-", 108))
	totalPeople := 0
	totalWAN := 0.0
	for _, p := range plans {
		totalPeople += p.TotalHeadcount()
		totalWAN += p.AllocatedWANGbps
		fmt.Printf("%-42s | %7d | %9.2f | %10.1f | %s\n",
			p.Name, p.TotalHeadcount(), p.RequiredPeakGbps(), p.AllocatedWANGbps, p.SelectedFirewall().VendorModel)
	}
	fmt.Println(strings.Repeat("-", 108))
	fmt.Printf("%-42s | %7d | %9s | %10.1f | %s\n", "GENEL TOPLAM (4 YERLESKE)", totalPeople, "-", totalWAN, "Zero-Trust Hibrit NGFW HA")

	fmt.Println("\n[2] SIFIR GUVEN KURAL MOTORU DOGRULAMA TESTLERI (12 KRITIK SENARYO)")
	fmt.Println(strings.Repeat("-", 108))
	for i, r := range results {
		status := "BASARILI"
		if !r.Passed {
			status = "HATA"
		}
		fmt.Printf(" %02d. [%s] %-6s | %-34s | %s\n",
			i+1, status, r.Decision.Action, r.Decision.MatchedRuleName, r.Title)
	}

	fmt.Println("\n[3] MERKEZ SOC SIEM KORELASYON, SOAR OTOMATIK KARANTINA VE YASAL UYUMLULUK")
	fmt.Println(strings.Repeat("-", 108))
	for _, inc := range siem.Incidents {
		fmt.Printf(" * [%s] %s (%s) -> %s\n   Ozet : %s\n   SOAR : %s\n",
			inc.Severity, inc.ID, inc.Branch, inc.AssignedTier, inc.SourceSummary, inc.SOARAction)
	}
	fmt.Printf("\n * SOAR Karantina Sonrasi Test (%s): Karar=%s | Kural=%s\n",
		rogueStudentIP, postQuarantineDec.Action, postQuarantineDec.MatchedRuleName)
	fmt.Printf(" * 5651 Sayili Kanun SHA-256 Log Zinciri Dogrulamasi : GECERLI=%v (%d imzali kayit)\n",
		siem.Verify5651Chain(), len(siem.Chain5651))
	fmt.Printf(" * UK-GDPR Kimlik Maskeleme Ornegi (uk.dba@...)      : %s\n",
		soc.PseudonymizeUKGDPR("uk.dba@aegisnova.co.uk"))
	fmt.Printf(" * Go Kural Motoru Paket Isleme Performansi          : %.0f paket/saniye (100.000 paketlik olcum: %s)\n",
		pps, elapsed.Round(time.Microsecond))
	fmt.Println(strings.Repeat("=", 108))

	if *serveAddr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/metrics", siem.MetricsHandler())
		fmt.Printf("\n[CANLI IZLEME] Prometheus telemetri sunucusu %s/metrics adresinde dinleniyor...\n", *serveAddr)
		if err := http.ListenAndServe(*serveAddr, mux); err != nil {
			fmt.Fprintf(os.Stderr, "Metrik sunucu hatasi: %v\n", err)
			os.Exit(1)
		}
	}
}

