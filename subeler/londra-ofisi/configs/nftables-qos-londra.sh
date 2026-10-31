#!/usr/bin/env bash
# Path: subeler/londra-ofisi/configs/nftables-qos-londra.sh
# AegisNova Global - Londra Ofisi Linux Laboratuvar / Doğrulama Betiği (nftables + tc HTB 4.0 Gbps & UK-GDPR Kilidi)

set -euo pipefail

WAN_IF="eth0"
VPN_HQ_IF="vti10"

echo "[1/2] Londra Ofisi (4.0 Gbps Dual Leased Line) tc HTB QoS sınıfları uygulanıyor..."
tc qdisc del dev "${WAN_IF}" root 2>/dev/null || true
tc qdisc add dev "${WAN_IF}" root handle 1: htb default 50
tc class add dev "${WAN_IF}" parent 1: classid 1:1 htb rate 4000mbit ceil 4000mbit

tc class add dev "${WAN_IF}" parent 1:1 classid 1:10 htb rate 250mbit ceil 900mbit prio 0
tc class add dev "${WAN_IF}" parent 1:1 classid 1:20 htb rate 600mbit ceil 2200mbit prio 1
tc class add dev "${WAN_IF}" parent 1:1 classid 1:30 htb rate 1200mbit ceil 3000mbit prio 2
tc class add dev "${WAN_IF}" parent 1:1 classid 1:50 htb rate 250mbit ceil 800mbit prio 3
tc class add dev "${WAN_IF}" parent 1:1 classid 1:90 htb rate 10mbit ceil 60mbit prio 7

echo "[2/2] Londra Ofisi UK-GDPR Sınır Ötesi Veri Korumalı nftables kuralları yükleniyor..."
nft -f - <<'EOF'
flush ruleset

table inet aegisnova_london {
    chain forward {
        type filter hook forward priority 0; policy drop;

        ct state vmap { established : accept, related : accept, invalid : drop }

        # UK-GDPR Madde 44-49: UK Kişisel Veri Sunucularından (10.30.50.0/24) yurt dışına DB/Dosya çıkışı YASAK!
        ip saddr 10.30.50.0/24 oifname { "vti10", "eth0" } tcp dport { 21, 22, 445, 873, 1433, 1521, 3306, 5432 } log prefix "[LON-UKGDPR-VIOLATION-DROP] " drop

        # Maskelenmiş Log Toplayıcıdan (10.30.50.25) İstanbul SIEM'e (10.10.50.25:6514) TLS Syslog İzni
        ip saddr 10.30.50.25 ip daddr 10.10.50.25 tcp dport 6514 accept

        # Cloud-Native Yazılım & SRE (50 Kişi): AWS eu-west-2 (172.28.0.0/16) ve HQ GitLab (10.10.50.40)
        ip saddr 10.30.30.0/24 ip daddr 172.28.0.0/16 tcp dport { 22, 443, 6443 } accept
        ip saddr 10.30.30.0/24 ip daddr 10.10.50.40 tcp dport { 22, 443 } accept
        ip saddr 10.30.30.0/24 oifname "eth0" tcp dport 443 meta priority set 1:30 accept

        # FinTech AR-GE (25 Kişi): Yalnızca HTTPS, iç ağlara yanal hareket kapalı
        ip saddr 10.30.40.0/24 ip daddr { 10.30.20.0/24, 172.16.30.0/24 } drop

        # İdari & Hukuk (45 Kişi)
        ip saddr 10.30.20.0/24 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:50 accept

        # Ziyaretçi Wi-Fi (Kişi başı 5 Mbps)
        ip saddr 172.16.30.0/24 ip daddr 10.0.0.0/8 drop
        ip saddr 172.16.30.0/24 oifname "eth0" tcp dport { 80, 443 } meter lon_guest { ip saddr limit rate over 625 kbytes/second } drop
        ip saddr 172.16.30.0/24 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:90 accept

        log prefix "[LON-FW-EXPLICIT-DENY] " drop
    }
}
EOF

echo "Londra Ofisi kural seti ve 4.0 Gbps QoS başarıyla uygulandı."

