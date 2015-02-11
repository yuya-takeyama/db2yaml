package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yuya-takeyama/db2yaml/model"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
)

const (
	defaultHost = "localhost"
	defaultPort = 3306
)

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
			Value: defaultHost,
			Usage: "MySQL server host name",
		},
		cli.IntFlag{
			Name:  "port, P",
			Value: defaultPort,
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

		os.Stdout.Write(yaml)
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
		return nil, fmt.Errorf("failed to load data structure: %s", err)
	}

	return yaml.Marshal(&database.Tables)
}

func loadDatabaseStructure(conn *sql.DB, databaseName string) (*model.Database, error) {
	database := &model.Database{
		Tables: make(map[string]*model.Table),
	}

	err := loadTables(conn, databaseName, database)
	if err != nil {
		return nil, fmt.Errorf("failed to load tables: %s", err)
	}

	err = loadColumns(conn, databaseName, database)
	if err != nil {
		return nil, fmt.Errorf("failed to load columns", err)
	}

	err = loadIndexes(conn, databaseName, database)
	if err != nil {
		return nil, fmt.Errorf("failed to load indexes", err)
	}

	return database, nil
}

func loadTables(conn *sql.DB, databaseName string, database *model.Database) error {
	stmt, err := conn.Prepare("SELECT `TABLES`.`TABLE_NAME`, `TABLES`.`TABLE_COMMENT` FROM `information_schema`.`TABLES` LEFT JOIN `information_schema`.`VIEWS` ON `TABLES`.`TABLE_SCHEMA` = `VIEWS`.`TABLE_SCHEMA` AND `TABLES`.`TABLE_NAME` = `VIEWS`.`TABLE_NAME` WHERE `TABLES`.`TABLE_SCHEMA` = ? AND `VIEWS`.`TABLE_NAME` IS NULL")

	if err != nil {
		return fmt.Errorf("failed to prepare statement to read table informations: %s", err)
	}

	rows, err := stmt.Query(databaseName)

	if err != nil {
		return fmt.Errorf("failed to execute query to read table informations: %s", err)
	}

	defer rows.Close()

	for rows.Next() {
		var tableName string
		var tableComment string

		err = rows.Scan(&tableName, &tableComment)
		if err != nil {
			return fmt.Errorf("failed to scan table information: %s", err)
		}

		var tbl string
		var ddl string

		err = conn.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)).Scan(&tbl, &ddl)
		if err != nil {
			return fmt.Errorf("failed to read DDL: %s", err)
		}

		database.Tables[tableName] = &model.Table{
			Name:    tableName,
			Comment: tableComment,
			DDL:     removeAutoIncrement(ddl),
		}
	}

	return nil
}

func removeAutoIncrement(ddl string) string {
	autoIncrementRegexp := regexp.MustCompile(`AUTO_INCREMENT=\d+ `)
	return autoIncrementRegexp.ReplaceAllString(ddl, "")
}

func loadColumns(conn *sql.DB, databaseName string, database *model.Database) error {
	stmt, err := conn.Prepare("SELECT `TABLE_NAME`, `COLUMN_NAME`, `IS_NULLABLE`, `DATA_TYPE`, `CHARACTER_MAXIMUM_LENGTH`, `COLUMN_DEFAULT`, `COLUMN_COMMENT`, `EXTRA` FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? ORDER BY `TABLE_NAME`, `ORDINAL_POSITION`")
	if err != nil {
		return fmt.Errorf("failed prepare statement to read column informations", err)
	}

	rows, err := stmt.Query(databaseName)
	if err != nil {
		return fmt.Errorf("failed to execute query to read column informations", err)
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

		column := &model.Column{
			Name:          columnName,
			Type:          dataType,
			Length:        int(length.Int64),
			AutoIncrement: autoIncrement,
			Nullable:      nullable,
			Default:       defaultValue.String,
			Comment:       columnComment.String,
		}

		table := database.Tables[tableName]

		if table != nil {
			table.AddColumn(column)
		}
	}

	return nil
}

func loadIndexes(conn *sql.DB, databaseName string, database *model.Database) error {
	stmt, err := conn.Prepare("SELECT `TABLE_NAME`, `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `information_schema`.`STATISTICS` WHERE `INDEX_SCHEMA` = ? ORDER BY `TABLE_NAME`, `NON_UNIQUE`, `INDEX_NAME` != 'PRIMARY', `INDEX_NAME`, `SEQ_IN_INDEX`")

	if err != nil {
		return fmt.Errorf("failed to prepare statement to read index informations", err)
	}

	rows, err := stmt.Query(databaseName)

	if err != nil {
		return fmt.Errorf("failed to execute query to read index informations", err)
	}

	prevTableName := ""
	prevIndexName := ""
	index := new(model.Index)

	for rows.Next() {
		var tableName string
		var indexName string
		var nonUnique int
		var unique bool
		var columnName string

		err = rows.Scan(&tableName, &indexName, &nonUnique, &columnName)

		if err != nil {
			return fmt.Errorf("failed to scan index information", err)
		}

		defer rows.Close()

		if !(prevTableName == tableName && prevIndexName == indexName) {
			if nonUnique == 0 {
				unique = true
			} else {
				unique = false
			}

			index = &model.Index{
				Name:    indexName,
				Unique:  unique,
				Columns: make([]*model.IndexColumn, 0),
			}

			table := database.Tables[tableName]

			table.AddIndex(index)
		}

		index.AddColumn(columnName)

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
