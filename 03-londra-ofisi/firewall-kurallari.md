<!-- Path: 03-londra-ofisi/firewall-kurallari.md -->
# Londra Ofisi (EMEA Operasyon, Cloud-Native Yazılım ve UK-GDPR) Güvenlik Duvarı Kural Seti

- **Hedef Cihaz:** `LON-EMEA-FW01 / FW02` — Palo Alto Networks PA-1410 (PAN-OS 11.1 LTS, Active-Passive HA)
- **Özel Tasarım Kriteri — UK-GDPR Sınır Ötesi Veri Koruma (Data Residency):** `VLAN 50 (UK_GDPR_SRV)` içindeki Birleşik Krallık vatandaşlarına ait ham kişisel veritabanlarının Birleşik Krallık / AEA dışına (İstanbul veya Ankara alt ağları dahil!) toplu aktarımı güvenlik duvarı seviyesinde engellenir; İstanbul SOC'ye yalnızca **anonimleştirilmiş/maskelenmiş güvenlik logları** gönderilir.

---

## 1. Sıralı Güvenlik Politikası Tablosu (Security Policy Rulebase)

| Sıra | Kural Adı | Kaynak Bölge / Alt Ağ | Kaynak Kimlik | Hedef Bölge / Alt Ağ | Uygulama / Servis | Güvenlik Profilleri | İşlem | Log |
|---:|---|---|---|---|---|---|---|---|
| **1** | `BLK-THREAT-NCSC-EDL` | `UNTRUST_WAN` | Any | Any | UK NCSC + küresel C2 / Tor / Botnet IP listesi | EDL-NCSC, DoS | **DROP** | SIEM |
| **2** | `ALLOW-IPSEC-HQ-AWS` | `UNTRUST_WAN` (İstanbul HQ + AWS `eu-west-2` VGW) | IKEv2 Cert | `UNTRUST_WAN` (Londra WAN) | `ike`, `ipsec-esp-udp` | Strict-DoS | **ALLOW** | SIEM |
| **3** | `DENY-UKGDPR-CROSSBORDER-DB` | `UK_GDPR_SRV` (`10.30.50.0/24`) | Any | `VPN_HQ` (`10.10.0.0/16`), `UNTRUST_WAN` | `postgresql`, `mysql`, `mssql-db`, `ftp`, `smb`, `rsync` | **UK-GDPR Veri Yerleşim Kilidi** | **DENY** | **DPO + SOC Kritik Alarm** |
| **4** | `ALLOW-PSEUDONYMIZED-SIEM` | `UK_GDPR_SRV` (`10.30.50.25` Maskeli Log Toplayıcı) | Sistem Sertifikası | `VPN_HQ` (`10.10.50.25` İstanbul SIEM) | `syslog` (TCP 6514 TLS) | Strict-IPS | **ALLOW** | SIEM |
| **5** | `ALLOW-SWCLOUD-TO-AWS-UK` | `SW_CLOUD` (`10.30.30.0/24`) | `AD\GRP_UK_CloudEng` (MFA) | `VPN_AWS_LON` (`172.28.0.0/16`) | `ssh`, `ssl`, `kubernetes`, `aws-api` | **SSL-Decrypt**, IPS, DLP-Secrets (API Key sızıntı engeli) | **ALLOW** | SIEM |
| **6** | `ALLOW-SWCLOUD-HQ-GIT` | `SW_CLOUD` (`10.30.30.0/24`) | `AD\GRP_UK_CloudEng` | `VPN_HQ` (`10.10.50.40` GitLab) | `git`, `ssh`, `ssl` | IPS, AV, WildFire | **ALLOW** | SIEM |
| **7** | `ALLOW-FINTECH-RND-SANDBOX` | `RND_FINTECH` (`10.30.40.0/24`) | `AD\GRP_UK_FinTech_RnD` | `UNTRUST_WAN` (Onaylı UK OpenBanking Sandbox FQDN) | `ssl` (TCP 443) | **SSL-Decrypt**, **DLP-UK-Financial-PII**, AV, WildFire | **ALLOW** | SIEM |
| **8** | `DENY-RND-UNAPPROVED-WAN` | `RND_FINTECH` (`10.30.40.0/24`) | Any | `UNTRUST_WAN`, `CORP_STAFF`, `GUEST_WIFI` | Any | — | **DENY** | SIEM |
| **9** | `ALLOW-CORP-SAAS-DLP` | `CORP_STAFF` (`10.30.20.0/24`) | `AD\GRP_UK_Staff` | `UNTRUST_WAN` | `ms-office365`, `salesforce`, `zoom`, `ssl`, `web-browsing` | **SSL-Decrypt**, **DLP-UK-GDPR** (Kişisel Bulut Yükleme Engeli: WeTransfer/Mega Drop), AV | **ALLOW** | Maskeli Yerel Log |
| **10** | `ALLOW-GLOBALPROTECT-ZTNA` | `UNTRUST_WAN` (UK/AB Coğrafi Bölgesi) | `AD\GRP_UK_Remote` + Cihaz Sertifikası + MFA | `DMZ_EMEA` (`10.30.60.15` GP Gateway) | `pan-os-global-protect`, `ssl` | HIP-Check (Disk Şifreleme & EDR Aktiflik Kontrolü) | **ALLOW** | SIEM |
| **11** | `ALLOW-GUEST-WIFI-UK` | `GUEST_WIFI` (`172.16.30.0/24`) | Captive Portal | `UNTRUST_WAN` | `web-browsing`, `ssl`, `dns` | URL-Guest, DNS-Sec, QoS 5 Mbps/IP | **ALLOW** | Yerel Log |
| **99** | `EXPLICIT-DENY-ALL-LOG` | `ANY` | `ANY` | `ANY` | `ANY` | — | **DROP** | **SIEM** |

---

## 2. PAN-OS CLI Konfigürasyon Bloğu (Londra PA-1410)

```text
configure

# UK-GDPR Veri Yerleşimi Kuralı: UK kişisel veri sunucularından Türkiye veya dış dünyaya veritabanı/dosya akışını kesin olarak engelle
set rulebase security rules DENY-UKGDPR-CROSSBORDER-DB from UK_GDPR_SRV to [ VPN_HQ UNTRUST_WAN ] source 10.30.50.0/24 destination any application [ postgresql mysql mssql-db oracle smb ftp rsync ] service any action deny log-start yes log-end yes log-setting LOCAL-DPO-AND-SIEM

# Yalnızca maskelenmiş (Pseudonymized) güvenlik loglarının İstanbul SIEM'e TLS üzerinden aktarımına izin ver
set rulebase security rules ALLOW-PSEUDONYMIZED-SIEM from UK_GDPR_SRV to VPN_HQ source 10.30.50.25 destination 10.10.50.25 application syslog service TCP-6514-TLS profile-setting group PG-STRICT-IPS action allow log-end yes log-setting LOCAL-DPO-AND-SIEM

# Cloud-Native Yazılım Ekibi (50 Kişi): AWS London VPC erişimi ve AWS Secret Key DLP denetimi
set rulebase security rules ALLOW-SWCLOUD-TO-AWS-UK from SW_CLOUD to VPN_AWS_LON source 10.30.30.0/24 source-user "AEGISNOVA\\GRP_UK_CloudEng" destination 172.28.0.0/16 application [ ssh ssl kubernetes ] service application-default profile-setting group PG-CLOUD-DLP-SECRETS action allow log-end yes log-setting LOCAL-DPO-AND-SIEM

# FinTech AR-GE Ekibi (25 Kişi): Yalnızca onaylı OpenBanking test API adreslerine çıkış; diğer dış çıkışlar kapalı
set rulebase security rules ALLOW-FINTECH-RND-SANDBOX from RND_FINTECH to UNTRUST_WAN source 10.30.40.0/24 source-user "AEGISNOVA\\GRP_UK_FinTech_RnD" destination FQDN-UK-OPENBANKING-SANDBOX application ssl service service-https profile-setting group PG-UKGDPR-DLP action allow log-end yes log-setting LOCAL-DPO-AND-SIEM

set rulebase security rules DENY-RND-UNAPPROVED-WAN from RND_FINTECH to [ UNTRUST_WAN CORP_STAFF GUEST_WIFI ] source 10.30.40.0/24 destination any application any service any action deny log-end yes log-setting LOCAL-DPO-AND-SIEM

# Kurumsal Personel (45 Kişi): Onaysız dosya paylaşım sitelerine (Shadow IT) yükleme engeli
set rulebase security rules ALLOW-CORP-SAAS-DLP from CORP_STAFF to UNTRUST_WAN source 10.30.20.0/24 source-user "AEGISNOVA\\GRP_UK_Staff" destination any application [ ms-office365 salesforce zoom web-browsing ssl ] service application-default profile-setting group PG-UKGDPR-DLP action allow log-end yes log-setting LOCAL-DPO-AND-SIEM

# En Alt Açık Reddetme ve Loglama
set rulebase security rules EXPLICIT-DENY-ALL-LOG from any to any source any destination any application any service any action drop log-end yes log-setting LOCAL-DPO-AND-SIEM

commit description "LON-EMEA UK-GDPR Compliant Rulebase v1.0 - CAB Ticket #SEC-1089"
```

