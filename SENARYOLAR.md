# Senaryolar: Sistem Gerçek Olaylarda Nasıl Çalışır

Her senaryo aynı sırayı izler: olay → güvenlik duvarı kararı → izleme ve müdahale → sonuç. Kural adları ilgili şubenin `firewall-kurallari.md` dosyasındadır.

| # | Olay | Güvenlik duvarı kararı | İzleme / müdahale | Sonuç |
| ---: | --- | --- | --- | --- |
| 1 | **Öğrenci OBS veritabanına bağlanmaya çalışır** (10.100.x.x → 10.40.50.10:5432) | Öğrenci–iç ağ izolasyon kuralı reddeder | İç tarama sayacı artar; SIEM'de vaka | Not verisine erişilemez |
| 2 | **Öğrenci aynı IP'ye art arda 5 port tarar** | Her deneme reddedilir | SIEM yanal tarama kuralı tetiklenir; SOAR kaynak IP'yi dinamik kara listeye ekler | Öğrencinin normal internet erişimi de kesilir, vaka SOC Tier-1'e düşer |
| 3 | **Sınav sırasında sınav bilgisayarı internete çıkmaya çalışır** | Sınav VLAN'ı yalnızca sınav sunucusuna izinli; diğer her yer reddedilir | Gözetmen/NOC alarmı, masa ve IP kaydı | Kopya girişimi engellenir ve kayıt altına alınır |
| 4 | **Öğrenci torrent indirir** | Uygulama kontrolü P2P'yi reddeder | 5651 loguna işlenir | Hat doluluğu korunur |
| 5 | **Öğrenci 40 Mbps çekmeye çalışır** | Kişi başı 15 Mbps şekillendirici sınırlar | Sınıf 6 doluluğu izlenir | Sınav ve ders trafiği etkilenmez |
| 6 | **Ankara savunma AR-GE bilgisayarı internete çıkmaya çalışır** | ISFW ve dış sınır birlikte reddeder | Kritik (Sev-1) vaka, uç nokta karantinaya alınır | Savunma verisi bölgeden çıkamaz |
| 7 | **Londra UK-GDPR sunucusu İstanbul'a PostgreSQL aktarır** | Sınır ötesi veri kilidi reddeder | DPO ve SOC Tier-2 kritik alarm; kimlik maskeli kayıt | Kişisel veri UK dışına çıkmaz |
| 8 | **Dışarıdan sahte iç IP veya bilinen zararlı IP gelir** | Anti-spoofing ve tehdit listesi sınırda düşürür | Sayaç ve log SIEM'de | DMZ'ye ulaşamaz |
| 9 | **İstanbul'da birincil güvenlik duvarının gücü kesilir** | İkincil cihaz stateful olarak devralır (< 1 sn) | HA senkron alarmı NOC'a gider | Oturum ve VPN düşmez |
| 10 | **ISP-1 fiber kablosu kopar** | SD-WAN sağlık eşiği aşılır, trafik ISP-2'ye kayar | Hat alarmı (1 dk içinde NOC) | VoIP ve canlı ders kesilmez |
| 11 | **Akademik şubede ders başlangıcında yeni oturum patlaması** | NP7 CGNAT port bloğu atar; oturum tablosu izlenir | %65 uyarı, %80 kritik alarm | NAT portu tükenmeden kapasite genişletilir |
| 12 | **Yazılımcı commit'e AWS anahtarı koyar** | DLP giden isteği engeller | Yazılım/SOC bildirimi | Gizli anahtar sızmaz |
| 13 | **5651 zaman damgası servisi 15 dk yanıt vermez** | – | Uyum ekibine kritik alarm | Yerel tamponla kayıp önlenir, servis dönünce imzalanır |
| 14 | **Güvenlik duvarına yetkisiz yönetici girişi denenir** | Yönetim yalnızca SOC alt ağı + MFA | 5 başarısız deneme / 5 dk → Tier-1 | Hesap kilitlenir |

## Kapsamlı bir olayın akışı (Senaryo 2 ayrıntılı)

1. Öğrenci cihazı OBS veritabanı portlarını taramaya başlar.
2. Güvenlik duvarı her paketi öğrenci–iç ağ izolasyon kuralıyla reddeder ve loglar.
3. Şube log toplayıcı olayları İstanbul SIEM'e syslog TLS ile iletir.
4. SIEM korelasyon kuralı eşiği aşan kaynağı yakalar ve MITRE T1046 (ağ servisi keşfi) olarak etiketler.
5. SOAR, kaynak IP'yi güvenlik duvarında dinamik listeye ekler; işlem insan beklemeden saniyeler içinde biter.
6. SOC Tier-1 vakayı alır (hedef: 10 dakika içinde üstlenme), kullanıcı kimliğini RSSO kaydından bulur ve gerekirse Tier-2'ye iletir.
7. Olay sonrası: playbook güncellenir, kural ve eşikler haftalık toplantıda gözden geçirilir.
