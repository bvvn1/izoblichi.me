package main

import (
	"database/sql"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
)

type Party struct {
	EIK             string  `json:"eik"`
	LegalName       *string `json:"legal_name"`
	DisplayName     *string `json:"display_name"`
	AddressLocality *string `json:"address_locality"`
	AddressRegion   *string `json:"address_region"`
	Country         *string `json:"country"`
	FirstSeen       *string `json:"first_seen"`
	LastSeen        *string `json:"last_seen"`
}

type PartiesResponse struct {
	Total   int64   `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
	Items   []Party `json:"items"`
}

type PartyRole struct {
	ContractCount int64 `json:"contract_count"`
}

type PartyDetail struct {
	Party
	AsBuyer    PartyRole `json:"as_buyer"`
	AsSupplier PartyRole `json:"as_supplier"`
}

func (a *App) listParties(c *gin.Context) {
	q := c.Query("q")
	role := c.Query("role")
	page, perPage := paginate(c)
	offset := (page - 1) * perPage

	qb := sb.Select(
		"eik", "legal_name", "display_name", "address_locality", "address_region",
		"country", "first_seen_date", "last_seen_date",
		"COUNT(*) OVER () AS total",
	).From("parties")

	if q != "" {
		pct := "%" + q + "%"
		qb = qb.Where(sq.Or{
			sq.Expr("legal_name ILIKE ?", pct),
			sq.Expr("display_name ILIKE ?", pct),
			sq.Expr("eik LIKE ?", pct),
		})
	}

	if role == "buyer" {
		qb = qb.Where("EXISTS (SELECT 1 FROM contracts_unified WHERE buyer_eik = parties.eik)")
	} else if role == "supplier" {
		qb = qb.Where("EXISTS (SELECT 1 FROM contracts_unified WHERE supplier_eik = parties.eik)")
	}
	qb = qb.OrderBy("display_name NULLS LAST", "legal_name NULLS LAST").
		Limit(uint64(perPage)).Offset(uint64(offset))

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

	var items []Party
	var total int64
	for rows.Next() {
		var (
			eik                       string
			legalName, displayName    sql.NullString
			locality, region, country sql.NullString
			firstSeen, lastSeen       sql.NullTime
			rowTotal                  int64
		)
		if err := rows.Scan(
			&eik, &legalName, &displayName, &locality, &region,
			&country, &firstSeen, &lastSeen, &rowTotal,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total = rowTotal
		items = append(items, Party{
			EIK:             eik,
			LegalName:       nullStr(legalName),
			DisplayName:     nullStr(displayName),
			AddressLocality: nullStr(locality),
			AddressRegion:   nullStr(region),
			Country:         nullStr(country),
			FirstSeen:       nullDate(firstSeen),
			LastSeen:        nullDate(lastSeen),
		})
	}
	if items == nil {
		items = []Party{}
	}

	c.JSON(http.StatusOK, PartiesResponse{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	})
}

func (a *App) getParty(c *gin.Context) {
	eik := c.Param("eik")

	var (
		legalName, displayName    sql.NullString
		locality, region, country sql.NullString
		firstSeen, lastSeen       sql.NullTime
	)
	err := a.db.QueryRow(
		`SELECT legal_name, display_name, address_locality, address_region, country,
		        first_seen_date, last_seen_date FROM parties WHERE eik = ?`, eik,
	).Scan(&legalName, &displayName, &locality, &region, &country, &firstSeen, &lastSeen)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "party not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var buyerCount, supplierCount int64
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM contracts_unified WHERE buyer_eik = ?`, eik).Scan(&buyerCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM contracts_unified WHERE supplier_eik = ?`, eik).Scan(&supplierCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PartyDetail{
		Party: Party{
			EIK:             eik,
			LegalName:       nullStr(legalName),
			DisplayName:     nullStr(displayName),
			AddressLocality: nullStr(locality),
			AddressRegion:   nullStr(region),
			Country:         nullStr(country),
			FirstSeen:       nullDate(firstSeen),
			LastSeen:        nullDate(lastSeen),
		},
		AsBuyer:    PartyRole{ContractCount: buyerCount},
		AsSupplier: PartyRole{ContractCount: supplierCount},
	})
}
