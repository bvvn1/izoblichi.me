package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Bulgarian procurement thresholds (BGN). Works threshold is higher.
const (
	ThresholdGoods         = 70_000.0
	ThresholdServices      = 70_000.0
	ThresholdWorks         = 264_033.0
	DominanceThreshold     = 0.80
	NearThresholdMargin    = 0.05
	RepeatedAwardMinYears  = 2
	RepeatedAwardMinWins   = 3
)

type NearThresholdAnomaly struct {
	ContractValue float64  `json:"contract_value"`
	Threshold     float64  `json:"threshold"`
	Gap           float64  `json:"gap"`
	GapPct        float64  `json:"gap_pct"`
	BuyerEIK      *string  `json:"buyer_eik"`
	BuyerName     *string  `json:"buyer_name"`
	SupplierEIK   *string  `json:"supplier_eik"`
	SupplierName  *string  `json:"supplier_name"`
	ContractDate  *string  `json:"contract_date"`
	Title         *string  `json:"title"`
	Category      *string  `json:"procurement_category"`
}

type NoBidAnomaly struct {
	BidCount      *int64   `json:"bid_count"`
	ContractValue *float64 `json:"contract_value"`
	Currency      *string  `json:"currency"`
	BuyerEIK      *string  `json:"buyer_eik"`
	BuyerName     *string  `json:"buyer_name"`
	SupplierEIK   *string  `json:"supplier_eik"`
	SupplierName  *string  `json:"supplier_name"`
	ContractDate  *string  `json:"contract_date"`
	Title         *string  `json:"title"`
}

type DominanceAnomaly struct {
	BuyerEIK            *string  `json:"buyer_eik"`
	BuyerName           *string  `json:"buyer_name"`
	SupplierEIK         *string  `json:"supplier_eik"`
	SupplierName        *string  `json:"supplier_name"`
	Wins                int64    `json:"wins"`
	TotalValue          *float64 `json:"total_value"`
	PctByCount          float64  `json:"pct_by_count"`
	TotalBuyerContracts int64    `json:"total_buyer_contracts"`
}

type RepeatedAwardAnomaly struct {
	BuyerEIK    *string  `json:"buyer_eik"`
	BuyerName   *string  `json:"buyer_name"`
	SupplierEIK *string  `json:"supplier_eik"`
	SupplierName *string `json:"supplier_name"`
	TotalWins   int64    `json:"total_wins"`
	YearsActive int64    `json:"years_active"`
	TotalValue  *float64 `json:"total_value"`
	Years       []int64  `json:"years"`
}

type AnomalyPage struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

type AnomaliesResponse struct {
	NearThreshold AnomalySection[NearThresholdAnomaly] `json:"near_threshold"`
	NoBid         AnomalySection[NoBidAnomaly]         `json:"no_bid"`
	Dominance     AnomalySection[DominanceAnomaly]     `json:"dominance"`
	RepeatedAward AnomalySection[RepeatedAwardAnomaly] `json:"repeated_award"`
}

type AnomalySection[T any] struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Items   []T   `json:"items"`
}

func (a *App) getAnomalies(c *gin.Context) {
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom := c.Query("year_from")
	yearTo := c.Query("year_to")
	types := c.QueryArray("type")
	if len(types) == 0 {
		types = []string{"near_threshold", "no_bid", "dominance", "repeated_award"}
	}
	typeSet := map[string]bool{}
	for _, t := range types {
		typeSet[t] = true
	}

	page, perPage := paginate(c)
	offset := (page - 1) * perPage

	baseConds := []string{}
	baseArgs := []any{}
	if buyerEIK != "" {
		baseConds = append(baseConds, "buyer_eik = ?")
		baseArgs = append(baseArgs, buyerEIK)
	}
	if supplierEIK != "" {
		baseConds = append(baseConds, "supplier_eik = ?")
		baseArgs = append(baseArgs, supplierEIK)
	}
	if yearFrom != "" {
		baseConds = append(baseConds, "year >= ?")
		baseArgs = append(baseArgs, yearFrom)
	}
	if yearTo != "" {
		baseConds = append(baseConds, "year <= ?")
		baseArgs = append(baseArgs, yearTo)
	}
	baseAnd := ""
	if len(baseConds) > 0 {
		baseAnd = "AND " + strings.Join(baseConds, " AND ")
	}

	resp := AnomaliesResponse{
		NearThreshold: AnomalySection[NearThresholdAnomaly]{Items: []NearThresholdAnomaly{}, Page: page, PerPage: perPage},
		NoBid:         AnomalySection[NoBidAnomaly]{Items: []NoBidAnomaly{}, Page: page, PerPage: perPage},
		Dominance:     AnomalySection[DominanceAnomaly]{Items: []DominanceAnomaly{}, Page: page, PerPage: perPage},
		RepeatedAward: AnomalySection[RepeatedAwardAnomaly]{Items: []RepeatedAwardAnomaly{}, Page: page, PerPage: perPage},
	}

	if typeSet["near_threshold"] {
		// Check against goods/services threshold (70k) and works threshold (264k).
		// A contract just below a threshold that would trigger a more competitive
		// procedure is suspicious regardless of category.
		q := `
			SELECT contract_value, buyer_eik, buyer_name, supplier_eik, supplier_name,
			       contract_date, title, procurement_category,
			       COUNT(*) OVER () AS total
			FROM contracts_unified
			WHERE currency = 'BGN'
			  AND (
			    (procurement_category NOT IN ('Строителство', 'works')
			     AND contract_value BETWEEN ? AND ?)
			    OR
			    (procurement_category IN ('Строителство', 'works')
			     AND contract_value BETWEEN ? AND ?)
			  ) ` + baseAnd + `
			ORDER BY contract_value DESC
			LIMIT ? OFFSET ?`
		args := append([]any{
			ThresholdGoods * (1 - NearThresholdMargin), ThresholdGoods,
			ThresholdWorks * (1 - NearThresholdMargin), ThresholdWorks,
		}, baseArgs...)
		args = append(args, perPage, offset)

		rows, err := a.db.Query(q, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for rows.Next() {
			var (
				val                      float64
				bEIK, bName, sEIK, sName sql.NullString
				cDate                    sql.NullTime
				title, cat               sql.NullString
				rowTotal                 int64
			)
			rows.Scan(&val, &bEIK, &bName, &sEIK, &sName, &cDate, &title, &cat, &rowTotal)
			resp.NearThreshold.Total = rowTotal

			threshold := ThresholdGoods
			catStr := ""
			if cat.Valid {
				catStr = cat.String
			}
			if catStr == "Строителство" || catStr == "works" {
				threshold = ThresholdWorks
			}
			gap := threshold - val
			resp.NearThreshold.Items = append(resp.NearThreshold.Items, NearThresholdAnomaly{
				ContractValue: val,
				Threshold:     threshold,
				Gap:           gap,
				GapPct:        gap / threshold * 100,
				BuyerEIK:      nullStr(bEIK),
				BuyerName:     nullStr(bName),
				SupplierEIK:   nullStr(sEIK),
				SupplierName:  nullStr(sName),
				ContractDate:  nullDate(cDate),
				Title:         nullStr(title),
				Category:      nullStr(cat),
			})
		}
		rows.Close()
	}

	if typeSet["no_bid"] {
		q := `
			SELECT bid_count, contract_value, currency,
			       buyer_eik, buyer_name, supplier_eik, supplier_name,
			       contract_date, title,
			       COUNT(*) OVER () AS total
			FROM contracts_unified
			WHERE bid_count <= 1 ` + baseAnd + `
			ORDER BY contract_value DESC NULLS LAST
			LIMIT ? OFFSET ?`
		args := append(baseArgs, perPage, offset)
		rows, err := a.db.Query(q, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for rows.Next() {
			var (
				bidCount                 sql.NullInt64
				val                      sql.NullFloat64
				curr                     sql.NullString
				bEIK, bName, sEIK, sName sql.NullString
				cDate                    sql.NullTime
				title                    sql.NullString
				rowTotal                 int64
			)
			rows.Scan(&bidCount, &val, &curr, &bEIK, &bName, &sEIK, &sName, &cDate, &title, &rowTotal)
			resp.NoBid.Total = rowTotal
			resp.NoBid.Items = append(resp.NoBid.Items, NoBidAnomaly{
				BidCount:      nullI64(bidCount),
				ContractValue: nullF64(val),
				Currency:      nullStr(curr),
				BuyerEIK:      nullStr(bEIK),
				BuyerName:     nullStr(bName),
				SupplierEIK:   nullStr(sEIK),
				SupplierName:  nullStr(sName),
				ContractDate:  nullDate(cDate),
				Title:         nullStr(title),
			})
		}
		rows.Close()
	}

	if typeSet["dominance"] {
		domConds := []string{}
		domArgs := []any{}
		if buyerEIK != "" {
			domConds = append(domConds, "buyer_eik = ?")
			domArgs = append(domArgs, buyerEIK)
		}
		if supplierEIK != "" {
			domConds = append(domConds, "supplier_eik = ?")
			domArgs = append(domArgs, supplierEIK)
		}
		if yearFrom != "" {
			domConds = append(domConds, "year >= ?")
			domArgs = append(domArgs, yearFrom)
		}
		if yearTo != "" {
			domConds = append(domConds, "year <= ?")
			domArgs = append(domArgs, yearTo)
		}
		domWhere := ""
		if len(domConds) > 0 {
			domWhere = "WHERE " + strings.Join(domConds, " AND ")
		}

		q := `
			WITH agg AS (
				SELECT
					buyer_eik, buyer_name,
					supplier_eik, supplier_name,
					COUNT(*) AS wins,
					SUM(contract_value) AS total_val,
					SUM(COUNT(*)) OVER (PARTITION BY buyer_eik) AS total_buyer
				FROM contracts_unified
				` + domWhere + `
				GROUP BY buyer_eik, buyer_name, supplier_eik, supplier_name
			),
			filtered AS (
				SELECT *, wins * 100.0 / total_buyer AS pct
				FROM agg
				WHERE wins >= 5 AND wins * 100.0 / total_buyer >= ?
			)
			SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
			       wins, total_val, pct, total_buyer,
			       COUNT(*) OVER () AS total
			FROM filtered
			ORDER BY pct DESC
			LIMIT ? OFFSET ?`
		domArgs = append(domArgs, DominanceThreshold*100, perPage, offset)
		rows, err := a.db.Query(q, domArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for rows.Next() {
			var (
				bEIK, bName, sEIK, sName sql.NullString
				wins, totalBuyer         int64
				val                      sql.NullFloat64
				pct                      float64
				rowTotal                 int64
			)
			rows.Scan(&bEIK, &bName, &sEIK, &sName, &wins, &val, &pct, &totalBuyer, &rowTotal)
			resp.Dominance.Total = rowTotal
			resp.Dominance.Items = append(resp.Dominance.Items, DominanceAnomaly{
				BuyerEIK:            nullStr(bEIK),
				BuyerName:           nullStr(bName),
				SupplierEIK:         nullStr(sEIK),
				SupplierName:        nullStr(sName),
				Wins:                wins,
				TotalValue:          nullF64(val),
				PctByCount:          pct,
				TotalBuyerContracts: totalBuyer,
			})
		}
		rows.Close()
	}

	if typeSet["repeated_award"] {
		raConds := []string{}
		raArgs := []any{}
		if buyerEIK != "" {
			raConds = append(raConds, "buyer_eik = ?")
			raArgs = append(raArgs, buyerEIK)
		}
		if supplierEIK != "" {
			raConds = append(raConds, "supplier_eik = ?")
			raArgs = append(raArgs, supplierEIK)
		}
		if yearFrom != "" {
			raConds = append(raConds, "year >= ?")
			raArgs = append(raArgs, yearFrom)
		}
		if yearTo != "" {
			raConds = append(raConds, "year <= ?")
			raArgs = append(raArgs, yearTo)
		}
		raWhere := ""
		if len(raConds) > 0 {
			raWhere = "WHERE " + strings.Join(raConds, " AND ")
		}

		// Pairs that won contracts in at least N distinct years with a combined
		// single-bid rate ≥ 50%, indicating a recurring no-competition pattern.
		q := `
			WITH pair_years AS (
				SELECT
					buyer_eik, buyer_name,
					supplier_eik, supplier_name,
					year,
					COUNT(*) AS year_wins,
					SUM(CASE WHEN bid_count <= 1 THEN 1 ELSE 0 END) AS year_no_bid,
					SUM(contract_value) AS year_value
				FROM contracts_unified
				` + raWhere + `
				GROUP BY buyer_eik, buyer_name, supplier_eik, supplier_name, year
			),
			pair_agg AS (
				SELECT
					buyer_eik, buyer_name,
					supplier_eik, supplier_name,
					SUM(year_wins) AS total_wins,
					COUNT(DISTINCT year) AS years_active,
					SUM(year_no_bid) AS total_no_bid,
					SUM(year_value) AS total_value,
					CAST(to_json(list(year ORDER BY year)) AS VARCHAR) AS years_list
				FROM pair_years
				GROUP BY buyer_eik, buyer_name, supplier_eik, supplier_name
			),
			filtered AS (
				SELECT *
				FROM pair_agg
				WHERE years_active >= ?
				  AND total_wins >= ?
				  AND total_no_bid * 100.0 / total_wins >= 50
			)
			SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
			       total_wins, years_active, total_value, years_list,
			       COUNT(*) OVER () AS total
			FROM filtered
			ORDER BY years_active DESC, total_wins DESC
			LIMIT ? OFFSET ?`
		raArgs = append(raArgs, RepeatedAwardMinYears, RepeatedAwardMinWins, perPage, offset)
		rows, err := a.db.Query(q, raArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for rows.Next() {
			var (
				bEIK, bName  sql.NullString
				sEIK, sName  sql.NullString
				totalWins    int64
				yearsActive  int64
				totalValue   sql.NullFloat64
				yearsJSON    sql.NullString
				rowTotal     int64
			)
			if err := rows.Scan(&bEIK, &bName, &sEIK, &sName,
				&totalWins, &yearsActive, &totalValue, &yearsJSON, &rowTotal); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			resp.RepeatedAward.Total = rowTotal
			var years []int64
			if yearsJSON.Valid && yearsJSON.String != "" {
				json.Unmarshal([]byte(yearsJSON.String), &years)
			}
			if years == nil {
				years = []int64{}
			}
			resp.RepeatedAward.Items = append(resp.RepeatedAward.Items, RepeatedAwardAnomaly{
				BuyerEIK:     nullStr(bEIK),
				BuyerName:    nullStr(bName),
				SupplierEIK:  nullStr(sEIK),
				SupplierName: nullStr(sName),
				TotalWins:    totalWins,
				YearsActive:  yearsActive,
				TotalValue:   nullF64(totalValue),
				Years:        years,
			})
		}
		rows.Close()
	}

	c.JSON(http.StatusOK, resp)
}
