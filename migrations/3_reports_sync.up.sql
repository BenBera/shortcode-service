CREATE TABLE reports_sync(
    id             BIGINT(20)                            NOT NULL AUTO_INCREMENT,
    report_name    VARCHAR(100)                          NOT NULL,
    last_timestamp DATETIME                              NOT NULL DEFAULT '1970-01-01 00:00:00',
    created        DATETIME                              NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated        TIMESTAMP on update CURRENT_TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY (report_name),
    INDEX (created),
    INDEX (updated),
    INDEX (last_timestamp)
);