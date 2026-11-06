<!-- Path: subeler/akademik-sube/firewall-kurallari.md -->
# Akademik Şube (10.000 Öğrenci, 320 Hoca, Sınav Ağı ve Kampüs) Güvenlik Duvarı Kural Seti

- **Hedef Cihaz:** `AKA-FW01 / FW02` — Fortinet FortiGate 1800F (FortiOS 7.4 LTS, Active-Passive HA, NP7 Hardware Offload)
- **Temel Kampüs Güvenlik İlkeleri:**
  1. **Öğrenci – İdare/Hoca Tam İzolasyonu:** 10.000 öğrencinin bulunduğu `STUDENT_WIFI (10.100.0.0/16)` ve `COMPUTER_LABS (10.40.80.0/21)` ağlarından Akademisyen (`FACULTY_NET`), İdari Personel (`ADMIN_STAFF`), İç Veritabanı (`CAMPUS_DC` - Not/OBS veritabanı!) ve Şirket Genel Merkezi VPN tüneline **doğrudan tek bir paket dahi geçemez**. Öğrenciler OBS ve LMS'e yalnızca `DMZ_EDU` içindeki **WAF (Web Application Firewall)** üzerinden HTTPS (TCP 443) ile erişir.
  2. **İstemciler Arası İzolasyon (Client Isolation):** Öğrencilerin birbirlerinin bilgisayarlarına ağ taraması veya ARP spoofing yapmaması için hem Kablosuz Erişim Noktalarında (AP Isolation) hem de güvenlik duvarı üzerinde *Intra-Zone Traffic Deny* aktiftir.
  3. **Sınav Salonu Kilidi (`EXAM_ISOLATED`):** Vize ve final sınavlarının yapıldığı bilgisayarlı sınav salonları yalnızca yerel Sınav Sunucusuna (`10.40.60.50`) erişebilir; dış internet, yapay zeka araçları ve mesajlaşma uygulamaları tamamen kapalıdır.
  4. **CGNAT PBA + 802.1X RSSO + 5651 İmzalama:** Her öğrencinin Eduroam kullanıcı adı (`ogrenci_no@kampus.edu.tr`), aldığı iç IP ve çıktığı genel CGNAT port bloğu eşleştirilerek TÜBİTAK Zaman Damgalı 5651 sunucusuna kaydedilir.

---

## 1. Sıralı Güvenlik Politikası Tablosu (Kampüs Rulebase)

| Sıra | Kural Adı | Kaynak Bölge / Alt Ağ | Kaynak Kimlik (RSSO/AD) | Hedef Bölge / Alt Ağ | İzin Verilen Uygulama / Servis | Güvenlik ve QoS Profilleri | İşlem | Log |
|---:|---|---|---|---|---|---|---|---|
| **1** | `BLK-WAN-BOGON-DDOS` | `WAN_ULAKNET`, `WAN_ISP` | Any | Any | USOM Zararlı IP + Botnet + Spoof | DoS Sensor (NP7 SYN Proxy) | **DROP** | SIEM |
| **2** | `ALLOW-NOC-MGMT-SIEM` | `CAMPUS_NOC` (`10.40.10.0/24`) | `AD\GRP_Campus_NOC` (2FA) | `CAMPUS_DC`, `DMZ_EDU`, `VPN_IST_SOC` | `SSH`, `HTTPS`, `SNMPv3`, `Syslog-TLS (6514)` | IPS-High | **ACCEPT** | SIEM |
| **3** | `DENY-STUDENT-TO-CORP-VPN` | `STUDENT_WIFI`, `COMPUTER_LABS`, `CAMPUS_GUEST` | Any | `VPN_IST_SOC` (`10.10.0.0/16`), `CAMPUS_DC`, `ADMIN_STAFF`, `FACULTY_NET`, `RND_HPC` | **ALL (Tam Kampüs İç İzolasyonu)** | — | **DENY** | **SIEM (İç Tarama Alarmı)** |
| **4** | `ALLOW-EXAM-ONLY-LMS` | `EXAM_ISOLATED` (`10.40.52.0/22`) | Sınav İstemci MAC/802.1X | `DMZ_EDU` (`10.40.60.50` Sınav Sunucusu) | `HTTPS (TCP 443)` | WAF-Protected, IPS | **ACCEPT** | SIEM + Sınav Denetim İzi |
| **5** | `DENY-EXAM-ALL-OTHER` | `EXAM_ISOLATED` | Any | `ANY` (İnternet dahil her yer) | `ALL` | — | **DENY** | **SIEM Alarm** |
| **6** | `ALLOW-STUDENT-TO-WAF-OBS` | `STUDENT_WIFI` (`10.100.0.0/16`), `COMPUTER_LABS` | `RADIUS\Role_Student` | `DMZ_EDU` (`10.40.60.10-20` FortiWeb WAF -> LMS/OBS) | `HTTPS (TCP 443)` | WAF OWASP Top-10, Rate-Limit | **ACCEPT** | SIEM |
| **7** | `DENY-STUDENT-P2P-TOR-MINE` | `STUDENT_WIFI`, `COMPUTER_LABS` | Any | `WAN_ULAKNET`, `WAN_ISP` | `BitTorrent`, `P2P`, `Tor`, `Psiphon/Proxy`, `Crypto-Mining`, `SMB/NetBIOS` | App-Control Block, DNS-Botnet | **DENY** | 5651 + SIEM |
| **8** | `ALLOW-STUDENT-INTERNET` | `STUDENT_WIFI` (`10.100.0.0/16`) | `RADIUS\Role_Student` | `WAN_ULAKNET`, `WAN_ISP` | `HTTP/HTTPS`, `DNS`, `QUIC/Zoom/Teams` (Eğitim & Genel Web) | WebFilter-Academic, AV, IPS, **Per-IP Shaper (15M DL / 8M UL, Max 400 Session)**, **PBA-CGNAT** | **ACCEPT** | **5651 Zaman Damgalı + SIEM** |
| **9** | `ALLOW-LABS-DEV-INTERNET` | `COMPUTER_LABS` (`10.40.80.0/21`) | `AD\Lab_Domain_Users` | `WAN_ULAKNET`, `WAN_ISP` | `HTTPS`, `SSH (GitHub/GitLab)`, `Paket Depoları` | Deep-SSL (Lab PC'lerde Sertifika Yüklü), AV, IPS, **30M Shaper** | **ACCEPT** | 5651 + SIEM |
| **10** | `ALLOW-FACULTY-ACADEMIC` | `FACULTY_NET` (`10.40.24.0/22` - 320 Hoca) | `RADIUS\Role_Faculty` | `DMZ_EDU`, `CAMPUS_DC` (Kütüphane/OBS Akademisyen Portalı), `WAN_ULAKNET` | `HTTPS`, `SSH`, `Zoom/Webex`, `IEEE/Scopus` | AV, IPS, DNS-Filter, QoS Garantili 2 Gbps (`AF41`) | **ACCEPT** | 5651 + SIEM |
| **11** | `ALLOW-ADMIN-STAFF-OBS` | `ADMIN_STAFF` (`10.40.20.0/23` - 190 Personel) | `AD\GRP_Campus_Admin` | `CAMPUS_DC` (`10.40.50.0/24`), `WAN_ISP` | `HTTPS`, `LDAPS`, `OBS-ERP`, `E-Devlet/YÖKSİS` | **Deep-SSL** (Resmi Kurumlar Muaf), DLP-StudentPII (Öğrenci KVKK Koruması), AV | **ACCEPT** | 5651 + SIEM |
| **12** | `ALLOW-CAMPUS-SWDEV` | `SW_CAMPUS` (`10.40.30.0/24` - 35 Yazılım) | `AD\GRP_Campus_Devs` | `CAMPUS_DC` (Staging), `DMZ_EDU`, `WAN_ULAKNET` | `Git`, `SSH`, `HTTPS`, `Docker` | Deep-SSL, IPS, AV | **ACCEPT** | SIEM + 5651 |
| **13** | `ALLOW-HPC-RND-ULAKNET` | `RND_HPC` (`10.40.40.0/24` - 45 AR-GE) | `AD\GRP_Academic_RnD` | `WAN_ULAKNET` (CERN / TRUBA / Akademik Grid) | `HTTPS`, `SSH`, `Globus-GridFTP` | IPS-High, AV (Yüksek Hızlı ASIC Offload) | **ACCEPT** | SIEM |
| **14** | `ALLOW-WAN-TO-DMZ-LMS` | `WAN_ULAKNET`, `WAN_ISP` | Any | `DMZ_EDU` (`10.40.60.10-25` LMS, OBS, Web, Eduroam RADIUS) | `HTTPS (443)`, `RADIUS (UDP 1812/1813 - Yalnızca ULAKBİM FLR Sunucuları)` | Inbound-SSL, WAF, DDoS Koruma | **ACCEPT** | SIEM |
| **15** | `ALLOW-GUEST-CONFERENCE` | `CAMPUS_GUEST` (`172.16.40.0/22`) | SMS / Etkinlik Kodu (5651) | `WAN_ISP` | `HTTP`, `HTTPS`, `DNS` | WebFilter-Strict, **Per-IP 5 Mbps Shaper** | **ACCEPT** | 5651 + SIEM |
| **99** | `EXPLICIT-DENY-CAMPUS-ALL` | `ANY` | `ANY` | `ANY` | `ALL` | — | **DENY** | **SIEM** |

---

## 2. FortiOS CLI Konfigürasyon Bloğu (Akademik Şube — FortiGate 1800F)

```text
# 1. 10.000 Öğrenci İçin Per-IP Bant Genişliği ve Oturum Sınırlayıcı (15 Mbps DL / 8 Mbps UL / 400 Max Session)
config firewall shaper per-ip-shaper
    edit "SHAPER_STUDENT_15M_400SES"
        set max-bandwidth 15360
        set bandwidth-unit kbps
        set max-concurrent-session 400
    next
    edit "SHAPER_GUEST_5M"
        set max-bandwidth 5120
        set bandwidth-unit kbps
        set max-concurrent-session 150
    next
end

# 2. NAT Port Tükenmesini Önleyen Donanım Hızlandırmalı PBA (Port Block Allocation) CGNAT Havuzu
config firewall ippool
    edit "CGNAT_POOL_STUDENTS_PBA"
        set type port-block-allocation
        set startip 193.140.210.65
        set endip 193.140.210.126
        set block-size 512
        set num-blocks-per-user 2
        set arp-reply enable
    next
end

# 3. P2P, BitTorrent, Tor, Proxy Atlama ve Kripto Madencilik Engelleme Uygulama Profili
config application list
    edit "APP_CTRL_CAMPUS_STUDENT"
        set other-application-log enable
        config entries
            edit 1
                set category 2 6
                set action block
                set comment "P2P (BitTorrent) ve Anonim Proxy/Tor Engeli"
            next
            edit 2
                set application 41668
                set action block
                set comment "Cryptomining Havuzları Engeli"
            next
            edit 3
                set category 1 3 5 12 15
                set action pass
                set comment "Eğitim, Web, Video Konferans ve Bulut Servisleri İzni"
            next
        end
    next
end

# 4. Öğrenci Wi-Fi Bölgesi İstemciler Arası Trafik Engeli (Intra-Zone Deny)
config system zone
    edit "ZONE_STUDENT_WIFI"
        set intrazone deny
        set interface "vlan100_stu_wifi"
    next
end

# 5. Kampüs Kritik Güvenlik Politikaları (İç İzolasyon, Sınav Kilidi ve PBA CGNAT Çıkışı)
config firewall policy
    edit 10
        set name "DENY-STUDENT-TO-INTERNAL-AND-HQVPN"
        set srcintf "ZONE_STUDENT_WIFI" "vlan80_labs" "vlan190_guest"
        set dstintf "vpn_ist_soc" "vlan50_campus_dc" "vlan20_admin" "vlan25_faculty" "vlan40_hpc"
        set srcaddr "all"
        set dstaddr "all"
        set action deny
        set schedule "always"
        set service "ALL"
        set logtraffic all
        set comments "10.000 Öğrenci ağından İdare, Hoca, OBS Veritabanı ve HQ VPN erişimi kesin yasaktır"
    next
    edit 20
        set name "ALLOW-EXAM-HALL-ONLY-LMS-SERVER"
        set srcintf "vlan55_exam"
        set dstintf "vlan60_dmz_edu"
        set srcaddr "NET_EXAM_HALLS_10_40_55_0"
        set dstaddr "VIP_EXAM_LMS_10_40_60_50"
        set action accept
        set schedule "always"
        set service "HTTPS"
        set utm-status enable
        set ips-sensor "high_security"
        set logtraffic all
    next
    edit 21
        set name "DENY-EXAM-HALL-ALL-INTERNET"
        set srcintf "vlan55_exam"
        set dstintf "any"
        set srcaddr "all"
        set dstaddr "all"
        set action deny
        set schedule "always"
        set service "ALL"
        set logtraffic all
    next
    edit 30
        set name "ALLOW-10K-STUDENTS-INTERNET-5651"
        set srcintf "ZONE_STUDENT_WIFI"
        set dstintf "sdwan_campus_uplinks"
        set srcaddr "NET_STUDENT_WIFI_10_100_0_0_16"
        set rsso enable
        set groups "RSSO_EDU_STUDENTS"
        set dstaddr "all"
        set action accept
        set schedule "always"
        set service "HTTP" "HTTPS" "DNS" "QUIC"
        set utm-status enable
        set av-profile "campus-av"
        set webfilter-profile "wf-academic-safe"
        set dnsfilter-profile "dns-botnet-sinkhole"
        set application-list "APP_CTRL_CAMPUS_STUDENT"
        set ips-sensor "campus-edge-ips"
        set per-ip-shaper "SHAPER_STUDENT_15M_400SES"
        set nat enable
        set ippool enable
        set poolname "CGNAT_POOL_STUDENTS_PBA"
        set logtraffic all
        set logtraffic-start enable
    next
    edit 999
        set name "EXPLICIT-DENY-CAMPUS-ALL"
        set srcintf "any"
        set dstintf "any"
        set srcaddr "all"
        set dstaddr "all"
        set action deny
        set schedule "always"
        set service "ALL"
        set logtraffic all
    next
end
```

