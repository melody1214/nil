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
	`
		CREATE TABLE IF NOT EXISTS cluster_job (
			clj_id int unsigned NOT NULL AUTO_INCREMENT,
			clj_type int unsigned NOT NULL,
			clj_state int unsigned NOT NULL,
			clj_event_type int unsigned NOT NULL,
			clj_event_affected int unsigned,
			clj_event_time varchar(32) CHARACTER SET ascii NOT NULL,
			clj_scheduled_at varchar(32) CHARACTER SET ascii,
			clj_finished_at varchar(32) CHARACTER SET ascii,
			clj_log varchar(64) CHARACTER SET ascii,
			PRIMARY KEY (clj_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_group (
			geg_id int unsigned NOT NULL AUTO_INCREMENT,
			geg_region_frst int unsigned NOT NULL,
			geg_region_secd int unsigned NOT NULL,
			geg_region_thrd int unsigned NOT NULL,
			geg_region_four int unsigned NOT NULL,
			geg_state int unsigned NOT NULL,
			PRIMARY KEY (geg_id),
			FOREIGN KEY (geg_region_frst) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_secd) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_thrd) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_four) REFERENCES region (rg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_table (
			get_id int unsigned NOT NULL AUTO_INCREMENT,
			get_global_encoding_group int unsigned NOT NULL,
			get_status int unsigned NOT NULL,
			PRIMARY KEY (get_id),
			FOREIGN KEY (get_global_encoding_group) REFERENCES global_encoding_group (geg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_table_eg (
			gete_id int unsigned NOT NULL AUTO_INCREMENT,
			gete_table_idx int unsigned NOT NULL,
			gete_encoding_group int unsigned NOT NULL,
			PRIMARY KEY (gete_id),
			FOREIGN KEY (gete_table_idx) REFERENCES global_encoding_table (get_id),
			FOREIGN KEY (gete_encoding_group) REFERENCES encoding_group (eg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_request (
			ger_id int unsigned NOT NULL AUTO_INCREMENT,
			ger_region int unsigned NOT NULL,
			ger_encoding_group_chunk int unsigned NOT NULL,
			PRIMARY KEY (ger_id),
			FOREIGN KEY (ger_region) REFERENCES region (rg_id)
		)
	`,
}
