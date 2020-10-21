CREATE DATABASE serverinfo;
CREATE USER 'serverinfo'@'localhost' IDENTIFIED BY 'serverinfo';
GRANT ALL PRIVILEGES on serverinfo.* TO 'serverinfo'@'localhost' IDENTIFIED BY 'serverinfo';
FLUSH PRIVILEGES;

USE serverinfo;

CREATE TABLE servers (
  email varchar(255) NOT NULL,
  version varchar(35) NOT NULL,
  UNIQUE (email, version)
)
