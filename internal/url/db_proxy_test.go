package url

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"

	"github.com/mattn/go-sqlite3"
)


var queryRexp = regexp.MustCompile(`\$[0-9]+`)

type proxyDriver struct {
	primaryDriver driver.Driver
}

type proxyConn struct {
	primaryConn driver.Conn
}

func (d *proxyDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.primaryDriver.Open(name)
	if err != nil {
		return nil, err
	}
	return &proxyConn{primaryConn: conn}, nil
}

func (c *proxyConn) Prepare(query string) (driver.Stmt, error) {
	sqliteQuery := queryRexp.ReplaceAllString(query, "?")
	return c.primaryConn.(driver.ConnPrepareContext).PrepareContext(context.Background(), sqliteQuery)
}

func (c *proxyConn) Close() error { return c.primaryConn.Close()}
func (c *proxyConn) Begin() (driver.Tx, error) { return c.BeginTx(context.Background(), driver.TxOptions{})}

func (c *proxyConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if connBeginTx, ok := c.primaryConn.(driver.ConnBeginTx); ok {
		return connBeginTx.BeginTx(ctx, opts)
	}

	return c.primaryConn.Begin()
}

func init() {
	sql.Register("sqlite3_proxy", &proxyDriver{primaryDriver: &sqlite3.SQLiteDriver{}})
}