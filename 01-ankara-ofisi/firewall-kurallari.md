<!-- Path: 01-ankara-ofisi/firewall-kurallari.md -->
# Ankara Ofisi (Kamu B2B, Yazılım ve Savunma AR-GE ISFW) Güvenlik Duvarı Kural Seti

Ankara şubesinde iki kademeli kural seti çalışır:
1. **Katman 1 — Dış Sınır (`ANK-EXT-FW01/02` - Palo Alto PA-1420):** İnternet, İstanbul HQ IPsec tüneli, Kamu B2B VPN tünelleri, İdari ve Yazılım ağlarını denetler.
2. **Katman 2 — İç Savunma AR-GE Sınırı (`ANK-ISFW-01/02` - FortiGate 200F):** 55 kişilik Savunma/Kriptografi AR-GE ekibini (`10.20.40.0/24`) Ankara ofisinin geri kalanından dahi yalıtır.

---

## 1. Dış Sınır Güvenlik Duvarı Kural Tablosu (PA-1420)

| Sıra | Kural Adı | Kaynak Bölge / Alt Ağ | Kaynak Kimlik | Hedef Bölge / Alt Ağ | Uygulama / Servis | Güvenlik Profilleri | İşlem | Log |
|---:|---|---|---|---|---|---|---|---|
| **1** | `BLK-THREAT-SPOOF` | `UNTRUST_WAN` | Any | Any | Bilinen Zararlı IP + RFC1918 Spoof | EDL-USOM, DoS | **DROP** | SIEM |
| **2** | `ALLOW-HQ-IPSEC-TUNNEL` | `UNTRUST_WAN` (İstanbul HQ IP) | IKEv2 Cert | `UNTRUST_WAN` (Ankara WAN IP) | `ike`, `ipsec-esp-udp` | Strict-DoS | **ALLOW** | SIEM |
| **3** | `ALLOW-GOV-B2B-VPN` | `UNTRUST_WAN` (Beyaz Listeli Kamu Statik IP'leri) | Kurum Sertifikası | `DMZ_GOV` (`10.20.60.0/24`) | `ike`, `ipsec`, `ssl` (TCP 8443 mTLS) | Gov-IPS-Strict, AV | **ALLOW** | SIEM |
| **4** | `ALLOW-MGMT-FROM-HQ-SOC` | `VPN_HQ` (`10.10.10.0/24`), `IT_MGMT` (`10.20.10.0/24`) | `AD\GRP_SecOps_Admins` | `LOCAL_SRV`, FW Mgmt | `ssh`, `ssl`, `snmpv3`, `syslog` | Strict-IPS | **ALLOW** | SIEM |
| **5** | `ALLOW-SWDEV-TO-HQ-GIT` | `SW_DEV` (`10.20.30.0/24`) | `AD\GRP_ANK_Software` | `VPN_HQ` (`10.10.50.40-49`) | `git`, `ssh`, `ssl`, `docker` | IPS, AV, WildFire | **ALLOW** | SIEM |
| **6** | `ALLOW-SWDEV-WAN-DEV` | `SW_DEV` (`10.20.30.0/24`) | `AD\GRP_ANK_Software` | `UNTRUST_WAN` | `ssl`, `github-base`, `apt-get` | **SSL-Decrypt**, DLP-Code, AV | **ALLOW** | SIEM + 5651 |
| **7** | `ALLOW-MIRROR-REPO-SYNC` | `LOCAL_SRV` (`10.20.50.30` Paket Aynası) | Sistem Hesabı | `VPN_HQ` (`10.10.50.60`) | `ssl` (TCP 443 - İmzalı Paket Çekimi) | **SSL-Decrypt**, AV, WildFire | **ALLOW** | SIEM |
| **8** | `DENY-ALL-RND-WAN` | `RND_TRANSIT` (`10.20.40.0/24`) | Any | `UNTRUST_WAN` | **ANY (Savunma AR-GE Doğrudan İnternete Çıkamaz)** | — | **DROP** | **SIEM Kritik Alarm** |
| **9** | `ALLOW-CORP-STAFF-ERP` | `CORP_STAFF` (`10.20.20.0/24`) | `AD\GRP_ANK_Staff` | `VPN_HQ` (`10.10.50.0/24`), `UNTRUST_WAN` | `sap`, `ms-office365`, `ssl`, `web-browsing` | **SSL-Decrypt**, URL-Corp, AV | **ALLOW** | SIEM + 5651 |
| **10** | `ALLOW-GUEST-ISOLATED` | `GUEST_WIFI` (`172.16.20.0/24`) | 5651 Captive Portal | `UNTRUST_WAN` | `web-browsing`, `ssl`, `dns` | URL-Guest, QoS 4 Mbps/IP | **ALLOW** | 5651 + SIEM |
| **99** | `EXPLICIT-DENY-ALL` | `ANY` | `ANY` | `ANY` | `ANY` | — | **DROP** | **SIEM** |

---

## 2. İç Segmentasyon Güvenlik Duvarı Kural Tablosu (Savunma AR-GE ISFW — FortiGate 200F)

Savunma AR-GE bölgesi (`VLAN 40 - 10.20.40.0/24`) için **Varsayılan Hava Boşluğuna Yakın (Near Air-Gap) Mikro-Segmentasyon** kuralları:

| Sıra | Kural Adı | Kaynak | Hedef | İzin Verilen Servis | Güvenlik Denetimi | İşlem |
|---:|---|---|---|---|---|---|
| **1** | `ALLOW-RND-LOCAL-AD-NTP` | `RND_DEFENSE` (`10.20.40.0/24`) | `LOCAL_SRV` (`10.20.50.10` RODC/NTP/DNS) | `LDAPS (636)`, `Kerberos (88)`, `DNS (53)`, `NTP (123)` | IPS-Strict, DNS-Filter | **ACCEPT** |
| **2** | `ALLOW-RND-SIGNED-REPO` | `RND_DEFENSE` (`10.20.40.0/24`) | `LOCAL_SRV` (`10.20.50.30` Yerel İmzalı Kütüphane Aynası) | `HTTPS (TCP 443)` — Yalnızca iç aynadan salt-okunur indirme | Deep-SSL, AV, DLP (Upload Block) | **ACCEPT** |
| **3** | `ALLOW-BASTION-VDI-ADMIN` | `IT_MGMT` (`10.20.10.50` Kayıtlı Bastion Sunucu) | `RND_DEFENSE` | `SSH (22)` / `RDP (3389)` (Ekran kaydı + 2FA zorunlu, pano/dosya kopyalama kapalı) | IPS-Strict | **ACCEPT** |
| **4** | `DENY-CORP-SW-TO-RND` | `CORP_STAFF`, `SW_DEV`, `DMZ_GOV`, `GUEST_WIFI` | `RND_DEFENSE` | `ALL` | — | **DENY + LOG** |
| **5** | `DENY-RND-OUTBOUND-ALL` | `RND_DEFENSE` | `ALL` (İnternet ve diğer tüm şubeler dahil) | `ALL` | — | **DENY + LOG** |

---

## 3. FortiOS CLI Konfigürasyonu (Ankara Savunma AR-GE ISFW — FortiGate 200F)

```text
config firewall address
    edit "NET_RND_DEFENSE"
        set subnet 10.20.40.0 255.255.255.0
    next
    edit "SRV_LOCAL_RODC_DNS"
        set subnet 10.20.50.10 255.255.255.255
    next
    edit "SRV_SIGNED_REPO_MIRROR"
        set subnet 10.20.50.30 255.255.255.255
    next
    edit "HOST_SECOPS_BASTION"
        set subnet 10.20.10.50 255.255.255.255
    next
end

config dlp sensor
    edit "DLP-DEFENSE-NO-EXFIL"
        set comment "Savunma AR-GE kaynak kod ve teknik çizim çıkış engeli"
        config entries
            edit 1
                set type file
                set file-type 1
                set action block
            next
        end
    next
end

config firewall policy
    edit 101
        set name "ALLOW-RND-LOCAL-AD-NTP"
        set srcintf "vlan40_rnd"
        set dstintf "vlan50_srv"
        set srcaddr "NET_RND_DEFENSE"
        set dstaddr "SRV_LOCAL_RODC_DNS"
        set action accept
        set schedule "always"
        set service "LDAPS" "KERBEROS" "DNS" "NTP"
        set utm-status enable
        set ips-sensor "high_security"
        set logtraffic all
    next
    edit 102
        set name "ALLOW-RND-SIGNED-REPO"
        set srcintf "vlan40_rnd"
        set dstintf "vlan50_srv"
        set srcaddr "NET_RND_DEFENSE"
        set dstaddr "SRV_SIGNED_REPO_MIRROR"
        set action accept
        set schedule "always"
        set service "HTTPS"
        set utm-status enable
        set ssl-ssh-profile "deep-inspection"
        set av-profile "strict-av"
        set dlp-sensor "DLP-DEFENSE-NO-EXFIL"
        set ips-sensor "high_security"
        set logtraffic all
    next
    edit 103
        set name "ALLOW-BASTION-VDI-ADMIN"
        set srcintf "vlan10_mgmt"
        set dstintf "vlan40_rnd"
        set srcaddr "HOST_SECOPS_BASTION"
        set dstaddr "NET_RND_DEFENSE"
        set action accept
        set schedule "always"
        set service "SSH"
        set utm-status enable
        set ips-sensor "high_security"
        set logtraffic all
    next
    edit 199
        set name "EXPLICIT-DENY-RND-ALL"
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

