#!/usr/bin/env bash
# Path: 01-ankara-ofisi/configs/nftables-qos-ankara.sh
# Kalkan Bilişim - Ankara Ofisi Linux Laboratuvar / Doğrulama Betiği (nftables + tc HTB 3.5 Gbps QoS)

set -euo pipefail

WAN_IF="eth0"

echo "[1/2] Ankara Ofisi (3.5 Gbps) tc HTB QoS sınıfları oluşturuluyor..."
tc qdisc del dev "${WAN_IF}" root 2>/dev/null || true
tc qdisc add dev "${WAN_IF}" root handle 1: htb default 50
tc class add dev "${WAN_IF}" parent 1: classid 1:1 htb rate 3500mbit ceil 3500mbit

tc class add dev "${WAN_IF}" parent 1:1 classid 1:10 htb rate 300mbit ceil 800mbit prio 0
tc class add dev "${WAN_IF}" parent 1:1 classid 1:20 htb rate 700mbit ceil 2000mbit prio 1
tc class add dev "${WAN_IF}" parent 1:1 classid 1:30 htb rate 800mbit ceil 2000mbit prio 2
tc class add dev "${WAN_IF}" parent 1:1 classid 1:50 htb rate 250mbit ceil 900mbit prio 3
tc class add dev "${WAN_IF}" parent 1:1 classid 1:90 htb rate 10mbit ceil 80mbit prio 7

echo "[2/2] Ankara Ofisi Savunma AR-GE İzolasyonlu nftables kuralları yükleniyor..."
nft -f - <<'EOF'
flush ruleset

table inet kalkan_ankara {
    chain forward {
        type filter hook forward priority 0; policy drop;

        ct state vmap { established : accept, related : accept, invalid : drop }

        # Savunma AR-GE (10.20.40.0/24) İnternete Kesinlikle Çıkamaz!
        ip saddr 10.20.40.0/24 oifname "eth0" log prefix "[ANK-CRITICAL-RND-WAN-BLOCK] " drop

        # Savunma AR-GE yalnızca yerel RODC (10.20.50.10) ve İmzalı Paket Aynasına (10.20.50.30) erişebilir
        ip saddr 10.20.40.0/24 ip daddr 10.20.50.10 tcp dport { 53, 88, 636 } accept
        ip saddr 10.20.40.0/24 ip daddr 10.20.50.10 udp dport { 53, 88, 123 } accept
        ip saddr 10.20.40.0/24 ip daddr 10.20.50.30 tcp dport 443 accept
        ip daddr 10.20.40.0/24 ip saddr != 10.20.10.50 log prefix "[ANK-ISFW-INBOUND-DENY] " drop

        # Gömülü & Kamu Yazılım Ekibi (85 Kişi)
        ip saddr 10.20.30.0/24 oifname "eth0" tcp dport 443 meta priority set 1:30 accept

        # İdari Personel (65 Kişi)
        ip saddr 10.20.20.0/24 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:50 accept

        # Misafir Ağı (Kişi başı 4 Mbps = 500 kbytes/sec)
        ip saddr 172.16.20.0/24 ip daddr 10.0.0.0/8 drop
        ip saddr 172.16.20.0/24 oifname "eth0" tcp dport { 80, 443 } meter ank_guest { ip saddr limit rate over 500 kbytes/second } drop
        ip saddr 172.16.20.0/24 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:90 accept

        log prefix "[ANK-FW-EXPLICIT-DENY] " drop
    }
}
EOF

echo "Ankara Ofisi kural seti ve 3.5 Gbps QoS başarıyla uygulandı."

