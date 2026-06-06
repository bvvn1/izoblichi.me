package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MapBuyer struct {
	EIK                string   `json:"eik"`
	Name               string   `json:"name"`
	AddressLocality    *string  `json:"address_locality"`
	AddressRegion      *string  `json:"address_region"`
	TotalContracts     int64    `json:"total_contracts"`
	TotalValueBGN      float64  `json:"total_value_bgn"`
	NoBidCount         int64    `json:"no_bid_count"`
	NearThresholdCount int64    `json:"near_threshold_count"`
	RiskScore          float64  `json:"risk_score"`
}

type MapBuyersResponse struct {
	Total int64      `json:"total"`
	Items []MapBuyer `json:"items"`
}

func (a *App) getMapBuyers(c *gin.Context) {
	rows, err := a.db.Query(`
		WITH buyer_stats AS (
			SELECT
				buyer_eik,
				MAX(buyer_name) AS buyer_name,
				COUNT(*) AS total_contracts,
				SUM(CASE WHEN currency = 'BGN' THEN contract_value ELSE 0 END) AS total_value_bgn,
				SUM(CASE WHEN bid_count <= 1 THEN 1 ELSE 0 END) AS no_bid_count,
				SUM(CASE WHEN currency = 'BGN'
				         AND contract_value BETWEEN ? AND ?
				         THEN 1 ELSE 0 END) AS near_threshold_count
			FROM contracts_unified
			WHERE buyer_eik IS NOT NULL
			GROUP BY buyer_eik
		)
		SELECT
			s.buyer_eik,
			COALESCE(p.display_name, p.legal_name, s.buyer_name, s.buyer_eik) AS name,
			p.address_locality,
			p.address_region,
			s.total_contracts,
			s.total_value_bgn,
			s.no_bid_count,
			s.near_threshold_count,
			LEAST(100.0,
				(s.no_bid_count * 60.0 / s.total_contracts) +
				(s.near_threshold_count * 40.0 / s.total_contracts)
			) AS risk_score
		FROM buyer_stats s
		LEFT JOIN parties p ON p.eik = s.buyer_eik
		ORDER BY s.total_value_bgn DESC NULLS LAST
		LIMIT 2000`,
		ThresholdGoods*(1-NearThresholdMargin), ThresholdGoods,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []MapBuyer
	for rows.Next() {
		var (
			eik              string
			name             sql.NullString
			locality, region sql.NullString
			totalContracts   int64
			totalValueBGN    float64
			noBidCount       int64
			nearThreshCount  int64
			riskScore        float64
		)
		if err := rows.Scan(&eik, &name, &locality, &region,
			&totalContracts, &totalValueBGN, &noBidCount, &nearThreshCount, &riskScore); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		displayName := eik
		if name.Valid && name.String != "" {
			displayName = name.String
		}
		items = append(items, MapBuyer{
			EIK:                eik,
			Name:               displayName,
			AddressLocality:    nullStr(locality),
			AddressRegion:      nullStr(region),
			TotalContracts:     totalContracts,
			TotalValueBGN:      totalValueBGN,
			NoBidCount:         noBidCount,
			NearThresholdCount: nearThreshCount,
			RiskScore:          riskScore,
		})
	}
	if items == nil {
		items = []MapBuyer{}
	}
	c.JSON(http.StatusOK, MapBuyersResponse{Total: int64(len(items)), Items: items})
}
