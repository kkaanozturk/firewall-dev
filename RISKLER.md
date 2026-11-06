# Riskler: Her Şubede Neyi, Neden, Nasıl Koruyoruz

| Şube | Ne koruyoruz | Neden (risk) | Nasıl (önlem) |
| --- | --- | --- | --- |
| İstanbul | Veri merkezi (AD, ERP, GitLab, SIEM) | Ele geçirilirse tüm şubeler etkilenir; hub olduğu için tüm tünellerin sonlandığı nokta | Bölgeler arası varsayılan reddet, yönetim yalnızca SOC VLAN'ı + MFA, HA, Panorama'da merkezi onaylı değişiklik |
| İstanbul | AR-GE verisi | Fikri mülkiyet sızıntısı | Doğrudan internet yok, yalnızca iç vekil, DLP yükleme engeli |
| İstanbul | Kaynak kod ve CI/CD | Gizli anahtar sızıntısı, tedarik zinciri saldırısı | DLP, SSL inceleme, yalnızca yazılım grubu erişir |
| İstanbul | Dışa açık DMZ | İnternetten gelen saldırı | WAF, IPS, DMZ'den iç ağa erişim yok, DoS koruma profili |
| Ankara | Savunma ve kriptografi AR-GE | Kamu/savunma verisinin sızması veya ele geçirilmesi | Çift üreticili iç güvenlik duvarı, internet yok, yalnızca imzalı yerel ayna, kayıtlı bastion, DLP |
| Ankara | Kamu B2B tünelleri | Zayıf kimlik veya yanlış yönlendirme | Beyaz liste IP, mTLS, ayrı DMZ bölgesi |
| Londra | UK müşteri/İK kişisel verisi | UK-GDPR sınır ötesi aktarım ihlali, ağır para cezası | UK-GDPR VLAN'ından yurt dışına veritabanı/dosya yasak, DPO alarmı, maskeli log |
| Londra | Bulut (AWS) erişimi | Sızan API anahtarı, aşırı yetki | Yalnızca VPN üzerinden, DLP (gizli anahtar), MFA, en az yetki |
| Londra | Uzaktan çalışanlar | Ele geçirilmiş cihaz | GlobalProtect + MFA + cihaz sağlık kontrolü |
| Akademik | 10.000 öğrenci ağı | Yanal hareket, port taraması, P2P, kötüye kullanım | Öğrenci–iç ağ tam izolasyonu, istemci izolasyonu, P2P/Tor engeli, kişi başı oturum ve hız sınırı, SOAR otomatik karantina |
| Akademik | OBS veritabanı (notlar) | Not/öğrenci verisi değiştirme veya sızdırma | Yalnızca WAF üzerinden HTTPS, veri merkezi VLAN'ına doğrudan erişim yok, KVKK DLP |
| Akademik | Dijital sınavlar | Kopya, dış erişim | Sınav VLAN'ı yalnızca sınav sunucusuna, internet kapalı, ihlalde gözetmen alarmı |
| Akademik | NAT/oturum tablosu | Port tükenmesi, SYN flood ile hizmet kesintisi | NP7 donanımsal CGNAT, port bloğu tahsisi, DoS koruma, %80 doluluk alarmı |
| Akademik | 5651 yasal kayıt | Delil niteliğinde log kaybı | Zaman damgalı, zincirli log; 72 saatlik yerel tampon |
| Tüm şubeler | Hat ve cihaz sürekliliği | Tek hata noktası | Çift ISP, çift güç, Active-Passive HA, SD-WAN |
| Tüm şubeler | Yönetim erişimi | Ele geçirilmiş yönetici hesabı | MFA, yönetim VLAN'ı, kayıtlı oturum, 4 göz ilkesi |

## Genel ağ riskleri ve karşılıkları

| Risk | Katman | Karşılık |
| --- | ---: | --- |
| ARP sahteciliği / ortadaki adam | 2 | DHCP snooping, dinamik ARP denetimi |
| VLAN atlama | 2 | Kullanılmayan portları kapat, native VLAN değiştir |
| Sahte erişim noktası (Evil Twin) | 2 | WPA3-Enterprise 802.1X, WIDS |
| Açık veya gereksiz portlar | 4 | Varsayılan reddet, düzenli Nmap taraması |
| DDoS | 3–7 | ISP koruması, WAF, güvenlik duvarında oran sınırlama |
| Yamasız yazılım | Tümü | Aylık yama, zafiyet tarayıcısı |
| Zayıf VPN / parola | 7 | MFA, güçlü kripto, hesap kilitleme |
| Yanal hareket | 3–7 | VLAN segmentasyonu, şubeler arası trafiğin merkezde denetimi |
| Oltalama | 7 | E-posta güvenlik geçidi, kum havuzu, farkındalık eğitimi |
| Web açıkları (SQLi, XSS) | 7 | DMZ + WAF |
