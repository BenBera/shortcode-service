CREATE TABLE inbox
(
    id       bigint       NOT NULL AUTO_INCREMENT,
    msisdn   bigint       NOT NULL,
    inbox_id int          NOT NULL,
    message  varchar(200) NOT NULL,
    response text         NULL     DEFAULT NULL,
    created  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY msisdn (msisdn),
    KEY inbox_id (inbox_id),
    KEY message (message),
    FULLTEXT KEY response (response),
    KEY created (created),
    KEY updated (updated)
);

CREATE TABLE sms_template
(
    id               bigint         NOT NULL AUTO_INCREMENT,
    name           VARCHAR(150)         NOT NULL,
    message         VARCHAR(300)         NOT NULL,
    created          datetime       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated          timestamp      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY name (name),
    KEY message (message),
    KEY created (created),
    KEY updated (updated)
);