<!-- Path: dokumanlar/02-top5-guvenlik-duvari-ve-tercih-analizi.md -->
# Şube ve Birim Bazlı En İyi Top 5 Güvenlik Duvarı Tavsiyeleri ve Tercih Analizi

Bu doküman, **AegisNova Global**'in dört ana yerleşkesi (İstanbul, Ankara, Londra, Akademik Şube) ve kritik iç parçaları (**Yazılım Ekibi Bulut/CI-CD Katmanı** ile **AR-GE İç Segmentasyon Katmanı**) için en uygun **Top 5 Yeni Nesil Güvenlik Duvarı (NGFW)** adaylarını gerçek dünya **Threat Protection / SSL Inspection bant genişliği** tahminleriyle karşılaştırır. Ayrıca her senaryoda **diğer güvenlik duvarlarının neden tercih edilmediğini** teknik ve mali gerekçeleriyle açıklar.

---

## 1. Değerlendirme Kriterleri: Pazarlama Verisi ve Gerçek Dünya Kapasitesi

Üretici kataloglarında yer alan "Firewall Throughput" değeri yalnızca Katman-3/Katman-4 paket başlığı kontrolünü (SPI) ifade eder. Modern kurumsal ağlarda trafiğin **%85'ten fazlası TLS 1.3 ile şifrelidir**. Güvenlik duvarı üzerinde **IPS + Application Control + Anti-Malware + URL Filtering + TLS/SSL Inspection** aynı anda açıldığında kapasite %50 ila %75 oranında düşer.

Bu çalışmadaki tüm donanım seçimleri aşağıdaki altın kurallara göre yapılmıştır:
1. **Gerçek Tehdit Önleme Kapasitesi (Threat Protection Throughput):** Şubenin toplam WAN kapasitesinin **en az 1.2–1.5 katı** olmalıdır.
2. **Eşzamanlı Oturum (Concurrent Sessions) ve Saniyedeki Yeni Oturum (New Sessions/sec):** Özellikle 10.000 öğrencili Akademik Şube için darboğaz Gbps değil, milyonlarca eşzamanlı NAT oturumudur.
3. **Merkezi Yönetim ve SOC Entegrasyonu:** Panorama / FortiManager desteği ve CEF/Syslog TLS üzerinden SIEM uyumu.

---

## 2. İstanbul Ofisi (Genel Merkez ve Ana DC — 10 Gbps WAN) İçin Top 5 Tavsiye

- **Hedef Kapasite:** 10 Gbps kurulu WAN (4.85 Gbps hesaplanan pik yük), yoğun SSL Inspection ve VPN Hub sonlandırma.

| Sıra | Model | L4 Firewall Throughput | NGFW / Threat Protection Throughput | Tahmini SSL Inspection Throughput | Eşzamanlı Oturum | Durum |
|---:|---|---:|---:|---:|---:|---|
| **1** | **Palo Alto Networks PA-3430** | 38.2 Gbps | **12.7 Gbps** | **~8.5 Gbps** | 3.0 Milyon | **SEÇİLDİ (Ana Çözüm)** |
| **2** | Fortinet FortiGate 600F | 139 Gbps | 10.5 Gbps | 9.0 Gbps | 8.0 Milyon | Alternatif #2 |
| **3** | Check Point Quantum 6700 | 42 Gbps | 7.6 Gbps (Maestro ile ölçeklenir) | ~5.8 Gbps | 4.0 Milyon | Alternatif #3 |
| **4** | Cisco Secure Firewall 3130 | 42 Gbps | 15.2 Gbps | ~6.4 Gbps | 4.0 Milyon | Alternatif #4 |
| **5** | Sophos XGS 5500 | 100 Gbps | 14.0 Gbps | 13.5 Gbps | 18.0 Milyon | Alternatif #5 |

### İstanbul İçin Neden PA-3430 Seçildi ve Diğerleri Neden Tercih Edilmedi?
- **Neden PA-3430 Seçildi:** Genel Merkez'de 140 kişilik yazılım ekibi, 65 kişilik AR-GE ve Ana Veri Merkezi bulunmaktadır. Palo Alto'nun tek geçişli mimarisi (*Single-Pass Architecture*), **App-ID** ve **User-ID** doğruluğu ile **Panorama** merkezi yönetimi kurumsal merkez, Ankara ve Londra şubelerini tek ekrandan yönetmek için endüstri standardıdır.
- **Diğerleri Neden Tercih Edilmedi:**
  - *FortiGate 600F:* Donanım performansı çok güçlü olmasına rağmen, son yıllarda FortiOS SSL-VPN ve yönetim arayüzünde yaşanan kritik CVE geçmişi nedeniyle dış sınırda (HQ Perimeter) Palo Alto tercih edilmiş; Fortinet ise Ankara iç segmentasyonunda (ISFW) ve Akademik Şube'de konumlandırılmıştır.
  - *Check Point Quantum 6700:* Tek cihazda Threat Prevention kapasitesi (7.6 Gbps), 10 Gbps çift ISP hattının tam yük altında şifreli trafiğini karşılamada üst sınıra fazla yaklaşmaktadır; bir üst seri (6900/7000) ise TCO'yu (Toplam Sahip Olma Maliyeti) %35 artırmaktadır.
  - *Cisco Secure Firewall 3130:* Snort 3 motoru güçlüdür ancak Firepower Management Center (FMC) kural dağıtım süreleri ve kullanıcı arayüzü ergonomisi, günde onlarca CI/CD ve mikro-segmentasyon değişikliği yapılan dinamik yazılım merkezinde operasyonel yavaşlık yaratmaktadır.
  - *Sophos XGS 5500:* KOBİ ve orta ölçekte başarılı olmakla birlikte, çok uluslu finans/AR-GE regülasyonlarında granüler BGP/SD-WAN ve ileri seviye tehdit avcılığı telemetrisinde Tier-1 kurumsal rakiplerinin gerisinde kalmaktadır.

---

## 3. Ankara Ofisi (Kamu ve Savunma AR-GE — 3.5 Gbps WAN) İçin Top 5 Tavsiye

- **Hedef Kapasite:** 3.5 Gbps toplam WAN (2.41 Gbps pik yük), Kamu B2B IPsec tünelleri ve yüksek güvenlikli AR-GE trafiği.

| Sıra | Model | L4 Firewall Throughput | NGFW / Threat Protection Throughput | Tahmini SSL Inspection Throughput | Eşzamanlı Oturum | Durum |
|---:|---|---:|---:|---:|---:|---|
| **1** | **Palo Alto Networks PA-1420** | 9.8 Gbps | **4.8 Gbps** | **~3.2 Gbps** | 1.4 Milyon | **SEÇİLDİ (Dış Sınır)** |
| **2** | Fortinet FortiGate 400F | 79.5 Gbps | 9.0 Gbps | 8.0 Gbps | 7.8 Milyon | Alternatif #2 |
| **3** | Check Point Quantum 6400 | 32 Gbps | 3.9 Gbps | ~2.9 Gbps | 2.0 Milyon | Alternatif #3 |
| **4** | Cisco Secure Firewall 3110 | 18 Gbps | 8.9 Gbps | ~3.8 Gbps | 2.0 Milyon | Alternatif #4 |
| **5** | WatchGuard Firebox M690 | 29.7 Gbps | 3.4 Gbps | 2.9 Gbps | 9.2 Milyon | Alternatif #5 |

### Ankara İçin Tercih ve Elenme Gerekçeleri
- **Neden PA-1420 Seçildi:** İstanbul Genel Merkez'deki Panorama sunucusuyla %100 uyumlu tekil politika yönetimi sağlar, 4.8 Gbps Threat Prevention kapasitesi 3.5 Gbps WAN hattını rahatlıkla taşır ve savunma projelerinde istenen sıkı uygulama kontrolünü sunar. *(Not: Savunma AR-GE'nin iç ağ ayrımı için içeride ayrıca FortiGate 200F kullanılarak çift üreticili savunma kurulmuştur.)*
- **Diğerleri Neden Tercih Edilmedi:**
  - *FortiGate 400F:* Dış sınırda kullanılsaydı İstanbul ve Londra (Palo Alto) ile tekil şablon yönetimi bozulacak, iki ayrı merkezi yönetim konsolunda sınır politikası senkronize edilmek zorunda kalınacaktı.
  - *Check Point Quantum 6400:* SSL Inspection açıkken ~2.9 Gbps sınırına inmesi, 3.5 Gbps tam hat kullanımında darboğaz riski taşımaktadır.
  - *Cisco 3110:* Merkezde FMC altyapısı bulunmadığı için tek şube adına ayrı FMC lisansı ve sunucusu kurmak mali açıdan verimsizdir.
  - *WatchGuard M690:* Savunma sanayii B2B entegrasyonlarında ve granüler uygulama imzalarında kurumsal uyumluluk beklentilerini tam karşılamamaktadır.

---

## 4. Londra Ofisi (UK ve EMEA Operasyon — 4 Gbps WAN) İçin Top 5 Tavsiye

- **Hedef Kapasite:** 4.0 Gbps Dual Leased Line (1.67 Gbps pik yük), Cloud-Native DevOps/SRE ve UK-GDPR DLP denetimi.

| Sıra | Model | L4 Firewall Throughput | NGFW / Threat Protection Throughput | Tahmini SSL Inspection Throughput | Eşzamanlı Oturum | Durum |
|---:|---|---:|---:|---:|---:|---|
| **1** | **Palo Alto Networks PA-1410** | 8.6 Gbps | **3.5 Gbps** | **~2.5 Gbps** | 980 Bin | **SEÇİLDİ (Ana Çözüm)** |
| **2** | Fortinet FortiGate 200F | 27 Gbps | 3.0 Gbps | 4.0 Gbps | 3.0 Milyon | Alternatif #2 |
| **3** | Juniper Networks SRX1600 | 25 Gbps | 4.6 Gbps | ~3.1 Gbps | 2.0 Milyon | Alternatif #3 |
| **4** | Cisco Secure Firewall 1220CX | 9.0 Gbps | 4.5 Gbps | ~2.2 Gbps | 600 Bin | Alternatif #4 |
| **5** | Check Point Quantum 6200 | 17.5 Gbps | 2.5 Gbps | ~1.8 Gbps | 2.0 Milyon | Alternatif #5 |

### Londra İçin Tercih ve Elenme Gerekçeleri
- **Neden PA-1410 Seçildi:** Enterprise DLP (Data Loss Prevention) modülü sayesinde UK-GDPR kapsamındaki kişisel verilerin yetkisiz bulut depolama servislerine sızmasını katman-7 seviyesinde engeller ve İstanbul Panorama üzerinden merkezi denetim izi (audit trail) üretir.
- **Diğerleri Neden Tercih Edilmedi:**
  - *FortiGate 200F:* Kurumsal sınır standartlaştırması (Panorama) dışına çıkmamak adına elenmiştir.
  - *Juniper SRX1600:* Yönlendirme (BGP/MPLS) yetenekleri mükemmeldir ancak L7 DLP ve kullanıcı bazlı SaaS görünürlüğü (CASB) Palo Alto kadar olgun değildir.
  - *Cisco 1220CX ve Check Point 6200:* SSL Inspection altındaki kapasite düşüşü ve ayrı yönetim konsolu maliyeti nedeniyle tercih edilmemiştir.

---

## 5. Akademik Şube (10.000 Öğrenci Kampüsü — 30 Gbps WAN) İçin Top 5 Tavsiye

- **Hedef Kapasite:** 30 Gbps toplam WAN (20 Gbps ULAKNET + 10 Gbps ISP, ~24.5 Gbps pik yük), **12.000+ eşzamanlı aktif cihaz**, milyonlarca NAT/CGNAT oturumu, ders başlangıç saatlerinde saniyede yüz binlerce yeni oturum isteği.

| Sıra | Model | L4 Firewall Throughput | NGFW / Threat Protection Throughput | Tahmini SSL Inspection Throughput | Eşzamanlı Oturum / Yeni Oturum | Durum |
|---:|---|---:|---:|---:|---:|---|
| **1** | **Fortinet FortiGate 1800F** | 198 Gbps | **31.0 Gbps** | **20.0 Gbps** | **12 Milyon / 450K sn** | **SEÇİLDİ (Kampüs Omurga)** |
| **2** | Palo Alto Networks PA-5420 | 78.6 Gbps | 30.2 Gbps | ~19.5 Gbps | 5.7 Milyon / 390K sn | Alternatif #2 |
| **3** | Check Point Quantum 16200 | 96 Gbps | 24.0 Gbps | ~16.0 Gbps | 10 Milyon / 340K sn | Alternatif #3 |
| **4** | Juniper Networks SRX4600 | 120 Gbps | 22.0 Gbps | ~14.5 Gbps | 20 Milyon / 450K sn | Alternatif #4 |
| **5** | Cisco Secure Firewall 4115 | 80 Gbps | 21.4 Gbps | ~12.0 Gbps | 15 Milyon / 250K sn | Alternatif #5 |

### Akademik Şube İçin Neden FortiGate 1800F Seçildi ve Diğerleri Neden Tercih Edilmedi?
- **Neden FortiGate 1800F Seçildi:** 10.000 öğrencinin aynı anda Wi-Fi (Eduroam) üzerinden internete çıktığı bir kampüste en büyük teknik risk **NAT Port Tükenmesi (Port Exhaustion)** ve **Conntrack / Oturum Tablosu Taşmasıdır**. FortiGate 1800F üzerindeki özel **NP7 (Network Processor 7) ASIC** yongaları, donanım hızlandırmalı CGNAT (Port Block Allocation), 12 Milyon eşzamanlı oturum, 40G/100G QSFP28 fiber portlar ve 31 Gbps Threat Protection kapasitesini rakip çözümlere kıyasla **yaklaşık %45 daha düşük 5 yıllık lisans ve donanım maliyetiyle** sunmaktadır.
- **Diğerleri Neden Tercih Edilmedi:**
  - *Palo Alto PA-5420:* Teknik olarak harika bir cihazdır ancak 30 Gbps kapasitede donanım + Threat Prevention + URL Filtering + DNS Security aboneliklerinin 5 yıllık toplam maliyeti eğitim bütçesi için aşırı yüksektir; ayrıca eşzamanlı oturum kapasitesi (5.7M), 1800F'in (12M) yarısından azdır.
  - *Check Point 16200, Juniper SRX4600 ve Cisco 4115:* Tam Threat Protection açıkken 21–24 Gbps bandında kalarak kampüsün 24.5 Gbps pik ihtiyacının ve 30 Gbps hat kapasitesinin altında kalmaktadır (darboğaz oluşturur).

---

## 6. Fonksiyonel Parçalar İçin Özel Top 5 Güvenlik Duvarı Tavsiyeleri

Şube dış sınırlarına ek olarak, iç ağdaki kritik "parçalar" (AR-GE ve Yazılım/Bulut) için değerlendirilen en iyi 5 çözüm aşağıdadır:

### 6.1. AR-GE Ekibi İç Segmentasyon Güvenlik Duvarı (ISFW — Doğu/Batı Trafiği)

Ankara (55 Savunma AR-GE) ve İstanbul (65 AI AR-GE) laboratuvarlarını kurum içi diğer VLAN'lardan ayırmak için yüksek port hızlı ve düşük gecikmeli ISFW adayları:

| Sıra | Model | Threat Protection / IPS | Neden Seçildi veya Neden Elenerek Tercih Edilmedi? |
|---:|---|---:|---|
| **1** | **Fortinet FortiGate 200F (SEÇİLDİ)** | **3.0 Gbps TP / 5.0 Gbps IPS** | Dış sınırda Palo Alto varken içeride Fortinet kullanılarak **Çift Üreticili Savunma (Dual-Vendor Defense-in-Depth)** sağlandı; 10GE SFP+ portlarla düşük gecikmeli mikro-segmentasyon sunar. |
| **2** | Palo Alto PA-460 | 2.4 Gbps TP | Sınır güvenlik duvarıyla aynı üretici olduğu için ortak sıfır-gün (0-day) riskini bertaraf etmez; 10G SFP+ portu yoktur. |
| **3** | Cisco Secure Firewall 1220CX | 4.5 Gbps TP | Güçlü donanım ancak ayrı FMC yönetim yükü nedeniyle elendi. |
| **4** | Check Point Quantum 3800 | 1.5 Gbps TP | Büyük veri seti aktarımlarında (Doğu-Batı 10 Gbps) 1.5 Gbps darboğaz yaratır. |
| **5** | Netgate 8200 (pfSense Plus) | ~1.8 Gbps (Suricata ile) | Açık kaynak tabanlı çözümlerde entegre ticari sıfır-gün kum havuzu (Sandbox), merkezi L7 uygulama imzası ve 4 saatte yerinde donanım değişim (RMA SLA) garantisi bulunmadığından kritik savunma AR-GE'de tercih edilmedi. |

### 6.2. Yazılım Ekibi Bulut, Konteyner ve Web Uygulama Katmanı (Virtual NGFW & WAF)

310 kişilik toplam yazılım ekibinin CI/CD hatlarını, Kubernetes kümelerini ve dışa açık API/LMS uygulamalarını korumak için değerlendirilen Top 5 sanal/bulut güvenlik duvarı:

| Sıra | Çözüm |Tahmini Kapasite | Durum ve Gerekçe |
|---:|---|---:|---|
| **1** | **Palo Alto VM-Series / CN-Series + Cloudflare Enterprise WAF (SEÇİLDİ)** | **2 – 16 Gbps (vCPU ölçekli)** | Merkez Panorama ile tekil kural yönetimi, Kubernetes Pod seviyesinde L7 görünürlük ve DDoS/OWASP Top 10 koruması sağlar. |
| **2** | Fortinet FortiGate-VM + FortiWeb | 2 – 20 Gbps | Güçlü alternatif; Akademik Şube LMS/OBS önünde **FortiWeb WAF** olarak konumlandırılmıştır. |
| **3** | AWS Network Firewall + AWS WAF | Bulut Yerel (Oto-Ölçek) | Yalnızca AWS içinde çalıştığı, çoklu bulut (Multi-Cloud) ve şirket içi (On-Prem) sunucuları kapsamadığı için ana çözüm olarak seçilmedi. |
| **4** | Check Point CloudGuard Network | 2 – 15 Gbps | Mevcut Panorama yönetim ekosistemiyle bütünleşmediği için elendi. |
| **5** | Tigera Calico Enterprise (eBPF) | 10+ Gbps (Kernel düzeyi) | Kubernetes içi Doğu-Batı politikalarında çok hızlıdır ancak tam teşekküllü IPS/Antimalware/TLS Inspection yapamadığı için NGFW yerine yalnızca yardımcı CNI katmanı olarak kullanılır. |

---

## 7. Genel Değerlendirme: Neden Standart Yönlendirici, L3 Switch ACL veya Açık Kaynak Güvenlik Duvarlarını Ana Omurgada Tercih Etmedik?

Üniversite ve kurumsal projelerde sıkça sorulan *"Neden maliyetli ticari NGFW yerine L3 Switch ACL, Linux `nftables` veya `OPNsense/pfSense` kullanmadık?"* sorusunun mühendislik yanıtı 5 maddede özetlenmektedir:

1. **Durumsuz (Stateless) ve Katman-4 Sınırlılığı:** L3 Switch ACL'leri yalnızca IP ve port numarasına bakar; TCP 443 (HTTPS) portundan tünel açan bir komuta-kontrol (C2) zararlısını veya BitTorrent/Tor trafiğini ayırt edemez.
2. **Donanımsal ASIC Hızlandırma Eksikliği:** Standart x86 sunucu üzerinde koşan açık kaynak güvenlik duvarlarında 30 Gbps kampüs trafiğinde Suricata/Snort IPS + TLS Inspection açıldığında CPU darboğazı ve paket düşmeleri (packet drop) yaşanır.
3. **Kimlik Farkındalığı (User-ID / 802.1X):** IP adresleri DHCP ile sürekli değişir. Kurumsal NGFW'ler Active Directory ve FreeRADIUS ile gerçek zamanlı konuşarak kuralı IP'ye değil, *"Yazılım Ekibi Grubu"* veya *"Öğrenci Rolü"*ne yazar.
4. **Sıfır-Gün Kum Havuzu (Cloud Sandbox) ve Tehdit İstihbaratı:** WildFire / FortiGuard gibi küresel sensör ağlarından her 5 dakikada bir güncellenen sıfır-gün imzaları açık kaynak listelerde bu hızda ve doğrulukta bulunmaz.
5. **7/24 Kurumsal SLA ve Regülatif Denetim:** 11.425 kişilik bir yapıda donanım arızasında 4 saat içinde yerinde parça değişimi (Next-Business-Day / 4-Hour Onsite RMA) ve ISO 27001 / KVKK / UK-GDPR denetçilerinin kabul ettiği sertifikalı (Common Criteria / FIPS 140-3) donanım zorunludur.

