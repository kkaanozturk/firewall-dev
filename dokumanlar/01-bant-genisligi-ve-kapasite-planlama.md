<!-- Path: dokumanlar/01-bant-genisligi-ve-kapasite-planlama.md -->
# Bant Genişliği Tanımlama, Kapasite Hesaplama ve QoS Mimarisi

Bu doküman, **AegisNova Global** bünyesindeki dört ana şubenin (İstanbul, Ankara, Londra, Akademik Şube) ve fonksiyonel birimlerin (İdari Personel, Yazılım Ekibi, AR-GE Ekibi, Teknik Ekip, Akademisyen/Hoca ve 10.000 Öğrenci) ihtiyaç duyduğu internet/WAN bant genişliğini mühendislik formülleriyle hesaplar ve güvenlik duvarları üzerinde uygulanacak **QoS (Quality of Service / Traffic Shaping)** profillerini tanımlar.

---

## 1. Kapasite Hesaplama Metodolojisi

Kurumsal ağ planlamasında her kullanıcının aynı saniyede hattı tam kapasiteyle kullanacağı varsayımı aşırı maliyet doğururken, eşzamanlılık ve protokol ek yükünü (overhead) ihmal etmek pik saatlerde paket kaybına yol açar. Bu nedenle aşağıdaki kapasite formülü uygulanmıştır:

$$B_{\text{toplam}} = \left( \sum_{i=1}^{k} N_i \times C_i \times M_i \times b_i \right) \times (1 + O_{\text{encap}}) \times (1 + H_{\text{burst}})$$

### Parametre Tanımları

- $N_i$: İlgili kullanıcı grubundaki toplam kişi sayısı.
- $C_i$: Pik saat eşzamanlı aktiflik oranı (*Concurrency Factor*, 0.0 – 1.0 arası).
- $M_i$: Kişi başı aktif uç nokta/oturum çarpanı (dizüstü + mobil + test VM).
- $b_i$: Aktif uç nokta başına sürdürülebilir ortalama bant genişliği ihtiyacı (Mbps).
- $O_{\text{encap}}$: Protokol ve şifreleme ek yükü (Ethernet/VLAN + IPsec/TLS + SD-WAN başlıkları için **%12**).
- $H_{\text{burst}}$: Ani trafik sıçramaları ve 3 yıllık büyüme payı (*Headroom*, hat doluluğunun pik anlarda %70–75 bandını aşmaması için **%30**).

---

## 2. Rol ve Ekip Bazlı Trafik Profili Katsayıları

Her ekibin ağ üzerindeki davranış modeli birbirinden farklıdır. Şirket genelinde standartlaştırılan profil değerleri aşağıdadır:

| Kullanıcı / Birim Rolü | Eşzamanlılık ($C_i$) | Cihaz Çarpanı ($M_i$) | Cihaz Başı Ortalama ($b_i$) | Kişi Başı Efektif Pik İhtiyaç | Baskın Trafik Türü |
|---|---:|---:|---:|---:|---|
| **İdari Personel** | %65 | 1.2 | 2.5 Mbps | ~1.95 Mbps | ERP, SaaS (M365/Workspace), VoIP, 1080p Toplantı |
| **Yazılım Ekibi** | %85 | 1.5 | 6.0 Mbps | ~7.65 Mbps | Git klonlama, CI/CD artefaktları, Docker/Container imaj çekme, Cloud SSH/API |
| **AR-GE Ekibi** | %85 | 1.6 | 8.0 Mbps | ~10.88 Mbps | Büyük veri setleri, yapay zeka model ağırlıkları, simülasyon telemetrisi |
| **Teknik / IT & SOC Ekibi** | %80 | 1.5 | 5.0 Mbps | ~6.00 Mbps | Yama dağıtımı (WSUS/Repo), SIEM log akışı, uzaktan erişim, imaj yedekleme |
| **Akademisyen (Hoca)** | %75 | 1.5 | 3.5 Mbps | ~3.94 Mbps | Akademik veri tabanları (IEEE/Springer), canlı ders yayını, bulut depolama |
| **Öğrenci (Kampüs)** | %65 | 1.8 | 1.7 Mbps | ~1.99 Mbps | LMS ders videoları, web araştırması, Eduroam Wi-Fi, çevrimiçi sınav/ödev |

---

## 3. Şube Bazlı Detaylı Bant Genişliği Hesapları

### 3.1. İstanbul Ofisi (Genel Merkez, Ana Veri Merkezi ve Merkez SOC)

İstanbul yerleşkesi hem 430 çalışanı barındırır hem de diğer tüm şubelerden gelen IPsec VPN tünellerini, merkezi SIEM loglarını ve dış dünyaya açık kurumsal servisleri (DMZ) sonlandırır.

| Birim / Kaynak | Kişi / Sistem | Ham Pik İhtiyaç (Mbps) | Açıklama |
|---|---:|---:|---|
| İdari Personel | 180 kişi | 351 Mbps | $180 \times 1.95\text{ Mbps}$ |
| Yazılım Ekibi | 140 kişi | 1.071 Mbps | $140 \times 7.65\text{ Mbps}$ (Merkez CI/CD ve Container trafiği) |
| AR-GE Ekibi | 65 kişi | 707 Mbps | $65 \times 10.88\text{ Mbps}$ (AI veri seti ve laboratuvar trafiği) |
| Teknik ve Merkez SOC | 45 kişi | 270 Mbps | $45 \times 6.00\text{ Mbps}$ |
| Ana Veri Merkezi (DC + DMZ + VPN Hub) | ~120 VM / Sunucu | 850 Mbps | Şubeler arası replikasyon, merkezi SIEM alımı, müşteri API trafiği |
| Misafir ve Stajyer Ağı | ~40 eşzamanlı | 80 Mbps | Kısıtlı misafir internet havuzu |
| **Alt Toplam (Ham Trafik)** | **430 Çalışan** | **3.329 Mbps** | |
| **Protokol Ek Yükü (%12) + Büyüme Payı (%30)** | — | **~4.847 Mbps (~4.85 Gbps)** | $3.329 \times 1.12 \times 1.30$ |

- **Tahsis Edilen Hat:** **2 × 5 Gbps Simetrik Metro Ethernet** (ISP-1: Türk Telekom Kurumsal + ISP-2: Turkcell Superonline, BGP + SD-WAN Aktif/Aktif).
- **Toplam Kurulu Kapasite:** **10 Gbps** (Tek ISP tamamen koptuğunda dahi 5 Gbps tek başına tüm pik yükü kesintisiz taşır).

---

### 3.2. Ankara Ofisi (Kamu Projeleri ve Savunma AR-GE)

| Birim / Kaynak | Kişi / Sistem | Ham Pik İhtiyaç (Mbps) | Açıklama |
|---|---:|---:|---|
| İdari ve Proje Personeli | 65 kişi | 127 Mbps | $65 \times 1.95\text{ Mbps}$ |
| Yazılım Ekibi | 85 kişi | 650 Mbps | $85 \times 7.65\text{ Mbps}$ (Kamu entegrasyon ve test ortamları) |
| Savunma AR-GE Ekibi | 55 kişi | 440 Mbps | Kontrollü/vekil sunucu (proxy) üzerinden onaylı paket ve veri akışı |
| Teknik Ekip | 25 kişi | 150 Mbps | $25 \times 6.00\text{ Mbps}$ |
| Kamu B2B VPN + İstanbul HQ Replikasyon | Yerel Sunucular | 250 Mbps | Şifreli IPsec tünel ve SIEM log aktarımı |
| Misafir Ağı | ~20 eşzamanlı | 40 Mbps | İzole misafir erişimi |
| **Alt Toplam (Ham Trafik)** | **230 Çalışan** | **1.657 Mbps** | |
| **Protokol Ek Yükü (%12) + Büyüme Payı (%30)** | — | **~2.413 Mbps (~2.41 Gbps)** | $1.657 \times 1.12 \times 1.30$ |

- **Tahsis Edilen Hat:** **1 × 2.5 Gbps Simetrik Metro Ethernet (Ana Hat)** + **1 × 1 Gbps Farklı Fiziksel Güzergah Fiber (Yedek/SD-WAN Aktif Hat)**.
- **Toplam Kurulu Kapasite:** **3.5 Gbps** (Ana hat kesintisinde kritik iş/VPN trafiği QoS ile 1 Gbps yedek hatta daralma olmadan aktarılır).

---

### 3.3. Londra Ofisi (UK ve EMEA Operasyon Merkezi)

| Birim / Kaynak | Kişi / Sistem | Ham Pik İhtiyaç (Mbps) | Açıklama |
|---|---:|---:|---|
| İdari, Satış ve Hukuk (GDPR) | 45 kişi | 88 Mbps | $45 \times 1.95\text{ Mbps}$ |
| Cloud-Native Yazılım ve SRE Ekibi | 50 kişi | 450 Mbps | AWS/Azure/GCP yoğunluklu geliştirme ($50 \times 9.0\text{ Mbps}$) |
| FinTech AR-GE Ekibi | 25 kişi | 272 Mbps | $25 \times 10.88\text{ Mbps}$ |
| Teknik Ekip | 15 kişi | 90 Mbps | $15 \times 6.00\text{ Mbps}$ |
| EMEA API Geçidi + İstanbul HQ VPN | Yerel/Bulut | 220 Mbps | GlobalProtect VPN ve merkez telemetri |
| Misafir Ağı | ~15 eşzamanlı | 30 Mbps | Ziyaretçi Wi-Fi |
| **Alt Toplam (Ham Trafik)** | **135 Çalışan** | **1.150 Mbps** | |
| **Protokol Ek Yükü (%12) + Büyüme Payı (%30)** | — | **~1.674 Mbps (~1.67 Gbps)** | $1.150 \times 1.12 \times 1.30$ |

- **Tahsis Edilen Hat:** **2 × 2 Gbps Dedicated Simetrik Leased Line** (ISP-1: BTnet + ISP-2: Colt Technology Services).
- **Toplam Kurulu Kapasite:** **4.0 Gbps** (Tam yedekli, tek hat arızasında 2 Gbps kesintisiz kapasite).

---

### 3.4. Akademik Şube (10.000 Öğrenci + 630 Akademik/İdari/Teknik Kadro)

Akademik yerleşkede ders araları ve vize/final haftalarında 10.000 öğrencinin oluşturduğu yoğun eşzamanlı oturum yükü, kapasite planlamasının ana belirleyicisidir.

| Birim / Kaynak | Kişi / Sistem | Ham Pik İhtiyaç (Mbps) | Açıklama |
|---|---:|---:|---|
| **Öğrenci Ağı (Eduroam + Lab)** | **10.000 öğrenci** | **13.000 Mbps (13.0 Gbps)** | Pik anda 6.500 aktif öğrenci (~11.700 cihaz), öğrenci başı ~2.0 Mbps (QoS ile P2P engelli) |
| **Akademisyen / Hoca Ağı** | **320 hoca** | **1.261 Mbps (1.26 Gbps)** | $320 \times 3.94\text{ Mbps}$ (Canlı yayın, makale/veri tabanı erişimi) |
| **İdari Personel** | **190 kişi** | **371 Mbps** | $190 \times 1.95\text{ Mbps}$ (Öğrenci işleri, kütüphane, mali işler) |
| **Kampüs Yazılım Ekibi** | **35 kişi** | **268 Mbps** | $35 \times 7.65\text{ Mbps}$ (OBS/LMS geliştirme) |
| **Akademik AR-GE ve HPC Kümesi** | **45 kişi** | **900 Mbps** | Akademik araştırma veri setleri ve üniversiteler arası işbirliği |
| **Kampüs Teknik / NOC Ekibi** | **40 kişi** | **240 Mbps** | $40 \times 6.00\text{ Mbps}$ |
| **Kampüs DMZ (LMS, OBS, Web, Eduroam)** | Kampüs Sunucuları | **1.500 Mbps (1.5 Gbps)** | Uzaktan eğitim öğrencilerinin kampüs LMS sunucularına dış erişimi |
| **Konferans / Kampüs Misafir Ağı** | ~250 eşzamanlı | **250 Mbps** | Sempozyum ve dış katılımcı erişimi |
| **Alt Toplam (Ham Trafik)** | **10.630 Kişi** | **17.790 Mbps (~17.79 Gbps)** | |
| **Protokol Ek Yükü (%10) + Büyüme Payı (%25)** | — | **~24.461 Mbps (~24.5 Gbps)** | $17.790 \times 1.10 \times 1.25$ |

- **Tahsis Edilen Hat:** **20 Gbps ULAKBİM / ULAKNET Simetrik Akademik Omurga Bağlantısı** + **10 Gbps Ticari ISP Metro Ethernet Yedek/İnternet Hattı**.
- **Toplam Kurulu Kapasite:** **30.0 Gbps**.

---

## 4. Güvenlik Duvarı Üzerinde QoS ve Trafik Şekillendirme (Traffic Shaping) Matrisi

Salt yüksek bant genişliği satın almak yeterli değildir; kontrolsüz bir yazılım imaj indirmesi veya öğrenci video trafiği, VoIP ses paketlerini veya kritik ERP/Sınav trafiğini boğabilir. Bu nedenle tüm güvenlik duvarlarında **8-Sınıflı DiffServ (DSCP) ve Garanti/Tavan (Guaranteed / Maximum Bandwidth)** politikası uygulanır:

| Öncelik Sınıfı | DSCP Etiketi | Trafik Kategorisi | Garanti Edilen Taban (%) | Tavan Sınırı (Max %) | Kuyruk / Düşürme Politikası |
|---|---|---|---|---|---|
| **Sınıf 1 (Gerçek Zamanlı)** | `EF (46)` | Kurumsal VoIP Ses, Acil Kampüs Anons | %10 | %20 | Strict Priority Queue (Düşük gecikme < 20 ms) |
| **Sınıf 2 (Ağ Kontrol ve HA)** | `CS6 (48)` | BGP, OSPF, IPsec IKE, SD-WAN Probe, SIEM TLS | %5 | %15 | Yüksek Öncelik (Asla düşürülmez) |
| **Sınıf 3 (İş Kritik / Sınav)** | `AF41 (34)` | ERP, Finans, OBS/LMS, Canlı Ders, B2B Kamu VPN | %25 | %60 | Garantili Bant Genişliği |
| **Sınıf 4 (Yazılım ve AR-GE)** | `AF31 (26)` | GitLab, CI/CD, Cloud API, Akademik Veri Tabanları | %25 | %70 | WRED (Adil Kuyruklama) |
| **Sınıf 5 (Kurumsal Standart)** | `AF21 (18)` | Kurumsal Web (HTTPS), E-posta, Bulut Ofis | %15 | %50 | Standart Kurumsal Kuyruk |
| **Sınıf 6 (Öğrenci Genel İnternet)** | `BE (0)` | Öğrenci Wi-Fi genel web ve eğitim dışı akış | %15 | %55 | Kişi başı **15 Mbps Tavan (Per-IP Shaper)** |
| **Sınıf 7 (Yedekleme ve Yama)** | `AF11 (10)` | Bulut yedekleme, OS güncellemeleri, büyük senkronizasyon | %5 | %25 | Mesai saatlerinde (08:30–18:00) %15 tavan sınırı |
| **Sınıf 8 (Misafir / Kısıtlı)** | `CS1 (8)` | Misafir Wi-Fi, sosyal medya/video (Akademik dışı) | %0 (Best Effort) | %10 | Kişi başı **5 Mbps Tavan**; sıkışıklıkta ilk düşürülen |

> [!TIP]
> **Akademik Şube Per-IP Shaper Kuralı:** 10.000 öğrencili ağda tek bir öğrencinin yüksek hızlı indirme yaparak bağlı bulunduğu Access Point veya NAT havuzunu tüketmesini engellemek için güvenlik duvarı üzerinde `Per-IP Traffic Shaper` tanımlanmıştır: Öğrenci başına anlık tavan **15 Mbps Download / 8 Mbps Upload**, eşzamanlı maksimum oturum sayısı ise **400 TCP/UDP session** ile sınırlandırılmıştır.

