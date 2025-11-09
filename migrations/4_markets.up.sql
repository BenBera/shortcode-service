CREATE TABLE markets (
                         id INT NOT NULL AUTO_INCREMENT ,
                         market_id INT NOT NULL ,
                         market_name VARCHAR(150) NOT NULL ,
                         priority INT NOT NULL DEFAULT '1' ,
                         created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                         updated TIMESTAMP on update CURRENT_TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                         PRIMARY KEY (id),
                         INDEX (market_name),
    INDEX (priority),
    INDEX (created),
    INDEX (updated),
    UNIQUE (market_id));

CREATE TABLE outcomes (
                          id INT NOT NULL AUTO_INCREMENT ,
                          market_id INT NOT NULL ,
                          outcome_id VARCHAR(150) NOT NULL ,
                          outcome_name VARCHAR(150) NOT NULL ,
                          created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                          updated TIMESTAMP on update CURRENT_TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                          PRIMARY KEY (id),
                          INDEX (market_id),
                         INDEX (outcome_name),
                         INDEX (outcome_id),
                         INDEX (created),
                         INDEX (updated),
                         UNIQUE (market_id,outcome_id));

CREATE TABLE specifiers (
                            id INT NOT NULL AUTO_INCREMENT ,
                            market_id INT NOT NULL ,
                            specifier VARCHAR(150) NOT NULL ,
                            data_type VARCHAR(150) NOT NULL ,
                            created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                            updated TIMESTAMP on update CURRENT_TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
                            PRIMARY KEY (id),
                            INDEX (market_id),
                          INDEX (specifier),
                          INDEX (data_type),
                          INDEX (created),
                          INDEX (updated),
                          UNIQUE (market_id,specifier));
