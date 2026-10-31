// Path: cmd/aegisnova-lab/main_test.go
// Runtime: Go 1.24+ | Interactive Web Lab REST & Drill Verification Tests

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebLabEndpointsAndDrills(t *testing.T) {
	lab := NewLabServer()
	handler := lab.Handler()

	// 1. Ana HTML arayüzü kontrolü
	reqUI := httptest.NewRequest(http.MethodGet, "/", nil)
	recUI := httptest.NewRecorder()
	handler.ServeHTTP(recUI, reqUI)
	if recUI.Code != http.StatusOK || !strings.Contains(recUI.Body.String(), "AegisNova Global") {
		t.Fatalf("UI beklenen yaniti donmedi: status=%d", recUI.Code)
	}

	// 2. Hazır Yanal Tarama + SOAR Tatbikatı kontrolü
	reqDrill := httptest.NewRequest(http.MethodPost, "/api/drill", bytes.NewBufferString(`{"drill":"lateral_soar"}`))
	recDrill := httptest.NewRecorder()
	handler.ServeHTTP(recDrill, reqDrill)
	if recDrill.Code != http.StatusOK {
		t.Fatalf("Drill calistirilamadi: status=%d", recDrill.Code)
	}
	var drillPayload struct {
		Summary string `json:"summary"`
		State   struct {
			Chain5651Valid bool `json:"chain_5651_valid"`
			Chain5651Count int  `json:"chain_5651_count"`
		} `json:"state"`
	}
	if err := json.NewDecoder(recDrill.Body).Decode(&drillPayload); err != nil {
		t.Fatalf("JSON cozulemedi: %v", err)
	}
	if !strings.Contains(drillPayload.Summary, "SOAR") || !drillPayload.State.Chain5651Valid || drillPayload.State.Chain5651Count < 5 {
		t.Fatalf("SOAR veya 5651 zincir sonucu beklenen gibi degil: %+v", drillPayload)
	}

	// 3. Özel Paket Enjektörü kontrolü (Ankara Savunma AR-GE -> İnternet DROP olmalı)
	pktJSON := `{"branch":"ankara-ofisi","src_zone":"RND_DEFENSE","dst_zone":"UNTRUST_WAN","src_ip":"10.20.40.12","dst_ip":"8.8.8.8","user_group":"GRP_Defense_RnD","application":"ssl","dst_port":443}`
	reqPkt := httptest.NewRequest(http.MethodPost, "/api/packet", bytes.NewBufferString(pktJSON))
	recPkt := httptest.NewRecorder()
	handler.ServeHTTP(recPkt, reqPkt)
	if recPkt.Code != http.StatusOK {
		t.Fatalf("Ozel paket gonderilemedi: status=%d", recPkt.Code)
	}
	var pktOut struct {
		LastEvent struct {
			Decision struct {
				Allowed bool   `json:"Allowed"`
				Action  string `json:"Action"`
			} `json:"decision"`
		} `json:"last_event"`
	}
	if err := json.NewDecoder(recPkt.Body).Decode(&pktOut); err != nil {
		t.Fatalf("Paket JSON cozulemedi: %v", err)
	}
	if pktOut.LastEvent.Decision.Allowed || pktOut.LastEvent.Decision.Action != "DROP" {
		t.Fatalf("Ankara Savunma AR-GE dis internet paketi DROP olmaliydi: %+v", pktOut)
	}

	// 4. HA Donanım Devretme (Failover) kontrolü
	reqHA := httptest.NewRequest(http.MethodPost, "/api/drill", bytes.NewBufferString(`{"drill":"ha_failover","branch":"istanbul-ofisi"}`))
	recHA := httptest.NewRecorder()
	handler.ServeHTTP(recHA, reqHA)
	if recHA.Code != http.StatusOK || !strings.Contains(recHA.Body.String(), "FW02-SECONDARY") {
		t.Fatalf("HA Failover beklenen sekilde FW02-SECONDARY dugumune gecmedi")
	}
}
