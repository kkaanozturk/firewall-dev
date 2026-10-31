#!/usr/bin/env bash
# Path: lab/scripts/init-fw-node.sh
# AegisNova Global - Konteyner Güvenlik Duvarı Düğümü Başlatıcı (IP Forwarding + nftables + tc QoS)

set -euo pipefail

ROLE="${1:-}"
echo "[INIT] Düğüm başlatılıyor: ROLE=${ROLE}"

sysctl -w net.ipv4.ip_forward=1 >/dev/null 2>&1 || true

case "${ROLE}" in
  istanbul-fw)
    nft -f - <<'EOF'
flush ruleset
table inet lab_ist_fw {
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept
        # Yazılım Ekibi (10.10.30.0/24) -> GitLab/SIEM (10.10.50.40:443) ve WAN (198.51.100.200:443) İzinli
        ip saddr 10.10.30.0/24 ip daddr 10.10.50.40 tcp dport 443 accept
        ip saddr 10.10.30.0/24 ip daddr 198.51.100.0/24 tcp dport 443 accept
        log prefix "[LAB-IST-DROP] " drop
    }
}
EOF
    ;;

  ankara-isfw)
    nft -f - <<'EOF'
flush ruleset
table inet lab_ank_isfw {
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept
        # Savunma AR-GE (10.20.40.0/24) yalnızca yerel İmzalı Paket Aynasına (10.20.50.30:443) erişebilir; WAN YASAK!
        ip saddr 10.20.40.0/24 ip daddr 10.20.50.30 tcp dport 443 accept
        ip saddr 10.20.40.0/24 ip daddr 198.51.100.0/24 log prefix "[LAB-ANK-RND-WAN-BLOCK] " drop
        log prefix "[LAB-ANK-ISFW-DROP] " drop
    }
}
EOF
    ;;

  londra-fw)
    nft -f - <<'EOF'
flush ruleset
table inet lab_lon_fw {
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept
        # UK-GDPR Kilidi: 10.30.50.0/24 alt ağından dışarıya PostgreSQL (5432) çıkışı KESİN YASAK!
        ip saddr 10.30.50.0/24 tcp dport 5432 log prefix "[LAB-LON-UKGDPR-BLOCK] " drop
        # Maskeli Syslog TLS (6514) İstanbul SIEM'e (10.10.50.25) İzinli
        ip saddr 10.30.50.0/24 ip daddr 198.51.100.0/24 tcp dport 6514 accept
        log prefix "[LAB-LON-DROP] " drop
    }
}
EOF
    ;;

  akademik-fw)
    # 30 Gbps simülasyonu için öğrenci arayüzünde kişi başı 15 Mbps tc HTB hız sınırı
    WAN_DEV="$(ip -o -4 route show to 198.51.100.0/24 | awk '{print $3}' | head -n1 || true)"
    if [[ -n "${WAN_DEV}" ]]; then
      tc qdisc del dev "${WAN_DEV}" root 2>/dev/null || true
      tc qdisc add dev "${WAN_DEV}" root handle 1: htb default 10
      tc class add dev "${WAN_DEV}" parent 1: classid 1:10 htb rate 15mbit ceil 15mbit
    fi

    nft -f - <<'EOF'
flush ruleset
table inet lab_aka_fw {
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept

        # 1. Dijital Sınav Salonu (10.40.55.0/24) -> Yalnızca Sınav LMS (10.40.60.50:443); İnternet YASAK!
        ip saddr 10.40.55.0/24 ip daddr 10.40.60.50 tcp dport 443 accept
        ip saddr 10.40.55.0/24 log prefix "[LAB-AKA-EXAM-LOCK] " drop

        # 2. 10.000 Öğrenci Ağı (10.100.0.0/16) -> İç OBS Veritabanı (10.40.50.10) KESİN YASAK!
        ip saddr 10.100.0.0/16 ip daddr 10.40.50.0/24 log prefix "[LAB-AKA-STU-DB-DENY] " drop

        # 3. Öğrenci -> DMZ LMS WAF (10.40.60.10:443) ve WAN HTTPS (443) İzinli; BitTorrent (6881) YASAK!
        ip saddr 10.100.0.0/16 tcp dport 6881 drop
        ip saddr 10.100.0.0/16 ip daddr 10.40.60.10 tcp dport 443 accept
        ip saddr 10.100.0.0/16 ip daddr 198.51.100.0/24 tcp dport 443 accept

        log prefix "[LAB-AKA-DROP] " drop
    }
}
EOF
    ;;
esac

echo "[READY] ${ROLE} kuralları aktif."
exec sleep infinity

