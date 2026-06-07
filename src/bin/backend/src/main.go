package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// sb is the package-level squirrel builder using DuckDB's '?' placeholder format.
var sb = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type App struct {
	db     *sqlx.DB
	dbPath string // cleaned path without query params, for os.Stat
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "../../../../data/procurement.duckdb?access_mode=read_only"
	}

	db, err := sqlx.Open("duckdb", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)

	cleanPath := dbPath
	if idx := strings.Index(cleanPath, "?"); idx >= 0 {
		cleanPath = cleanPath[:idx]
	}
	app := &App{db: db, dbPath: cleanPath}

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		stats := app.db.Stats()

		var lastUpdated sql.NullTime
		if err := app.db.QueryRow(`SELECT MAX(source_file_date) FROM ocds_releases`).Scan(&lastUpdated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": err.Error()})
			return
		}

		var dbSizeMB float64
		if info, err := os.Stat(app.dbPath); err == nil {
			dbSizeMB = float64(info.Size()) / (1024 * 1024)
		}

		resp := gin.H{
			"status":             "ok",
			"db_size_mb":         dbSizeMB,
			"active_connections": stats.InUse,
			"last_updated":       nil,
		}
		if lastUpdated.Valid {
			resp["last_updated"] = lastUpdated.Time.Format(time.RFC3339)
		}
		c.JSON(http.StatusOK, resp)
	})

	v1 := r.Group("/api/v1")
	v1.GET("/", apiGuide)
	v1.GET("/contracts", app.listContracts)
	v1.GET("/contracts/:source/:row_key", app.getContract)
	v1.GET("/annexes", app.listAnnexes)
	v1.GET("/buyers/:eik", app.getBuyer)
	v1.GET("/suppliers/:eik", app.getSupplier)
	v1.GET("/anomalies", app.getAnomalies)
	v1.GET("/parties", app.listParties)
	v1.GET("/parties/:eik", app.getParty)
	v1.GET("/stats", app.getStats)
	v1.GET("/stats/timeseries", app.getTimeseries)
	v1.GET("/network/buyer/:eik", app.getBuyerNetwork)
	v1.GET("/network/supplier/:eik", app.getSupplierNetwork)
	v1.GET("/map/buyers", app.getMapBuyers)
	v1.GET("/search/autocomplete", app.autocomplete)
	v1.GET("/export/contracts", app.exportContracts)
	v1.GET("/export/anomalies/:type", app.exportAnomalies)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func apiGuide(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "Изобличи.ме API",
		"description": "Public procurement transparency API for Bulgaria. Exposes contracts, buyers, suppliers, anomalies, and statistics from public procurement data (2020-2026).",
		"version":     "v1",
		"base_url":    "https://api.izoblichi.me/api/v1",
		"endpoints": []gin.H{
			{
				"path":   "/",
				"method": "GET",
				"desc":   "This guide",
			},
			{
				"path":   "/stats",
				"method": "GET",
				"desc":   "Aggregate statistics — total contracts, total value, unique buyers/suppliers, yearly breakdown",
			},
			{
				"path":   "/contracts",
				"method": "GET",
				"desc":   "List contracts with pagination and filtering",
				"params": gin.H{
					"page":         "int (default 1)",
					"per_page":     "int (default 20, max 100)",
					"q":            "search on title, buyer name, supplier name",
					"buyer_eik":    "filter by buyer EIK",
					"supplier_eik": "filter by supplier EIK",
					"year_from":    "filter by year (inclusive)",
					"year_to":      "filter by year (inclusive)",
					"min_value":    "filter by minimum contract value",
					"max_value":    "filter by maximum contract value",
					"category":     "filter by procurement_category",
					"source":       "filter by data_source ('legacy' or 'ocds')",
					"sort_by":      "contract_date (default) | contract_value | year | buyer_name | supplier_name",
					"sort_dir":     "desc (default) | asc",
				},
			},
			{
				"path":   "/contracts/:source/:row_key",
				"method": "GET",
				"desc":   "Single contract detail by source ('legacy' or 'ocds') and row_key",
			},
			{
				"path":   "/annexes",
				"method": "GET",
				"desc":   "Contract amendments — value changes, reasons, and circumstances (2020-2023 legacy data)",
				"params": gin.H{
					"page":               "int (default 1)",
					"per_page":           "int (default 20, max 100)",
					"q":                  "search on subject, buyer name, supplier name",
					"buyer_eik":          "filter by buyer EIK",
					"supplier_eik":       "filter by supplier EIK",
					"year_from":          "filter by year",
					"year_to":            "filter by year",
					"procurement_number": "filter by procurement number",
				},
			},
			{
				"path":   "/buyers/:eik",
				"method": "GET",
				"desc":   "Buyer profile — top suppliers, year breakdown, flags (dominance, no-bid, near-threshold)",
			},
			{
				"path":   "/suppliers/:eik",
				"method": "GET",
				"desc":   "Supplier profile — top buyers, year breakdown, concentration flags",
			},
			{
				"path":   "/parties",
				"method": "GET",
				"desc":   "List all parties (buyers and suppliers) with search and pagination",
				"params": gin.H{"page": "int (default 1)", "per_page": "int (default 20, max 100)", "q": "search by name or EIK"},
			},
			{
				"path":   "/parties/:eik",
				"method": "GET",
				"desc":   "Single party detail by EIK — includes contract counts as buyer and supplier",
			},
			{
				"path":   "/anomalies",
				"method": "GET",
				"desc":   "Detected anomalies — paginated sections for near_threshold, no_bid, dominance, repeated_award",
				"params": gin.H{
					"type":         "repeated: near_threshold | no_bid | dominance | repeated_award (default: all)",
					"buyer_eik":    "filter by buyer EIK",
					"supplier_eik": "filter by supplier EIK",
					"year_from":    "filter by year",
					"year_to":      "filter by year",
					"page":         "int (default 1) — applies to all sections equally",
					"per_page":     "int (default 20, max 100)",
				},
				"thresholds": gin.H{
					"goods_services_bgn": ThresholdGoods,
					"works_bgn":          ThresholdWorks,
					"near_threshold_pct": NearThresholdMargin * 100,
					"dominance_pct":      DominanceThreshold * 100,
				},
			},
			{
				"path":   "/stats/timeseries",
				"method": "GET",
				"desc":   "Monthly breakdown of contract volume, BGN value, no-bid count, and competitive count — ready for time-series charts",
			},
			{
				"path":   "/network/buyer/:eik",
				"method": "GET",
				"desc":   "Radial network graph around a buyer — nodes (buyer + up to 50 suppliers) and weighted edges, ready for D3/Cytoscape",
			},
			{
				"path":   "/network/supplier/:eik",
				"method": "GET",
				"desc":   "Radial network graph around a supplier — nodes (supplier + up to 50 buyers) and weighted edges",
			},
			{
				"path":   "/map/buyers",
				"method": "GET",
				"desc":   "Aggregated buyer data for map visualisation — includes address_locality, address_region, total_value_bgn, and a composite risk_score (0-100)",
			},
			{
				"path":   "/search/autocomplete",
				"method": "GET",
				"desc":   "Instant autocomplete — returns up to 5 buyers, 5 suppliers, and 5 contracts matching the query",
				"params": gin.H{"q": "search string (min 2 chars)"},
			},
			{
				"path":   "/export/contracts",
				"method": "GET",
				"desc":   "Download filtered contracts as CSV (max 10 000 rows). Accepts the same filters as /contracts",
			},
			{
				"path":   "/export/anomalies/:type",
				"method": "GET",
				"desc":   "Download an anomaly list as CSV. :type is one of near_threshold | no_bid | dominance | repeated_award. Accepts buyer_eik, supplier_eik, year_from, year_to filters",
			},
		},
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func paginate(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return
}

func nullStr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullF64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func nullI64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullBool(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

func nullDate(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	s := nt.Time.Format("2006-01-02")
	return &s
}

// parseYear parses a query-param year string to int64. Returns 0 and false when empty/unparseable.
func parseYear(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// buildFilters returns squirrel conditions for the standard buyer/supplier/year filters.
func buildFilters(buyerEIK, supplierEIK string, yearFrom, yearTo int64) sq.And {
	var conds sq.And
	if buyerEIK != "" {
		conds = append(conds, sq.Eq{"buyer_eik": buyerEIK})
	}
	if supplierEIK != "" {
		conds = append(conds, sq.Eq{"supplier_eik": supplierEIK})
	}
	if yearFrom != 0 {
		conds = append(conds, sq.GtOrEq{"year": yearFrom})
	}
	if yearTo != 0 {
		conds = append(conds, sq.LtOrEq{"year": yearTo})
	}
	return conds
}

// whereClause converts an sq.And into a "WHERE ..." string + args for template injection.
// Returns ("", nil) when there are no conditions.
func whereClause(conds sq.And) (string, []any) {
	if len(conds) == 0 {
		return "", nil
	}
	wSQL, wArgs, _ := conds.ToSql()
	return "WHERE " + wSQL, wArgs
}
