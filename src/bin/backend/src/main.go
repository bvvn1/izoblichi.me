package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/duckdb/duckdb-go/v2"
)

type App struct {
	db *sql.DB
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "../../../../data/procurement.duckdb?access_mode=read_only"
	}

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)

	app := &App{db: db}

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	v1.GET("/", apiGuide)
	v1.GET("/contracts", app.listContracts)
	v1.GET("/buyers/:eik", app.getBuyer)
	v1.GET("/suppliers/:eik", app.getSupplier)
	v1.GET("/anomalies", app.getAnomalies)
	v1.GET("/parties", app.listParties)
	v1.GET("/parties/:eik", app.getParty)
	v1.GET("/stats", app.getStats)

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
				"path":    "/",
				"method":  "GET",
				"desc":    "This guide",
			},
			{
				"path":    "/stats",
				"method":  "GET",
				"desc":    "Aggregate statistics — total contracts, total value, unique buyers/suppliers, yearly breakdown",
			},
			{
				"path":    "/contracts",
				"method":  "GET",
				"desc":    "List contracts with pagination and filtering",
				"params":  gin.H{"page": "int (default 1)", "per_page": "int (default 20, max 100)", "buyer_eik": "filter by buyer EIK", "supplier_eik": "filter by supplier EIK", "year": "filter by year (e.g. 2024)", "q": "full-text search on contract title"},
			},
			{
				"path":    "/buyers/:eik",
				"method":  "GET",
				"desc":    "Buyer profile — all contracts where this entity is the buyer",
			},
			{
				"path":    "/suppliers/:eik",
				"method":  "GET",
				"desc":    "Supplier profile — all contracts where this entity is the supplier",
			},
			{
				"path":    "/parties",
				"method":  "GET",
				"desc":    "List all parties (buyers and suppliers) with search and pagination",
				"params":  gin.H{"page": "int (default 1)", "per_page": "int (default 20, max 100)", "q": "search by name or EIK"},
			},
			{
				"path":    "/parties/:eik",
				"method":  "GET",
				"desc":    "Get a single party by EIK",
			},
			{
				"path":    "/anomalies",
				"method":  "GET",
				"desc":    "Detected anomalies — high concentration, threshold proximity, single-bidder contracts, repeated award patterns",
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
