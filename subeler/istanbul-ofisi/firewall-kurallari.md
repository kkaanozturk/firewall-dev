<!-- Path: subeler/istanbul-ofisi/firewall-kurallari.md -->
# İstanbul Ofisi (HQ, Core DC, Yazılım ve AR-GE) Güvenlik Duvarı Kural Seti

- **Hedef Cihaz:** `IST-HQ-FW01 / FW02` — Palo Alto Networks PA-3430 (PAN-OS 11.1 LTS, Active-Passive HA)
- **Tasarım İlkesi:** Sıfır Güven (*Zero Trust*), Kimlik Bazlı Erişim (*Active Directory User-ID*), Uygulama Katmanı Denetimi (*App-ID*) ve Açık Kayıtlı Reddetme (*Explicit Deny & Log*).

---

## 1. Sıralı Güvenlik Politikası Tablosu (Security Policy Rulebase)

Kurallar yukarıdan aşağıya (*Top-Down First Match*) işlenir:

| Sıra | Kural Adı | Kaynak Bölge / Alt Ağ | Kaynak Kimlik (AD Grubu) | Hedef Bölge / Alt Ağ | Uygulama (App-ID) / Port | Güvenlik Profilleri | İşlem | Log |
|---:|---|---|---|---|---|---|---|---|
| **1** | `BLK-THREAT-GEO-EDL` | `UNTRUST_WAN` (Any) | Any | Any | Any | EDL (USOM + Tor + Bilinen C2 IP Listesi) | **DROP** | SIEM + Başlangıç |
| **2** | `BLK-ANTISPOOF-RFC1918` | `UNTRUST_WAN` (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) | Any | Any | Any | Anti-Spoofing Koruma | **DROP** | SIEM |
| **3** | `ALLOW-MGMT-SOC-JUMP` | `IT_MGMT` (`10.10.10.0/24`) | `AD\GRP_SecOps_Admins` (MFA Zorunlu) | `CORE_DC`, Tüm FW Mgmt | `ssh`, `ssl`, `pan-os-panorama`, `snmpv3` | Strict-IPS, File-Block | **ALLOW** | SIEM (Oturum Sonu + Komut) |
| **4** | `ALLOW-IPSEC-BRANCHES` | `UNTRUST_WAN` (Ankara, Londra, Akademik Statik WAN IP) | Machine Cert | `UNTRUST_WAN` (HQ WAN IP) | `ike`, `ipsec-esp-udp` (UDP 500/4500, ESP) | DoS-Protection | **ALLOW** | SIEM |
| **5** | `ALLOW-BRANCH-SIEM-INGEST` | `VPN_TUNNEL` (`10.20.50.0/24`, `10.30.50.0/24`, `10.40.50.0/24`) | Sistem Sertifikası | `CORE_DC` (`10.10.50.25` - SIEM) | `syslog` (TLS/6514), `netflow` | Strict-IPS | **ALLOW** | SIEM |
| **6** | `ALLOW-CORP-TO-ERP-AD` | `CORP_STAFF` (`10.10.20.0/23`) | `AD\GRP_HQ_Staff` | `CORE_DC` (`10.10.50.0/24`) | `active-directory`, `kerberos`, `dns`, `sap`, `ssl` | Corp-AV, Strict-IPS | **ALLOW** | SIEM |
| **7** | `DENY-CORP-TO-SW-RND` | `CORP_STAFF` | Any | `SW_DEV`, `RND_SEC` | Any | — | **DENY** | SIEM (Yanal Hareket Alarmı) |
| **8** | `ALLOW-SWDEV-CICD-GIT` | `SW_DEV` (`10.10.30.0/23`) | `AD\GRP_Software_Eng` | `CORE_DC` (`10.10.50.40-49` GitLab/Harbor/Vault) | `git`, `ssl`, `ssh`, `docker` | Dev-IPS, AV-Malware, WildFire | **ALLOW** | SIEM |
| **9** | `ALLOW-SWDEV-CLOUD-REPOS` | `SW_DEV` (`10.10.30.0/23`) | `AD\GRP_Software_Eng` | `UNTRUST_WAN` | `github-base`, `aws-console`, `npmjs`, `pypi`, `apt-get`, `ssl` (TCP 443) | **SSL-Decrypt**, AV, WildFire, Anti-Spyware, DLP-SourceCode | **ALLOW** | SIEM + 5651 |
| **10** | `ALLOW-RND-PROXY-ONLY` | `RND_SEC` (`10.10.40.0/24`) | `AD\GRP_RnD_Researchers` | `CORE_DC` (`10.10.50.60` Güvenli Artefakt Vekili) | `web-browsing`, `ssl` (TCP 8443) | **SSL-Decrypt**, Strict-IPS, **DLP-Strict-Block-Upload**, WildFire | **ALLOW** | SIEM (Tam Denetim) |
| **11** | `DENY-RND-DIRECT-WAN` | `RND_SEC` (`10.10.40.0/24`) | Any | `UNTRUST_WAN`, `CORP_STAFF`, `GUEST_WIFI` | Any | — | **DENY** | SIEM (Yüksek Öncelikli Alarm) |
| **12** | `ALLOW-CORP-INTERNET` | `CORP_STAFF` (`10.10.20.0/23`) | `AD\GRP_HQ_Staff` | `UNTRUST_WAN` | `web-browsing`, `ssl`, `ms-office365`, `zoom`, `slack` | **SSL-Decrypt** (Finans/Sağlık Muaf), URL-Corp, AV, DNS-Sec | **ALLOW** | SIEM + 5651 |
| **13** | `ALLOW-VOICE-SIP-QOS` | `VOICE_IOT` (`10.10.70.0/23`) | MAC/802.1X Doğrulamalı | `CORE_DC` (`10.10.50.70` IP-PBX) | `sip`, `rtp` | VoIP-IPS, QoS Class-1 (`EF`) | **ALLOW** | SIEM |
| **14** | `DENY-IOT-TO-INTERNAL` | `VOICE_IOT` | Any | `CORP_STAFF`, `SW_DEV`, `RND_SEC`, `CORE_DC` | Any | — | **DENY** | SIEM |
| **15** | `ALLOW-WAN-TO-DMZ-WAF` | `UNTRUST_WAN` | Any | `DMZ_PUB` (`10.10.60.10-20` WAF / API GW / GlobalProtect) | `ssl` (TCP 443), `pan-os-global-protect` | Inbound-SSL-Inspect, Strict-IPS, DoS-Zone-Profile | **ALLOW** | SIEM |
| **16** | `ALLOW-DMZ-TO-APP-DB` | `DMZ_PUB` (`10.10.60.10` WAF) | Servis Hesabı | `CORE_DC` (`10.10.50.80` App Kümesi) | `ssl` ( yalnızca TCP 8443 — Doğrudan DB erişimi YASAK) | Strict-IPS, SQLi/XSS Koruma | **ALLOW** | SIEM |
| **17** | `ALLOW-GUEST-INTERNET` | `GUEST_WIFI` (`172.16.10.0/23`) | Captive Portal (SMS/TCKN 5651) | `UNTRUST_WAN` | `web-browsing`, `ssl`, `dns` (P2P/Tor/SMB Engelli) | URL-Guest, DNS-Sinkhole, QoS 5 Mbps/IP | **ALLOW** | 5651 İmza + SIEM |
| **18** | `DENY-GUEST-RFC1918` | `GUEST_WIFI` | Any | Tüm İç Bölgeler (`10.0.0.0/8`) | Any | — | **DENY** | SIEM |
| **99** | `EXPLICIT-DENY-ALL-LOG` | `ANY` | `ANY` | `ANY` | `ANY` | — | **DROP** | **SIEM (Zorunlu Log)** |

---

## 2. Doğrudan Uygulanabilir PAN-OS CLI Konfigürasyon Bloğu

```text
configure

# Güvenlik Profilleri Grubu Tanımı (Strict Zero-Trust)
set profile-group PG-ZERO-TRUST virus AV-STRICT spyware AS-STRICT-SINKHOLE vulnerability IPS-CRITICAL-HIGH url-filtering URL-CORP-POLICY file-blocking FB-BLOCK-EXE-SCR wildfire-analysis WF-PUBLIC-CLOUD

# Kural 1: Dış Tehdit İstihbaratı (USOM & Bilinen C2) Engelleme
set rulebase security rules BLK-THREAT-GEO-EDL from UNTRUST_WAN to any source [ EDL-USOM-MALICIOUS EDL-TOR-EXIT-NODES ] destination any application any service any action drop log-start yes log-end yes log-setting SIEM-TLS-PROFILE

# Kural 8 & 9: Yazılım Ekibi (140 Kişi) İç GitLab/Harbor ve Kontrollü Bulut Erişimi
set rulebase security rules ALLOW-SWDEV-CICD-GIT from SW_DEV to CORE_DC source 10.10.30.0/23 source-user "AEGISNOVA\\GRP_Software_Eng" destination [ 10.10.50.40 10.10.50.41 10.10.50.42 ] application [ git ssh ssl docker ] service application-default profile-setting group PG-ZERO-TRUST action allow log-end yes log-setting SIEM-TLS-PROFILE

set rulebase security rules ALLOW-SWDEV-CLOUD-REPOS from SW_DEV to UNTRUST_WAN source 10.10.30.0/23 source-user "AEGISNOVA\\GRP_Software_Eng" destination any application [ github-base aws-console npmjs pypi apt-get ssl ] service service-https profile-setting group PG-ZERO-TRUST action allow log-end yes log-setting SIEM-TLS-PROFILE

# Kural 10 & 11: AR-GE Ekibi (65 Kişi) Doğrudan İnternet Yasağı ve İzole Vekil Zorunluluğu
set rulebase security rules ALLOW-RND-PROXY-ONLY from RND_SEC to CORE_DC source 10.10.40.0/24 source-user "AEGISNOVA\\GRP_RnD_Researchers" destination 10.10.50.60 application [ ssl web-browsing ] service TCP-8443 profile-setting group PG-ZERO-TRUST action allow log-end yes log-setting SIEM-TLS-PROFILE

set rulebase security rules DENY-RND-DIRECT-WAN from RND_SEC to [ UNTRUST_WAN CORP_STAFF GUEST_WIFI ] source 10.10.40.0/24 destination any application any service any action deny log-start yes log-end yes log-setting SIEM-TLS-PROFILE

# Kural 17 & 18: Misafir Ağı İzolasyonu ve 5651 Loglamalı İnternet
set rulebase security rules DENY-GUEST-RFC1918 from GUEST_WIFI to [ IT_MGMT CORP_STAFF SW_DEV RND_SEC CORE_DC DMZ_PUB VOICE_IOT ] source 172.16.10.0/23 destination any application any service any action deny log-end yes log-setting SIEM-TLS-PROFILE

set rulebase security rules ALLOW-GUEST-INTERNET from GUEST_WIFI to UNTRUST_WAN source 172.16.10.0/23 destination any application [ web-browsing ssl dns ] service application-default profile-setting group PG-ZERO-TRUST action allow log-start yes log-end yes log-setting SIEM-5651-PROFILE

# Kural 99: En Alt Açık Reddetme ve Loglama Kuralı (Sessiz Düşürme Bırakılmaz)
set rulebase security rules EXPLICIT-DENY-ALL-LOG from any to any source any destination any application any service any action drop log-start no log-end yes log-setting SIEM-TLS-PROFILE

# Dinamik Kaynak NAT (DIPP) Tanımı
set rulebase nat rules NAT-OUTBOUND-ISP1 from [ CORP_STAFF SW_DEV GUEST_WIFI ] to UNTRUST_WAN source any destination any service any source-translation dynamic-ip-and-port translated-address [ 198.51.100.10/30 ]

commit description "IST-HQ Zero-Trust Rulebase v1.0 - CAB Ticket #SEC-1042"
```

