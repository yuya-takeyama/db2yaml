# db2yaml

Generate YAML file from MySQL database tables

## Usage

```
$ db2yaml --host HOST --user USER -D DATABASE -p PASSWORD
```

### Output examle

```yaml
pictures:
  name: pictures
  columns:
  - name: id
    type: int
    auto_increment: true
  - name: hash
    type: varchar
    length: 32
    comment: MD5 hash identifies picture
  - name: user_id
    type: varchar
    length: 32
    comment: User ID
  - name: body
    type: mediumblob
    length: 16777215
    comment: Picture data (PNG)
  - name: created_at
    type: datetime
    default: 0000-00-00 00:00:00
  - name: updated_at
    type: datetime
    default: 0000-00-00 00:00:00
  indexes:
  - name: PRIMARY
    unique: true
    columns:
    - name: id
  - name: hash
    unique: true
    columns:
    - name: hash
  - name: user_id
    unique: false
    columns:
    - name: user_id
  comment: Uploaded pictures
posts:
  name: posts
  columns:
  - name: id
    type: int
    auto_increment: true
  - name: title
    type: varchar
    length: 64
    nullable: true
    comment: Title of post
  - name: body
    type: text
    length: 65535
    nullable: true
    comment: Body of post
  - name: created_at
    type: datetime
    nullable: true
  - name: updated_at
    type: datetime
    nullable: true
  indexes:
  - name: PRIMARY
    unique: true
    columns:
    - name: id
  comment: Posts on board
```

## Author

Yuya Takeyama
