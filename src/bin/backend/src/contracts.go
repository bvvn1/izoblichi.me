package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Contract struct {
	DataSource          string   `json:"data_source"`
	RowKey              string   `json:"row_key"`
	OCID                *string  `json:"ocid"`
	Year                *int64   `json:"year"`
	ContractDate        *string  `json:"contract_date"`
	BuyerEIK            *string  `json:"buyer_eik"`
	BuyerName           *string  `json:"buyer_name"`
	SupplierEIK         *string  `json:"supplier_eik"`
	SupplierName        *string  `json:"supplier_name"`
	ContractValue       *float64 `json:"contract_value"`
	Currency            *string  `json:"currency"`
	ProcurementNumber   *string  `json:"procurement_number"`
	Title               *string  `json:"title"`
	ProcurementCategory *string  `json:"procurement_category"`
	BidCount            *int64   `json:"bid_count"`
	ProcurementMethod   *string  `json:"procurement_method"`
	EUFunded            *bool    `json:"eu_funded"`
}

type ContractsResponse struct {
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Items   []Contract `json:"items"`
}

func (a *App) listContracts(c *gin.Context) {
	q := c.Query("q")
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom := c.Query("year_from")
	yearTo := c.Query("year_to")
	minValue := c.Query("min_value")
	maxValue := c.Query("max_value")
	category := c.Query("category")
	source := c.Query("source")

	page, perPage := paginate(c)
	offset := (page - 1) * perPage

	conds := []string{}
	args := []any{}

	if q != "" {
		conds = append(conds, "(title ILIKE ? OR buyer_name ILIKE ? OR supplier_name ILIKE ?)")
		pct := "%" + q + "%"
		args = append(args, pct, pct, pct)
	}
	if buyerEIK != "" {
		conds = append(conds, "buyer_eik = ?")
		args = append(args, buyerEIK)
	}
	if supplierEIK != "" {
		conds = append(conds, "supplier_eik = ?")
		args = append(args, supplierEIK)
	}
	if yearFrom != "" {
		conds = append(conds, "year >= ?")
		args = append(args, yearFrom)
	}
	if yearTo != "" {
		conds = append(conds, "year <= ?")
		args = append(args, yearTo)
	}
	if minValue != "" {
		conds = append(conds, "contract_value >= ?")
		args = append(args, minValue)
	}
	if maxValue != "" {
		conds = append(conds, "contract_value <= ?")
		args = append(args, maxValue)
	}
	if category != "" {
		conds = append(conds, "procurement_category = ?")
		args = append(args, category)
	}
	if source != "" {
		conds = append(conds, "data_source = ?")
		args = append(args, source)
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	query := `
		SELECT
			data_source, row_key, ocid, year, contract_date,
			buyer_eik, buyer_name, supplier_eik, supplier_name,
			contract_value, currency, procurement_number, title,
			procurement_category, bid_count, procurement_method, eu_funded,
			COUNT(*) OVER () AS total
		FROM contracts_unified
		` + where + `
		ORDER BY contract_date DESC NULLS LAST
		LIMIT ? OFFSET ?`

	queryArgs := append(args, perPage, offset)
	rows, err := a.db.Query(query, queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []Contract
	var total int64
	for rows.Next() {
		var (
			dataSource, rowKey         string
			ocid                       sql.NullString
			year                       sql.NullInt64
			contractDate               sql.NullTime
			bEIK, bName                sql.NullString
			sEIK, sName                sql.NullString
			contractValue              sql.NullFloat64
			currency                   sql.NullString
			procNumber, title          sql.NullString
			procCategory               sql.NullString
			bidCount                   sql.NullInt64
			procMethod                 sql.NullString
			euFunded                   sql.NullBool
			rowTotal                   int64
		)
		if err := rows.Scan(
			&dataSource, &rowKey, &ocid, &year, &contractDate,
			&bEIK, &bName, &sEIK, &sName,
			&contractValue, &currency, &procNumber, &title,
			&procCategory, &bidCount, &procMethod, &euFunded,
			&rowTotal,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total = rowTotal
		items = append(items, Contract{
			DataSource:          dataSource,
			RowKey:              rowKey,
			OCID:                nullStr(ocid),
			Year:                nullI64(year),
			ContractDate:        nullDate(contractDate),
			BuyerEIK:            nullStr(bEIK),
			BuyerName:           nullStr(bName),
			SupplierEIK:         nullStr(sEIK),
			SupplierName:        nullStr(sName),
			ContractValue:       nullF64(contractValue),
			Currency:            nullStr(currency),
			ProcurementNumber:   nullStr(procNumber),
			Title:               nullStr(title),
			ProcurementCategory: nullStr(procCategory),
			BidCount:            nullI64(bidCount),
			ProcurementMethod:   nullStr(procMethod),
			EUFunded:            nullBool(euFunded),
		})
	}
	if items == nil {
		items = []Contract{}
	}

	c.JSON(http.StatusOK, ContractsResponse{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	})
}
