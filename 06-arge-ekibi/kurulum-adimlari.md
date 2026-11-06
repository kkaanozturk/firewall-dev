# AR-GE Ekibi – Kurulum ve Testler

Ortak adımlar: [KURULUM.md](../KURULUM.md).

## Kurulacak katmanlar

| Şube | Katman | Çözüm |
| --- | --- | --- |
| İstanbul | Şube sınırı | PA-3430; AR-GE yalnızca iç vekil (10.10.50.60:8443) |
| Ankara | İç segmentasyon (ISFW) | FortiGate 200F HA, [konfig](../01-ankara-ofisi/configs/fg200f-rnd-isfw.conf) |
| Londra | Şube sınırı | PA-1410; FQDN beyaz liste + DLP |
| Akademik | Şube sınırı | FG-1800F; ULAKNET öncelikli |

## Adımlar

1. VLAN 40'ı her şubede ayrı bölge (`RND_SEC` / `RND_DEFENSE` / `RND_FINTECH` / `RND_HPC`) olarak tanımla.
2. Ankara'da FG-200F'i dış sınır ve AR-GE arasına HA çifti olarak koy; transit VLAN 45 (10.20.45.0/30) kullan.
3. Kimlik grubunu bağla (`GRP_RnD_Researchers`, `GRP_Defense_RnD`); kurallar gruba yazılır.
4. İç vekil / imzalı paket aynası sunucularını kur; AR-GE yalnızca bunlara erişebilsin.
5. DLP profilini ata (kaynak kod, şema, finansal kişisel veri); yükleme engeli aç.
6. Bastion sunucusunu yönetim VLAN'ında kur; oturum kaydı + 2FA zorunlu, pano ve dosya kopyalama kapalı.
7. Tüm reddedilen AR-GE çıkış denemelerini SIEM'de kritik alarma bağla.

## Kabul testleri

| # | Test | Beklenen |
| ---: | --- | --- |
| 1 | AR-GE VLAN'ından 1.1.1.1:443 (İstanbul) | Reddet + kritik alarm |
| 2 | Ankara AR-GE'den imzalı aynaya HTTPS | İzin |
| 3 | Ankara AR-GE'den 8.8.8.8:443 | Reddet (ISFW ve dış sınır) + Sev-1 vaka |
| 4 | Personel VLAN'ından AR-GE'ye erişim | Reddet |
| 5 | AR-GE'den kaynak kod dosyası yükleme | DLP engeller |
| 6 | Bastion dışındaki makineden SSH | Reddet |
| 7 | Londra FinTech'ten onaylı olmayan siteye çıkış | Reddet |
