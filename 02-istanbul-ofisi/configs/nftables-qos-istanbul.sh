#!/usr/bin/env bash
# Path: 02-istanbul-ofisi/configs/nftables-qos-istanbul.sh
# Kalkan Bilişim - İstanbul Ofisi Linux Laboratuvar / Doğrulama Betiği (nftables + tc HTB 10 Gbps QoS)
# Kullanım (Linux Lab VM / Simülasyon): sudo bash nftables-qos-istanbul.sh

set -euo pipefail

WAN_IF="eth0"
LAN_TRUNK="bond0"

echo "[1/3] İstanbul Ofisi (10 Gbps) Linux tc HTB Bant Genişliği Şekillendirme uygulanıyor..."
tc qdisc del dev "${WAN_IF}" root 2>/dev/null || true
tc qdisc add dev "${WAN_IF}" root handle 1: htb default 50

# Toplam Hat Kapasitesi: 10 Gbps (10000mbit)
tc class add dev "${WAN_IF}" parent 1: classid 1:1 htb rate 10000mbit ceil 10000mbit

# Sınıf 10: VoIP ve Ağ Kontrol (Garanti 800 Mbps, Tavan 2000 Mbps, En Yüksek Öncelik prio 0)
tc class add dev "${WAN_IF}" parent 1:1 classid 1:10 htb rate 800mbit ceil 2000mbit prio 0
# Sınıf 20: Core DC, SIEM ve DMZ İş Kritik (Garanti 2500 Mbps, Tavan 5000 Mbps, prio 1)
tc class add dev "${WAN_IF}" parent 1:1 classid 1:20 htb rate 2500mbit ceil 5000mbit prio 1
# Sınıf 30: Yazılım (140 Kişi) ve AR-GE Vekil (65 Kişi) (Garanti 2500 Mbps, Tavan 4500 Mbps, prio 2)
tc class add dev "${WAN_IF}" parent 1:1 classid 1:30 htb rate 2500mbit ceil 4500mbit prio 2
# Sınıf 50: İdari Personel (180 Kişi) Genel Kurumsal (Garanti 600 Mbps, Tavan 2500 Mbps, prio 3)
tc class add dev "${WAN_IF}" parent 1:1 classid 1:50 htb rate 600mbit ceil 2500mbit prio 3
# Sınıf 90: Misafir ve Stajyer Wi-Fi (Garanti 10 Mbps, Katı Tavan 250 Mbps, prio 7)
tc class add dev "${WAN_IF}" parent 1:1 classid 1:90 htb rate 10mbit ceil 250mbit prio 7

echo "[2/3] İstanbul Ofisi Zero-Trust nftables kuralları yükleniyor..."
nft -f - <<'EOF'
flush ruleset

table inet kalkan_istanbul {
    set threat_edl_ips {
        type ipv4_addr
        flags interval
        elements = { 185.220.101.0/24, 198.98.50.0/24, 45.155.205.0/24 }
    }

    set rfc1918_nets {
        type ipv4_addr
        flags interval
        elements = { 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 }
    }

    chain forward {
        type filter hook forward priority 0; policy drop;

        # 1. Durumsal (Stateful) Bağlantı Takibi
        ct state vmap { established : accept, related : accept, invalid : drop }

        # 2. Anti-Spoofing & Tehdit İstihbaratı (EDL) Düşürme
        iifname "eth0" ip saddr @rfc1918_nets log prefix "[IST-FW-SPOOF-DROP] " drop
        iifname "eth0" ip saddr @threat_edl_ips log prefix "[IST-FW-EDL-DROP] " drop

        # 3. Teknik & Merkez SOC (10.10.10.0/24) Yönetim İzni
        ip saddr 10.10.10.0/24 ip daddr 10.10.50.0/24 tcp dport { 22, 443, 6514 } accept

        # 4. İdari Personel (180 Kişi): ERP/AD İzni, Yazılım ve AR-GE'ye Yanal Hareket Yasağı
        ip saddr 10.10.20.0/23 ip daddr { 10.10.30.0/23, 10.10.40.0/24 } log prefix "[IST-FW-LATERAL-DENY] " drop
        ip saddr 10.10.20.0/23 ip daddr 10.10.50.0/24 tcp dport { 53, 88, 389, 443, 636, 3200 } accept
        ip saddr 10.10.20.0/23 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:50 accept

        # 5. Yazılım Ekibi (140 Kişi): GitLab/Harbor ve Kontrollü HTTPS Çıkışı
        ip saddr 10.10.30.0/23 ip daddr { 10.10.50.40, 10.10.50.41, 10.10.50.42 } tcp dport { 22, 443, 5000 } accept
        ip saddr 10.10.30.0/23 oifname "eth0" tcp dport 443 meta priority set 1:30 accept

        # 6. AR-GE Ekibi (65 Kişi): Yalnızca İç Vekil Sunucu (10.10.50.60:8443), Doğrudan WAN Yasak!
        ip saddr 10.10.40.0/24 ip daddr 10.10.50.60 tcp dport 8443 accept
        ip saddr 10.10.40.0/24 oifname "eth0" log prefix "[IST-FW-RND-WAN-VIOLATION] " drop

        # 7. Misafir Ağı: Tüm İç Ağlara Kapalı, İnternete Kişi Başı 5 Mbps Sınırlı
        ip saddr 172.16.10.0/23 ip daddr @rfc1918_nets log prefix "[IST-FW-GUEST-RFC1918-DENY] " drop
        ip saddr 172.16.10.0/23 oifname "eth0" tcp dport { 80, 443 } meter guest_per_ip { ip saddr limit rate over 625 kbytes/second } drop
        ip saddr 172.16.10.0/23 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:90 accept

        # 99. Kayıtlı Açık Reddetme (Explicit Deny & Log)
        log prefix "[IST-FW-EXPLICIT-DENY] " flags all drop
    }
}
EOF

echo "[3/3] İstanbul Ofisi güvenlik duvarı ve 10 Gbps QoS başarıyla uygulandı."

