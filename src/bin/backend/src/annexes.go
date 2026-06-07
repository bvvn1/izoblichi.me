package main

import (
	"database/sql"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
)

type Annex struct {
	RowKey               int64    `json:"row_key"`
	Year                 int16    `json:"year"`
	DocNumber            *string  `json:"doc_number"`
	ContractNumber       *string  `json:"contract_number"`
	ContractDate         *string  `json:"contract_date"`
	PublishedDate        *string  `json:"published_date"`
	ProcurementNumber    *string  `json:"procurement_number"`
	BuyerEIK             *string  `json:"buyer_eik"`
	BuyerName            *string  `json:"buyer_name"`
	ProcurementSubject   *string  `json:"procurement_subject"`
	ProcurementCategory  *string  `json:"procurement_category"`
	EUFunded             *bool    `json:"eu_funded"`
	ContractSubject      *string  `json:"contract_subject"`
	SupplierEIK          *string  `json:"supplier_eik"`
	SupplierName         *string  `json:"supplier_name"`
	ValueBefore          *float64 `json:"value_before"`
	ValueAfter           *float64 `json:"value_after"`
	ValueChange          *float64 `json:"value_change"`
	Currency             *string  `json:"currency"`
	AmendmentDescription *string  `json:"amendment_description"`
	AmendmentReason      *string  `json:"amendment_reason"`
	Circumstances        *string  `json:"circumstances"`
}

type AnnexesResponse struct {
	Total   int64   `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
	Items   []Annex `json:"items"`
}

func (a *App) listAnnexes(c *gin.Context) {
	q := c.Query("q")
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom, _ := parseYear(c.Query("year_from"))
	yearTo, _ := parseYear(c.Query("year_to"))
	procNum := c.Query("procurement_number")

	page, perPage := paginate(c)
	offset := (page - 1) * perPage

	qb := sb.Select(
		"rowid", "source_year", "doc_number", "contract_number",
		"contract_date", "published_date", "procurement_number",
		"buyer_eik", "buyer_name", "procurement_subject", "procurement_category",
		"eu_funded", "contract_subject",
		"supplier_eik", "supplier_name",
		"value_before", "value_after", "value_change", "currency",
		"amendment_description", "amendment_reason", "circumstances",
		"COUNT(*) OVER () AS total",
	).From("legacy_annexes")

	if q != "" {
		pct := "%" + q + "%"
		qb = qb.Where(sq.Or{
			sq.Expr("procurement_subject ILIKE ?", pct),
			sq.Expr("contract_subject ILIKE ?", pct),
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
		qb = qb.Where(sq.GtOrEq{"source_year": yearFrom})
	}
	if yearTo != 0 {
		qb = qb.Where(sq.LtOrEq{"source_year": yearTo})
	}
	if procNum != "" {
		qb = qb.Where(sq.Eq{"procurement_number": procNum})
	}
	qb = qb.OrderBy("contract_date DESC NULLS LAST").Limit(uint64(perPage)).Offset(uint64(offset))

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

	var items []Annex
	var total int64
	for rows.Next() {
		var (
			rowKey               int64
			year                 int16
			docNum               sql.NullString
			contractNum          sql.NullString
			contractDate         sql.NullTime
			publishedDate        sql.NullTime
			procNumber           sql.NullString
			bEIK, bName          sql.NullString
			procSubj             sql.NullString
			procCat              sql.NullString
			euFunded             sql.NullBool
			contractSubj         sql.NullString
			sEIK, sName          sql.NullString
			vBefore, vAfter, vCh sql.NullFloat64
			currency             sql.NullString
			aDesc, aReason, circ sql.NullString
			rowTotal             int64
		)
		if err := rows.Scan(
			&rowKey, &year, &docNum, &contractNum,
			&contractDate, &publishedDate, &procNumber,
			&bEIK, &bName, &procSubj, &procCat,
			&euFunded, &contractSubj,
			&sEIK, &sName,
			&vBefore, &vAfter, &vCh, &currency,
			&aDesc, &aReason, &circ,
			&rowTotal,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total = rowTotal
		items = append(items, Annex{
			RowKey:               rowKey,
			Year:                 year,
			DocNumber:            nullStr(docNum),
			ContractNumber:       nullStr(contractNum),
			ContractDate:         nullDate(contractDate),
			PublishedDate:        nullDate(publishedDate),
			ProcurementNumber:    nullStr(procNumber),
			BuyerEIK:             nullStr(bEIK),
			BuyerName:            nullStr(bName),
			ProcurementSubject:   nullStr(procSubj),
			ProcurementCategory:  nullStr(procCat),
			EUFunded:             nullBool(euFunded),
			ContractSubject:      nullStr(contractSubj),
			SupplierEIK:          nullStr(sEIK),
			SupplierName:         nullStr(sName),
			ValueBefore:          nullF64(vBefore),
			ValueAfter:           nullF64(vAfter),
			ValueChange:          nullF64(vCh),
			Currency:             nullStr(currency),
			AmendmentDescription: nullStr(aDesc),
			AmendmentReason:      nullStr(aReason),
			Circumstances:        nullStr(circ),
		})
	}
	if items == nil {
		items = []Annex{}
	}

	c.JSON(http.StatusOK, AnnexesResponse{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	})
}
