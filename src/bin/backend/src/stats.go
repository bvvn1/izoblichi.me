package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonthStat struct {
	Year             int64   `json:"year"`
	Month            int64   `json:"month"`
	ContractCount    int64   `json:"contract_count"`
	TotalValueBGN    float64 `json:"total_value_bgn"`
	NoBidCount       int64   `json:"no_bid_count"`
	CompetitiveCount int64   `json:"competitive_count"`
}

type TimeseriesResponse struct {
	Months []MonthStat `json:"months"`
}

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

func (a *App) getTimeseries(c *gin.Context) {
	rows, err := a.db.Query(`
		SELECT
			year,
			EXTRACT(MONTH FROM contract_date)::BIGINT AS month,
			COUNT(*) AS contract_count,
			SUM(CASE WHEN currency = 'BGN' THEN contract_value ELSE 0 END) AS total_value_bgn,
			SUM(CASE WHEN bid_count <= 1 THEN 1 ELSE 0 END) AS no_bid_count,
			SUM(CASE WHEN bid_count > 1 THEN 1 ELSE 0 END) AS competitive_count
		FROM contracts_unified
		WHERE year IS NOT NULL AND contract_date IS NOT NULL
		GROUP BY year, EXTRACT(MONTH FROM contract_date)
		ORDER BY year, month`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var months []MonthStat
	for rows.Next() {
		var s MonthStat
		var monthF float64
		if err := rows.Scan(&s.Year, &monthF, &s.ContractCount, &s.TotalValueBGN, &s.NoBidCount, &s.CompetitiveCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		s.Month = int64(monthF)
		months = append(months, s)
	}
	if months == nil {
		months = []MonthStat{}
	}
	c.JSON(http.StatusOK, TimeseriesResponse{Months: months})
}

func (a *App) getStats(c *gin.Context) {
	var totalContracts, uniqueBuyers, uniqueSuppliers int64
	var totalValue sql.NullFloat64

	if err := a.db.QueryRow(
		`
		SELECT COUNT(*),
		       SUM(CASE WHEN currency = 'BGN' THEN contract_value ELSE NULL END),
		       COUNT(DISTINCT buyer_eik),
		       COUNT(DISTINCT supplier_eik)
		FROM contracts_unified`,
	).Scan(&totalContracts, &totalValue, &uniqueBuyers, &uniqueSuppliers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := a.db.Query(`
		SELECT year, COUNT(*), SUM(contract_value)
		FROM contracts_unified WHERE year IS NOT NULL
		GROUP BY year ORDER BY year`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var yearBreakdown []YearStat
	for rows.Next() {
		var yr, cnt int64
		var val sql.NullFloat64
		if err := rows.Scan(&yr, &cnt, &val); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		yearBreakdown = append(yearBreakdown, YearStat{Year: yr, ContractCount: cnt, TotalValue: nullF64(val)})
	}
	rows.Close()

	var dataThrough sql.NullTime
	if err := a.db.QueryRow(`SELECT MAX(source_file_date) FROM ocds_releases`).Scan(&dataThrough); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
