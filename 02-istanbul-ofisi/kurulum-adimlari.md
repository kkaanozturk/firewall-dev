# İstanbul Ofisi – Şubeye Özel Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md). Bu dosya yalnızca İstanbul'a özgü ayarları ve kabul testlerini içerir.

## Cihazlar

| Cihaz | Görev | Konfigürasyon |
| --- | --- | --- |
| 2 × Palo Alto PA-3430 (HA) | Sınır, VPN hub, DMZ, Panorama ile üç şubenin merkezi yönetimi | [pa3430-running.conf](configs/pa3430-running.conf) |

## Şubeye özel ayarlar

1. **Hatlar:** `ethernet1/1` Türk Telekom 5 Gbps, `ethernet1/2` Superonline 5 Gbps; BGP + SD-WAN aktif/aktif.
2. **LACP:** `ae1` = `ethernet1/21` + `ethernet1/22` (2 × 25G) merkez omurgaya; VLAN 10, 20, 30, 40, 50, 60, 70, 90.
3. **HA:** HA1 `ethernet1/19`, HA2 `ethernet1/20`; öncelik 90/110, preemption kapalı.
4. **VPN hub:** Ankara (`tunnel.20`), Londra (`tunnel.30`), Akademik (`tunnel.40`, yalnızca log/yönetim).
5. **Panorama:** Ankara ve Londra kural şablonları buradan dağıtılır.
6. **SSL inceleme:** Personel, yazılım ve AR-GE giden HTTPS çözülür; finans, sağlık ve kamu kategorileri muaf (KVKK).
7. **AR-GE:** Doğrudan internet yok, yalnızca iç vekil sunucu (10.10.50.60:8443), yükleme DLP ile engelli.
8. **SIEM:** 10.10.50.25 (syslog TLS 6514), 5651 imzalayıcı 10.10.50.26.

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | PA-3430 birincil gücü kesilir | İkincil < 1 sn devralır, SSH ve IPsec düşmez |
| 2 | ISP-1 çıkarılır | Trafik ISP-2'ye geçer, VoIP ve VPN kesilmez |
| 3 | Yazılım VLAN'ından GitLab (10.10.50.40:443) | İzin |
| 4 | AR-GE VLAN'ından 1.1.1.1:443 | Reddet + kritik alarm |
| 5 | Personel VLAN'ından AR-GE VLAN'ına erişim | Reddet + yanal hareket alarmı |
| 6 | Dış ağdan RFC1918 kaynaklı paket | Anti-spoofing ile düşer |
| 7 | EICAR dosyasını HTTPS ile indirme | SSL inceleme + antivirüs engeller |
| 8 | Misafir VLAN'ında hız testi | Kişi başı 5 Mbps |
