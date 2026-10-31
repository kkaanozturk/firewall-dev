// Path: internal/firewall/engine_test.go
// Runtime: Go 1.24+ | Automated Architecture & Security Policy Verification Suite

package firewall_test

import (
	"net/netip"
	"testing"

	"github.com/aegisnova-global/ngfw-architecture/internal/capacity"
	"github.com/aegisnova-global/ngfw-architecture/internal/firewall"
	"github.com/aegisnova-global/ngfw-architecture/internal/soc"
)

func TestEnterpriseHeadcountAndBandwidthPlan(t *testing.T) {
	plans := capacity.BuildEnterprisePlans()
	if len(plans) != 4 {
		t.Fatalf("4 sube bekleniyordu, bulunan: %d", len(plans))
	}

	totalPeople := 0
	campusStudents := 0
	for _, p := range plans {
		totalPeople += p.TotalHeadcount()
		if p.ID == "akademik-sube" {
			campusStudents = p.Students
		}
		if len(p.Top5Firewalls) != 5 {
			t.Errorf("%s subesinde Top 5 guvenlik duvari listesi eksik: %d", p.ID, len(p.Top5Firewalls))
		}
		if p.AllocatedWANGbps <= p.RequiredPeakGbps() {
			t.Errorf("%s subesinde tahsis edilen WAN (%.2f Gbps) gereken pik yukun (%.2f Gbps) uzerinde olmali",
				p.ID, p.AllocatedWANGbps, p.RequiredPeakGbps())
		}
		sel := p.SelectedFirewall()
		if sel.VendorModel == "" {
			t.Errorf("%s subesinde secili #1 guvenlik duvari tanimli degil", p.ID)
		}
	}

	if campusStudents != 10000 {
		t.Errorf("Akademik Sube ogrenci sayisi 10000 olmali, bulunan: %d", campusStudents)
	}
	if totalPeople != 11425 {
		t.Errorf("Sirket genel toplam kisi sayisi 11425 olmali, bulunan: %d", totalPeople)
	}
}

func TestZeroTrustBranchRulesAndCompliance(t *testing.T) {
	fws := firewall.BuildAllBranchFirewalls()
	siem := soc.NewSIEMEngine(fws)

	// 1. Akademik Şube 10.000 Öğrenci -> İç OBS Veritabanı erişimi kesinlikle engellenmelidir
	stuToDB := firewall.Packet{
		Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "CAMPUS_DC",
		SrcIP: netip.MustParseAddr("10.100.15.99"), DstIP: netip.MustParseAddr("10.40.50.10"),
		UserGroup: "Role_Student", Application: "postgresql", Protocol: "TCP", DstPort: 5432,
	}
	dec := fws["akademik-sube"].Evaluate(stuToDB)
	siem.Ingest(stuToDB, dec)
	if dec.Allowed {
		t.Fatal("Ogrenci agindan OBS veritabanina erisim engellenmeliydi!")
	}

	// 2. Akademik Şube Öğrenci > 400 eşzamanlı oturum açtığında CGNAT koruma kuralı devreye girmelidir
	stuSessionFlood := firewall.Packet{
		Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET",
		SrcIP: netip.MustParseAddr("10.100.20.44"), DstIP: netip.MustParseAddr("193.140.100.10"),
		UserGroup: "Role_Student", Application: "ssl", Protocol: "TCP", DstPort: 443,
		ActiveHostSessions: 650, HostBandwidthMbps: 4.0,
	}
	decFlood := fws["akademik-sube"].Evaluate(stuSessionFlood)
	if decFlood.Allowed {
		t.Fatal("400 eszamanli oturum sinirini asan ogrenci trafigi sinirlandirilmaliydi!")
	}

	// 3. Ankara Savunma AR-GE -> Dış İnternet kesinlikle engellenmelidir
	rndLeak := firewall.Packet{
		Branch: "ankara-ofisi", SrcZone: "RND_DEFENSE", DstZone: "UNTRUST_WAN",
		SrcIP: netip.MustParseAddr("10.20.40.22"), DstIP: netip.MustParseAddr("1.1.1.1"),
		UserGroup: "GRP_Defense_RnD", Application: "ssl", Protocol: "TCP", DstPort: 443,
	}
	decRnd := fws["ankara-ofisi"].Evaluate(rndLeak)
	siem.Ingest(rndLeak, decRnd)
	if decRnd.Allowed {
		t.Fatal("Ankara Savunma AR-GE dogrudan internete cikamamali!")
	}

	// 4. Londra UK-GDPR -> Sınır ötesi veritabanı aktarımı engellenmeli ve SOC/DPO alarmı üretmelidir
	gdprViolation := firewall.Packet{
		Branch: "londra-ofisi", SrcZone: "UK_GDPR_SRV", DstZone: "VPN_HQ",
		SrcIP: netip.MustParseAddr("10.30.50.10"), DstIP: netip.MustParseAddr("10.10.50.80"),
		UserGroup: "john.smith@aegisnova.co.uk", Application: "postgresql", Protocol: "TCP", DstPort: 5432,
	}
	decGDPR := fws["londra-ofisi"].Evaluate(gdprViolation)
	siem.Ingest(gdprViolation, decGDPR)
	if decGDPR.Allowed {
		t.Fatal("UK-GDPR sinir otesi veritabani aktarimi engellenmeliydi!")
	}

	// 5. WAN Anti-Spoofing kontrolü
	spoofPkt := firewall.Packet{
		Branch: "istanbul-ofisi", SrcZone: "UNTRUST_WAN", DstZone: "DMZ_PUB",
		SrcIP: netip.MustParseAddr("10.10.10.15"), DstIP: netip.MustParseAddr("10.10.60.10"),
		Application: "ssl", Protocol: "TCP", DstPort: 443,
	}
	if fws["istanbul-ofisi"].Evaluate(spoofPkt).Allowed {
		t.Fatal("WAN uzerinden gelen sahte RFC1918 IP paketi dusurulmeliydi!")
	}

	// 6. 5651 Log Zinciri Bütünlük Doğrulaması
	if !siem.Verify5651Chain() {
		t.Fatal("5651 SHA-256 log zinciri dogrulamasi basarisiz!")
	}
	if len(siem.Incidents) < 2 {
		t.Fatalf("Kritik ihlaller icin SOC vakasi olusmaliydi, bulunan: %d", len(siem.Incidents))
	}
}

func BenchmarkNGFWPacketEvaluation(b *testing.B) {
	fws := firewall.BuildAllBranchFirewalls()
	campus := fws["akademik-sube"]
	pkt := firewall.Packet{
		Branch: "akademik-sube", SrcZone: "STUDENT_WIFI", DstZone: "WAN_ULAKNET",
		SrcIP: netip.MustParseAddr("10.100.42.19"), DstIP: netip.MustParseAddr("193.140.100.10"),
		UserGroup: "Role_Student", Application: "ssl", Protocol: "TCP", DstPort: 443,
		ActiveHostSessions: 120, HostBandwidthMbps: 8.0,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = campus.Evaluate(pkt)
	}
}

