package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AutocompleteParty struct {
	EIK  string `json:"eik"`
	Name string `json:"name"`
}

type AutocompleteContract struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Title  string `json:"title"`
}

type AutocompleteResponse struct {
	Buyers    []AutocompleteParty    `json:"buyers"`
	Suppliers []AutocompleteParty    `json:"suppliers"`
	Contracts []AutocompleteContract `json:"contracts"`
}

func (a *App) autocomplete(c *gin.Context) {
	q := c.Query("q")
	resp := AutocompleteResponse{
		Buyers:    []AutocompleteParty{},
		Suppliers: []AutocompleteParty{},
		Contracts: []AutocompleteContract{},
	}
	if len([]rune(q)) < 2 {
		c.JSON(http.StatusOK, resp)
		return
	}
	pct := "%" + q + "%"

	buyerRows, err := a.db.Query(`
		SELECT buyer_eik, MAX(buyer_name) AS name
		FROM contracts_unified
		WHERE buyer_eik IS NOT NULL AND buyer_name ILIKE ?
		GROUP BY buyer_eik
		ORDER BY COUNT(*) DESC
		LIMIT 5`, pct)
	if err == nil {
		for buyerRows.Next() {
			var eik, name sql.NullString
			buyerRows.Scan(&eik, &name)
			if eik.Valid && name.Valid {
				resp.Buyers = append(resp.Buyers, AutocompleteParty{EIK: eik.String, Name: name.String})
			}
		}
		buyerRows.Close()
	}

	supplierRows, err := a.db.Query(`
		SELECT supplier_eik, MAX(supplier_name) AS name
		FROM contracts_unified
		WHERE supplier_eik IS NOT NULL AND supplier_name ILIKE ?
		GROUP BY supplier_eik
		ORDER BY COUNT(*) DESC
		LIMIT 5`, pct)
	if err == nil {
		for supplierRows.Next() {
			var eik, name sql.NullString
			supplierRows.Scan(&eik, &name)
			if eik.Valid && name.Valid {
				resp.Suppliers = append(resp.Suppliers, AutocompleteParty{EIK: eik.String, Name: name.String})
			}
		}
		supplierRows.Close()
	}

	contractRows, err := a.db.Query(`
		SELECT row_key, data_source, title
		FROM contracts_unified
		WHERE title IS NOT NULL AND title ILIKE ?
		ORDER BY contract_date DESC NULLS LAST
		LIMIT 5`, pct)
	if err == nil {
		for contractRows.Next() {
			var rowKey, source string
			var title sql.NullString
			contractRows.Scan(&rowKey, &source, &title)
			if title.Valid {
				resp.Contracts = append(resp.Contracts, AutocompleteContract{
					ID:     rowKey,
					Source: source,
					Title:  title.String,
				})
			}
		}
		contractRows.Close()
	}

	c.JSON(http.StatusOK, resp)
}
