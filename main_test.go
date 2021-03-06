package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	MysqlDatabase = os.Getenv("DB2YAML_MYSQL_DATABASE")
)

func TestEmptyDatabase(t *testing.T) {
	conn, err := setupDB()
	if err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	yaml, err := generateYaml(conn, MysqlDatabase)

	if string(yaml) != "{}\n" {
		t.Fatal("not matched")
	}
}

func TestSingleTable(t *testing.T) {
	conn, err := setupDB()
	if err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE users (
		  id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
		  PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Users table'
	`)
	if err != nil {
		t.Fatalf("failed to create table: %s", err)
	}

	yaml, err := generateYaml(conn, MysqlDatabase)

	expected :=
		`users:
  name: users
  columns:
  - name: id
    type: int
    auto_increment: true
    comment: User ID
  indexes:
  - name: PRIMARY
    unique: true
    columns:
    - name: id
  comment: Users table
  ddl: |-
    CREATE TABLE ` + "`users`" + ` (
      ` + "`id`" + ` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
      PRIMARY KEY (` + "`id`" + `)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Users table'
`

	assert.Equal(t, string(yaml), expected)
}

func TestRegressionGitHubIssues1(t *testing.T) {
	conn, err := setupDB()
	if err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE users (
		  id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
		  PRIMARY  KEY (id)
		) COMMENT = 'Users table';
	`)
	if err != nil {
		t.Fatalf("failed to create table: %s", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE users2 (
		  id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
		  PRIMARY  KEY (id)
		) COMMENT = 'Users table 2';
	`)
	if err != nil {
		t.Fatalf("failed to create table: %s", err)
	}

	yaml, err := generateYaml(conn, MysqlDatabase)

	expected :=
		`users:
  name: users
  columns:
  - name: id
    type: int
    auto_increment: true
    comment: User ID
  indexes:
  - name: PRIMARY
    unique: true
    columns:
    - name: id
  comment: Users table
  ddl: |-
    CREATE TABLE ` + "`users`" + ` (
      ` + "`id`" + ` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
      PRIMARY KEY (` + "`id`" + `)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Users table'
users2:
  name: users2
  columns:
  - name: id
    type: int
    auto_increment: true
    comment: User ID
  indexes:
  - name: PRIMARY
    unique: true
    columns:
    - name: id
  comment: Users table 2
  ddl: |-
    CREATE TABLE ` + "`users2`" + ` (
      ` + "`id`" + ` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'User ID',
      PRIMARY KEY (` + "`id`" + `)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Users table 2'
`

	assert.Equal(t, string(yaml), expected)
}

func setupDB() (*sql.DB, error) {
	dest := fmt.Sprintf("tcp(%s:%s)", os.Getenv("DB2YAML_MYSQL_HOST"), os.Getenv("DB2YAML_MYSQL_PORT"))
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8", os.Getenv("DB2YAML_MYSQL_USERNAME"), os.Getenv("DB2YAML_MYSQL_PASSWORD"), dest, MysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", MysqlDatabase))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", MysqlDatabase))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf("USE `%s`", MysqlDatabase))
	if err != nil {
		return nil, err
	}

	return db, err
}
