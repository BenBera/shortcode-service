-- Create the sms_categories table
CREATE TABLE sms_categories (
    id int(11) NOT NULL AUTO_INCREMENT,
    category_name varchar(255) NOT NULL,
    description text,
    is_active tinyint(1) NOT NULL DEFAULT '1',
    created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_category_name (category_name)
);

-- Create the sms_keywords table
CREATE TABLE sms_keywords (
    id int(11) NOT NULL AUTO_INCREMENT,
    category_id int(11) NOT NULL,
    keyword varchar(255) NOT NULL,
    is_active tinyint(1) NOT NULL DEFAULT '1',
    created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_category_keyword (category_id, keyword),
    KEY idx_keyword (keyword),
    CONSTRAINT fk_sms_keywords_category FOREIGN KEY (category_id) REFERENCES sms_categories (id) ON DELETE CASCADE ON UPDATE CASCADE
);

