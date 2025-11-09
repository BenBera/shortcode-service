CREATE TABLE inbox_dlr (
                           id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                           inbox_id BIGINT NOT NULL,
                           status INT NOT NULL,
                           description VARCHAR(150) NOT NULL,
                           message_type VARCHAR(50) DEFAULT 'ondemand',
                           created          datetime       NOT NULL DEFAULT CURRENT_TIMESTAMP,
                           updated          timestamp      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);