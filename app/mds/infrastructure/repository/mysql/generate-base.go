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
			vl_name varchar(32) CHARACTER SET ascii NOT NULL,
			vl_status varchar(32) CHARACTER SET ascii NOT NULL,
			vl_node int unsigned NOT NULL,
			vl_size int unsigned NOT NULL,
			vl_encoding_group int unsigned NOT NULL,
			vl_max_encoding_group int unsigned NOT NULL,
			vl_speed varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (vl_id),
			UNIQUE KEY (vl_node, vl_name),
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
			egv_move_to int unsigned NOT NULL DEFAULT 0,
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
			obj_role int unsigned NOT NULL,
			PRIMARY KEY (obj_id),
			FOREIGN KEY (obj_bucket) REFERENCES bucket (bk_id),
			FOREIGN KEY (obj_encoding_group) REFERENCES encoding_group (eg_id)
		) ENGINE=InnoDB DEFAULT CHARSET=ascii
	`,
	`
		CREATE TABLE IF NOT EXISTS chunk (
			chk_id int unsigned NOT NULL AUTO_INCREMENT,
			chk_encoding_group int unsigned NOT NULL,
			chk_status varchar(32) charset ascii NOT NULL,
			PRIMARY KEY (chk_id),
			FOREIGN KEY (chk_encoding_group) REFERENCES encoding_group (eg_id)
		)
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
			clj_log varchar(256) CHARACTER SET ascii,
			PRIMARY KEY (clj_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_group (
			geg_id int unsigned NOT NULL AUTO_INCREMENT,
			geg_region_first int unsigned NOT NULL,
			geg_region_second int unsigned NOT NULL,
			geg_region_third int unsigned NOT NULL,
			geg_region_parity int unsigned NOT NULL,
			geg_state int unsigned NOT NULL,
			PRIMARY KEY (geg_id),
			FOREIGN KEY (geg_region_first) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_second) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_third) REFERENCES region (rg_id),
			FOREIGN KEY (geg_region_parity) REFERENCES region (rg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoded_chunk (
			gec_id int unsigned NOT NULL AUTO_INCREMENT,
			gec_global_encoding_group int unsigned NOT NULL,
			gec_local_chunk_first int unsigned NOT NULL,
			gec_local_chunk_second int unsigned NOT NULL,
			gec_local_chunk_third int unsigned NOT NULL,
			gec_local_chunk_parity int unsigned NOT NULL,
			PRIMARY KEY (gec_id),
			FOREIGN KEY (gec_global_encoding_group) REFERENCES global_encoding_group (geg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_job (
			gej_id int unsigned NOT NULL AUTO_INCREMENT,
			gej_status int unsigned NOT NULL,
			PRIMARY KEY (gej_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS global_encoding_chunk (
			guc_id int unsigned NOT NULL AUTO_INCREMENT,
			guc_job int unsigned NOT NULL,
			guc_role int unsigned NOT NULL,
			guc_region int unsigned NOT NULL,
			guc_node int unsigned NOT NULL,
			guc_volume int unsigned NOT NULL,
			guc_encgrp int unsigned NOT NULL,
			guc_chunk varchar(32) CHARACTER SET ascii NOT NULL,
			PRIMARY KEY (guc_id),
			UNIQUE KEY (guc_region, guc_encgrp, guc_chunk),
			FOREIGN KEY (guc_job) REFERENCES global_encoding_job (gej_id),
			FOREIGN KEY (guc_region) REFERENCES region (rg_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS recovery (
			rc_id int unsigned NOT NULL,
			PRIMARY KEY (rc_id),
			FOREIGN KEY (rc_id) REFERENCES cluster_job (clj_id)
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS recovery_volume (
			rcv_id int unsigned NOT NULL AUTO_INCREMENT,
			rcv_job int unsigned NOT NULL,
			rcv_failed_volume int unsigned NOT NULL,
			rcv_replace_volume int unsigned NOT NULL,
			PRIMARY KEY (rcv_id),
			FOREIGN KEY (rcv_job) REFERENCES recovery (rc_id),
			FOREIGN KEY (rcv_failed_volume) REFERENCES volume (vl_id),
			FOREIGN KEY (rcv_replace_volume) REFERENCES volume (vl_id)
		)
	`,
}
