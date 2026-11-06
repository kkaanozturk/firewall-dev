# Akademik Şube – Şubeye Özel Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md). Bu dosya yalnızca kampüse özgü ayarları ve kabul testlerini içerir.

## Cihazlar

| Cihaz | Görev | Konfigürasyon |
| --- | --- | --- |
| 2 × FortiGate 1800F (HA) | Kampüs omurga güvenlik duvarı, CGNAT, 802.1X RSSO | [fg1800f-campus.conf](configs/fg1800f-campus.conf) |

## Şubeye özel ayarlar

1. **Hatlar:** `port25` ULAKNET 20 Gbps, `port26` ticari ISP 10 Gbps. Akademik veri tabanları ve ULAKBİM trafiği ULAKNET'e, genel video/CDN trafiği QoS ile ISP'ye yönlendirilir.
2. **LACP:** `agg_campus_core` = `port37` + `port38` (2 × 40G); VLAN 10, 20, 25, 30, 40, 50, 55, 60, 80, 100, 190.
3. **HA:** 2 × 40G senkron, `session-pickup` açık (12 milyon oturum).
4. **Öğrenci ağı:** Bölge içi trafik kapalı (`intrazone deny`), erişim noktalarında istemci izolasyonu.
5. **CGNAT:** 193.140.210.65–126 havuzu, port bloğu 512, kullanıcı başına 2 blok; her blok 5651 için zaman damgalı loglanır.
6. **802.1X RSSO:** Eduroam FreeRADIUS (10.40.50.15) hesap bilgisi güvenlik duvarına gönderilir; kural IP'ye değil öğrenci rolüne bağlanır.
7. **Kişi başı sınır:** Öğrenci 15 Mbps indirme, 8 Mbps yükleme, 400 oturum; lab 30 Mbps; misafir 5 Mbps.
8. **Sınav salonları (VLAN 55):** Yalnızca sınav sunucusu 10.40.60.50:443; sınav sırasında internet kapalı.
9. **DMZ:** FortiWeb WAF arkasında LMS ve OBS; öğrenciler yalnızca buraya HTTPS ile ulaşır.
10. **İstanbul tüneli:** Yalnızca NOC ve SIEM alt ağları; öğrenci ağı tünele dahil değildir.

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | Öğrenciden OBS veritabanına (10.40.50.10:5432) erişim | Reddet + iç tarama alarmı |
| 2 | Öğrenciden BitTorrent (6881) | Reddet |
| 3 | Öğrenci hız testi | 15 Mbps'de sınırlanır |
| 4 | Aynı öğrenciden 400'den fazla oturum | Fazlası reddedilir |
| 5 | Sınav salonundan sınav sunucusu | İzin |
| 6 | Sınav salonundan 8.8.8.8:443 | Reddet + gözetmen alarmı |
| 7 | Öğrenciden aynı IP'ye art arda 5 port taraması | SOAR dinamik kara listeye alır |
| 8 | Öğrenciden öğrenciye doğrudan erişim | Reddet (istemci izolasyonu) |
| 9 | HA devretme | < 1 sn, oturum kaybı yok |
| 10 | 5651 log zinciri doğrulaması | Zincir geçerli, TSA yanıtı alınıyor |
