CREATE TABLE IF NOT EXISTS sms_error (
  id BIGINT(20) NOT NULL AUTO_INCREMENT,
  error_type VARCHAR(255) NOT NULL,
  number_of_sends INT NOT NULL DEFAULT 0,
  no_of_errors INT NOT NULL DEFAULT 0,
  previous_errors INT NOT NULL DEFAULT 0,
  average_latency FLOAT NOT NULL DEFAULT 0,
  created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY number_of_sends (number_of_sends),
  KEY no_of_errors (no_of_errors),
  KEY average_latency (average_latency),
  KEY previous_errors (previous_errors),
  KEY error_type (error_type),
  KEY created (created),
  KEY updated (updated)
);
