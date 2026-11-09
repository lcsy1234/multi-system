CREATE TABLE IF NOT EXISTS material_requisition (
 id CHAR(36) NOT NULL PRIMARY KEY,
 number VARCHAR(80) NOT NULL UNIQUE,
 request_hash CHAR(64) NOT NULL,
 process_id VARCHAR(64) NOT NULL,
 warehouse_code VARCHAR(64) NOT NULL,
 requested_date DATE NOT NULL,
 remark VARCHAR(1000) NOT NULL,
 sync_status VARCHAR(32) NOT NULL,
 last_error TEXT NOT NULL,
 attempts INT NOT NULL DEFAULT 0,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_created(created_at)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS material_requisition_line (
 requisition_id CHAR(36) NOT NULL,
 line_no INT NOT NULL,
 material_code VARCHAR(64) NOT NULL,
 material_name VARCHAR(255) NOT NULL,
 spec VARCHAR(255) NOT NULL,
 unit VARCHAR(32) NOT NULL,
 requested_qty DECIMAL(18,4) NOT NULL,
 remark VARCHAR(500) NOT NULL,
 PRIMARY KEY(requisition_id,line_no),
 UNIQUE KEY uk_material(requisition_id,material_code),
 FOREIGN KEY(requisition_id) REFERENCES material_requisition(id),
 CHECK(requested_qty>0)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS integration_event_log (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 requisition_id CHAR(36) NOT NULL,
 message TEXT NOT NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 FOREIGN KEY(requisition_id) REFERENCES material_requisition(id),
 INDEX idx_request(requisition_id,id)
) ENGINE=InnoDB;
