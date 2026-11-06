# Kaynaklar

## Cihaz hızları hakkında önemli not

Bu projedeki güvenlik duvarı modelleri ve hız değerleri **yaklaşık sınıf göstergesidir**. Değerler üreticilerin yayımladığı veri sayfalarındaki ürün sınıflarına dayanır; satın almadan önce üreticinin güncel veri sayfasından doğrulanmalıdır. Veri sayfalarında dikkat edilecek satır "firewall throughput" değil, IPS + uygulama kontrolü + antivirüs (+ SSL inceleme) açıkken verilen **threat protection / NGFW throughput** değeridir.

## Üretici ürün sayfaları ve veri sayfaları

| Ürün ailesi | Kullanıldığı yer | Kaynak (üretici) |
| --- | --- | --- |
| Palo Alto Networks PA-3400, PA-1400, PA-5400, PA-400 serisi | İstanbul, Ankara, Londra | paloaltonetworks.com (Resources → Datasheets) |
| Palo Alto Panorama, VM-Series, CN-Series | Merkezi yönetim, yazılım ekibi bulut katmanı | paloaltonetworks.com |
| Fortinet FortiGate 1800F, 600F, 400F, 200F | Akademik şube, Ankara ISFW | fortinet.com (Products → FortiGate datasheet) |
| Fortinet FortiWeb, FortiManager | Kampüs WAF, merkezi yönetim | fortinet.com |
| Check Point Quantum 6000/16000 serisi | Karşılaştırma adayı | checkpoint.com |
| Cisco Secure Firewall 1200/3100/4100 serisi | Karşılaştırma adayı | cisco.com |
| Juniper SRX serisi | Karşılaştırma adayı | juniper.net |
| Sophos XGS, WatchGuard Firebox, Netgate | Karşılaştırma adayı | sophos.com, watchguard.com, netgate.com |
| Cloudflare WAF, AWS WAF / Network Firewall | Web ve bulut koruması | cloudflare.com, aws.amazon.com |

Analist raporu olarak Gartner Magic Quadrant (Network Firewalls / Hybrid Mesh Firewall) üretici seçimini destekleyen ikincil kaynaktır.

## Standartlar ve mevzuat

| Konu | Kaynak |
| --- | --- |
| 5651 sayılı kanun | mevzuat.gov.tr |
| KVKK (6698 sayılı kanun) | kvkk.gov.tr |
| UK-GDPR / Data Protection Act 2018 | legislation.gov.uk, ico.org.uk |
| ISO/IEC 27001 | iso.org |
| IKEv2 (RFC 7296) | rfc-editor.org/rfc/rfc7296 |
| Syslog (RFC 5424) ve TLS ile syslog (RFC 5425) | rfc-editor.org |
| IEEE 802.1X, 802.1Q, 802.3ad | ieee.org |
| Olay müdahalesi: NIST SP 800-61 | csrc.nist.gov |
| MITRE ATT&CK | attack.mitre.org |
| OWASP Top 10 | owasp.org |
| Eduroam | eduroam.org |
| ULAKBİM / ULAKNET | ulakbim.tubitak.gov.tr |
