package mysql

// generateSQLBase is the query list of SQL statements required to build the nil backend.
var generateSQLBase = []string{
	`
		CREATE TABLE IF NOT EXISTS region (
			rg_id int unsigned NOT NULL AUTO_INCREMENT,
			rg_name varchar(32) CHARACTER SET ascii NOT NULL,
			rg_end_point varchar(128) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (rg_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS user (
			user_id int unsigned NOT NULL AUTO_INCREMENT,
			user_name varchar(128) CHARACTER SET ascii NOT NULL,
			user_access_key varchar(32) charset ascii NOT NULL,
			user_secret_key varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (user_id),
			UNIQUE KEY (user_access_key)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS bucket (
			bk_id int unsigned NOT NULL AUTO_INCREMENT,
			bk_name varchar(32) CHARACTER SET ascii NOT NULL,
			bk_user int unsigned NOT NULL,
			bk_region int unsigned NOT NULL,
			PRIMARY KEY (bk_id),
			UNIQUE KEY (bk_name, bk_region),
			FOREIGN KEY (bk_user) REFERENCES user (user_id),
			FOREIGN KEY (bk_region) REFERENCES region (rg_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS node (
			node_id int unsigned NOT NULL AUTO_INCREMENT,
			node_name varchar(32) CHARACTER SET ascii NOT NULL,
			node_type varchar(32) CHARACTER SET ascii NOT NULL,
			node_status varchar(32) CHARACTER SET ascii NOT NULL,
			node_address varchar(32) CHARACTER SET ascii NOT NULL,
			node_size int unsigned NOT NULL DEFAULT '0',
			PRIMARY KEY (node_id),
			UNIQUE KEY (node_name)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS cmap (
			cmap_id int unsigned NOT NULL AUTO_INCREMENT,
			cmap_time varchar(128) CHARACTER SET ascii,
			PRIMARY KEY (cmap_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS cmap_encoding_matrix (
			cem_id int unsigned NOT NULL,
			PRIMARY KEY (cem_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
}
