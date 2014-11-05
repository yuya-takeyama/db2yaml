package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"os"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = 3306
)

type Database struct {
	Tables map[string]*Table
}

type Table struct {
	Name    string
	Columns []*Column
	Indexes []*Index
	Comment string
}

func (t *Table) addColumn(c *Column) {
	t.Columns = append(t.Columns, c)
}

func (t *Table) addIndex(i *Index) {
	t.Indexes = append(t.Indexes, i)
}

type Column struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	Length        int    `yaml:"length,omitempty"`
	AutoIncrement bool   `yaml:"auto_increment,omitempty"`
	Nullable      bool   `yaml:"nullable,omitempty"`
	Default       string `yaml:"default,omitempty"`
	Comment       string `yaml:"comment,omitempty"`
}

type Index struct {
	Name    string
	Unique  bool
	Columns []*IndexColumn
}

func (i *Index) addColumn(name string) {
	column := &IndexColumn{
		Name: name,
	}
	i.Columns = append(i.Columns, column)
}

type IndexColumn struct {
	Name string
}

func main() {
	app := cli.NewApp()
	app.Name = "db2yaml"
	app.Usage = "Generate YAML file from database tables"
	app.HideHelp = true

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [options] [arguments...]

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Value: "root",
			Usage: "MySQL user name",
		},
		cli.StringFlag{
			Name:  "host, h",
			Value: DEFAULT_HOST,
			Usage: "MySQL server host name",
		},
		cli.IntFlag{
			Name:  "port, P",
			Value: DEFAULT_PORT,
			Usage: "MySQL server port number",
		},
		cli.StringFlag{
			Name:  "database, D",
			Usage: "Database to use",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "Password to connect to the MySQL server",
		},
		cli.BoolFlag{
			Name:  "help",
			Usage: "Show usage",
		},
	}

	app.Action = func(c *cli.Context) {
		dsn := getDsn(c)
		conn, err := sql.Open("mysql", dsn)
		panicIf(err)

		databaseName := c.String("database")

		yaml, err := generateYaml(conn, databaseName)
		panicIf(err)

		fmt.Print(string(yaml))
	}
	app.Run(os.Args)
}

func getDsn(c *cli.Context) string {
	dest := fmt.Sprintf("tcp(%s:%d)", c.String("host"), c.Int("port"))
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8", c.String("user"), c.String("password"), dest, c.String("database"))
}

func generateYaml(conn *sql.DB, databaseName string) ([]byte, error) {
	database, err := loadDatabaseStructure(conn, databaseName)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(&database.Tables)
}

func loadDatabaseStructure(conn *sql.DB, databaseName string) (*Database, error) {
	database := &Database{
		Tables: make(map[string]*Table),
	}

	err := loadTables(conn, databaseName, database)
	if err != nil {
		return nil, err
	}

	err = loadColumns(conn, databaseName, database)
	if err != nil {
		return nil, err
	}

	err = loadIndexes(conn, databaseName, database)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func loadTables(conn *sql.DB, databaseName string, database *Database) error {
	stmt, err := conn.Prepare("SELECT `TABLE_NAME`, `TABLE_COMMENT` FROM `information_schema`.`TABLES` WHERE `TABLE_SCHEMA` = ?")

	if err != nil {
		return err
	}

	rows, err := stmt.Query(databaseName)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var tableName string
		var tableComment string

		err = rows.Scan(&tableName, &tableComment)

		if err != nil {
			return err
		}

		database.Tables[tableName] = &Table{
			Name:    tableName,
			Comment: tableComment,
		}
	}

	return nil
}

func loadColumns(conn *sql.DB, databaseName string, database *Database) error {
	stmt, err := conn.Prepare("SELECT `TABLE_NAME`, `COLUMN_NAME`, `IS_NULLABLE`, `DATA_TYPE`, `CHARACTER_MAXIMUM_LENGTH`, `COLUMN_DEFAULT`, `COLUMN_COMMENT`, `EXTRA` FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? ORDER BY `TABLE_NAME`, `ORDINAL_POSITION`")
	if err != nil {
		return err
	}

	rows, err := stmt.Query(databaseName)
	if err != nil {
		return err
	}

	for rows.Next() {
		var tableName string
		var columnName string
		var isNullable string
		var nullable bool
		var dataType string
		var length sql.NullInt64
		var defaultValue sql.NullString
		var columnComment sql.NullString
		var extra sql.NullString
		var autoIncrement bool

		err = rows.Scan(&tableName, &columnName, &isNullable, &dataType, &length, &defaultValue, &columnComment, &extra)

		defer rows.Close()

		if isNullable == "YES" {
			nullable = true
		} else {
			nullable = false
		}

		if extra.String == "auto_increment" {
			autoIncrement = true
		} else {
			autoIncrement = false
		}

		column := &Column{
			Name:          columnName,
			Type:          dataType,
			Length:        int(length.Int64),
			AutoIncrement: autoIncrement,
			Nullable:      nullable,
			Default:       defaultValue.String,
			Comment:       columnComment.String,
		}

		table := database.Tables[tableName]

		table.addColumn(column)
	}

	return nil
}

func loadIndexes(conn *sql.DB, databaseName string, database *Database) error {
	stmt, err := conn.Prepare("SELECT `TABLE_NAME`, `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `information_schema`.`STATISTICS` WHERE `INDEX_SCHEMA` = ? ORDER BY `TABLE_NAME`, `NON_UNIQUE`, `INDEX_NAME` != 'PRIMARY', `INDEX_NAME`, `SEQ_IN_INDEX`")

	if err != nil {
		return err
	}

	rows, err := stmt.Query(databaseName)

	if err != nil {
		return err
	}

	prevTableName := ""
	prevIndexName := ""
	index := new(Index)

	for rows.Next() {
		var tableName string
		var indexName string
		var nonUnique int
		var unique bool
		var columnName string

		err = rows.Scan(&tableName, &indexName, &nonUnique, &columnName)

		if err != nil {
			return err
		}

		defer rows.Close()

		if !(prevTableName == tableName && prevIndexName == indexName) {
			if nonUnique == 0 {
				unique = true
			} else {
				unique = false
			}

			index = &Index{
				Name:    indexName,
				Unique:  unique,
				Columns: make([]*IndexColumn, 0),
			}

			table := database.Tables[tableName]

			table.addIndex(index)
		}

		index.addColumn(columnName)

		prevTableName = tableName
		prevIndexName = indexName
	}

	return nil
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
