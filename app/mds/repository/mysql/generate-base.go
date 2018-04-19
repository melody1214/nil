package mysql

// generateSQLBase is the query list of SQL statements required to build the nil backend.
var generateSQLBase = []string{
	`
		CREATE TABLE IF NOT EXISTS cluster (
			cl_id int unsigned NOT NULL AUTO_INCREMENT,
			cl_local_parity_shards int unsigned NOT NULL,
			PRIMARY KEY (cl_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS user (
			user_id int unsigned NOT NULL AUTO_INCREMENT,
			user_name varchar(128) CHARACTER SET ascii NOT NULL,
			user_access_key varchar(32) charset ascii NOT NULL,
			user_secret_key varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS region (
			rg_id int unsigned NOT NULL AUTO_INCREMENT,
			rg_name varchar(32) CHARACTER SET ascii NOT NULL,
			rg_end_point varchar(128) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (rg_id)
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
			PRIMARY KEY (node_id),
			UNIQUE KEY (node_name)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS volume (
			vl_id int unsigned NOT NULL AUTO_INCREMENT,
			vl_status varchar(32) CHARACTER SET ascii NOT NULL,
			vl_node int unsigned NOT NULL,
			vl_size int unsigned NOT NULL,
			vl_free int unsigned NOT NULL,
			vl_used int unsigned NOT NULL,
			vl_encoding_group int unsigned NOT NULL,
			vl_max_encoding_group int unsigned NOT NULL,
			vl_speed varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (vl_id),
			FOREIGN KEY (vl_node) REFERENCES node (node_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS encoding_group (
			eg_id int unsigned NOT NULL AUTO_INCREMENT,
			eg_status varchar(32) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (eg_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS encoding_group_volume (
			egv_id int unsigned NOT NULL AUTO_INCREMENT,
			egv_encoding_group int unsigned NOT NULL,
			egv_volume int unsigned NOT NULL,
			egv_role int unsigned NOT NULL,
			PRIMARY KEY (egv_id),
			FOREIGN KEY (egv_encoding_group) REFERENCES encoding_group (eg_id),
			FOREIGN KEY (egv_volume) REFERENCES volume (vl_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS cmap (
			cmap_id int unsigned NOT NULL AUTO_INCREMENT,
			PRIMARY KEY (cmap_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS object (
			obj_id int unsigned NOT NULL AUTO_INCREMENT,
			obj_name varchar(255) NOT NULL,
			obj_bucket int unsigned NOT NULL,
			obj_encoding_group int unsigned NOT NULL,
			obj_volume int unsigned NOT NULL,
			PRIMARY KEY (obj_id),
			FOREIGN KEY (obj_bucket) REFERENCES bucket (bk_id),
			FOREIGN KEY (obj_encoding_group) REFERENCES encoding_group (eg_id),
			FOREIGN KEY (obj_volume) REFERENCES volume (vl_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
}
