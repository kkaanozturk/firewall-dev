# Yazılım Ekibi – Kişi Sayıları, VLAN ve Bant Genişliği

Yazılım ekibi dört şubede toplam **310 kişidir**. Trafiği (Git, CI/CD, konteyner imajı, bulut API) diğer ekiplerden belirgin biçimde yoğundur; bu yüzden ayrı VLAN, ayrı QoS sınıfı ve ayrı güvenlik katmanı gerekir. En iyi 5 güvenlik duvarı: [top5-firewall.md](top5-firewall.md). Kurulum ve testler: [kurulum-adimlari.md](kurulum-adimlari.md).

## Şube dağılımı

| Şube | Kişi | VLAN | Alt ağ | Ham pik (Mbps) | Formüllü (×1,12×1,30; akademik ×1,10×1,25) |
| --- | ---: | ---: | --- | ---: | ---: |
| İstanbul | 140 | 30 | 10.10.30.0/23 | 1.071 | 1.559 Mbps |
| Ankara | 85 | 30 | 10.20.30.0/24 | 650 | 946 Mbps |
| Londra | 50 | 30 | 10.30.30.0/24 | 450 | 655 Mbps |
| Akademik | 35 | 30 | 10.40.30.0/24 | 268 | 369 Mbps |
| **Toplam** | **310** | | | **2.439** | **≈ 3,5 Gbps** |

Kişi başı pik: 7,65 Mbps (eşzamanlılık %85, cihaz çarpanı 1,5, cihaz başı 6 Mbps).

## QoS ve politika

- QoS sınıfı 4 (AF31): garanti %25, tavan %70.
- Erişim: yalnızca kendi VLAN'ından GitLab/Harbor (İstanbul 10.10.50.40–42) ve onaylı bulut depolarına (GitHub, npm, PyPI, AWS) HTTPS.
- Londra yazılımı AWS eu-west-2 VPC'sine (172.28.0.0/16) SSH/HTTPS/Kubernetes ile erişir; API anahtarı sızıntısı DLP ile engellenir.
- İdari personel VLAN'ından yazılım VLAN'ına erişim reddedilir.
