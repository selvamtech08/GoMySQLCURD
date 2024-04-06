create database gobasicdb;

use gobasicdb;

create table users (
	id int primary key auto_increment,
    name varchar(50) not null,
    email varchar(50) not null,
    location varchar(50) not null    
);
