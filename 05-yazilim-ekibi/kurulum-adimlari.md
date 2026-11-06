# Yazılım Ekibi – Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md).

## Kurulacak katmanlar

| Katman | Çözüm | Görev |
| --- | --- | --- |
| Şube sınırı | Şubenin NGFW'si (PA-3430 / PA-1420 / PA-1410 / FG-1800F) | VLAN 30'dan çıkan trafiğin denetimi |
| Bulut / konteyner | Palo Alto VM-Series / CN-Series | Kubernetes pod seviyesinde görünürlük, Panorama ile aynı kurallar |
| Web uygulamaları | Cloudflare WAF (kampüste FortiWeb) | OWASP Top 10, bot ve DDoS koruması |

## Adımlar

1. VLAN 30'u her şubede ayrı bölge (`SW_DEV`) olarak tanımla.
2. Kimlik kaynağını bağla (AD grubu `GRP_Software_Eng`); kurallar IP'ye değil gruba yazılır.
3. GitLab, Harbor ve Vault sunucularını veri merkezi VLAN'ında yalnızca yazılım bölgesinden erişilebilir yap.
4. Giden HTTPS için SSL inceleme + antivirüs + DLP (kaynak kod ve gizli anahtar) profilini ata.
5. QoS sınıf 4 (AF31) şekillendirmesini giriş/çıkış arayüzlerine bağla.
6. Sanal NGFW'yi bulut VPC'sinin girişine koy ve Panorama'ya ekle.
7. WAF'ı API ve LMS önüne al.

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | Yazılım VLAN'ından GitLab HTTPS | İzin |
| 2 | Yazılım VLAN'ından onaylı olmayan dosya paylaşım sitesi | Reddet |
| 3 | Commit içinde AWS erişim anahtarı gönderme | DLP engeller |
| 4 | İdari VLAN'dan yazılım VLAN'ına erişim | Reddet + alarm |
| 5 | CI hattında 1 GB imaj çekme | QoS sınıf 4 tavanı içinde tamamlanır, VoIP etkilenmez |
