package model

type Database struct {
	Tables map[string]*Table
}

type Table struct {
	Name    string
	Columns []*Column
	Indexes []*Index
	Comment string
	DDL     string
}

func (t *Table) AddColumn(c *Column) {
	t.Columns = append(t.Columns, c)
}

func (t *Table) AddIndex(i *Index) {
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

func (i *Index) AddColumn(name string) {
	column := &IndexColumn{
		Name: name,
	}
	i.Columns = append(i.Columns, column)
}

type IndexColumn struct {
	Name string
}
