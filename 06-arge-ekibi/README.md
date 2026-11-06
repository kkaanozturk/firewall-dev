# AR-GE Ekibi – Kişi Sayıları, VLAN ve Bant Genişliği

AR-GE ekibi dört şubede toplam **190 kişidir** ve en hassas veriyi (savunma, kriptografi, finansal modeller, akademik araştırma) taşır. Bu yüzden her şubede ayrı ve kapalı bir bölgedir. En iyi 5 güvenlik duvarı: [top5-firewall.md](top5-firewall.md). Kurulum ve testler: [kurulum-adimlari.md](kurulum-adimlari.md).

## Şube dağılımı

| Şube | Kişi | VLAN | Alt ağ | Ham pik (Mbps) | Formüllü | Erişim modeli |
| --- | ---: | ---: | --- | ---: | ---: | --- |
| İstanbul | 65 | 40 | 10.10.40.0/24 | 707 | 1.029 Mbps | Yalnızca iç vekil sunucu (8443) |
| Ankara (savunma) | 55 | 40 | 10.20.40.0/24 | 440 | 641 Mbps | FG-200F ISFW arkasında, internet yok |
| Londra (FinTech) | 25 | 40 | 10.30.40.0/24 | 272 | 396 Mbps | Yalnızca onaylı sandbox FQDN'leri |
| Akademik (HPC) | 45 | 40 | 10.40.40.0/24 | 900 | 1.238 Mbps | ULAKNET araştırma ağı |
| **Toplam** | **190** | | | **2.319** | **≈ 3,3 Gbps** | |

Kişi başı pik: 10,88 Mbps (eşzamanlılık %85, cihaz çarpanı 1,6, cihaz başı 8 Mbps).

## QoS ve politika

- QoS sınıfı 4 (AF31); akademik HPC için ULAKNET öncelikli.
- Hiçbir şubede AR-GE doğrudan internete çıkamaz (Akademik HPC yalnızca ULAKNET araştırma uçlarına gider).
- Çıkış DLP ile denetlenir: kaynak kod, şema, teknik çizim ve FPGA bitstream yükleme engellenir.
- Diğer VLAN'lardan AR-GE'ye erişim reddedilir; yönetim yalnızca kayıtlı bastion üzerinden (2FA) yapılır.
- Ankara'da çift üreticili koruma: dışta Palo Alto PA-1420, içte Fortinet FG-200F.
