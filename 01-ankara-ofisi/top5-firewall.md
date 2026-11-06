# Ankara Ofisi – En İyi 5 Güvenlik Duvarı Adayı ve Seçim


### A) Dış Sınır (Perimeter — 3.5 Gbps WAN) Top 5 Tavsiyesi

| Sıra | Model | L4 Firewall Throughput | Threat Protection Throughput | Tahmini SSL Inspection | Karar ve Gerekçe |
|---:|---|---:|---:|---:|---|
| **1** | **Palo Alto PA-1420** | 9.8 Gbps | **4.8 Gbps** | **~3.2 Gbps** | **SEÇİLDİ (Dış Sınır):** Merkez Panorama uyumu ve 3.5 Gbps hatta tam kapasite. |
| **2** | Fortinet FortiGate 400F | 79.5 Gbps | 9.0 Gbps | 8.0 Gbps | Dış sınırda Panorama bütünlüğünü bozmamak için seçilmedi. |
| **3** | Check Point Quantum 6400 | 32.0 Gbps | 3.9 Gbps | ~2.9 Gbps | SSL Inspection kapasitesi 3.5 Gbps hatta sınırda kaldığı için elendi. |
| **4** | Cisco Secure Firewall 3110 | 18.0 Gbps | 8.9 Gbps | ~3.8 Gbps | Ayrı FMC yönetim sunucusu maliyeti nedeniyle tercih edilmedi. |
| **5** | WatchGuard Firebox M690 | 29.7 Gbps | 3.4 Gbps | 2.9 Gbps | Kamu/Savunma B2B regülasyon derinliği açısından elendi. |

### B) Savunma AR-GE İç Segmentasyon (ISFW — Doğu/Batı 10 Gbps) Top 5 Tavsiyesi

| Sıra | Model | Threat Protection / IPS | Karar ve Gerekçe |
|---:|---|---:|---|
| **1** | **Fortinet FortiGate 200F** | **3.0 Gbps TP / 5.0 Gbps IPS** | **SEÇİLDİ (İç AR-GE ISFW):** Dış sınırda Palo Alto varken içeride farklı üretici (Fortinet ASIC + 10GE SFP+) kullanılarak tekil sıfır-gün zafiyeti riski sıfırlandı. |
| **2** | Palo Alto PA-460 | 2.4 Gbps TP | Dış sınırla aynı üretici olduğu (monokültür riski) ve 10G SFP+ portu bulunmadığı için elendi. |
| **3** | Cisco Secure Firewall 1220CX | 4.5 Gbps TP | Ek yönetim konsolu yükü nedeniyle tercih edilmedi. |
| **4** | Check Point Quantum 3800 | 1.5 Gbps TP | AR-GE içi büyük imaj aktarımlarında darboğaz oluşturacağı için elendi. |
| **5** | Netgate 8200 (pfSense Plus) | ~1.8 Gbps | 4-saat yerinde donanım değişim SLA'i ve FIPS/CC sertifikalı donanım hızlandırması olmadığı için tercih edilmedi. |

