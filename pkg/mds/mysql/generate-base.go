package mysql

// generateSQLBase is the query list of SQL statements required to build the nil backend.
var generateSQLBase = []string{
	`
		CREATE TABLE IF NOT EXISTS user (
			user_id bigint unsigned NOT NULL AUTO_INCREMENT,
			user_name varchar(128) CHARACTER SET ascii NOT NULL,
			access_key varchar(32) charset ascii not null,
			secret_key varchar(32) charset ascii not null,
			PRIMARY KEY (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS region (
			region_id int unsigned NOT NULL AUTO_INCREMENT,
			region_name varchar(32) CHARACTER SET ascii NOT NULL,
			end_point varchar(128) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (region_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS bucket (
			bucket_id bigint unsigned NOT NULL AUTO_INCREMENT,
			bucket_name varchar(32) CHARACTER SET ascii NOT NULL,
			user_id bigint unsigned NOT NULL,
			region_id int unsigned NOT NULL,
			PRIMARY KEY (bucket_id),
			UNIQUE KEY (bucket_name, region_id),
			FOREIGN KEY (user_id) REFERENCES user (user_id),
			FOREIGN KEY (region_id) REFERENCES region (region_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
}
