package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type YearStat struct {
	Year          int64    `json:"year"`
	ContractCount int64    `json:"contract_count"`
	TotalValue    *float64 `json:"total_value"`
}

type StatsResponse struct {
	TotalContracts  int64      `json:"total_contracts"`
	TotalValueBGN   *float64   `json:"total_value_bgn"`
	UniqueBuyers    int64      `json:"unique_buyers"`
	UniqueSuppliers int64      `json:"unique_suppliers"`
	YearBreakdown   []YearStat `json:"year_breakdown"`
	DataThrough     *string    `json:"data_through"`
}

func (a *App) getStats(c *gin.Context) {
	var totalContracts, uniqueBuyers, uniqueSuppliers int64
	var totalValue sql.NullFloat64

	a.db.QueryRow(`
		SELECT COUNT(*),
		       SUM(CASE WHEN currency = 'BGN' THEN contract_value ELSE NULL END),
		       COUNT(DISTINCT buyer_eik),
		       COUNT(DISTINCT supplier_eik)
		FROM contracts_unified`,
	).Scan(&totalContracts, &totalValue, &uniqueBuyers, &uniqueSuppliers)

	rows, err := a.db.Query(`
		SELECT year, COUNT(*), SUM(contract_value)
		FROM contracts_unified WHERE year IS NOT NULL
		GROUP BY year ORDER BY year`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var yearBreakdown []YearStat
	for rows.Next() {
		var yr, cnt int64
		var val sql.NullFloat64
		rows.Scan(&yr, &cnt, &val)
		yearBreakdown = append(yearBreakdown, YearStat{Year: yr, ContractCount: cnt, TotalValue: nullF64(val)})
	}
	rows.Close()

	var dataThrough sql.NullTime
	a.db.QueryRow(`SELECT MAX(source_file_date) FROM ocds_releases`).Scan(&dataThrough)

	if yearBreakdown == nil {
		yearBreakdown = []YearStat{}
	}

	c.JSON(http.StatusOK, StatsResponse{
		TotalContracts:  totalContracts,
		TotalValueBGN:   nullF64(totalValue),
		UniqueBuyers:    uniqueBuyers,
		UniqueSuppliers: uniqueSuppliers,
		YearBreakdown:   yearBreakdown,
		DataThrough:     nullDate(dataThrough),
	})
}
