-- https://dba.stackexchange.com/questions/117109/how-to-manage-default-privileges-for-users-on-a-database-vs-schema/117661#117661
CREATE DATABASE db_chat_server OWNER chat_server_admin;
REVOKE ALL ON DATABASE db_chat_server FROM public;

GRANT CONNECT ON DATABASE db_chat_server TO chat_server_user;

\connect db_chat_server

CREATE SCHEMA chat_server_schema AUTHORIZATION chat_server_admin;

SET search_path = chat_server_schema;

ALTER ROLE chat_server_admin IN DATABASE db_chat_server SET search_path = chat_server_schema;
ALTER ROLE chat_server_manager IN DATABASE db_chat_server SET search_path = chat_server_schema;
ALTER ROLE chat_server_user IN DATABASE db_chat_server SET search_path = chat_server_schema;

GRANT USAGE  ON SCHEMA chat_server_schema TO chat_server_user;
GRANT CREATE ON SCHEMA chat_server_schema TO chat_server_admin;

ALTER DEFAULT PRIVILEGES FOR ROLE chat_server_admin
GRANT SELECT ON TABLES TO chat_server_user;

ALTER DEFAULT PRIVILEGES FOR ROLE chat_server_admin
GRANT INSERT, UPDATE, DELETE ON TABLES TO chat_server_manager;
