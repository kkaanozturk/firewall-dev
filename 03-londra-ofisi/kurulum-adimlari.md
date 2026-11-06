# Londra Ofisi – Şubeye Özel Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md). Bu dosya yalnızca Londra'ya özgü ayarları ve kabul testlerini içerir.

## Cihazlar

| Cihaz | Görev | Konfigürasyon |
| --- | --- | --- |
| 2 × Palo Alto PA-1410 (HA) | Sınır, UK-GDPR DLP, AWS ve İstanbul tünelleri | [pa1410-ukgdpr.conf](configs/pa1410-ukgdpr.conf) |

## Şubeye özel ayarlar

1. **Hatlar:** `ethernet1/1` BTnet 2 Gbps, `ethernet1/2` Colt 2 Gbps (özel hat, farklı fiziksel güzergah).
2. **LACP:** İç omurgaya 2 × 10G; VLAN 10, 20, 30, 40, 50 (UK-GDPR), 60, 70, 90.
3. **Tüneller:** İstanbul `tunnel.10`; AWS eu-west-2 VPC (172.28.0.0/16) `tunnel.99`.
4. **UK-GDPR kilidi:** VLAN 50'den İstanbul ve internete veritabanı/dosya protokolleri (PostgreSQL, MySQL, SMB, FTP, rsync) reddedilir.
5. **Log maskeleme:** Ham kimlikler Londra'daki toplayıcıda kalır; İstanbul SIEM'e kullanıcı adı maskelenmiş (SHA-256) log gider. Ham kimlik yalnızca DPO ve SOC yöneticisinin çift onayıyla açılır.
6. **Yazılım:** AWS ve İstanbul GitLab erişimi; API anahtarı sızıntısı DLP ile engellenir.
7. **FinTech AR-GE:** Yalnızca onaylı sandbox FQDN'lerine HTTPS.
8. **Uzaktan erişim:** GlobalProtect + MFA + cihaz sağlığı kontrolü (disk şifreleme, EDR).

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | UK-GDPR sunucusundan İstanbul'a PostgreSQL 5432 | Reddet + DPO ve SOC alarmı |
| 2 | Maskeli toplayıcıdan SIEM'e syslog 6514 | İzin, kimlikler maskeli |
| 3 | Yazılım VLAN'ından AWS VPC SSH/HTTPS | İzin |
| 4 | Personelin kişisel bulut sitesine dosya yüklemesi | DLP engeller |
| 5 | ISP-1 kesintisi | ISP-2'ye geçiş |
| 6 | HA devretme | < 1 sn, oturum kaybı yok |
