package mysql

// generateSQLBase is the query list of SQL statements required to build the nil backend.
var generateSQLBase = []string{
	`
		CREATE TABLE IF NOT EXISTS user (
			user_id bigint unsigned NOT NULL AUTO_INCREMENT,
			user_name varchar(128) CHARACTER SET ascii NOT NULL,
			access_key varchar(32) charset ascii NOT NULL,
			secret_key varchar(32) charset ascii NOT NULL,
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
	`
		CREATE TABLE IF NOT EXISTS node (
			node_id bigint unsigned NOT NULL AUTO_INCREMENT,
			node_name varchar(32) CHARACTER SET ascii NOT NULL,
			node_type varchar(32) CHARACTER SET ascii NOT NULL,
			node_status varchar(32) CHARACTER SET ascii NOT NULL,
			node_address varchar(32) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (node_id),
			UNIQUE KEY (node_name)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS volume (
			volume_id bigint unsigned NOT NULL AUTO_INCREMENT,
			volume_status varchar(32) CHARACTER SET ascii NOT NULL,
			node_id bigint unsigned NOT NULL,
			size int unsigned NOT NULL,
			free int unsigned NOT NULL,
			used int unsigned NOT NULL,
			max_chain int unsigned NOT NULL,
			speed varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (volume_id),
			FOREIGN KEY (node_id) REFERENCES node (node_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS local_chain (
			local_chain_id bigint unsigned NOT NULL AUTO_INCREMENT,
			local_chain_status varchar(32) CHARACTER SET ascii NOT NULL,
			first_volume_id bigint unsigned NOT NULL,
			second_volume_id bigint unsigned NOT NULL,
			third_volume_id bigint unsigned NOT NULL,
			parity_volume_id bigint unsigned NOT NULL,
			PRIMARY KEY (local_chain_id),
			FOREIGN KEY (first_volume_id) REFERENCES volume (volume_id),
			FOREIGN KEY (second_volume_id) REFERENCES volume (volume_id),
			FOREIGN KEY (third_volume_id) REFERENCES volume (volume_id),
			FOREIGN KEY (parity_volume_id) REFERENCES volume (volume_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS cmap (
			cmap_id bigint unsigned NOT NULL AUTO_INCREMENT,
			PRIMARY KEY (cmap_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
}
