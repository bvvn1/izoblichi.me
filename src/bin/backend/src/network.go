package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NetworkNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	Type  string  `json:"type"`
	Size  float64 `json:"size"`
}

type NetworkEdge struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Weight float64 `json:"weight"`
	Label  string  `json:"label"`
}

type NetworkResponse struct {
	Nodes []NetworkNode `json:"nodes"`
	Edges []NetworkEdge `json:"edges"`
}

func contractLabel(n int64) string {
	if n == 1 {
		return "1 договор"
	}
	return fmt.Sprintf("%d договора", n)
}

func resolvePartyName(fromParties, fromContracts sql.NullString, fallback string) string {
	if fromParties.Valid && fromParties.String != "" {
		return fromParties.String
	}
	if fromContracts.Valid && fromContracts.String != "" {
		return fromContracts.String
	}
	return fallback
}

func (a *App) getBuyerNetwork(c *gin.Context) {
	eik := c.Param("eik")

	var nameFromParties sql.NullString
	if err := a.db.QueryRow(`SELECT COALESCE(display_name, legal_name) FROM parties WHERE eik = ?`, eik).Scan(&nameFromParties); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalContracts int64
	var nameFromContracts sql.NullString
	if err := a.db.QueryRow(`SELECT COUNT(*), MAX(buyer_name) FROM contracts_unified WHERE buyer_eik = ?`, eik).
		Scan(&totalContracts, &nameFromContracts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if totalContracts == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "buyer not found"})
		return
	}

	rows, err := a.db.Query(`
		SELECT
			supplier_eik,
			MAX(supplier_name) AS supplier_name,
			COUNT(*) AS contract_count,
			SUM(contract_value) AS total_value
		FROM contracts_unified
		WHERE buyer_eik = ? AND supplier_eik IS NOT NULL
		GROUP BY supplier_eik
		ORDER BY contract_count DESC
		LIMIT 50`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	nodes := []NetworkNode{{
		ID:    eik,
		Label: resolvePartyName(nameFromParties, nameFromContracts, eik),
		Type:  "buyer",
		Size:  float64(totalContracts),
	}}
	edges := []NetworkEdge{}

	for rows.Next() {
		var sEIK, sName sql.NullString
		var count int64
		var val sql.NullFloat64
		rows.Scan(&sEIK, &sName, &count, &val)
		if !sEIK.Valid {
			continue
		}
		label := sEIK.String
		if sName.Valid && sName.String != "" {
			label = sName.String
		}
		weight := 0.0
		if val.Valid {
			weight = val.Float64
		}
		nodes = append(nodes, NetworkNode{ID: sEIK.String, Label: label, Type: "supplier", Size: weight})
		edges = append(edges, NetworkEdge{
			Source: eik,
			Target: sEIK.String,
			Weight: weight,
			Label:  contractLabel(count),
		})
	}

	c.JSON(http.StatusOK, NetworkResponse{Nodes: nodes, Edges: edges})
}

func (a *App) getSupplierNetwork(c *gin.Context) {
	eik := c.Param("eik")

	var nameFromParties sql.NullString
	if err := a.db.QueryRow(`SELECT COALESCE(display_name, legal_name) FROM parties WHERE eik = ?`, eik).Scan(&nameFromParties); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}

	var totalWins int64
	var nameFromContracts sql.NullString
	if err := a.db.QueryRow(`SELECT COUNT(*), MAX(supplier_name) FROM contracts_unified WHERE supplier_eik = ?`, eik).
		Scan(&totalWins, &nameFromContracts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if totalWins == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "supplier not found"})
		return
	}

	rows, err := a.db.Query(`
		SELECT
			buyer_eik,
			MAX(buyer_name) AS buyer_name,
			COUNT(*) AS contract_count,
			SUM(contract_value) AS total_value
		FROM contracts_unified
		WHERE supplier_eik = ? AND buyer_eik IS NOT NULL
		GROUP BY buyer_eik
		ORDER BY contract_count DESC
		LIMIT 50`, eik)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	nodes := []NetworkNode{{
		ID:    eik,
		Label: resolvePartyName(nameFromParties, nameFromContracts, eik),
		Type:  "supplier",
		Size:  float64(totalWins),
	}}
	edges := []NetworkEdge{}

	for rows.Next() {
		var bEIK, bName sql.NullString
		var count int64
		var val sql.NullFloat64
		if err := rows.Scan(&bEIK, &bName, &count, &val); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !bEIK.Valid {
			continue
		}
		label := bEIK.String
		if bName.Valid && bName.String != "" {
			label = bName.String
		}
		weight := 0.0
		if val.Valid {
			weight = val.Float64
		}
		nodes = append(nodes, NetworkNode{ID: bEIK.String, Label: label, Type: "buyer", Size: weight})
		edges = append(edges, NetworkEdge{
			Source: bEIK.String,
			Target: eik,
			Weight: weight,
			Label:  contractLabel(count),
		})
	}

	c.JSON(http.StatusOK, NetworkResponse{Nodes: nodes, Edges: edges})
}
