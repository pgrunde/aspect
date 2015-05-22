package aspect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO Should this be an illegal schema?
var noColumns = Table("none")

var singleColumn = Table("single",
	Column("name", String{}),
)

// Declare schemas that can be used package-wide
var users = Table("users",
	Column("id", Integer{NotNull: true}),
	Column("name", String{Length: 32, Unique: true, NotNull: true}),
	Column("password", String{Length: 128}),
	PrimaryKey("id"),
)

type user struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Password string `db:"password"`
}

var views = Table("views",
	Column("id", Integer{PrimaryKey: true}),
	Column("user_id", Integer{}),
	Column("url", String{}),
	Column("ip", String{}),
	Column("timestamp", Timestamp{}),
)

var edges = Table("edges",
	Column("a", Integer{}),
	Column("b", Integer{}),
	PrimaryKey("a", "b"),
)

type edge struct {
	A int64 `db:"a"`
	B int64 `db:"b"`
}

var attrs = Table("attrs",
	Column("id", Integer{PrimaryKey: true}),
	Column("a", Integer{}),
	Column("b", Integer{}),
	Unique("a", "b"),
)

func TestTableSchema(t *testing.T) {
	// Test table properties
	assert.Equal(t, "users", users.Name)
	assert.Equal(t, "users", users.String())

	// Test the accessor methods
	userID := users.C["id"]
	assert.Equal(t, "id", userID.name)
	assert.Equal(t, users, userID.table)

	// And primary key(s) should exist
	assert.Equal(t, PrimaryKeyArray{"id"}, users.pk)

	// Unique constraints should be declared
	assert.Equal(t, []UniqueConstraint{{"a", "b"}}, attrs.uniques)

	// Primary keys can also be specified through types
	assert.Equal(t, PrimaryKeyArray{"id"}, attrs.pk)

	// As well as uniques
	assert.Equal(t, []UniqueConstraint{{"name"}}, users.uniques)

	// Test improper schemas
	assert.Panics(t, func() {
		Table("bad", Column("a", String{}), Column("a", String{}))
	},
		"failed to panic when duplicate columns were created",
	)

	assert.Panics(t, func() {
		Table("bad", Column("", String{}))
	},
		"failed to panic when a column without a name was created",
	)

	assert.Panics(t, func() {
		Table("bad", Column("okay", String{}), PrimaryKey("not"))
	},
		"failed to panic when a primary key used a non-existent column",
	)

	assert.Panics(t, func() {
		Table("bad", Column("okay", String{}), Unique("not"))
	},
		"failed to panic when a unique constraint used a non-existent column",
	)

	assert.Panics(t, func() {
		Table("bad",
			Column("id", Integer{PrimaryKey: true}),
			Column("id2", Integer{PrimaryKey: true}),
		)
	},
		"failed to panic when given multiple primary keys",
	)
}

// Test the InsertStmt generated by table.Insert()
func TestTableInsert(t *testing.T) {
	expect := NewTester(t, &defaultDialect{})

	// An example user
	admin := user{1, "admin", "secret"}

	// Insert a single value into the table
	expect.SQL(
		`INSERT INTO "users" ("id", "name", "password") VALUES ($1, $2, $3)`,
		users.Insert().Values(&admin),
		1,
		"admin",
		"secret",
	)

	// Test a single column
	expect.SQL(
		`INSERT INTO "single" ("name") VALUES ($1)`,
		singleColumn.Insert().Values(struct{ Name string }{Name: "hello"}),
		"hello",
	)
}

// Test DeleteStmt generated by table.Delete()
func TestTableDelete(t *testing.T) {
	expect := NewTester(t, &defaultDialect{})

	// Delete the entire table
	expect.SQL(`DELETE FROM "users"`, users.Delete())

	// Delete using a conditional
	expect.SQL(
		`DELETE FROM "users" WHERE "users"."id" = $1`,
		users.Delete().Where(users.C["id"].Equals(1)),
		1,
	)
}

// Test UpdateStmt generated by table.Update()
func TestTableUpdate(t *testing.T) {
	expect := NewTester(t, &defaultDialect{})

	expect.SQL(
		`UPDATE "users" SET "name" = $1`,
		users.Update().Values(Values{"name": "Jabroni"}),
		"Jabroni",
	)
}
