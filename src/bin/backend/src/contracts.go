package main

import (
	"database/sql"
	"net/http"

	sq "github.com/Masterminds/squirrel"
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
	LegacyType          *string  `json:"legacy_type,omitempty"`
}

type ContractsResponse struct {
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Items   []Contract `json:"items"`
}

var allowedSortCols = map[string]string{
	"contract_date":  "contract_date",
	"contract_value": "contract_value",
	"year":           "year",
	"buyer_name":     "buyer_name",
	"supplier_name":  "supplier_name",
}

func (a *App) listContracts(c *gin.Context) {
	q := c.Query("q")
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom, _ := parseYear(c.Query("year_from"))
	yearTo, _ := parseYear(c.Query("year_to"))
	minValue := c.Query("min_value")
	maxValue := c.Query("max_value")
	category := c.Query("category")
	source := c.Query("source")
	sortBy := c.DefaultQuery("sort_by", "contract_date")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	page, perPage := paginate(c)
	offset := (page - 1) * perPage

	col, ok := allowedSortCols[sortBy]
	if !ok {
		col = "contract_date"
	}
	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}
	orderClause := col + " " + dir + " NULLS LAST"

	qb := sb.Select(
		"data_source", "row_key", "ocid", "year", "contract_date",
		"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
		"contract_value", "currency", "procurement_number", "title",
		"procurement_category", "bid_count", "procurement_method", "eu_funded",
		"legacy_type", "COUNT(*) OVER () AS total",
	).From("contracts_unified")

	if q != "" {
		pct := "%" + q + "%"
		qb = qb.Where(sq.Or{
			sq.Expr("title ILIKE ?", pct),
			sq.Expr("buyer_name ILIKE ?", pct),
			sq.Expr("supplier_name ILIKE ?", pct),
		})
	}
	if buyerEIK != "" {
		qb = qb.Where(sq.Eq{"buyer_eik": buyerEIK})
	}
	if supplierEIK != "" {
		qb = qb.Where(sq.Eq{"supplier_eik": supplierEIK})
	}
	if yearFrom != 0 {
		qb = qb.Where(sq.GtOrEq{"year": yearFrom})
	}
	if yearTo != 0 {
		qb = qb.Where(sq.LtOrEq{"year": yearTo})
	}
	if minValue != "" {
		qb = qb.Where(sq.GtOrEq{"contract_value": minValue})
	}
	if maxValue != "" {
		qb = qb.Where(sq.LtOrEq{"contract_value": maxValue})
	}
	if category != "" {
		qb = qb.Where(sq.Eq{"procurement_category": category})
	}
	if source != "" {
		qb = qb.Where(sq.Eq{"data_source": source})
	}
	qb = qb.OrderBy(orderClause).Limit(uint64(perPage)).Offset(uint64(offset))

	query, queryArgs, err := qb.ToSql()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := a.db.Query(query, queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []Contract
	var total int64
	for rows.Next() {
		var ct Contract
		var (
			ocid         sql.NullString
			year         sql.NullInt64
			contractDate sql.NullTime
			bEIK, bName  sql.NullString
			sEIK, sName  sql.NullString
			cv           sql.NullFloat64
			currency     sql.NullString
			procNum      sql.NullString
			title        sql.NullString
			procCat      sql.NullString
			bidCount     sql.NullInt64
			procMethod   sql.NullString
			euFunded     sql.NullBool
			legacyType   sql.NullString
			rowTotal     int64
		)
		if err := rows.Scan(
			&ct.DataSource, &ct.RowKey, &ocid, &year, &contractDate,
			&bEIK, &bName, &sEIK, &sName,
			&cv, &currency, &procNum, &title,
			&procCat, &bidCount, &procMethod, &euFunded,
			&legacyType,
			&rowTotal,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total = rowTotal
		ct.OCID = nullStr(ocid)
		ct.Year = nullI64(year)
		ct.ContractDate = nullDate(contractDate)
		ct.BuyerEIK = nullStr(bEIK)
		ct.BuyerName = nullStr(bName)
		ct.SupplierEIK = nullStr(sEIK)
		ct.SupplierName = nullStr(sName)
		ct.ContractValue = nullF64(cv)
		ct.Currency = nullStr(currency)
		ct.ProcurementNumber = nullStr(procNum)
		ct.Title = nullStr(title)
		ct.ProcurementCategory = nullStr(procCat)
		ct.BidCount = nullI64(bidCount)
		ct.ProcurementMethod = nullStr(procMethod)
		ct.EUFunded = nullBool(euFunded)
		ct.LegacyType = nullStr(legacyType)
		items = append(items, ct)
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

func (a *App) getContract(c *gin.Context) {
	source := c.Param("source")
	rowKey := c.Param("row_key")

	if source != "legacy" && source != "ocds" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source must be 'legacy' or 'ocds'"})
		return
	}

	query := `
		SELECT
			data_source, row_key, ocid, year, contract_date,
			buyer_eik, buyer_name, supplier_eik, supplier_name,
			contract_value, currency, procurement_number, title,
			procurement_category, bid_count, procurement_method, eu_funded,
			legacy_type
		FROM contracts_unified
		WHERE data_source = ? AND row_key = ?
		LIMIT 1`

	var ct Contract
	var (
		ocid         sql.NullString
		year         sql.NullInt64
		contractDate sql.NullTime
		bEIK, bName  sql.NullString
		sEIK, sName  sql.NullString
		cv           sql.NullFloat64
		currency     sql.NullString
		procNum      sql.NullString
		title        sql.NullString
		procCat      sql.NullString
		bidCount     sql.NullInt64
		procMethod   sql.NullString
		euFunded     sql.NullBool
		legacyType   sql.NullString
	)
	err := a.db.QueryRow(query, source, rowKey).Scan(
		&ct.DataSource, &ct.RowKey, &ocid, &year, &contractDate,
		&bEIK, &bName, &sEIK, &sName,
		&cv, &currency, &procNum, &title,
		&procCat, &bidCount, &procMethod, &euFunded,
		&legacyType,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ct.OCID = nullStr(ocid)
	ct.Year = nullI64(year)
	ct.ContractDate = nullDate(contractDate)
	ct.BuyerEIK = nullStr(bEIK)
	ct.BuyerName = nullStr(bName)
	ct.SupplierEIK = nullStr(sEIK)
	ct.SupplierName = nullStr(sName)
	ct.ContractValue = nullF64(cv)
	ct.Currency = nullStr(currency)
	ct.ProcurementNumber = nullStr(procNum)
	ct.Title = nullStr(title)
	ct.ProcurementCategory = nullStr(procCat)
	ct.BidCount = nullI64(bidCount)
	ct.ProcurementMethod = nullStr(procMethod)
	ct.EUFunded = nullBool(euFunded)
	ct.LegacyType = nullStr(legacyType)

	c.JSON(http.StatusOK, ct)
}
