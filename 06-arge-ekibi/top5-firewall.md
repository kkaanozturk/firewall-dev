# AR-GE Ekibi – En İyi 5 Güvenlik Duvarı Adayı ve Seçim

AR-GE ekibi dört şubede toplam 190 kişidir (İstanbul 65, Ankara 55, Londra 25, Akademik 45). Ankara savunma AR-GE bölgesi iç segmentasyon güvenlik duvarının (ISFW) arkasındadır.

Ankara (55 Savunma AR-GE) ve İstanbul (65 AI AR-GE) laboratuvarlarını kurum içi diğer VLAN'lardan ayırmak için yüksek port hızlı ve düşük gecikmeli ISFW adayları:

| Sıra | Model | Threat Protection / IPS | Neden Seçildi veya Neden Elenerek Tercih Edilmedi? |
|---:|---|---:|---|
| **1** | **Fortinet FortiGate 200F (SEÇİLDİ)** | **3.0 Gbps TP / 5.0 Gbps IPS** | Dış sınırda Palo Alto varken içeride Fortinet kullanılarak **Çift Üreticili Savunma (Dual-Vendor Defense-in-Depth)** sağlandı; 10GE SFP+ portlarla düşük gecikmeli mikro-segmentasyon sunar. |
| **2** | Palo Alto PA-460 | 2.4 Gbps TP | Sınır güvenlik duvarıyla aynı üretici olduğu için ortak sıfır-gün (0-day) riskini bertaraf etmez; 10G SFP+ portu yoktur. |
| **3** | Cisco Secure Firewall 1220CX | 4.5 Gbps TP | Güçlü donanım ancak ayrı FMC yönetim yükü nedeniyle elendi. |
| **4** | Check Point Quantum 3800 | 1.5 Gbps TP | Büyük veri seti aktarımlarında (Doğu-Batı 10 Gbps) 1.5 Gbps darboğaz yaratır. |
| **5** | Netgate 8200 (pfSense Plus) | ~1.8 Gbps (Suricata ile) | Açık kaynak tabanlı çözümlerde entegre ticari sıfır-gün kum havuzu (Sandbox), merkezi L7 uygulama imzası ve 4 saatte yerinde donanım değişim (RMA SLA) garantisi bulunmadığından kritik savunma AR-GE'de tercih edilmedi. |

