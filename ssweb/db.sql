DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
    `id` INT(10) NOT NULL AUTO_INCREMENT,
    `email` VARCHAR(191) NOT NULL,
    `password` VARCHAR(191) NOT NULL,
    `is_verified` TINYINT(2) NOT NULL DEFAULT 0,
    `verify_code` VARCHAR(191)  NOT NULL,
    `verify_email_resends` SMALLINT(2)  NOT NULL DEFAULT 0,
    `enable` TINYINT(1) NOT NULL DEFAULT 1,
    `reg_time` INT(10) NOT NULL DEFAULT 0,
    `reg_ip` INT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `server`;
CREATE TABLE `server` (
    `id` INT(10) NOT NULL AUTO_INCREMENT,
    `ip` VARCHAR(50) NOT NULL,
    `area` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '0:jp, 1:us',
    `is_up` TINYINT(2) NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `ssport`;
CREATE TABLE `ssport` (
    `port` SMALLINT(5) NOT NULL,
    `user_id` INT(10) NOT NULL,
    `password` VARCHAR(191) NOT NULL,
    `enable` TINYINT(1) NOT NULL DEFAULT 1,
    PRIMARY KEY (`port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `transfer`;
CREATE TABLE `transfer` (
    `port` SMALLINT(5) NOT NULL,
    `transfer` BIGINT(19) NOT NULL DEFAULT 0,
    `last_update` INT(10) NOT NULL DEFAULT 0,
    PRIMARY KEY (`port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `token`;
CREATE TABLE `token` (
    `user_id` INT(10) NOT NULL,
    `token` VARCHAR(191) NOT NULL,
    PRIMARY KEY(`user_id`),
    UNIQUE KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `helper`;
CREATE TABLE `helper` (
    `last_port` SMALLINT(5) NOT NULL DEFAULT '10000'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `helper` (`last_port`) VALUES(10000);
