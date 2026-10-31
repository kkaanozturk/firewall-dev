// Path: internal/firewall/engine.go
// Runtime: Go 1.24+ (amd64/arm64) | Module: firewall
// AegisNova Global - Deterministik Bölge, Kimlik (User-ID/802.1X) ve Uygulama (App-ID) Tabanlı NGFW Kural Motoru

package firewall

import (
	"fmt"
	"net/netip"
	"sync"
	"sync/atomic"
)

// Action güvenlik duvarı kararını temsil eder.
type Action string

const (
	ActionAllow Action = "ALLOW"
	ActionDeny  Action = "DENY"
	ActionDrop  Action = "DROP"
)

// Packet simüle edilen veya denetlenen Katman 3/4/7 ağ paketini ve kimlik bağlamını temsil eder.
type Packet struct {
	Branch             string
	SrcZone            string
	DstZone            string
	SrcIP              netip.Addr
	DstIP              netip.Addr
	UserGroup          string // AD User-ID veya FreeRADIUS 802.1X RSSO grubu
	Application        string // App-ID (ssl, git, postgresql, bittorrent, syslog, ssh vb.)
	Protocol           string // TCP, UDP, ESP
	DstPort            uint16
	ActiveHostSessions int     // İstemcinin anlık eşzamanlı oturum sayısı (Per-IP Shaper kontrolü)
	HostBandwidthMbps  float64 // İstemcinin anlık bant genişliği tüketimi (Per-IP QoS kontrolü)
}

// Rule NGFW üzerindeki tek bir sıralı güvenlik kuralını temsil eder.
type Rule struct {
	ID                 int
	Name               string
	SrcZones           []string
	DstZones           []string
	SrcPrefixes        []netip.Prefix
	DstPrefixes        []netip.Prefix
	RequiredUserGroups []string
	Applications       []string
	DstPorts           []uint16
	MaxHostSessions    int     // >0 ise Per-IP eşzamanlı oturum tavanı uygular
	MaxHostMbps        float64 // >0 ise Per-IP bant genişliği tavanı uygular
	SecurityProfile    string
	ComplianceTag      string // SIEM, 5651-SIGNER, UK-GDPR-LOCK, RND-ISFW, EXAM-LOCK
	Action             Action
	LogToSIEM          bool
	HitCount           atomic.Uint64
}

// Decision bir paketin güvenlik duvarından geçiş sonucunu döndürür.
type Decision struct {
	Allowed         bool
	Action          Action
	MatchedRuleID   int
	MatchedRuleName string
	SecurityProfile string
	ComplianceTag   string
	Reason          string
}

// BranchFirewall tek bir şubenin NGFW örneğini, SOAR dinamik kara listesini (EDL) ve kurallarını tutar.
type BranchFirewall struct {
	BranchID   string
	DeviceName string
	mu         sync.RWMutex
	edlBlocked map[netip.Addr]string
	Rules      []*Rule
}

// NewBranchFirewall yeni bir şube güvenlik duvarı oluşturur ve başlangıç USOM/C2 tehdit beslemesini yükler.
func NewBranchFirewall(branchID, deviceName string, rules []*Rule) *BranchFirewall {
	fw := &BranchFirewall{
		BranchID:   branchID,
		DeviceName: deviceName,
		edlBlocked: make(map[netip.Addr]string),
		Rules:      rules,
	}
	// Başlangıç Tehdit İstihbaratı (EDL - Bilinen C2 ve Tor Çıkış Düğümleri)
	fw.AddDynamicBlockIP(netip.MustParseAddr("185.220.101.45"), "EDL-TOR-EXIT-NODE")
	fw.AddDynamicBlockIP(netip.MustParseAddr("45.155.205.99"), "EDL-USOM-MALICIOUS-C2")
	return fw
}

// AddDynamicBlockIP SOAR platformu tarafından tespit edilen saldırgan IP'yi anında (0-saniye commit ile) kara listeye alır.
func (fw *BranchFirewall) AddDynamicBlockIP(ip netip.Addr, reason string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.edlBlocked[ip] = reason
}

// Evaluate paketi yukarıdan aşağıya (Top-Down First Match) değerlendirir.
func (fw *BranchFirewall) Evaluate(pkt Packet) Decision {
	// 0. SOAR / EDL Dinamik Tehdit İstihbaratı Kontrolü
	fw.mu.RLock()
	edlReason, blocked := fw.edlBlocked[pkt.SrcIP]
	fw.mu.RUnlock()
	if blocked {
		return Decision{
			Allowed:         false,
			Action:          ActionDrop,
			MatchedRuleID:   0,
			MatchedRuleName: "SOAR-DYNAMIC-EDL-BLOCK",
			ComplianceTag:   "SOAR-EDL",
			Reason:          fmt.Sprintf("Kaynak IP (%s) dinamik tehdit listesinde engellendi: %s", pkt.SrcIP, edlReason),
		}
	}

	// 1. WAN Anti-Spoofing Kontrolü (Dış ağdan RFC1918 özel IP ile gelen sahte paketler)
	if (pkt.SrcZone == "UNTRUST_WAN" || pkt.SrcZone == "WAN_ULAKNET" || pkt.SrcZone == "WAN_ISP") && pkt.SrcIP.IsPrivate() {
		return Decision{
			Allowed:         false,
			Action:          ActionDrop,
			MatchedRuleID:   1,
			MatchedRuleName: "BLK-ANTISPOOF-RFC1918",
			ComplianceTag:   "ANTI-SPOOF",
			Reason:          "WAN arayüzünden özel (RFC1918) kaynak IP ile gelen sahte paket düşürüldü",
		}
	}

	// 2. Kural Tabanı Sıralı Değerlendirme
	for _, r := range fw.Rules {
		if !matchStringSlice(r.SrcZones, pkt.SrcZone) {
			continue
		}
		if !matchStringSlice(r.DstZones, pkt.DstZone) {
			continue
		}
		if !matchPrefixSlice(r.SrcPrefixes, pkt.SrcIP) {
			continue
		}
		if !matchPrefixSlice(r.DstPrefixes, pkt.DstIP) {
			continue
		}
		if !matchStringSlice(r.RequiredUserGroups, pkt.UserGroup) {
			continue
		}
		if !matchStringSlice(r.Applications, pkt.Application) {
			continue
		}
		if !matchPortSlice(r.DstPorts, pkt.DstPort) {
			continue
		}

		r.HitCount.Add(1)

		// Per-IP Oturum Sınırı Kontrolü (Örn: Akademik Şube 10.000 Öğrenci Max 400 Oturum)
		if r.Action == ActionAllow && r.MaxHostSessions > 0 && pkt.ActiveHostSessions > r.MaxHostSessions {
			return Decision{
				Allowed:         false,
				Action:          ActionDrop,
				MatchedRuleID:   r.ID,
				MatchedRuleName: r.Name + "-SESSION-CAP",
				SecurityProfile: r.SecurityProfile,
				ComplianceTag:   "QOS-SESSION-LIMIT",
				Reason:          fmt.Sprintf("İstemci eşzamanlı oturum sınırı aşıldı (%d > %d)", pkt.ActiveHostSessions, r.MaxHostSessions),
			}
		}

		// Per-IP Bant Genişliği Şekillendirme Kontrolü (Örn: Öğrenci > 15 Mbps veya Misafir > 5 Mbps)
		if r.Action == ActionAllow && r.MaxHostMbps > 0 && pkt.HostBandwidthMbps > r.MaxHostMbps {
			return Decision{
				Allowed:         false,
				Action:          ActionDrop,
				MatchedRuleID:   r.ID,
				MatchedRuleName: r.Name + "-QOS-SHAPER",
				SecurityProfile: r.SecurityProfile,
				ComplianceTag:   "QOS-BW-SHAPER",
				Reason:          fmt.Sprintf("Per-IP QoS tavanı aşıldı (%.1f Mbps > %.1f Mbps tavan)", pkt.HostBandwidthMbps, r.MaxHostMbps),
			}
		}

		return Decision{
			Allowed:         r.Action == ActionAllow,
			Action:          r.Action,
			MatchedRuleID:   r.ID,
			MatchedRuleName: r.Name,
			SecurityProfile: r.SecurityProfile,
			ComplianceTag:   r.ComplianceTag,
			Reason:          fmt.Sprintf("%s kuralı ile eşleşti (%s)", r.Name, r.Action),
		}
	}

	return Decision{
		Allowed:         false,
		Action:          ActionDrop,
		MatchedRuleID:   999,
		MatchedRuleName: "IMPLICIT-FALLBACK-DROP",
		ComplianceTag:   "SIEM",
		Reason:          "Hiçbir kuralla eşleşmedi, varsayılan güvenlik politikası gereği düşürüldü",
	}
}

// UnusedRules 0 isabet (hit) alan atıl kuralları raporlar (Tufin / Panorama Policy Optimizer karşılığı).
func (fw *BranchFirewall) UnusedRules() []string {
	var unused []string
	for _, r := range fw.Rules {
		if r.HitCount.Load() == 0 {
			unused = append(unused, r.Name)
		}
	}
	return unused
}

func pfx(cidr string) netip.Prefix { return netip.MustParsePrefix(cidr) }

func matchStringSlice(allowed []string, val string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if item == "*" || item == "ANY" || item == val {
			return true
		}
	}
	return false
}

func matchPrefixSlice(prefixes []netip.Prefix, ip netip.Addr) bool {
	if len(prefixes) == 0 {
		return true
	}
	for _, p := range prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

func matchPortSlice(ports []uint16, port uint16) bool {
	if len(ports) == 0 {
		return true
	}
	for _, p := range ports {
		if p == port {
			return true
		}
	}
	return false
}

// BuildAllBranchFirewalls 4 ana şubenin tamamı için üretim kural tabanlarını oluşturur.
func BuildAllBranchFirewalls() map[string]*BranchFirewall {
	return map[string]*BranchFirewall{
		"istanbul-ofisi": NewBranchFirewall("istanbul-ofisi", "IST-HQ-PA3430-HA", []*Rule{
			{ID: 10, Name: "ALLOW-MGMT-SOC-JUMP", SrcZones: []string{"IT_MGMT"}, DstZones: []string{"CORE_DC", "VPN_TUNNEL"}, SrcPrefixes: []netip.Prefix{pfx("10.10.10.0/24")}, RequiredUserGroups: []string{"GRP_SecOps_Admins"}, Applications: []string{"ssh", "ssl", "syslog"}, DstPorts: []uint16{22, 443, 6514}, SecurityProfile: "Strict-IPS+MFA", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 20, Name: "DENY-CORP-TO-SW-RND", SrcZones: []string{"CORP_STAFF"}, DstZones: []string{"SW_DEV", "RND_SEC"}, ComplianceTag: "LATERAL_DENY", Action: ActionDeny, LogToSIEM: true},
			{ID: 30, Name: "ALLOW-SWDEV-CICD-GIT", SrcZones: []string{"SW_DEV"}, DstZones: []string{"CORE_DC"}, SrcPrefixes: []netip.Prefix{pfx("10.10.30.0/23")}, DstPrefixes: []netip.Prefix{pfx("10.10.50.40/30")}, RequiredUserGroups: []string{"GRP_Software_Eng"}, Applications: []string{"git", "ssh", "ssl", "docker"}, DstPorts: []uint16{22, 443, 5000}, SecurityProfile: "Dev-IPS+WildFire", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 35, Name: "ALLOW-SWDEV-CLOUD-REPOS", SrcZones: []string{"SW_DEV"}, DstZones: []string{"UNTRUST_WAN"}, SrcPrefixes: []netip.Prefix{pfx("10.10.30.0/23")}, RequiredUserGroups: []string{"GRP_Software_Eng"}, Applications: []string{"github-base", "ssl", "npmjs", "pypi"}, DstPorts: []uint16{443}, SecurityProfile: "SSL-Decrypt+DLP-Code", ComplianceTag: "5651-SIGNER", Action: ActionAllow, LogToSIEM: true},
			{ID: 40, Name: "ALLOW-RND-PROXY-ONLY", SrcZones: []string{"RND_SEC"}, DstZones: []string{"CORE_DC"}, SrcPrefixes: []netip.Prefix{pfx("10.10.40.0/24")}, DstPrefixes: []netip.Prefix{pfx("10.10.50.60/32")}, RequiredUserGroups: []string{"GRP_RnD_Researchers"}, Applications: []string{"ssl"}, DstPorts: []uint16{8443}, SecurityProfile: "SSL-Decrypt+DLP-NoUpload", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 41, Name: "DENY-RND-DIRECT-WAN", SrcZones: []string{"RND_SEC"}, DstZones: []string{"UNTRUST_WAN", "CORP_STAFF", "GUEST_WIFI"}, SrcPrefixes: []netip.Prefix{pfx("10.10.40.0/24")}, ComplianceTag: "RND_ISOLATION", Action: ActionDeny, LogToSIEM: true},
			{ID: 50, Name: "ALLOW-CORP-INTERNET", SrcZones: []string{"CORP_STAFF"}, DstZones: []string{"UNTRUST_WAN"}, SrcPrefixes: []netip.Prefix{pfx("10.10.20.0/23")}, RequiredUserGroups: []string{"GRP_HQ_Staff"}, Applications: []string{"ssl", "web-browsing", "ms-office365", "zoom"}, DstPorts: []uint16{80, 443}, SecurityProfile: "SSL-Decrypt+URL-Corp", ComplianceTag: "5651-SIGNER", Action: ActionAllow, LogToSIEM: true},
			{ID: 80, Name: "DENY-GUEST-RFC1918", SrcZones: []string{"GUEST_WIFI"}, DstZones: []string{"IT_MGMT", "CORP_STAFF", "SW_DEV", "RND_SEC", "CORE_DC"}, ComplianceTag: "LATERAL_DENY", Action: ActionDeny, LogToSIEM: true},
			{ID: 85, Name: "ALLOW-GUEST-INTERNET", SrcZones: []string{"GUEST_WIFI"}, DstZones: []string{"UNTRUST_WAN"}, SrcPrefixes: []netip.Prefix{pfx("172.16.10.0/23")}, Applications: []string{"ssl", "web-browsing", "dns"}, DstPorts: []uint16{53, 80, 443}, MaxHostMbps: 5.0, SecurityProfile: "URL-Guest+5Mbps", ComplianceTag: "5651-SIGNER", Action: ActionAllow, LogToSIEM: true},
			{ID: 99, Name: "EXPLICIT-DENY-ALL-LOG", SrcZones: []string{"*"}, DstZones: []string{"*"}, ComplianceTag: "SIEM", Action: ActionDrop, LogToSIEM: true},
		}),

		"ankara-ofisi": NewBranchFirewall("ankara-ofisi", "ANK-PA1420-EXT+FG200F-ISFW", []*Rule{
			{ID: 10, Name: "ALLOW-RND-SIGNED-REPO-ISFW", SrcZones: []string{"RND_DEFENSE"}, DstZones: []string{"LOCAL_SRV"}, SrcPrefixes: []netip.Prefix{pfx("10.20.40.0/24")}, DstPrefixes: []netip.Prefix{pfx("10.20.50.30/32")}, RequiredUserGroups: []string{"GRP_Defense_RnD"}, Applications: []string{"ssl"}, DstPorts: []uint16{443}, SecurityProfile: "FG200F-DeepSSL+DLP-NoExfil", ComplianceTag: "RND-ISFW", Action: ActionAllow, LogToSIEM: true},
			{ID: 11, Name: "ALLOW-BASTION-TO-RND-ISFW", SrcZones: []string{"IT_MGMT"}, DstZones: []string{"RND_DEFENSE"}, SrcPrefixes: []netip.Prefix{pfx("10.20.10.50/32")}, DstPrefixes: []netip.Prefix{pfx("10.20.40.0/24")}, RequiredUserGroups: []string{"GRP_SecOps_Admins"}, Applications: []string{"ssh"}, DstPorts: []uint16{22}, SecurityProfile: "Bastion-Recorded-2FA", ComplianceTag: "RND-ISFW", Action: ActionAllow, LogToSIEM: true},
			{ID: 12, Name: "DENY-ALL-RND-OUTBOUND-AND-INBOUND", SrcZones: []string{"RND_DEFENSE", "CORP_STAFF", "SW_DEV", "GUEST_WIFI"}, DstZones: []string{"UNTRUST_WAN", "RND_DEFENSE"}, ComplianceTag: "RND_ISOLATION", Action: ActionDrop, LogToSIEM: true},
			{ID: 20, Name: "ALLOW-GOV-B2B-VPN", SrcZones: []string{"UNTRUST_WAN"}, DstZones: []string{"DMZ_GOV"}, DstPrefixes: []netip.Prefix{pfx("10.20.60.0/24")}, Applications: []string{"ike", "ipsec", "ssl"}, DstPorts: []uint16{500, 4500, 8443}, SecurityProfile: "Gov-mTLS-IPS", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 30, Name: "ALLOW-ANK-SWDEV-HQ-GIT", SrcZones: []string{"SW_DEV"}, DstZones: []string{"VPN_HQ"}, SrcPrefixes: []netip.Prefix{pfx("10.20.30.0/24")}, DstPrefixes: []netip.Prefix{pfx("10.10.50.40/30")}, RequiredUserGroups: []string{"GRP_ANK_Software"}, Applications: []string{"git", "ssh", "ssl"}, DstPorts: []uint16{22, 443}, SecurityProfile: "IPS+WildFire", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 99, Name: "EXPLICIT-DENY-ALL", SrcZones: []string{"*"}, DstZones: []string{"*"}, ComplianceTag: "SIEM", Action: ActionDrop, LogToSIEM: true},
		}),

		"londra-ofisi": NewBranchFirewall("londra-ofisi", "LON-EMEA-PA1410-HA", []*Rule{
			{ID: 10, Name: "DENY-UKGDPR-CROSSBORDER-DB", SrcZones: []string{"UK_GDPR_SRV"}, DstZones: []string{"VPN_HQ", "UNTRUST_WAN"}, SrcPrefixes: []netip.Prefix{pfx("10.30.50.0/24")}, Applications: []string{"postgresql", "mysql", "mssql-db", "smb", "ftp", "rsync"}, ComplianceTag: "UK-GDPR-LOCK", Action: ActionDeny, LogToSIEM: true},
			{ID: 15, Name: "ALLOW-PSEUDONYMIZED-SIEM", SrcZones: []string{"UK_GDPR_SRV"}, DstZones: []string{"VPN_HQ"}, SrcPrefixes: []netip.Prefix{pfx("10.30.50.25/32")}, DstPrefixes: []netip.Prefix{pfx("10.10.50.25/32")}, Applications: []string{"syslog"}, DstPorts: []uint16{6514}, SecurityProfile: "TLS1.3+Pseudonymized", ComplianceTag: "UK-GDPR-PSEUDO", Action: ActionAllow, LogToSIEM: true},
			{ID: 25, Name: "ALLOW-SWCLOUD-TO-AWS-UK", SrcZones: []string{"SW_CLOUD"}, DstZones: []string{"VPN_AWS_LON"}, SrcPrefixes: []netip.Prefix{pfx("10.30.30.0/24")}, DstPrefixes: []netip.Prefix{pfx("172.28.0.0/16")}, RequiredUserGroups: []string{"GRP_UK_CloudEng"}, Applications: []string{"ssh", "ssl", "kubernetes"}, DstPorts: []uint16{22, 443, 6443}, SecurityProfile: "DLP-CloudSecrets+IPS", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 40, Name: "ALLOW-CORP-SAAS-DLP", SrcZones: []string{"CORP_STAFF"}, DstZones: []string{"UNTRUST_WAN"}, SrcPrefixes: []netip.Prefix{pfx("10.30.20.0/24")}, RequiredUserGroups: []string{"GRP_UK_Staff"}, Applications: []string{"ms-office365", "salesforce", "ssl"}, DstPorts: []uint16{443}, SecurityProfile: "DLP-UK-GDPR-BlockShadowIT", ComplianceTag: "UK-GDPR-PSEUDO", Action: ActionAllow, LogToSIEM: true},
			{ID: 99, Name: "EXPLICIT-DENY-ALL-LOG", SrcZones: []string{"*"}, DstZones: []string{"*"}, ComplianceTag: "SIEM", Action: ActionDrop, LogToSIEM: true},
		}),

		"akademik-sube": NewBranchFirewall("akademik-sube", "AKA-CAMPUS-FG1800F-HA", []*Rule{
			{ID: 10, Name: "DENY-STUDENT-TO-INTERNAL-AND-HQVPN", SrcZones: []string{"STUDENT_WIFI", "COMPUTER_LABS", "CAMPUS_GUEST"}, DstZones: []string{"VPN_IST_SOC", "CAMPUS_DC", "ADMIN_STAFF", "FACULTY_NET", "RND_HPC", "STUDENT_WIFI"}, ComplianceTag: "LATERAL_DENY", Action: ActionDeny, LogToSIEM: true},
			{ID: 20, Name: "ALLOW-EXAM-ONLY-LMS", SrcZones: []string{"EXAM_ISOLATED"}, DstZones: []string{"DMZ_EDU"}, SrcPrefixes: []netip.Prefix{pfx("10.40.55.0/22")}, DstPrefixes: []netip.Prefix{pfx("10.40.60.50/32")}, Applications: []string{"ssl"}, DstPorts: []uint16{443}, SecurityProfile: "Exam-Lockdown-WAF", ComplianceTag: "EXAM-AUDIT", Action: ActionAllow, LogToSIEM: true},
			{ID: 21, Name: "DENY-EXAM-ALL-OTHER", SrcZones: []string{"EXAM_ISOLATED"}, DstZones: []string{"*"}, ComplianceTag: "EXAM_VIOLATION", Action: ActionDeny, LogToSIEM: true},
			{ID: 25, Name: "ALLOW-STUDENT-TO-WAF-OBS", SrcZones: []string{"STUDENT_WIFI", "COMPUTER_LABS"}, DstZones: []string{"DMZ_EDU"}, DstPrefixes: []netip.Prefix{pfx("10.40.60.10/32")}, RequiredUserGroups: []string{"Role_Student"}, Applications: []string{"ssl"}, DstPorts: []uint16{443}, SecurityProfile: "FortiWeb-OWASP-Top10", ComplianceTag: "SIEM", Action: ActionAllow, LogToSIEM: true},
			{ID: 28, Name: "DENY-STUDENT-P2P-TOR-MINE", SrcZones: []string{"STUDENT_WIFI", "COMPUTER_LABS"}, DstZones: []string{"WAN_ULAKNET", "WAN_ISP"}, Applications: []string{"bittorrent", "tor", "cryptomining", "smb"}, ComplianceTag: "APP-BLOCK-5651", Action: ActionDeny, LogToSIEM: true},
			{ID: 30, Name: "ALLOW-10K-STUDENTS-INTERNET-5651", SrcZones: []string{"STUDENT_WIFI"}, DstZones: []string{"WAN_ULAKNET", "WAN_ISP"}, SrcPrefixes: []netip.Prefix{pfx("10.100.0.0/16")}, RequiredUserGroups: []string{"Role_Student"}, Applications: []string{"ssl", "web-browsing", "dns", "zoom"}, DstPorts: []uint16{53, 80, 443}, MaxHostSessions: 400, MaxHostMbps: 15.0, SecurityProfile: "PBA-CGNAT+15Mbps+400Ses+WebFilter", ComplianceTag: "5651-SIGNER", Action: ActionAllow, LogToSIEM: true},
			{ID: 40, Name: "ALLOW-320-FACULTY-ACADEMIC", SrcZones: []string{"FACULTY_NET"}, DstZones: []string{"WAN_ULAKNET", "DMZ_EDU"}, SrcPrefixes: []netip.Prefix{pfx("10.40.24.0/22")}, RequiredUserGroups: []string{"Role_Faculty"}, Applications: []string{"ssl", "ssh", "ieee-xplore", "zoom"}, DstPorts: []uint16{22, 443}, MaxHostMbps: 100.0, SecurityProfile: "ULAKNET-Priority-AF41", ComplianceTag: "5651-SIGNER", Action: ActionAllow, LogToSIEM: true},
			{ID: 99, Name: "EXPLICIT-DENY-CAMPUS-ALL", SrcZones: []string{"*"}, DstZones: []string{"*"}, ComplianceTag: "SIEM", Action: ActionDeny, LogToSIEM: true},
		}),
	}
}

