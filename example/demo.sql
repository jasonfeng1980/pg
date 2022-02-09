
CREATE TABLE `company` (
  `company_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '公司ID',
  `company_name` varchar(50) DEFAULT NULL COMMENT '公司名称',
  `company_money` int(11) NOT NULL DEFAULT '88' COMMENT '公司收益',
  `company_update_at` datetime DEFAULT NULL COMMENT '更新时间',
  `company_create_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`company_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='公司';

CREATE TABLE `company_member` (
  `company_member_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `company_member_ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `company_id` int(11) NOT NULL COMMENT '公司id',
  `company_member_name` varchar(255) NOT NULL COMMENT '成员名称',
  `company_member_birthday` date DEFAULT NULL COMMENT '成员生日',
  `company_member_id_parent` int(11) NOT NULL DEFAULT '0' COMMENT '成员上级ID',
  PRIMARY KEY (`company_member_id`),
  KEY `company_id` (`company_id`),
  KEY `company_member_id_parent` (`company_member_id_parent`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司_成员';


