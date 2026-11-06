#!/usr/bin/env bash
# Path: 04-akademik-sube/configs/nftables-qos-akademik.sh
# AegisNova Global - Akademik Şube (10.000 Öğrenci, 30 Gbps WAN) Linux Lab Betiği (nftables + tc HTB + Per-IP Conntrack)

set -euo pipefail

ULAKNET_IF="eth0"

echo "[1/2] Akademik Şube (30 Gbps = 20G ULAKNET + 10G ISP) tc HTB QoS sınıfları oluşturuluyor..."
tc qdisc del dev "${ULAKNET_IF}" root 2>/dev/null || true
tc qdisc add dev "${ULAKNET_IF}" root handle 1: htb default 60
tc class add dev "${ULAKNET_IF}" parent 1: classid 1:1 htb rate 30000mbit ceil 30000mbit

# Hocalar (320 Kişi) & Akademik AR-GE HPC (45 Kişi): Garanti 4 Gbps, Tavan 13 Gbps
tc class add dev "${ULAKNET_IF}" parent 1:1 classid 1:20 htb rate 4000mbit ceil 13000mbit prio 1
# Kampüs DMZ (LMS/OBS Dış Erişim): Garanti 3 Gbps, Tavan 10 Gbps
tc class add dev "${ULAKNET_IF}" parent 1:1 classid 1:30 htb rate 3000mbit ceil 10000mbit prio 1
# 10.000 Öğrenci Wi-Fi + Lab Havuzu: Garanti 11.5 Gbps, Tavan 18 Gbps
tc class add dev "${ULAKNET_IF}" parent 1:1 classid 1:60 htb rate 11500mbit ceil 18000mbit prio 3
# Konferans Misafir Ağı: Tavan 500 Mbps
tc class add dev "${ULAKNET_IF}" parent 1:1 classid 1:90 htb rate 50mbit ceil 500mbit prio 7

echo "[2/2] 10.000 Öğrenci İzolasyonlu ve Sınav Salonu Kilitli nftables kuralları yükleniyor..."
nft -f - <<'EOF'
flush ruleset

table inet aegisnova_campus {
    chain forward {
        type filter hook forward priority 0; policy drop;

        ct state vmap { established : accept, related : accept, invalid : drop }

        # 1. Dijital Sınav Salonları (10.40.52.0/22): Yalnızca Yerel Sınav Sunucusu (10.40.60.50:443), İNTERNET KAPALI!
        ip saddr 10.40.52.0/22 ip daddr 10.40.60.50 tcp dport 443 accept
        ip saddr 10.40.52.0/22 log prefix "[AKA-EXAM-VIOLATION-DROP] " drop

        # 2. 10.000 Öğrenci (10.100.0.0/16) -> İdare, Hoca, İç OBS DB ve HQ VPN Kesin Yasak!
        ip saddr 10.100.0.0/16 ip daddr { 10.10.0.0/16, 10.40.10.0/24, 10.40.20.0/23, 10.40.24.0/22, 10.40.40.0/24, 10.40.50.0/24 } log prefix "[AKA-STUDENT-LATERAL-SCAN-ALERT] " drop

        # 3. Öğrenciler Arası Doğrudan Trafik Engeli (Client Isolation: 10.100.0.0/16 -> 10.100.0.0/16)
        ip saddr 10.100.0.0/16 ip daddr 10.100.0.0/16 drop

        # 4. Öğrenci -> Kampüs LMS/OBS WAF (10.40.60.10:443) İzni
        ip saddr { 10.100.0.0/16, 10.40.80.0/21 } ip daddr 10.40.60.10 tcp dport 443 accept

        # 5. Öğrenci İnternet Çıkışı: P2P Portları Kapalı, Kişi Başı Max 400 Oturum ve 15 Mbps (1875 KB/s) Tavan!
        ip saddr 10.100.0.0/16 tcp dport { 445, 6881-6889 } drop
        ip saddr 10.100.0.0/16 ct count over 400 log prefix "[AKA-STUDENT-SESSION-CAP] " drop
        ip saddr 10.100.0.0/16 oifname "eth0" tcp dport { 80, 443 } meter stu_bw_cap { ip saddr limit rate over 1875 kbytes/second } drop
        ip saddr 10.100.0.0/16 oifname "eth0" tcp dport { 80, 443 } meta priority set 1:60 accept
        ip saddr 10.100.0.0/16 oifname "eth0" udp dport { 53, 443 } meta priority set 1:60 accept

        # 6. Akademisyen / Hocalar (320 Kişi - 10.40.24.0/22) ve HPC AR-GE (45 Kişi - 10.40.40.0/24)
        ip saddr { 10.40.24.0/22, 10.40.40.0/24 } oifname "eth0" tcp dport { 22, 80, 443 } meta priority set 1:20 accept

        log prefix "[AKA-FW-EXPLICIT-DENY] " drop
    }
}
EOF

echo "Akademik Şube (10.000 Öğrenci) kural seti ve 30 Gbps QoS başarıyla uygulandı."

