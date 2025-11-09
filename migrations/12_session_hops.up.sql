CREATE TABLE session_hops
(
    id            bigint       NOT NULL AUTO_INCREMENT,
    user_input    varchar(30)   NOT NULL,
    session_id    varchar(100) NOT NULL,
    msisdn        bigint  NOT NULL,
    response TEXT NULL,
    created  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY msisdn (msisdn),
    KEY user_input (user_input),
    KEY  response ( response),
    KEY session_id (session_id),
    KEY created (created),
    KEY updated (updated)
);