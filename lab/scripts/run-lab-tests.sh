#!/usr/bin/env bash
# Path: lab/scripts/run-lab-tests.sh
# AegisNova Global - Çok Konteynerli Docker Ağ Laboratuvarı Otomatik Kabul Testi (UAT)

set -uo pipefail

PASS_COUNT=0
FAIL_COUNT=0

expect_allow() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then
    echo " [BASARILI - ALLOW] ${desc}"
    ((PASS_COUNT++))
  else
    echo " [HATA - BEKLENEN ALLOW] ${desc}"
    ((FAIL_COUNT++))
  fi
}

expect_block() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then
    echo " [HATA - BEKLENEN BLOCK] ${desc}"
    ((FAIL_COUNT++))
  else
    echo " [BASARILI - BLOCK] ${desc}"
    ((PASS_COUNT++))
  fi
}

echo "========================================================================================="
echo " AEGISNOVA GLOBAL - GERCEK LINUX NFTABLES + TC QOS DOCKER LAB TESTLERI BASLIYOR"
echo "========================================================================================="

expect_allow "1. İstanbul Yazılım Ekibi -> İç GitLab Sunucusu (10.10.50.40:443)" \
  docker exec lab-ist-swdev-client curl -s --max-time 2 http://10.10.50.40:443

expect_allow "2. Ankara Savunma AR-GE -> İç İmzalı Paket Aynası (10.20.50.30:443)" \
  docker exec lab-ank-rnd-pc curl -s --max-time 2 http://10.20.50.30:443

expect_block "3. Ankara Savunma AR-GE -> Dış İnternet Çıkış Denemesi (ISFW Hava Boşluğu Kilidi)" \
  docker exec lab-ank-rnd-pc curl -s --max-time 2 http://198.51.100.200:443

expect_block "4. Londra UK-GDPR Sunucusu -> Sınır Ötesi PostgreSQL (5432) Aktarımı" \
  docker exec lab-lon-gdpr-srv nc -z -w 2 198.51.100.200 5432

expect_allow "5. Londra UK-GDPR Sunucusu -> Maskeli SIEM TLS Syslog (6514) Gönderimi" \
  docker exec lab-lon-gdpr-srv nc -z -w 2 198.51.100.200 6514

expect_allow "6. Akademik Öğrenci PC -> Dış İnternet HTTPS (15 Mbps tc HTB QoS Altında)" \
  docker exec lab-aka-student-pc curl -s --max-time 2 http://198.51.100.200:443

expect_block "7. Akademik Öğrenci PC -> İç OBS Veritabanı (10.40.50.10:5432) Erişim Girişimi" \
  docker exec lab-aka-student-pc nc -z -w 2 10.40.50.10 5432

expect_block "8. Akademik Öğrenci PC -> BitTorrent Portu (6881) Engeli" \
  docker exec lab-aka-student-pc nc -z -w 2 198.51.100.200 6881

expect_allow "9. Akademik Dijital Sınav PC -> Yerel Sınav LMS Sunucusu (10.40.60.50:443)" \
  docker exec lab-aka-exam-pc curl -s --max-time 2 http://10.40.60.50:443

expect_block "10. Akademik Dijital Sınav PC -> Dış İnternet Kopya Girişimi (Sınav Kilidi)" \
  docker exec lab-aka-exam-pc curl -s --max-time 2 http://198.51.100.200:443

echo "========================================================================================="
echo " SONUC: ${PASS_COUNT} BASARILI | ${FAIL_COUNT} HATA"
echo "========================================================================================="
exit "${FAIL_COUNT}"

