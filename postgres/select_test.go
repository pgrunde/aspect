package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aodin/aspect"
)

// Perform some selections against an actual postgres database
// Note: A db.json file must be set in this package

type incompleteUser struct {
	Name     string `db:"name"`
	Password string `db:"password"`
	IsActive bool   `db:"is_active"`
}

// Select incomplete structs
type testUser struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Password string `db:"password"`
	IsActive bool   `db:"is_active"`
	Contacts []string
	manager  string
}

func TestSelect(t *testing.T) {
	assert := assert.New(t)

	// Connect to the database specified in the test db.json config
	// Default to the Travis CI settings if no file is found
	conf, err := aspect.ParseTestConfig("./db.json")
	if err != nil {
		t.Fatalf(
			"postgres: failed to parse test configuration, test aborted: %s",
			err,
		)
	}

	db, err := aspect.Connect(conf.Driver, conf.Credentials())
	require.Nil(t, err)
	defer db.Close()

	// Perform all tests in a transaction and always rollback
	tx, err := db.Begin()
	require.Nil(t, err)
	defer tx.Rollback()

	_, err = tx.Execute(users.Create())
	require.Nil(t, err)

	// Insert users as values or pointers
	admin := user{Name: "admin", IsActive: true}
	stmt := aspect.Insert(
		users.C["name"],
		users.C["password"],
		users.C["is_active"],
	)
	_, err = tx.Execute(stmt.Values(admin))
	require.Nil(t, err)

	_, err = tx.Execute(stmt.Values(&admin))
	require.Nil(t, err)

	var u user
	require.Nil(t, tx.QueryOne(users.Select(), &u))
	assert.Equal("admin", u.Name)
	assert.Equal(true, u.IsActive)

	// Select using a returning clause
	client := user{Name: "client", Password: "1234", IsActive: false}
	returningStmt := Insert(
		users.C["name"],
		users.C["password"],
		users.C["is_active"],
	).Returning(
		users.C["id"],
	)
	require.Nil(t, tx.QueryOne(returningStmt.Values(client), &client.ID))
	assert.NotEqual(0, client.ID) // The ID should be anything but zero

	// Select into a struct that has extra columns
	// TODO Skip unexported fields
	var extraField testUser
	require.Nil(t, tx.QueryOne(users.Select().Where(users.C["name"].Equals("client")), &extraField))
	assert.Equal("client", extraField.Name)
	assert.Equal(false, extraField.IsActive)

	// Query multiple users
	var extraFields []testUser
	assert.Nil(tx.QueryAll(users.Select(), &extraFields))
	assert.Equal(3, len(extraFields))

	// Query ids directly
	var ids []int64
	orderBy := aspect.Select(users.C["id"]).OrderBy(users.C["id"].Desc())
	assert.Nil(tx.QueryAll(orderBy, &ids))
	assert.Equal(3, len(ids))
	var prev int64
	for _, id := range ids {
		if prev != 0 {
			if prev < id {
				t.Errorf("id results returned out of order")
			}
		}
	}

	// TODO destination types that don't match the result

	// Update
	// ------

	updateStmt := users.Update().Values(
		aspect.Values{"name": "HELLO", "password": "BYE"},
	).Where(
		users.C["id"].Equals(client.ID),
	)
	result, err := tx.Execute(updateStmt)
	require.Nil(t, err)

	rowsAffected, err := result.RowsAffected()
	assert.Nil(err)
	assert.Equal(1, rowsAffected)

	// Delete
	// ------

	result, err = tx.Execute(users.Delete())
	require.Nil(t, err)

	rowsAffected, err = result.RowsAffected()
	assert.Nil(err)
	assert.Equal(3, rowsAffected)
}
