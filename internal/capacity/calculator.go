// Path: internal/capacity/calculator.go
// Runtime: Go 1.24+ (amd64/arm64) | Module: capacity
// AegisNova Global - Matematiksel Bant Genişliği, QoS ve Top 5 NGFW Kapasite Hesaplama Motoru

package capacity

import "math"

// RoleProfile bir kullanıcı grubunun eşzamanlılık ve cihaz başı trafik katsayılarını tanımlar.
type RoleProfile struct {
	Name              string  `json:"name"`
	Count             int     `json:"count"`
	ConcurrencyFactor float64 `json:"concurrency_factor"`
	DeviceMultiplier  float64 `json:"device_multiplier"`
	PerDeviceMbps     float64 `json:"per_device_mbps"`
}

// EffectivePeakMbps ilgili rolün ham pik bant genişliği ihtiyacını (Mbps) döndürür.
func (r RoleProfile) EffectivePeakMbps() float64 {
	return float64(r.Count) * r.ConcurrencyFactor * r.DeviceMultiplier * r.PerDeviceMbps
}

// NGFWCandidate şube için değerlendirilen Top 5 güvenlik duvarı adayını temsil eder.
type NGFWCandidate struct {
	Rank                   int     `json:"rank"`
	VendorModel            string  `json:"vendor_model"`
	L4StatefulGbps         float64 `json:"l4_stateful_gbps"`
	ThreatProtectionGbps   float64 `json:"threat_protection_gbps"`
	SSLInspectionGbps      float64 `json:"ssl_inspection_gbps"`
	MaxConcurrentSessionsM float64 `json:"max_concurrent_sessions_millions"`
	Selected               bool    `json:"selected"`
	DecisionRationale      string  `json:"decision_rationale"`
}

// BranchPlan bir yerleşkenin kadro, bant genişliği ve güvenlik duvarı planını tutar.
type BranchPlan struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	AdminStaff          int             `json:"admin_staff"`
	Faculty             int             `json:"faculty"`
	Students            int             `json:"students"`
	SoftwareTeam        int             `json:"software_team"`
	RnDTeam             int             `json:"rnd_team"`
	TechnicalTeam       int             `json:"technical_team"`
	ServerAndGuestMbps  float64         `json:"server_and_guest_mbps"`
	EncapOverheadRatio  float64         `json:"encap_overhead_ratio"`
	BurstHeadroomRatio  float64         `json:"burst_headroom_ratio"`
	AllocatedWANGbps    float64         `json:"allocated_wan_gbps"`
	WANArchitectureDesc string          `json:"wan_architecture_desc"`
	Roles               []RoleProfile   `json:"roles"`
	Top5Firewalls       []NGFWCandidate `json:"top5_firewalls"`
}

// TotalHeadcount şubedeki toplam kişi sayısını hesaplar.
func (b BranchPlan) TotalHeadcount() int {
	return b.AdminStaff + b.Faculty + b.Students + b.SoftwareTeam + b.RnDTeam + b.TechnicalTeam
}

// RawPeakMbps tüm kullanıcı rolleri ve sunucu/DMZ/misafir yükünün ham toplamını döndürür.
func (b BranchPlan) RawPeakMbps() float64 {
	total := b.ServerAndGuestMbps
	for _, r := range b.Roles {
		total += r.EffectivePeakMbps()
	}
	return math.Round(total*100) / 100
}

// RequiredPeakGbps şifreleme ek yükü (IPsec/TLS) ve büyüme payı (Headroom) dahil gereken pik Gbps değerini hesaplar.
func (b BranchPlan) RequiredPeakGbps() float64 {
	withOverhead := b.RawPeakMbps() * (1.0 + b.EncapOverheadRatio) * (1.0 + b.BurstHeadroomRatio)
	return math.Round((withOverhead/1000.0)*100) / 100
}

// SelectedFirewall şube için seçilen #1 güvenlik duvarı modelini döndürür.
func (b BranchPlan) SelectedFirewall() NGFWCandidate {
	for _, fw := range b.Top5Firewalls {
		if fw.Selected {
			return fw
		}
	}
	return NGFWCandidate{}
}

// BuildEnterprisePlans AegisNova Global'in 4 ana şubesinin eksiksiz kapasite ve Top 5 NGFW verisini üretir.
func BuildEnterprisePlans() []BranchPlan {
	return []BranchPlan{
		{
			ID:                  "istanbul-ofisi",
			Name:                "İstanbul Ofisi (HQ, Core DC & Merkez SOC)",
			AdminStaff:          180,
			Faculty:             0,
			Students:            0,
			SoftwareTeam:        140,
			RnDTeam:             65,
			TechnicalTeam:       45,
			ServerAndGuestMbps:  930, // 850 Mbps DC/DMZ/VPN Hub + 80 Mbps Misafir/Stajyer
			EncapOverheadRatio:  0.12,
			BurstHeadroomRatio:  0.30,
			AllocatedWANGbps:    10.0,
			WANArchitectureDesc: "2 x 5 Gbps Simetrik Metro Ethernet (TT + Superonline BGP/SD-WAN)",
			Roles: []RoleProfile{
				{Name: "İdari Personel", Count: 180, ConcurrencyFactor: 0.65, DeviceMultiplier: 1.2, PerDeviceMbps: 2.5},
				{Name: "Yazılım Ekibi", Count: 140, ConcurrencyFactor: 0.85, DeviceMultiplier: 1.5, PerDeviceMbps: 6.0},
				{Name: "AR-GE Ekibi", Count: 65, ConcurrencyFactor: 0.85, DeviceMultiplier: 1.6, PerDeviceMbps: 8.0},
				{Name: "Teknik & SOC Ekibi", Count: 45, ConcurrencyFactor: 0.80, DeviceMultiplier: 1.5, PerDeviceMbps: 5.0},
			},
			Top5Firewalls: []NGFWCandidate{
				{Rank: 1, VendorModel: "Palo Alto PA-3430 (HA)", L4StatefulGbps: 38.2, ThreatProtectionGbps: 12.7, SSLInspectionGbps: 8.5, MaxConcurrentSessionsM: 3.0, Selected: true, DecisionRationale: "12.7 Gbps TP kapasitesi, App-ID/User-ID hassasiyeti ve Panorama merkezi yönetim standardı."},
				{Rank: 2, VendorModel: "Fortinet FortiGate 600F", L4StatefulGbps: 139.0, ThreatProtectionGbps: 10.5, SSLInspectionGbps: 9.0, MaxConcurrentSessionsM: 8.0, Selected: false, DecisionRationale: "Dış sınır SSL-VPN geçmiş zafiyet profili nedeniyle HQ sınırında tercih edilmedi."},
				{Rank: 3, VendorModel: "Check Point Quantum 6700", L4StatefulGbps: 42.0, ThreatProtectionGbps: 7.6, SSLInspectionGbps: 5.8, MaxConcurrentSessionsM: 4.0, Selected: false, DecisionRationale: "SSL Inspection altında 5.8 Gbps'e inerek 10 Gbps çift ISP hattına dar kalmaktadır."},
				{Rank: 4, VendorModel: "Cisco Secure Firewall 3130", L4StatefulGbps: 42.0, ThreatProtectionGbps: 15.2, SSLInspectionGbps: 6.4, MaxConcurrentSessionsM: 4.0, Selected: false, DecisionRationale: "FMC kural dağıtım süreleri ve dinamik CI/CD operasyonlarında yönetim yükü nedeniyle elendi."},
				{Rank: 5, VendorModel: "Sophos XGS 5500", L4StatefulGbps: 100.0, ThreatProtectionGbps: 14.0, SSLInspectionGbps: 13.5, MaxConcurrentSessionsM: 18.0, Selected: false, DecisionRationale: "Çok uluslu finans/AR-GE regülasyonu ve ileri seviye BGP/SOC telemetri derinliği nedeniyle elendi."},
			},
		},
		{
			ID:                  "ankara-ofisi",
			Name:                "Ankara Ofisi (Kamu B2B & Savunma AR-GE)",
			AdminStaff:          65,
			Faculty:             0,
			Students:            0,
			SoftwareTeam:        85,
			RnDTeam:             55,
			TechnicalTeam:       25,
			ServerAndGuestMbps:  290, // 250 Mbps Kamu VPN/Log + 40 Mbps Misafir
			EncapOverheadRatio:  0.12,
			BurstHeadroomRatio:  0.30,
			AllocatedWANGbps:    3.5,
			WANArchitectureDesc: "2.5 Gbps Ana Metro + 1.0 Gbps Farklı Fiziksel Güzergah Fiber",
			Roles: []RoleProfile{
				{Name: "İdari Personel", Count: 65, ConcurrencyFactor: 0.65, DeviceMultiplier: 1.2, PerDeviceMbps: 2.5},
				{Name: "Yazılım Ekibi", Count: 85, ConcurrencyFactor: 0.85, DeviceMultiplier: 1.5, PerDeviceMbps: 6.0},
				{Name: "Savunma AR-GE (Vekil Senkron)", Count: 55, ConcurrencyFactor: 0.80, DeviceMultiplier: 1.25, PerDeviceMbps: 8.0},
				{Name: "Teknik Ekip", Count: 25, ConcurrencyFactor: 0.80, DeviceMultiplier: 1.5, PerDeviceMbps: 5.0},
			},
			Top5Firewalls: []NGFWCandidate{
				{Rank: 1, VendorModel: "Palo Alto PA-1420 (Dış) + FortiGate 200F (AR-GE ISFW)", L4StatefulGbps: 9.8, ThreatProtectionGbps: 4.8, SSLInspectionGbps: 3.2, MaxConcurrentSessionsM: 1.4, Selected: true, DecisionRationale: "Panorama uyumu + İç Savunma AR-GE'de Fortinet ile çift üreticili (Dual-Vendor) izolasyon."},
				{Rank: 2, VendorModel: "Fortinet FortiGate 400F", L4StatefulGbps: 79.5, ThreatProtectionGbps: 9.0, SSLInspectionGbps: 8.0, MaxConcurrentSessionsM: 7.8, Selected: false, DecisionRationale: "Dış sınırda Panorama tekil şablon yönetimini bozmamak adına elendi."},
				{Rank: 3, VendorModel: "Check Point Quantum 6400", L4StatefulGbps: 32.0, ThreatProtectionGbps: 3.9, SSLInspectionGbps: 2.9, MaxConcurrentSessionsM: 2.0, Selected: false, DecisionRationale: "SSL Inspection kapasitesi 3.5 Gbps tam hat yükünde sınırda kalmaktadır."},
				{Rank: 4, VendorModel: "Cisco Secure Firewall 3110", L4StatefulGbps: 18.0, ThreatProtectionGbps: 8.9, SSLInspectionGbps: 3.8, MaxConcurrentSessionsM: 2.0, Selected: false, DecisionRationale: "Tek şube için ayrı FMC yönetim altyapısı maliyeti nedeniyle tercih edilmedi."},
				{Rank: 5, VendorModel: "WatchGuard Firebox M690", L4StatefulGbps: 29.7, ThreatProtectionGbps: 3.4, SSLInspectionGbps: 2.9, MaxConcurrentSessionsM: 9.2, Selected: false, DecisionRationale: "Savunma sanayii B2B entegrasyon ve granüler L7 imza gereksinimleri nedeniyle elendi."},
			},
		},
		{
			ID:                  "londra-ofisi",
			Name:                "Londra Ofisi (UK & EMEA Cloud Operasyon)",
			AdminStaff:          45,
			Faculty:             0,
			Students:            0,
			SoftwareTeam:        50,
			RnDTeam:             25,
			TechnicalTeam:       15,
			ServerAndGuestMbps:  250, // 220 Mbps EMEA API/VPN + 30 Mbps Misafir
			EncapOverheadRatio:  0.12,
			BurstHeadroomRatio:  0.30,
			AllocatedWANGbps:    4.0,
			WANArchitectureDesc: "2 x 2 Gbps Dedicated Leased Line (BTnet + Colt Tier-1)",
			Roles: []RoleProfile{
				{Name: "İdari, Satış & Hukuk (DPO)", Count: 45, ConcurrencyFactor: 0.65, DeviceMultiplier: 1.2, PerDeviceMbps: 2.5},
				{Name: "Cloud-Native Yazılım & SRE", Count: 50, ConcurrencyFactor: 0.90, DeviceMultiplier: 1.5, PerDeviceMbps: 6.67},
				{Name: "FinTech AR-GE Ekibi", Count: 25, ConcurrencyFactor: 0.85, DeviceMultiplier: 1.6, PerDeviceMbps: 8.0},
				{Name: "Teknik & CloudSec Ekibi", Count: 15, ConcurrencyFactor: 0.80, DeviceMultiplier: 1.5, PerDeviceMbps: 5.0},
			},
			Top5Firewalls: []NGFWCandidate{
				{Rank: 1, VendorModel: "Palo Alto PA-1410 (HA)", L4StatefulGbps: 8.6, ThreatProtectionGbps: 3.5, SSLInspectionGbps: 2.5, MaxConcurrentSessionsM: 0.98, Selected: true, DecisionRationale: "UK-GDPR yerleşik Enterprise DLP koruması, Prisma/AWS entegrasyonu ve Panorama uyumu."},
				{Rank: 2, VendorModel: "Fortinet FortiGate 200F", L4StatefulGbps: 27.0, ThreatProtectionGbps: 3.0, SSLInspectionGbps: 4.0, MaxConcurrentSessionsM: 3.0, Selected: false, DecisionRationale: "Kurumsal dış sınırda Panorama standardını korumak amacıyla elendi."},
				{Rank: 3, VendorModel: "Juniper Networks SRX1600", L4StatefulGbps: 25.0, ThreatProtectionGbps: 4.6, SSLInspectionGbps: 3.1, MaxConcurrentSessionsM: 2.0, Selected: false, DecisionRationale: "L7 SaaS/CASB ve UK-GDPR DLP olgunluğu Palo Alto kadar kapsamlı olmadığı için elendi."},
				{Rank: 4, VendorModel: "Cisco Secure Firewall 1220CX", L4StatefulGbps: 9.0, ThreatProtectionGbps: 4.5, SSLInspectionGbps: 2.2, MaxConcurrentSessionsM: 0.6, Selected: false, DecisionRationale: "Ayrı yönetim konsolu ve düşük oturum tablosu nedeniyle tercih edilmedi."},
				{Rank: 5, VendorModel: "Check Point Quantum 6200", L4StatefulGbps: 17.5, ThreatProtectionGbps: 2.5, SSLInspectionGbps: 1.8, MaxConcurrentSessionsM: 2.0, Selected: false, DecisionRationale: "SSL Inspection altında 1.8 Gbps'e inerek büyüme payını kısıtladığı için elendi."},
			},
		},
		{
			ID:                  "akademik-sube",
			Name:                "Akademik Şube (10.000 Öğrencili Teknoloji Kampüsü)",
			AdminStaff:          190,
			Faculty:             320,
			Students:            10000,
			SoftwareTeam:        35,
			RnDTeam:             45,
			TechnicalTeam:       40,
			ServerAndGuestMbps:  1750, // 1500 Mbps DMZ LMS/OBS + 250 Mbps Sempozyum Misafir
			EncapOverheadRatio:  0.10,
			BurstHeadroomRatio:  0.25,
			AllocatedWANGbps:    30.0,
			WANArchitectureDesc: "20 Gbps ULAKBİM/ULAKNET Simetrik + 10 Gbps Ticari ISP Metro Ethernet",
			Roles: []RoleProfile{
				{Name: "Öğrenci Ağı (10.000 Öğrenci Eduroam/Lab)", Count: 10000, ConcurrencyFactor: 0.65, DeviceMultiplier: 1.8, PerDeviceMbps: 1.111},
				{Name: "Akademisyen / Hoca Kadrosu", Count: 320, ConcurrencyFactor: 0.75, DeviceMultiplier: 1.5, PerDeviceMbps: 3.5},
				{Name: "İdari Personel", Count: 190, ConcurrencyFactor: 0.65, DeviceMultiplier: 1.2, PerDeviceMbps: 2.5},
				{Name: "Kampüs Yazılım Ekibi", Count: 35, ConcurrencyFactor: 0.85, DeviceMultiplier: 1.5, PerDeviceMbps: 6.0},
				{Name: "Akademik AR-GE & HPC", Count: 45, ConcurrencyFactor: 0.85, DeviceMultiplier: 2.0, PerDeviceMbps: 11.76},
				{Name: "Kampüs Teknik / NOC Ekibi", Count: 40, ConcurrencyFactor: 0.80, DeviceMultiplier: 1.5, PerDeviceMbps: 5.0},
			},
			Top5Firewalls: []NGFWCandidate{
				{Rank: 1, VendorModel: "Fortinet FortiGate 1800F (HA - NP7 ASIC)", L4StatefulGbps: 198.0, ThreatProtectionGbps: 31.0, SSLInspectionGbps: 20.0, MaxConcurrentSessionsM: 12.0, Selected: true, DecisionRationale: "31 Gbps TP, NP7 donanımsal PBA-CGNAT, 12 Milyon eşzamanlı oturum ve %45 TCO avantajı."},
				{Rank: 2, VendorModel: "Palo Alto PA-5420", L4StatefulGbps: 78.6, ThreatProtectionGbps: 30.2, SSLInspectionGbps: 19.5, MaxConcurrentSessionsM: 5.7, Selected: false, DecisionRationale: "Yarı yarıya düşük oturum tablosu (5.7M) ve eğitim bütçesini aşan yüksek abonelik maliyeti."},
				{Rank: 3, VendorModel: "Check Point Quantum 16200", L4StatefulGbps: 96.0, ThreatProtectionGbps: 24.0, SSLInspectionGbps: 16.0, MaxConcurrentSessionsM: 10.0, Selected: false, DecisionRationale: "24 Gbps TP kapasitesi kampüsün 24.5 Gbps pik ihtiyacında darboğaz riski taşımaktadır."},
				{Rank: 4, VendorModel: "Juniper Networks SRX4600", L4StatefulGbps: 120.0, ThreatProtectionGbps: 22.0, SSLInspectionGbps: 14.5, MaxConcurrentSessionsM: 20.0, Selected: false, DecisionRationale: "Threat Protection açıkken 22 Gbps'e düşerek 30 Gbps hattın gerisinde kalmaktadır."},
				{Rank: 5, VendorModel: "Cisco Secure Firewall 4115", L4StatefulGbps: 80.0, ThreatProtectionGbps: 21.4, SSLInspectionGbps: 12.0, MaxConcurrentSessionsM: 15.0, Selected: false, DecisionRationale: "21.4 Gbps TP ve 12 Gbps SSL kapasitesi kampüs ölçeği için yetersiz kalmaktadır."},
			},
		},
	}
}

