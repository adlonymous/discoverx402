package state

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adlonymous/discoverx402/internal/types"
	_ "modernc.org/sqlite"
)

type Repo struct{ db *sql.DB }

func OpenSQLite(path string) (*Repo, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode;`); err != nil {
		_ = db.Close()
		dsn = fmt.Sprintf("file:%s?_pragma=journal_mode(DELETE)&_pragma=busy_timeout(5000)", path)
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		dsn = fmt.Sprintf("file:%s?_pragma=journal_mode(DELETE)&_pragma=busy_timeout(5000)", path)
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
		if err := migrate(db); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &Repo{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS listings(
  id TEXT PRIMARY KEY,
  resource TEXT NOT NULL,
  type TEXT NOT NULL,
  x402_version INTEGER NOT NULL,
  last_updated TEXT NOT NULL,
  metadata_json TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS accepts(
  listing_id TEXT NOT NULL,
  asset TEXT NOT NULL,
  network TEXT NOT NULL,
  scheme TEXT NOT NULL,
  pay_to TEXT NOT NULL,
  max_amount_required TEXT NOT NULL,
  max_timeout_seconds INTEGER NOT NULL,
  mime_type TEXT,
  output_schema_json TEXT,
  description TEXT,
  extra_json TEXT,
  resource TEXT NOT NULL,
  PRIMARY KEY(listing_id, asset, network, pay_to),
  FOREIGN KEY(listing_id) REFERENCES listings(id) ON DELETE CASCADE
);`)
	return err
}

func idFor(l types.Listing) string {
	sum := sha256.Sum256([]byte(l.Resource))
	return hex.EncodeToString(sum[:])
}

func (r *Repo) Upsert(ctx context.Context, l types.Listing) error {
	id := idFor(l)

	metaJSON, _ := json.Marshal(l.Metadata)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO listings(id, resource, type, x402_version, last_updated, metadata_json)
VALUES(?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
 resource=excluded.resource, type=excluded.type, x402_version=excluded.x402_version,
 last_updated=excluded.last_updated, metadata_json=excluded.metadata_json
`, id, l.Resource, l.Type, l.X402Version, l.LastUpdated, string(metaJSON))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM accepts WHERE listing_id=?`, id)
	if err != nil {
		return err
	}

	for _, a := range l.Accepts {
		outJSON, _ := json.Marshal(a.OutputSchema)
		extraJSON, _ := json.Marshal(a.Extra)
		_, err = tx.ExecContext(ctx, `
INSERT INTO accepts(listing_id, asset, network, scheme, pay_to, max_amount_required, max_timeout_seconds,
                    mime_type, output_schema_json, description, extra_json, resource)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			id, a.Asset, a.Network, a.Scheme, a.PayTo, a.MaxAmountRequired, a.MaxTimeoutSeconds,
			nullStr(a.MimeType), nullStr(string(outJSON)), nullStr(a.Description), nullStr(string(extraJSON)), a.Resource,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repo) List(ctx context.Context) ([]types.Listing, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, resource, type, x402_version, last_updated, metadata_json FROM listings ORDER BY last_updated DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Listing, 0)
	for rows.Next() {
		var id, resource, typ, lastUpdated, metaJSON string
		var ver int
		if err := rows.Scan(&id, &resource, &typ, &ver, &lastUpdated, &metaJSON); err != nil {
			return nil, err
		}
		meta := map[string]any{}
		_ = json.Unmarshal([]byte(metaJSON), &meta)

		// load accepts
		aRows, err := r.db.QueryContext(ctx, `SELECT asset,network,scheme,pay_to,max_amount_required,max_timeout_seconds,mime_type,output_schema_json,description,extra_json,resource FROM accepts WHERE listing_id=?`, id)
		if err != nil {
			return nil, err
		}
		accepts := make([]types.Accept, 0)
		for aRows.Next() {
			var a types.Accept
			var outJSON, extraJSON, mime, desc sql.NullString
			if err := aRows.Scan(&a.Asset, &a.Network, &a.Scheme, &a.PayTo, &a.MaxAmountRequired, &a.MaxTimeoutSeconds, &mime, &outJSON, &desc, &extraJSON, &a.Resource); err != nil {
				aRows.Close()
				return nil, err
			}
			if mime.Valid {
				a.MimeType = mime.String
			}
			if desc.Valid {
				a.Description = desc.String
			}
			if outJSON.Valid && outJSON.String != "" {
				var os types.OutputSchema
				_ = json.Unmarshal([]byte(outJSON.String), &os)
				a.OutputSchema = &os
			}
			if extraJSON.Valid && extraJSON.String != "" {
				var ex types.AcceptExtra
				_ = json.Unmarshal([]byte(extraJSON.String), &ex)
				a.Extra = &ex
			}
			accepts = append(accepts, a)
		}
		aRows.Close()

		out = append(out, types.Listing{
			Accepts: accepts, LastUpdated: lastUpdated, Metadata: meta,
			Resource: resource, Type: typ, X402Version: ver,
		})
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *Repo) Close() error {
	return r.db.Close()
}
