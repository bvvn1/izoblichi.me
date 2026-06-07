package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
)

func fmtF64(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *v)
}

func fmtI64(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func fmtBool(v *bool) string {
	if v == nil {
		return ""
	}
	if *v {
		return "true"
	}
	return "false"
}

func fmtStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (a *App) exportContracts(c *gin.Context) {
	q := c.Query("q")
	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom, _ := parseYear(c.Query("year_from"))
	yearTo, _ := parseYear(c.Query("year_to"))
	minValue := c.Query("min_value")
	maxValue := c.Query("max_value")
	category := c.Query("category")
	source := c.Query("source")

	qb := sb.Select(
		"data_source", "row_key", "year", "contract_date",
		"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
		"contract_value", "currency", "procurement_number", "title",
		"procurement_category", "bid_count", "procurement_method", "eu_funded",
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
	qb = qb.OrderBy("contract_date DESC NULLS LAST").Limit(10000)

	query, args, err := qb.ToSql()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := a.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="contracts.csv"`)
	c.Header("X-Content-Type-Options", "nosniff")

	w := csv.NewWriter(c.Writer)
	w.Write([]string{
		"data_source", "row_key", "year", "contract_date",
		"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
		"contract_value", "currency", "procurement_number", "title",
		"procurement_category", "bid_count", "procurement_method", "eu_funded",
	})

	for rows.Next() {
		var (
			dataSource, rowKey string
			year               sql.NullInt64
			contractDate       sql.NullTime
			bEIK, bName        sql.NullString
			sEIK, sName        sql.NullString
			contractValue      sql.NullFloat64
			currency           sql.NullString
			procNum, title     sql.NullString
			procCat            sql.NullString
			bidCount           sql.NullInt64
			procMethod         sql.NullString
			euFunded           sql.NullBool
		)
		if err := rows.Scan(
			&dataSource, &rowKey, &year, &contractDate,
			&bEIK, &bName, &sEIK, &sName,
			&contractValue, &currency, &procNum, &title,
			&procCat, &bidCount, &procMethod, &euFunded,
		); err != nil {
			continue
		}
		dateStr := ""
		if contractDate.Valid {
			dateStr = contractDate.Time.Format("2006-01-02")
		}
		yearStr := ""
		if year.Valid {
			yearStr = fmt.Sprintf("%d", year.Int64)
		}
		w.Write([]string{
			dataSource, rowKey, yearStr, dateStr,
			fmtStr(nullStr(bEIK)), fmtStr(nullStr(bName)),
			fmtStr(nullStr(sEIK)), fmtStr(nullStr(sName)),
			fmtF64(nullF64(contractValue)),
			fmtStr(nullStr(currency)),
			fmtStr(nullStr(procNum)),
			fmtStr(nullStr(title)),
			fmtStr(nullStr(procCat)),
			fmtI64(nullI64(bidCount)),
			fmtStr(nullStr(procMethod)),
			fmtBool(nullBool(euFunded)),
		})
	}
	w.Flush()
}

var validAnomalyTypes = map[string]bool{
	"near_threshold": true,
	"no_bid":         true,
	"dominance":      true,
	"repeated_award": true,
}

func (a *App) exportAnomalies(c *gin.Context) {
	anomalyType := c.Param("type")
	if !validAnomalyTypes[anomalyType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be one of: near_threshold, no_bid, dominance, repeated_award"})
		return
	}

	buyerEIK := c.Query("buyer_eik")
	supplierEIK := c.Query("supplier_eik")
	yearFrom, _ := parseYear(c.Query("year_from"))
	yearTo, _ := parseYear(c.Query("year_to"))

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="anomalies_%s.csv"`, anomalyType))
	c.Header("X-Content-Type-Options", "nosniff")
	w := csv.NewWriter(c.Writer)

	switch anomalyType {
	case "near_threshold":
		w.Write([]string{
			"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"contract_date", "title", "procurement_category",
			"contract_value", "threshold", "gap", "gap_pct",
		})
		qb := sb.Select(
			"contract_value", "buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"contract_date", "title", "procurement_category",
		).From("contracts_unified").
			Where("currency = 'BGN'").
			Where(sq.Or{
				sq.And{
					sq.Expr("procurement_category NOT IN ('Строителство', 'works')"),
					sq.Expr("contract_value BETWEEN ? AND ?", ThresholdGoods*(1-NearThresholdMargin), ThresholdGoods),
				},
				sq.And{
					sq.Expr("procurement_category IN ('Строителство', 'works')"),
					sq.Expr("contract_value BETWEEN ? AND ?", ThresholdWorks*(1-NearThresholdMargin), ThresholdWorks),
				},
			})
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
		qb = qb.OrderBy("contract_value DESC").Limit(10000)
		ntSQL, ntArgs, _ := qb.ToSql()
		rows, err := a.db.Query(ntSQL, ntArgs...)
		if err != nil {
			return
		}
		for rows.Next() {
			var (
				val                      float64
				bEIK, bName, sEIK, sName sql.NullString
				cDate                    sql.NullTime
				title, cat               sql.NullString
			)
			if err := rows.Scan(&val, &bEIK, &bName, &sEIK, &sName, &cDate, &title, &cat); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			threshold := ThresholdGoods
			if cat.Valid && (cat.String == "Строителство" || cat.String == "works") {
				threshold = ThresholdWorks
			}
			gap := threshold - val
			dateStr := ""
			if cDate.Valid {
				dateStr = cDate.Time.Format("2006-01-02")
			}
			w.Write([]string{
				fmtStr(nullStr(bEIK)), fmtStr(nullStr(bName)),
				fmtStr(nullStr(sEIK)), fmtStr(nullStr(sName)),
				dateStr, fmtStr(nullStr(title)), fmtStr(nullStr(cat)),
				fmt.Sprintf("%.2f", val), fmt.Sprintf("%.2f", threshold),
				fmt.Sprintf("%.2f", gap), fmt.Sprintf("%.4f", gap/threshold*100),
			})
		}
		rows.Close()

	case "no_bid":
		w.Write([]string{
			"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"contract_date", "title", "bid_count", "contract_value", "currency",
		})
		qb := sb.Select(
			"bid_count", "contract_value", "currency",
			"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"contract_date", "title",
		).From("contracts_unified").Where("bid_count <= 1")
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
		qb = qb.OrderBy("contract_value DESC NULLS LAST").Limit(10000)
		nbSQL, nbArgs, _ := qb.ToSql()
		rows, err := a.db.Query(nbSQL, nbArgs...)
		if err != nil {
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
			if err := rows.Scan(&bidCount, &val, &curr, &bEIK, &bName, &sEIK, &sName, &cDate, &title); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			dateStr := ""
			if cDate.Valid {
				dateStr = cDate.Time.Format("2006-01-02")
			}
			w.Write([]string{
				fmtStr(nullStr(bEIK)), fmtStr(nullStr(bName)),
				fmtStr(nullStr(sEIK)), fmtStr(nullStr(sName)),
				dateStr, fmtStr(nullStr(title)),
				fmtI64(nullI64(bidCount)), fmtF64(nullF64(val)), fmtStr(nullStr(curr)),
			})
		}
		rows.Close()

	case "dominance":
		w.Write([]string{
			"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"wins", "total_value", "pct_by_count", "total_buyer_contracts",
		})
		domWhere, domArgs := whereClause(buildFilters(buyerEIK, supplierEIK, yearFrom, yearTo))
		q := `
			WITH agg AS (
				SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
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
			       wins, total_val, pct, total_buyer
			FROM filtered
			ORDER BY pct DESC
			LIMIT 10000`
		args := append(domArgs, DominanceThreshold*100)
		rows, err := a.db.Query(q, args...)
		if err != nil {
			return
		}
		for rows.Next() {
			var (
				bEIK, bName, sEIK, sName sql.NullString
				wins, totalBuyer         int64
				val                      sql.NullFloat64
				pct                      float64
			)
			if err := rows.Scan(&bEIK, &bName, &sEIK, &sName, &wins, &val, &pct, &totalBuyer); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			w.Write([]string{
				fmtStr(nullStr(bEIK)), fmtStr(nullStr(bName)),
				fmtStr(nullStr(sEIK)), fmtStr(nullStr(sName)),
				fmt.Sprintf("%d", wins), fmtF64(nullF64(val)),
				fmt.Sprintf("%.2f", pct), fmt.Sprintf("%d", totalBuyer),
			})
		}
		rows.Close()

	case "repeated_award":
		w.Write([]string{
			"buyer_eik", "buyer_name", "supplier_eik", "supplier_name",
			"total_wins", "years_active", "total_value", "years",
		})
		raWhere, raArgs := whereClause(buildFilters(buyerEIK, supplierEIK, yearFrom, yearTo))
		q := `
			WITH pair_years AS (
				SELECT buyer_eik, buyer_name, supplier_eik, supplier_name, year,
				       COUNT(*) AS year_wins,
				       SUM(CASE WHEN bid_count <= 1 THEN 1 ELSE 0 END) AS year_no_bid,
				       SUM(contract_value) AS year_value
				FROM contracts_unified
				` + raWhere + `
				GROUP BY buyer_eik, buyer_name, supplier_eik, supplier_name, year
			),
			pair_agg AS (
				SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
				       SUM(year_wins) AS total_wins,
				       COUNT(DISTINCT year) AS years_active,
				       SUM(year_no_bid) AS total_no_bid,
				       SUM(year_value) AS total_value,
				       CAST(to_json(list(year ORDER BY year)) AS VARCHAR) AS years_list
				FROM pair_years
				GROUP BY buyer_eik, buyer_name, supplier_eik, supplier_name
			),
			filtered AS (
				SELECT * FROM pair_agg
				WHERE years_active >= ? AND total_wins >= ?
				  AND total_no_bid * 100.0 / total_wins >= 50
			)
			SELECT buyer_eik, buyer_name, supplier_eik, supplier_name,
			       total_wins, years_active, total_value, years_list
			FROM filtered
			ORDER BY years_active DESC, total_wins DESC
			LIMIT 10000`
		args := append(raArgs, RepeatedAwardMinYears, RepeatedAwardMinWins)
		rows, err := a.db.Query(q, args...)
		if err != nil {
			return
		}
		for rows.Next() {
			var (
				bEIK, bName sql.NullString
				sEIK, sName sql.NullString
				totalWins   int64
				yearsActive int64
				totalValue  sql.NullFloat64
				yearsJSON   sql.NullString
			)
			if err := rows.Scan(&bEIK, &bName, &sEIK, &sName, &totalWins, &yearsActive, &totalValue, &yearsJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			yearsStr := ""
			if yearsJSON.Valid {
				yearsStr = yearsJSON.String
			}
			w.Write([]string{
				fmtStr(nullStr(bEIK)), fmtStr(nullStr(bName)),
				fmtStr(nullStr(sEIK)), fmtStr(nullStr(sName)),
				fmt.Sprintf("%d", totalWins), fmt.Sprintf("%d", yearsActive),
				fmtF64(nullF64(totalValue)), yearsStr,
			})
		}
		rows.Close()
	}

	w.Flush()
}
