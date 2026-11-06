# Ankara Ofisi – Şubeye Özel Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md). Bu dosya yalnızca Ankara'ya özgü ayarları ve kabul testlerini içerir.

## Cihazlar

| Katman | Cihaz | Görev | Konfigürasyon |
| --- | --- | --- | --- |
| Dış sınır | 2 × Palo Alto PA-1420 (HA) | İnternet, İstanbul tüneli, Kamu B2B VPN | [pa1420-perimeter.conf](configs/pa1420-perimeter.conf) |
| İç segmentasyon | 2 × FortiGate 200F (HA) | Savunma AR-GE bölgesi (VLAN 40) | [fg200f-rnd-isfw.conf](configs/fg200f-rnd-isfw.conf) |

Dışta Palo Alto, içte Fortinet kullanılır: tek üretici zafiyeti iki katmanı birden etkilemez.

## Şubeye özel ayarlar

1. **Hatlar:** `ethernet1/1` ISP-1 2,5 Gbps metro, `ethernet1/2` ISP-2 1 Gbps farklı güzergah fiber (SD-WAN yedek).
2. **LACP:** İç omurgaya 2 × 10G; VLAN 10, 20, 30, 45 (ISFW transit, 10.20.45.0/30), 50, 60, 70, 90 alt arayüzleri.
3. **ISFW:** FG-200F, VLAN 40 (10.20.40.0/24) ile dış sınır arasında; `x1` transit, `x2` savunma AR-GE.
4. **Kamu B2B:** Beyaz listedeki kamu IP'lerinden IPsec ve mTLS 8443; yalnızca DMZ VLAN 60'a.
5. **İstanbul tüneli:** IKEv2, AES-256-GCM, SHA-384, DH 20; Panorama'dan merkezi yönetim.
6. **AR-GE erişimi:** Yalnızca yerel RODC/DNS/NTP, imzalı paket aynası (10.20.50.30:443) ve bastion SSH; internete çıkış yok.

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | HA birincil cihaz gücü kesilir | İkincil < 1 sn devralır, tünel düşmez |
| 2 | ISP-1 kablosu çıkarılır | Trafik ISP-2'ye geçer, kayıp < %0,5 |
| 3 | AR-GE bilgisayarından 8.8.8.8:443 | Reddet + SIEM kritik alarm |
| 4 | AR-GE bilgisayarından imzalı aynaya HTTPS | İzin |
| 5 | Personel VLAN'ından AR-GE VLAN'ına erişim | Reddet |
| 6 | Kamu B2B tüneli kopması | 1 dk içinde NOC alarmı |
| 7 | Misafir VLAN'ından iç ağ taraması | Reddet, 5 sn içinde SIEM alarmı |
