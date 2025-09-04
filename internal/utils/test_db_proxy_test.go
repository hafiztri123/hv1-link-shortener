package utils

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriver_RegistrationAndConnection(t *testing.T) {
	db, err := sql.Open("sqlite3_proxy", "file::memory:")
	require.NoError(t, err, "Should be able to open a connection with the proxy driver")
	require.NotNil(t, db, "DB object should not be nil")
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err, "Should be able to ping the database")
}

// Test 2: Verify the core logic: placeholder translation ($1 -> ?).
func TestProxyConn_PlaceholderTranslation(t *testing.T) {
	db, err := sql.Open("sqlite3_proxy", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_items (id INTEGER, name TEXT)`)
	require.NoError(t, err)

	insertQuery := `INSERT INTO test_items (id, name) VALUES ($1, $2)`
	_, err = db.Exec(insertQuery, 1, "widget")
	require.NoError(t, err, "INSERT with $ placeholders should work")

	var itemName string
	selectQuery := `SELECT name FROM test_items WHERE id = $1`
	err = db.QueryRow(selectQuery, 1).Scan(&itemName)

	require.NoError(t, err)
	assert.Equal(t, "widget", itemName, "Should retrieve the correct value inserted via proxy")
}

func TestProxyConn_TransactionSupport(t *testing.T) {
	db, err := sql.Open("sqlite3_proxy", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_tx (id INTEGER)`)
	require.NoError(t, err)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	_, err = tx.Exec(`INSERT INTO test_tx (id) VALUES ($1)`, 100)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	var id int
	err = db.QueryRow(`SELECT id FROM test_tx WHERE id = $1`, 100).Scan(&id)
	require.NoError(t, err)
	assert.Equal(t, 100, id)
}
