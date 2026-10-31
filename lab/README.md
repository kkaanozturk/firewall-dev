<!-- Path: lab/README.md -->
# AegisNova Global — Test ve Doğrulama Laboratuvarı (Canlı Web SOC Lab + Docker Multi-Subnet Lab)

Bu klasör, ödev kapsamında tasarlanan **4 şubeli güvenlik duvarı mimarisini, bant genişliği/QoS sınırlarını, Sıfır Güven kural setlerini ve 7/24 SOC/SOAR izleme yapısını** uygulamalı olarak test etmek için iki tamamlayıcı laboratuvar ortamı sunar:

1. **Mod 1 — Sıfır Ek Kurulumlu Tarayıcı Tabanlı Canlı NGFW & SOC Web Laboratuvarı (`http://127.0.0.1:8080`):**
   - Docker veya sanal makine gerektirmeden doğrudan `Go` ile Windows/macOS/Linux üzerinde saniyeler içinde açılır.
   - Özel paket oluşturup güvenlik duvarına gönderme, tek tıkla **Sınav Ağı Kilidi**, **Öğrenci Yanal Tarama + SOAR Otomatik Karantina**, **Savunma AR-GE İzolasyonu**, **Londra UK-GDPR Kilidi** ve **Active-Passive HA Donanım Devretme (Failover)** tatbikatlarını canlı SOC ekranında görselleştirir.
2. **Mod 2 — Docker Çoklu Alt Ağ (Multi-Subnet) Gerçek Linux `nftables` + `tc HTB QoS` Laboratuvarı:**
   - 14 ayrı Linux konteyneri ve 10 izole alt ağ (`10.10.x.x`, `10.20.x.x`, `10.30.x.x`, `10.100.0.0/16`) oluşturarak gerçek Linux çekirdeği üzerinde paket yönlendirme, `nftables` filtreleme ve `tc` bant genişliği şekillendirme testlerini koşturur.

---

## 1. Mod 1: Canlı Web Laboratuvarını Başlatma (Önerilen — Hemen Çalışır)

Proje ana dizininde aşağıdaki komutu çalıştırın (veya doğrudan `.\start-lab.ps1` betiğini çağırın):

```powershell
# PowerShell ile tek komutta başlatma ve tarayıcıda açma
.\start-lab.ps1

# Veya doğrudan Go komutu ile başlatma
go run ./cmd/aegisnova-lab -addr 127.0.0.1:8080
```

Tarayıcınızda `http://127.0.0.1:8080` adresini açtığınızda:
- **Üst Durum Kartları:** İstanbul (`10 Gbps`), Ankara (`3.5 Gbps`), Londra (`4 Gbps`) ve Akademik Şube (`30 Gbps / 10.000 Öğrenci`) güvenlik duvarlarının anlık HA (`FW01` / `FW02`), ISP-1/ISP-2 ve eşzamanlı oturum durumlarını izleyin.
- **Hazır Tatbikat Senaryoları:** Tek tıkla saldırı/ihlal veya HA kesinti senaryolarını tetikleyin; sağ panelde **SOC Tier-1/2/3 vakalarının** ve **5651 SHA-256 zaman damgası zincirinin** canlı oluşumunu gözlemleyin.
- **Özel Paket Enjektörü:** İstediğiniz kaynak/hedef IP, kullanıcı grubu (`Role_Student`, `GRP_Software_Eng`, `GRP_Defense_RnD`), port, uygulama ve Mbps değeriyle kendi paketinizi ateşleyin.

---

## 2. Mod 2: Docker Çoklu Alt Ağ (Multi-Subnet) Laboratuvarını Başlatma

Docker Desktop veya Linux Docker Engine kurulu bir ortamda gerçek çekirdek (`CAP_NET_ADMIN`) topolojisini ayağa kaldırmak için:

```bash
# 1. 4 şube güvenlik duvarını, sunucuları ve istemcileri ayağa kaldır
docker compose -f lab/docker-compose.lab.yml up -d --build

# 2. 10 adımlı otomatik uçtan uca paket doğrulama testini çalıştır
bash lab/scripts/run-lab-tests.sh

# 3. İsteğe bağlı: Akademik Şube güvenlik duvarında engellenen paket loglarını canlı izle
docker exec -it lab-aka-campus-fw nft list ruleset

# 4. Laboratuvarı kapat ve ağları temizle
docker compose -f lab/docker-compose.lab.yml down
```

### Docker Lab Alt Ağ ve Konteyner Haritası

| Şube | Güvenlik Duvarı Konteyneri | İç İstemci Konteyneri | İç Hedef / Sunucu Konteyneri | Doğrulanan Kural |
|---|---|---|---|---|
| **İstanbul** | `lab-ist-hq-fw` | `lab-ist-swdev-client` (`10.10.30.25`) | `lab-ist-gitlab` (`10.10.50.40`) | Yazılım CI/CD HTTPS izni |
| **Ankara** | `lab-ank-isfw` | `lab-ank-rnd-pc` (`10.20.40.12`) | `lab-ank-repo-mirror` (`10.20.50.30`) | İmzalı ayna izni & WAN çıkış engeli |
| **Londra** | `lab-lon-emea-fw` | `lab-lon-gdpr-srv` (`10.30.50.14`) | `lab-wan-target` (`198.51.100.200`) | PostgreSQL 5432 UK-GDPR engeli & TLS 6514 izni |
| **Akademik** | `lab-aka-campus-fw` | `lab-aka-student-pc` (`10.100.42.19`) & `lab-aka-exam-pc` (`10.40.55.101`) | `lab-aka-obs-db` (`10.40.50.10`) & `lab-aka-exam-lms` (`10.40.60.50`) | OBS DB engeli, P2P engeli, 15 Mbps `tc` sınırı ve Sınav Salonu kilidi |

