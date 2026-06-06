"""
ETL script: loads raw_data/ JSON files into data/procurement.duckdb

Two source formats:
  - Legacy CSV-JSON (2020-2023): JSON arrays, row[0]=headers, row[1..n]=data
  - OCDS 1.1 (2026): proper OCDS JSON with releases[] arrays

Run with: python3 -u scripts/load_data.py
"""

import csv
import json
import os
import re
import tempfile
from datetime import datetime
from pathlib import Path

import duckdb

RAW_DATA = Path("raw_data")
DB_PATH = Path("data/procurement.duckdb")

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def strip_quotes(val: str | None) -> str:
    """Strip spurious ASCII double-quote wrapping: '"COMPANY"' → 'COMPANY'.
    Does NOT touch Bulgarian „ " quote characters."""
    val = (val or "").strip()
    while val.startswith('"') and val.endswith('"') and len(val) > 1:
        val = val[1:-1]
    return val.strip()


def parse_date(s: str | None) -> str | None:
    s = strip_quotes(s or "").strip()
    if not s:
        return None
    for fmt in ("%d/%m/%Y", "%Y-%m-%d"):
        try:
            return datetime.strptime(s, fmt).strftime("%Y-%m-%d")
        except ValueError:
            pass
    return None


def parse_bool(s: str | None) -> bool | None:
    s = strip_quotes(s or "").strip().lower()
    if s in ("true", "1", "yes"):
        return True
    if s in ("false", "0", "no"):
        return False
    return None


def parse_float(s: str | None) -> float | None:
    s = strip_quotes(s or "").strip()
    if not s:
        return None
    try:
        return float(s)
    except ValueError:
        return None


def parse_int(s: str | None) -> int | None:
    v = parse_float(s)
    return int(v) if v is not None else None


# ---------------------------------------------------------------------------
# Bulk insert via temp CSV + COPY FROM (100x faster than executemany)
# ---------------------------------------------------------------------------

def _to_csv_val(v) -> str:
    """Convert a Python value to a CSV cell string. Empty string = NULL."""
    if v is None:
        return ""
    if isinstance(v, bool):
        return "true" if v else "false"
    return str(v)


def bulk_copy(conn: duckdb.DuckDBPyConnection, table: str, rows: list[tuple]) -> None:
    """Write rows to a temp CSV file and COPY INTO table. Fastest bulk load."""
    if not rows:
        return
    with tempfile.NamedTemporaryFile(
        suffix=".csv", mode="w", encoding="utf-8", newline="", delete=False
    ) as f:
        tmpfile = f.name
        writer = csv.writer(f, quoting=csv.QUOTE_MINIMAL)
        for row in rows:
            writer.writerow([_to_csv_val(v) for v in row])

    try:
        conn.execute(
            f"COPY {table} FROM '{tmpfile}' (FORMAT CSV, NULLSTR '', HEADER FALSE)"
        )
    finally:
        os.unlink(tmpfile)


# ---------------------------------------------------------------------------
# Schema DDL
# ---------------------------------------------------------------------------

DDL = """
CREATE TABLE IF NOT EXISTS legacy_contracts (
    source_type          VARCHAR NOT NULL,
    source_year          SMALLINT NOT NULL,
    doc_number           VARCHAR,
    contract_number      VARCHAR,
    contract_date        DATE,
    published_date       DATE,
    procurement_number   VARCHAR,
    buyer_eik            VARCHAR,
    buyer_name           VARCHAR,
    procurement_subject  VARCHAR,
    procurement_category VARCHAR,
    eu_funded            BOOLEAN,
    bid_count            SMALLINT,
    contract_subject     VARCHAR,
    supplier_eik         VARCHAR,
    supplier_name        VARCHAR,
    value_at_signing     DOUBLE,
    currency             VARCHAR(3),
    vat_amount           DOUBLE,
    is_sme               BOOLEAN
);

CREATE TABLE IF NOT EXISTS legacy_annexes (
    source_year           SMALLINT NOT NULL,
    doc_number            VARCHAR,
    contract_number       VARCHAR,
    contract_date         DATE,
    published_date        DATE,
    procurement_number    VARCHAR,
    buyer_eik             VARCHAR,
    buyer_name            VARCHAR,
    procurement_subject   VARCHAR,
    procurement_category  VARCHAR,
    eu_funded             BOOLEAN,
    contract_subject      VARCHAR,
    supplier_eik          VARCHAR,
    supplier_name         VARCHAR,
    value_before          DOUBLE,
    value_after           DOUBLE,
    value_change          DOUBLE,
    currency              VARCHAR(3),
    amendment_description VARCHAR,
    amendment_reason      VARCHAR,
    circumstances         VARCHAR,
    is_sme                BOOLEAN
);

CREATE TABLE IF NOT EXISTS ocds_releases (
    release_id             VARCHAR PRIMARY KEY,
    ocid                   VARCHAR NOT NULL,
    tender_aop_id          VARCHAR,
    release_date           TIMESTAMPTZ,
    tags                   VARCHAR,
    source_file_date       DATE NOT NULL,
    tender_title           VARCHAR,
    tender_status          VARCHAR,
    procurement_method     VARCHAR,
    procurement_category   VARCHAR,
    legal_basis_id         VARCHAR,
    tender_value_amount    DOUBLE,
    tender_value_currency  VARCHAR(3),
    lot_count              SMALLINT,
    bid_count              SMALLINT,
    electronic_bid_count   SMALLINT,
    buyer_org_id           VARCHAR,
    buyer_name             VARCHAR,
    buyer_eik              VARCHAR
);

CREATE INDEX IF NOT EXISTS idx_ocds_releases_ocid      ON ocds_releases(ocid);
CREATE INDEX IF NOT EXISTS idx_ocds_releases_buyer_eik ON ocds_releases(buyer_eik);
CREATE INDEX IF NOT EXISTS idx_ocds_releases_date      ON ocds_releases(release_date);

CREATE TABLE IF NOT EXISTS ocds_award_contracts (
    release_id              VARCHAR NOT NULL,
    ocid                    VARCHAR NOT NULL,
    contract_id             VARCHAR,
    contract_value_amount   DOUBLE,
    contract_value_currency VARCHAR(3),
    contract_date_signed    DATE,
    contract_period_start   DATE,
    contract_period_end     DATE,
    has_amendments          BOOLEAN,
    award_id                VARCHAR,
    award_status            VARCHAR,
    related_lot_id          VARCHAR,
    supplier_org_id         VARCHAR,
    supplier_name           VARCHAR,
    supplier_eik            VARCHAR
);

CREATE INDEX IF NOT EXISTS idx_oac_ocid         ON ocds_award_contracts(ocid);
CREATE INDEX IF NOT EXISTS idx_oac_supplier_eik ON ocds_award_contracts(supplier_eik);

CREATE TABLE IF NOT EXISTS parties (
    eik              VARCHAR PRIMARY KEY,
    legal_name       VARCHAR,
    display_name     VARCHAR,
    address_locality VARCHAR,
    address_region   VARCHAR,
    country          VARCHAR(3),
    first_seen_date  DATE,
    last_seen_date   DATE
);
"""

VIEW_DDL = """
CREATE OR REPLACE VIEW contracts_unified AS

SELECT
    'legacy'                        AS data_source,
    CAST(rowid AS VARCHAR)          AS row_key,
    NULL::VARCHAR                   AS ocid,
    source_year                     AS year,
    contract_date,
    buyer_eik,
    buyer_name,
    supplier_eik,
    supplier_name,
    value_at_signing                AS contract_value,
    currency,
    procurement_number,
    procurement_subject             AS title,
    procurement_category,
    bid_count,
    NULL::VARCHAR                   AS procurement_method,
    eu_funded,
    source_type                     AS legacy_type
FROM legacy_contracts

UNION ALL

SELECT
    'ocds'                                                              AS data_source,
    CAST(oac.rowid AS VARCHAR)                                          AS row_key,
    oac.ocid,
    CAST(EXTRACT(YEAR FROM r.release_date) AS SMALLINT)                AS year,
    oac.contract_date_signed                                            AS contract_date,
    r.buyer_eik,
    r.buyer_name,
    oac.supplier_eik,
    oac.supplier_name,
    COALESCE(oac.contract_value_amount, r.tender_value_amount)         AS contract_value,
    COALESCE(oac.contract_value_currency, r.tender_value_currency)     AS currency,
    r.tender_aop_id                                                     AS procurement_number,
    r.tender_title                                                      AS title,
    r.procurement_category,
    r.bid_count,
    r.procurement_method,
    NULL::BOOLEAN                                                       AS eu_funded,
    NULL::VARCHAR                                                       AS legacy_type
FROM ocds_award_contracts oac
JOIN ocds_releases r ON oac.release_id = r.release_id
WHERE oac.award_status = 'active' OR oac.award_status IS NULL;
"""

# ---------------------------------------------------------------------------
# Column name aliases (Bulgarian variants across years → canonical field name)
# ---------------------------------------------------------------------------

CONTRACT_COL_MAP = {
    "Номер на документ": "doc_number",
    "Номер на договор": "contract_number",
    "Дата на договор": "contract_date",
    "Публикуван на": "published_date",
    "Уникален номер на поръчката": "procurement_number",  # contracts
    "Номер на поръчката": "procurement_number",            # excl
    "ЕИК на възложителя": "buyer_eik",
    "Възложител": "buyer_name",
    "Предмет на поръчката": "procurement_subject",
    "Обект на поръчката": "procurement_category",
    "EU финансиране": "eu_funded",
    "Брой оферти": "bid_count",
    "Предмет на договора": "contract_subject",
    "ЕИК на изпълнителя": "supplier_eik",
    "Изпълнител": "supplier_name",
    "Стойност при сключване": "value_at_signing",
    "Валута": "currency",
    "ДДС": "vat_amount",
    "Малко или средно предприятие (МСП)": "is_sme",
}

ANNEX_COL_MAP = {
    "Номер на документ": "doc_number",
    "Номер на договор": "contract_number",
    "Дата на договор": "contract_date",
    "Дата на договора": "contract_date",   # 2020 variant
    "Публикуван на": "published_date",
    "Уникален номер на поръчката": "procurement_number",
    "ЕИК на възложителя": "buyer_eik",
    "Възложител": "buyer_name",
    "Ime na възложителя": "buyer_name",    # 2020 variant
    "Предмет на поръчката": "procurement_subject",
    "Обект на поръчката": "procurement_category",
    "EU финансиране": "eu_funded",
    "Предмет на договора": "contract_subject",
    "ЕИК на изпълнителя": "supplier_eik",
    "Изпълнител": "supplier_name",
    "Ime na изпълнител": "supplier_name",  # 2020 variant
    "Стойност преди изменението": "value_before",
    "Стойност след изменението": "value_after",
    "Изменение на стойността": "value_change",
    "Валута": "currency",
    "Описание на измененията": "amendment_description",
    "Причини за изменение": "amendment_reason",
    "Обстоятелства": "circumstances",
    "Малко или средно предприятие (МСП)": "is_sme",
}


def build_col_index(headers: list[str], col_map: dict) -> dict[str, int]:
    idx = {}
    for i, h in enumerate(headers):
        field = col_map.get(h)
        if field and field not in idx:
            idx[field] = i
    return idx


def get(row: list, col_idx: dict, field: str) -> str | None:
    i = col_idx.get(field)
    if i is None or i >= len(row):
        return None
    return row[i]


# ---------------------------------------------------------------------------
# Legacy loaders
# ---------------------------------------------------------------------------

def classify_legacy_file(name: str) -> str | None:
    if "dogovori-sklyucheni" in name:
        return "contracts"
    if "izmeneniya" in name:
        return "annexes"
    if "izvn" in name:
        return "excl"
    return None


def handle_legacy_year(conn: duckdb.DuckDBPyConnection, year_dir: Path) -> None:
    year = int(year_dir.name)
    for f in sorted(year_dir.glob("*.json")):
        kind = classify_legacy_file(f.name)
        if kind is None:
            print(f"  [skip] {f.name}", flush=True)
            continue

        print(f"  {year} {kind} ...", end=" ", flush=True)
        data = json.loads(f.read_text(encoding="utf-8"))

        if kind == "annexes":
            n = _load_annexes(conn, data, year)
        else:
            n = _load_contracts(conn, data, year, kind)
        print(f"{n:,} rows", flush=True)


def _load_contracts(conn: duckdb.DuckDBPyConnection, data: list, year: int, source_type: str) -> int:
    headers = data[0]
    col_idx = build_col_index(headers, CONTRACT_COL_MAP)
    rows = []
    for raw in data[1:]:
        g = lambda field, _r=raw: get(_r, col_idx, field)
        rows.append((
            source_type,
            year,
            strip_quotes(g("doc_number")),
            strip_quotes(g("contract_number")),
            parse_date(g("contract_date")),
            parse_date(g("published_date")),
            strip_quotes(g("procurement_number")),
            strip_quotes(g("buyer_eik")),
            strip_quotes(g("buyer_name")),
            strip_quotes(g("procurement_subject")),
            strip_quotes(g("procurement_category")),
            parse_bool(g("eu_funded")),
            parse_int(g("bid_count")),
            strip_quotes(g("contract_subject")),
            strip_quotes(g("supplier_eik")),
            strip_quotes(g("supplier_name")),
            parse_float(g("value_at_signing")),
            strip_quotes(g("currency")) or "BGN",
            parse_float(g("vat_amount")),
            parse_bool(g("is_sme")),
        ))
    bulk_copy(conn, "legacy_contracts", rows)
    return len(rows)


def _load_annexes(conn: duckdb.DuckDBPyConnection, data: list, year: int) -> int:
    headers = data[0]
    col_idx = build_col_index(headers, ANNEX_COL_MAP)
    rows = []
    for raw in data[1:]:
        g = lambda field, _r=raw: get(_r, col_idx, field)
        rows.append((
            year,
            strip_quotes(g("doc_number")),
            strip_quotes(g("contract_number")),
            parse_date(g("contract_date")),
            parse_date(g("published_date")),
            strip_quotes(g("procurement_number")),
            strip_quotes(g("buyer_eik")),
            strip_quotes(g("buyer_name")),
            strip_quotes(g("procurement_subject")),
            strip_quotes(g("procurement_category")),
            parse_bool(g("eu_funded")),
            strip_quotes(g("contract_subject")),
            strip_quotes(g("supplier_eik")),
            strip_quotes(g("supplier_name")),
            parse_float(g("value_before")),
            parse_float(g("value_after")),
            parse_float(g("value_change")),
            strip_quotes(g("currency")) or "BGN",
            strip_quotes(g("amendment_description")),
            strip_quotes(g("amendment_reason")),
            strip_quotes(g("circumstances")),
            parse_bool(g("is_sme")),
        ))
    bulk_copy(conn, "legacy_annexes", rows)
    return len(rows)


# ---------------------------------------------------------------------------
# OCDS loaders
# ---------------------------------------------------------------------------

def handle_ocds_file(conn: duckdb.DuckDBPyConnection, dated_dir: Path) -> None:
    source_file_date = dated_dir.name
    json_file = next(dated_dir.glob("*.json"))
    print(f"  OCDS {source_file_date} ...", end=" ", flush=True)

    data = json.loads(json_file.read_text(encoding="utf-8"))
    releases = data.get("releases", [])

    release_rows = []
    award_rows = []

    for r in releases:
        release_row = _parse_release(r, source_file_date)
        if release_row is None:
            continue
        release_rows.append(release_row)
        collect_parties(r, source_file_date)
        award_rows.extend(_parse_award_contracts(r))

    # Deduplicate releases by release_id before inserting
    seen = set(
        row[0] for row in
        conn.execute("SELECT release_id FROM ocds_releases").fetchall()
    )
    new_releases = [r for r in release_rows if r[0] not in seen]

    bulk_copy(conn, "ocds_releases", new_releases)
    bulk_copy(conn, "ocds_award_contracts", award_rows)

    print(f"{len(new_releases):,} releases, {len(award_rows):,} awards", flush=True)


def _parse_release(r: dict, source_file_date: str) -> tuple | None:
    release_id = r.get("id")
    if not release_id:
        return None

    ocid = r.get("ocid", "")
    tender = r.get("tender", {})
    bids = r.get("bids", {})

    bid_stats: dict[str, int] = {}
    for s in bids.get("statistics", []):
        measure = s.get("measure", "")
        val = s.get("value")
        if measure and val is not None:
            bid_stats[measure] = val

    buyer_org_id = (r.get("buyer") or {}).get("id")
    buyer_name = (r.get("buyer") or {}).get("name")
    buyer_eik = None
    for party in r.get("parties", []):
        if party.get("id") == buyer_org_id:
            ident = party.get("identifier", {})
            if ident.get("scheme") == "BG-EIK":
                buyer_eik = ident.get("id")
            break

    val_obj = tender.get("value") or {}
    legal_basis = tender.get("legalBasis") or {}

    tags = ";".join(r.get("tag", []))  # stored as semicolon-delimited VARCHAR

    return (
        release_id,
        ocid,
        tender.get("id"),
        r.get("date"),
        tags or None,
        source_file_date,
        tender.get("title"),
        tender.get("status"),
        tender.get("procurementMethod"),
        tender.get("mainProcurementCategory"),
        legal_basis.get("id"),
        val_obj.get("amount"),
        val_obj.get("currency"),
        len(tender.get("lots", [])) or None,
        bid_stats.get("bids"),
        bid_stats.get("electronicBids"),
        buyer_org_id,
        buyer_name,
        buyer_eik,
    )


def _parse_award_contracts(r: dict) -> list[tuple]:
    release_id = r.get("id", "")
    ocid = r.get("ocid", "")
    awards_by_id = {a["id"]: a for a in r.get("awards", []) if "id" in a}

    party_eik: dict[str, str] = {}
    for p in r.get("parties", []):
        ident = p.get("identifier", {})
        if ident.get("scheme") == "BG-EIK" and p.get("id"):
            party_eik[p["id"]] = ident.get("id", "")

    rows = []
    emitted_award_ids: set[str] = set()

    def _date(s: str | None) -> str | None:
        return s[:10] if s and len(s) >= 10 else None

    for c in r.get("contracts", []):
        award_id = c.get("awardID")
        award = awards_by_id.get(award_id, {})
        suppliers = award.get("suppliers", [])
        supplier = suppliers[0] if suppliers else {}
        supplier_org_id = supplier.get("id")
        period = c.get("period") or {}
        val = c.get("value") or {}
        rows.append((
            release_id, ocid,
            c.get("id"),
            val.get("amount"), val.get("currency"),
            _date(c.get("dateSigned")),
            _date(period.get("startDate")), _date(period.get("endDate")),
            bool(c.get("amendments")),
            award_id, award.get("status"),
            (award.get("relatedLots") or [None])[0],
            supplier_org_id, supplier.get("name"),
            party_eik.get(supplier_org_id) if supplier_org_id else None,
        ))
        if award_id:
            emitted_award_ids.add(award_id)

    for a in r.get("awards", []):
        if a.get("id") in emitted_award_ids:
            continue
        suppliers = a.get("suppliers", [])
        supplier = suppliers[0] if suppliers else {}
        supplier_org_id = supplier.get("id")
        rows.append((
            release_id, ocid,
            None, None, None, None, None, None, False,
            a.get("id"), a.get("status"),
            (a.get("relatedLots") or [None])[0],
            supplier_org_id, supplier.get("name"),
            party_eik.get(supplier_org_id) if supplier_org_id else None,
        ))

    return rows


# ---------------------------------------------------------------------------
# Parties — collected in-memory during OCDS pass, flushed at end
# ---------------------------------------------------------------------------

_parties_buffer: dict[str, dict] = {}


def collect_parties(r: dict, source_file_date: str) -> None:
    for p in r.get("parties", []):
        ident = p.get("identifier", {})
        if ident.get("scheme") != "BG-EIK":
            continue
        eik = ident.get("id")
        if not eik:
            continue
        addr = p.get("address", {})
        existing = _parties_buffer.get(eik)
        if existing is None or source_file_date > existing["last_seen"]:
            _parties_buffer[eik] = {
                "eik": eik,
                "legal_name": ident.get("legalName"),
                "display_name": p.get("name"),
                "address_locality": addr.get("locality"),
                "address_region": addr.get("region"),
                "country": addr.get("countryName"),
                "first_seen": source_file_date,
                "last_seen": source_file_date,
            }
        else:
            if source_file_date < existing["first_seen"]:
                existing["first_seen"] = source_file_date


def flush_parties(conn: duckdb.DuckDBPyConnection) -> None:
    print("Parties ...", end=" ", flush=True)
    rows = [
        (p["eik"], p["legal_name"], p["display_name"],
         p["address_locality"], p["address_region"], p["country"],
         p["first_seen"], p["last_seen"])
        for p in _parties_buffer.values()
    ]
    # COPY FROM handles upsert via temp table + MERGE
    bulk_copy(conn, "parties", rows)
    print(f"{len(rows):,} unique parties", flush=True)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    if DB_PATH.exists():
        DB_PATH.unlink()
        print(f"Removed existing {DB_PATH}", flush=True)

    conn = duckdb.connect(str(DB_PATH))
    conn.execute(DDL)

    print("=== Legacy years (2020–2023) ===", flush=True)
    for subdir in sorted(RAW_DATA.iterdir()):
        if re.match(r"^\d{4}$", subdir.name):
            handle_legacy_year(conn, subdir)

    print("\n=== OCDS biweekly files (2026) ===", flush=True)
    for subdir in sorted(RAW_DATA.iterdir()):
        if re.match(r"^\d{4}-\d{2}-\d{2}$", subdir.name):
            handle_ocds_file(conn, subdir)

    print()
    flush_parties(conn)

    print("Creating view ...", end=" ", flush=True)
    conn.execute(VIEW_DDL)
    print("done", flush=True)

    conn.execute("ANALYZE")
    print("\n=== Row counts ===", flush=True)
    for table in ("legacy_contracts", "legacy_annexes", "ocds_releases",
                  "ocds_award_contracts", "parties", "contracts_unified"):
        n = conn.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]
        print(f"  {table:30s}: {n:>8,}", flush=True)

    conn.close()
    size_mb = DB_PATH.stat().st_size / 1024 / 1024
    print(f"\nDatabase: {DB_PATH}  ({size_mb:.1f} MB)", flush=True)


if __name__ == "__main__":
    main()
