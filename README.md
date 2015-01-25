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
  ddl: |-
    CREATE TABLE `pictures` (
      `id` int(11) NOT NULL AUTO_INCREMENT,
      `hash` varchar(32) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'MD5 hash identifies picture',
      `user_id` varchar(32) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'User ID',
      `body` mediumblob NOT NULL COMMENT 'Picture data (PNG)',
      `created_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
      `updated_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
      PRIMARY KEY (`id`),
      UNIQUE KEY `hash` (`hash`),
      KEY `user_id` (`user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='Uploaded pictures'
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
  ddl: |-
    CREATE TABLE `posts` (
      `id` int(11) NOT NULL AUTO_INCREMENT,
      `title` varchar(64) DEFAULT NULL COMMENT 'Title of post',
      `body` text COMMENT 'Body of post',
      `created_at` datetime DEFAULT NULL,
      `updated_at` datetime DEFAULT NULL,
      PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Posts on board'
```

## Author

Yuya Takeyama
