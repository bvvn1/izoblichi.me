package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TopCounterpart struct {
	EIK        *string  `json:"eik"`
	Name       *string  `json:"name"`
	Wins       int64    `json:"wins"`
	TotalValue *float64 `json:"total_value"`
	PctByCount float64  `json:"pct_by_count"`
}

type YearBreakdown struct {
	Year          int64    `json:"year"`
	ContractCount int64    `json:"contract_count"`
	TotalValue    *float64 `json:"total_value"`
}

type BuyerFlags struct {
	SupplierDominance   bool    `json:"supplier_dominance"`
	DominantSupplierEIK *string `json:"dominant_supplier_eik,omitempty"`
	DominantSupplierPct float64 `json:"dominant_supplier_pct"`
	NoBidCount          int64   `json:"no_bid_count"`
	NearThresholdCount  int64   `json:"near_threshold_count"`
}

type BuyerProfile struct {
	EIK             string           `json:"eik"`
	Name            string           `json:"name"`
	AddressLocality *string          `json:"address_locality"`
	AddressRegion   *string          `json:"address_region"`
	TotalContracts  int64            `json:"total_contracts"`
	TotalValue      *float64         `json:"total_value"`
	Currency        string           `json:"currency"`
	YearsActive     []int64          `json:"years_active"`
	TopSuppliers    []TopCounterpart `json:"top_suppliers"`
	YearBreakdown   []YearBreakdown  `json:"year_breakdown"`
	Flags           BuyerFlags       `json:"flags"`
}

type SupplierFlags struct {
	BuyerConcentration bool    `json:"buyer_concentration"`
	TopBuyerPct        float64 `json:"top_buyer_pct"`
}

type SupplierProfile struct {
	EIK             string           `json:"eik"`
	Name            string           `json:"name"`
	AddressLocality *string          `json:"address_locality"`
	AddressRegion   *string          `json:"address_region"`
	TotalWins       int64            `json:"total_wins"`
	TotalValue      *float64         `json:"total_value"`
	Currency        string           `json:"currency"`
	TopBuyers       []TopCounterpart `json:"top_buyers"`
	YearBreakdown   []YearBreakdown  `json:"year_breakdown"`
	Flags           SupplierFlags    `json:"flags"`
}

func (a *App) getBuyer(c *gin.Context) {
	eik := c.Param("eik")

	var name, locality, region sql.NullString
	if err := a.db.QueryRow(
		`SELECT COALESCE(display_name, legal_name, ''), address_locality, address_region FROM parties WHERE eik = ?`, eik,
	).Scan(&name, &locality, &region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalContracts int64
	var totalValue sql.NullFloat64
	if err := a.db.QueryRow(
		`SELECT COUNT(*), SUM(contract_value) FROM contracts_unified WHERE buyer_eik = ?`, eik,
	).Scan(&totalContracts, &totalValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if totalContracts == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "buyer not found"})
		return
	}

	rows, err := a.db.Query(`
		SELECT
			supplier_eik, supplier_name,
			COUNT(*) AS wins,
			SUM(contract_value) AS total_val,
			COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () AS pct
		FROM contracts_unified
		WHERE buyer_eik = ? AND supplier_eik IS NOT NULL
		GROUP BY supplier_eik, supplier_name
		ORDER BY wins DESC
		LIMIT 10`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var topSuppliers []TopCounterpart
	var dominantEIK *string
	var dominantPct float64
	for rows.Next() {
		var sEIK, sName sql.NullString
		var wins int64
		var val sql.NullFloat64
		var pct float64
		if err := rows.Scan(&sEIK, &sName, &wins, &val, &pct); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if dominantEIK == nil && pct >= DominanceThreshold*100 {
			dominantEIK = nullStr(sEIK)
			dominantPct = pct
		}
		topSuppliers = append(topSuppliers, TopCounterpart{
			EIK:        nullStr(sEIK),
			Name:       nullStr(sName),
			Wins:       wins,
			TotalValue: nullF64(val),
			PctByCount: pct,
		})
	}
	rows.Close()

	yearRows, err := a.db.Query(`
		SELECT year, COUNT(*), SUM(contract_value)
		FROM contracts_unified WHERE buyer_eik = ? AND year IS NOT NULL
		GROUP BY year ORDER BY year`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer yearRows.Close()

	var yearBreakdown []YearBreakdown
	var yearsActive []int64
	for yearRows.Next() {
		var yr, cnt int64
		var val sql.NullFloat64
		if err := yearRows.Scan(&yr, &cnt, &val); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		yearsActive = append(yearsActive, yr)
		yearBreakdown = append(yearBreakdown, YearBreakdown{Year: yr, ContractCount: cnt, TotalValue: nullF64(val)})
	}
	yearRows.Close()

	var noBidCount int64
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM contracts_unified WHERE buyer_eik = ? AND bid_count <= 1`, eik).Scan(&noBidCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nearThresholdCount int64
	if err := a.db.QueryRow(
		`
		SELECT COUNT(*) FROM contracts_unified
		WHERE buyer_eik = ? AND currency = 'BGN'
		AND contract_value BETWEEN ? AND ?`,
		eik, ThresholdGoods*(1-NearThresholdMargin), ThresholdGoods,
	).Scan(&nearThresholdCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	displayName := name.String
	if displayName == "" {
		displayName = eik
	}
	if topSuppliers == nil {
		topSuppliers = []TopCounterpart{}
	}
	if yearBreakdown == nil {
		yearBreakdown = []YearBreakdown{}
	}
	if yearsActive == nil {
		yearsActive = []int64{}
	}

	c.JSON(http.StatusOK, BuyerProfile{
		EIK:             eik,
		Name:            displayName,
		AddressLocality: nullStr(locality),
		AddressRegion:   nullStr(region),
		TotalContracts:  totalContracts,
		TotalValue:      nullF64(totalValue),
		Currency:        "BGN",
		YearsActive:     yearsActive,
		TopSuppliers:    topSuppliers,
		YearBreakdown:   yearBreakdown,
		Flags: BuyerFlags{
			SupplierDominance:   dominantEIK != nil,
			DominantSupplierEIK: dominantEIK,
			DominantSupplierPct: dominantPct,
			NoBidCount:          noBidCount,
			NearThresholdCount:  nearThresholdCount,
		},
	})
}

func (a *App) getSupplier(c *gin.Context) {
	eik := c.Param("eik")

	var name, locality, region sql.NullString
	if err := a.db.QueryRow(
		`SELECT COALESCE(display_name, legal_name, ''), address_locality, address_region FROM parties WHERE eik = ?`, eik,
	).Scan(&name, &locality, &region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalWins int64
	var totalValue sql.NullFloat64
	if err := a.db.QueryRow(
		`SELECT COUNT(*), SUM(contract_value) FROM contracts_unified WHERE supplier_eik = ?`, eik,
	).Scan(&totalWins, &totalValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if totalWins == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "supplier not found"})
		return
	}

	rows, err := a.db.Query(`
		SELECT
			buyer_eik, buyer_name,
			COUNT(*) AS wins,
			SUM(contract_value) AS total_val,
			COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () AS pct
		FROM contracts_unified
		WHERE supplier_eik = ? AND buyer_eik IS NOT NULL
		GROUP BY buyer_eik, buyer_name
		ORDER BY wins DESC
		LIMIT 10`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var topBuyers []TopCounterpart
	var topPct float64
	for rows.Next() {
		var bEIK, bName sql.NullString
		var wins int64
		var val sql.NullFloat64
		var pct float64
		if err := rows.Scan(&bEIK, &bName, &wins, &val, &pct); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(topBuyers) == 0 {
			topPct = pct
		}
		topBuyers = append(topBuyers, TopCounterpart{
			EIK:        nullStr(bEIK),
			Name:       nullStr(bName),
			Wins:       wins,
			TotalValue: nullF64(val),
			PctByCount: pct,
		})
	}
	rows.Close()

	yearRows, err := a.db.Query(`
		SELECT year, COUNT(*), SUM(contract_value)
		FROM contracts_unified WHERE supplier_eik = ? AND year IS NOT NULL
		GROUP BY year ORDER BY year`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer yearRows.Close()

	var yearBreakdown []YearBreakdown
	for yearRows.Next() {
		var yr, cnt int64
		var val sql.NullFloat64
		if err := yearRows.Scan(&yr, &cnt, &val); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		yearBreakdown = append(yearBreakdown, YearBreakdown{Year: yr, ContractCount: cnt, TotalValue: nullF64(val)})
	}
	yearRows.Close()

	displayName := name.String
	if displayName == "" {
		displayName = eik
	}
	if topBuyers == nil {
		topBuyers = []TopCounterpart{}
	}
	if yearBreakdown == nil {
		yearBreakdown = []YearBreakdown{}
	}

	c.JSON(http.StatusOK, SupplierProfile{
		EIK:             eik,
		Name:            displayName,
		AddressLocality: nullStr(locality),
		AddressRegion:   nullStr(region),
		TotalWins:       totalWins,
		TotalValue:      nullF64(totalValue),
		Currency:        "BGN",
		TopBuyers:       topBuyers,
		YearBreakdown:   yearBreakdown,
		Flags: SupplierFlags{
			BuyerConcentration: topPct >= DominanceThreshold*100,
			TopBuyerPct:        topPct,
		},
	})
}
