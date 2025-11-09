CREATE TABLE ussd_request
(
    id            bigint       NOT NULL AUTO_INCREMENT,
    ussd_level    bigint       NOT NULL,
    session_id    varchar(100) NOT NULL,
    msisdn        bigint  NOT NULL,
    text          varchar(200) NULL,
    created  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY msisdn (msisdn),
    KEY ussd_level (ussd_level),
    KEY text(text),
    KEY session_id (session_id),
    KEY created (created),
    KEY updated (updated)
);