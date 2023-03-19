## Message Index

| Field      | Type             | **Null** | Key  | **Default** | Extra          |
| ---------- | ---------------- | -------- | ---- | ----------- | -------------- |
| id         | bigint           | NO       | PRI  | <null>      | auto_increment |
| account_a  | varchar(60)      | NO       | MUL  | <null>      |                |
| account_b  | varchar(60)      | NO       |      | <null>      |                |
| direction  | tinyint unsigned | NO       |      | 0           |                |
| message_id | bigint           | NO       |      | <null>      |                |
| group      | varchar(30)      | YES      |      | <null>      |                |
| send_time  | bigint           | NO       | MUL  | <null>      |                |

## Message Content

| Field     | Type             | **Null** | Key  | **Default** | Extra          |
| --------- | ---------------- | -------- | ---- | ----------- | -------------- |
| id        | bigint           | NO       | PRI  | <null>      | auto_increment |
| type      | tinyint unsigned | YES      |      | 0           |                |
| body      | varchar(5000)    | NO       |      | <null>      |                |
| extra     | varchar(500)     | YES      |      | <null>      |                |
| send_time | bigint           | YES      | MUL  | <null>      |                |

## User

| Field    | Type         | **Null** | Key  | **Default** | Extra          |
| -------- | ------------ | -------- | ---- | ----------- | -------------- |
| id       | bigint       | NO       | PRI  | <null>      | auto_increment |
| app      | varchar(30)  | YES      |      | <null>      |                |
| account  | varchar(60)  | YES      | UNI  | <null>      |                |
| password | varchar(30)  | YES      |      | <null>      |                |
| avatar   | varchar(200) | YES      |      | <null>      |                |
| nickname | varchar(20)  | YES      |      | <null>      |                |

## Group

| Field        | Type         | **Null** | Key  | **Default** | Extra          |
| ------------ | ------------ | -------- | ---- | ----------- | -------------- |
| id           | bigint       | NO       | PRI  | <null>      | auto_increment |
| group        | varchar(30)  | YES      | UNI  | <null>      |                |
| app          | varchar(30)  | YES      |      | <null>      |                |
| name         | varchar(50)  | YES      |      | <null>      |                |
| owner        | varchar(60)  | YES      |      | <null>      |                |
| avatar       | varchar(200) | YES      |      | <null>      |                |
| introduction | varchar(300) | YES      |      | <null>      |                |

## GroupMember

| Field   | Type        | **Null** | Key  | **Default** | Extra          |
| ------- | ----------- | -------- | ---- | ----------- | -------------- |
| id      | bigint      | NO       | PRI  | <null>      | auto_increment |
| account | varchar(60) | YES      | UNI  | <null>      |                |
| group   | varchar(30) | YES      | MUL  | <null>      |                |
| alias   | varchar(30) | YES      |      | <null>      |                |