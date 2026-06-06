package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

type AnomaliesResponse struct {
	NearThreshold []NearThresholdAnomaly `json:"near_threshold"`
	NoBid         []NoBidAnomaly         `json:"no_bid"`
	Dominance     []DominanceAnomaly     `json:"dominance"`
}

func (a *App) getAnomalies(c *gin.Context) {
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom := c.Query("year_from")
	yearTo := c.Query("year_to")
	types := c.QueryArray("type")
	if len(types) == 0 {
		types = []string{"near_threshold", "no_bid", "dominance"}
	}
	typeSet := map[string]bool{}
	for _, t := range types {
		typeSet[t] = true
	}

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
		NearThreshold: []NearThresholdAnomaly{},
		NoBid:         []NoBidAnomaly{},
		Dominance:     []DominanceAnomaly{},
	}

	if typeSet["near_threshold"] {
		low := ThresholdGoods * (1 - NearThresholdMargin)
		q := `
			SELECT contract_value, buyer_eik, buyer_name, supplier_eik, supplier_name,
			       contract_date, title
			FROM contracts_unified
			WHERE currency = 'BGN' AND contract_value BETWEEN ? AND ? ` + baseAnd + `
			ORDER BY contract_value DESC LIMIT 50`
		args := append([]any{low, ThresholdGoods}, baseArgs...)
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
				title                    sql.NullString
			)
			rows.Scan(&val, &bEIK, &bName, &sEIK, &sName, &cDate, &title)
			gap := ThresholdGoods - val
			resp.NearThreshold = append(resp.NearThreshold, NearThresholdAnomaly{
				ContractValue: val,
				Threshold:     ThresholdGoods,
				Gap:           gap,
				GapPct:        gap / ThresholdGoods * 100,
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

	if typeSet["no_bid"] {
		q := `
			SELECT bid_count, contract_value, currency,
			       buyer_eik, buyer_name, supplier_eik, supplier_name,
			       contract_date, title
			FROM contracts_unified
			WHERE bid_count <= 1 ` + baseAnd + `
			ORDER BY contract_value DESC NULLS LAST LIMIT 50`
		rows, err := a.db.Query(q, baseArgs...)
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
			)
			rows.Scan(&bidCount, &val, &curr, &bEIK, &bName, &sEIK, &sName, &cDate, &title)
			resp.NoBid = append(resp.NoBid, NoBidAnomaly{
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
			)
			SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
			       wins, total_val, wins * 100.0 / total_buyer AS pct, total_buyer
			FROM agg
			WHERE wins >= 5 AND wins * 100.0 / total_buyer >= ?
			ORDER BY pct DESC LIMIT 50`
		domArgs = append(domArgs, DominanceThreshold*100)
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
			)
			rows.Scan(&bEIK, &bName, &sEIK, &sName, &wins, &val, &pct, &totalBuyer)
			resp.Dominance = append(resp.Dominance, DominanceAnomaly{
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

	c.JSON(http.StatusOK, resp)
}
