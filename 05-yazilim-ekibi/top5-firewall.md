# Yazılım Ekibi – En İyi 5 Güvenlik Duvarı Adayı ve Seçim

Yazılım ekibi dört şubede toplam 310 kişidir (İstanbul 140, Ankara 85, Londra 50, Akademik 35). CI/CD, konteyner ve bulut API trafiği nedeniyle klasik sınır güvenlik duvarına ek olarak sanal NGFW + WAF katmanı gerekir.

310 kişilik toplam yazılım ekibinin CI/CD hatlarını, Kubernetes kümelerini ve dışa açık API/LMS uygulamalarını korumak için değerlendirilen Top 5 sanal/bulut güvenlik duvarı:

| Sıra | Çözüm |Tahmini Kapasite | Durum ve Gerekçe |
|---:|---|---:|---|
| **1** | **Palo Alto VM-Series / CN-Series + Cloudflare Enterprise WAF (SEÇİLDİ)** | **2 – 16 Gbps (vCPU ölçekli)** | Merkez Panorama ile tekil kural yönetimi, Kubernetes Pod seviyesinde L7 görünürlük ve DDoS/OWASP Top 10 koruması sağlar. |
| **2** | Fortinet FortiGate-VM + FortiWeb | 2 – 20 Gbps | Güçlü alternatif; Akademik Şube LMS/OBS önünde **FortiWeb WAF** olarak konumlandırılmıştır. |
| **3** | AWS Network Firewall + AWS WAF | Bulut Yerel (Oto-Ölçek) | Yalnızca AWS içinde çalıştığı, çoklu bulut (Multi-Cloud) ve şirket içi (On-Prem) sunucuları kapsamadığı için ana çözüm olarak seçilmedi. |
| **4** | Check Point CloudGuard Network | 2 – 15 Gbps | Mevcut Panorama yönetim ekosistemiyle bütünleşmediği için elendi. |
| **5** | Tigera Calico Enterprise (eBPF) | 10+ Gbps (Kernel düzeyi) | Kubernetes içi Doğu-Batı politikalarında çok hızlıdır ancak tam teşekküllü IPS/Antimalware/TLS Inspection yapamadığı için NGFW yerine yalnızca yardımcı CNI katmanı olarak kullanılır. |

---

